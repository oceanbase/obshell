/*
 * Copyright (c) 2024 OceanBase.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package tenant

import (
	"context"
	stderrors "errors"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/oceanbase/obshell/ob/agent/constant"
	"github.com/oceanbase/obshell/ob/agent/errors"
	"github.com/oceanbase/obshell/ob/agent/global"
	libhttp "github.com/oceanbase/obshell/ob/agent/lib/http"
	"github.com/oceanbase/obshell/ob/agent/meta"
	"github.com/oceanbase/obshell/ob/agent/repository/model/bo"
	"github.com/oceanbase/obshell/ob/agent/secure"
	coordinatorservice "github.com/oceanbase/obshell/ob/agent/service/coordinator"
	"github.com/oceanbase/obshell/ob/agent/service/tenant"
	"github.com/oceanbase/obshell/ob/param"
	"golang.org/x/text/language"
)

const passwordPersistAttempts = 5
const passwordPersistTimeout = 20 * time.Second

// Read the election record without changing the coordinator's identity or
// renewing its lease. A watcher's cached expiration is not authoritative.
func getPasswordMaintainer(ctx context.Context) (meta.AgentInfoInterface, error) {
	service := coordinatorservice.CoordinatorService{}
	maintainer, err := service.GetMaintainerFromObWithContext(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "get password maintainer from OceanBase failed")
	}
	if !maintainer.IsActive || maintainer.AgentIp == "" || maintainer.AgentPort == 0 {
		return nil, errors.Occur(errors.ErrAgentMaintainerNotActive)
	}
	return meta.NewAgentInfo(maintainer.AgentIp, maintainer.AgentPort), nil
}

// Dependencies belong to each request, so tests do not replace mutable global
// functions while production requests may be running.
type passwordPersister struct {
	self    meta.AgentInfoInterface
	cache   *tenant.PasswordMap
	resolve func(context.Context) (meta.AgentInfoInterface, error)
	check   func(context.Context, string, string) (bool, error)
	forward func(context.Context, meta.AgentInfoInterface, string, string) error
	delay   time.Duration
}

func newPasswordPersister() passwordPersister {
	return passwordPersister{
		self: meta.OCS_AGENT, cache: tenant.GetPasswordMap(), resolve: getPasswordMaintainer,
		check: checkTenantPassword, forward: forwardTenantPassword, delay: time.Second,
	}
}

func (p passwordPersister) persist(ctx context.Context, name, password string, allowForward bool) error {
	ctx, cancel := context.WithTimeout(ctx, passwordPersistTimeout)
	defer cancel()
	attempts := passwordPersistAttempts
	if !allowForward {
		// An already-forwarded request must not forward or retry recursively.
		// Its sender re-resolves the maintainer and owns the retry budget.
		attempts = 1
	}
	var err error
	for attempt := 0; attempt < attempts; attempt++ {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		var maintainer meta.AgentInfoInterface
		maintainer, err = p.resolve(ctx)
		if err == nil {
			if samePasswordAgent(p.self, maintainer) {
				err = p.cache.SetValidated(ctx, name, password, func() error {
					// Waiting for the tenant's lock may outlive this caller.
					if err := ctx.Err(); err != nil {
						return err
					}
					connectable, checkErr := p.check(ctx, name, password)
					if checkErr != nil {
						return checkErr
					}
					if !connectable {
						return errors.Occur(errors.ErrObTenantRootPasswordIncorrect)
					}
					current, resolveErr := p.resolve(ctx)
					if resolveErr != nil {
						return resolveErr
					}
					if !samePasswordAgent(p.self, current) {
						return errors.Occur(errors.ErrAgentMaintainerNotActive)
					}
					return ctx.Err()
				})
				if err == nil {
					// A lease may change while storing. Retry against its new owner
					// instead of acknowledging only the previous owner's cache.
					current, resolveErr := p.resolve(ctx)
					err = resolveErr
					if err == nil && !samePasswordAgent(p.self, current) {
						err = errors.Occur(errors.ErrAgentMaintainerNotActive)
					}
				}
			} else if allowForward {
				err = p.forward(ctx, maintainer, name, password)
			} else {
				err = errors.Occur(errors.ErrAgentMaintainerNotActive)
			}
		}
		if err == nil || isTenantPasswordIncorrect(err) {
			return err
		}
		if attempt+1 < attempts {
			timer := time.NewTimer(p.delay)
			select {
			case <-ctx.Done():
				timer.Stop()
				return ctx.Err()
			case <-timer.C:
			}
		}
	}
	return err
}

func samePasswordAgent(left, right meta.AgentInfoInterface) bool {
	return left.GetIp() == right.GetIp() && left.GetPort() == right.GetPort()
}

func passwordRequestContext(c *gin.Context) context.Context {
	if c != nil && c.Request != nil {
		return c.Request.Context()
	}
	return context.Background()
}

func passwordCompletionContext(c *gin.Context) context.Context {
	// ALTER USER has already committed. Finish its cache synchronization even
	// if the client disconnected; persist still owns its timeout and retry
	// budget. No background goroutine is started: this handler completes it.
	return context.WithoutCancel(passwordRequestContext(c))
}

func isTenantPasswordIncorrect(err error) bool {
	var apiErr *libhttp.ApiError
	if stderrors.As(err, &apiErr) {
		return apiErr.ErrCode == errors.ErrObTenantRootPasswordIncorrect.Code
	}
	var agentErr errors.OcsAgentErrorInterface
	return stderrors.As(err, &agentErr) && agentErr.ErrorCode().Code == errors.ErrObTenantRootPasswordIncorrect.Code
}

func forwardTenantPassword(ctx context.Context, maintainer meta.AgentInfoInterface, name, password string) error {
	uri := constant.URI_TENANT_API_PREFIX + "/" + name + constant.URI_ROOTPASSWORD + constant.URI_PERSIST
	return sendPasswordRequest(ctx, maintainer, libhttp.POST, uri, param.PersistTenantRootPasswordParam{Password: password}, nil)
}

func checkTenantPassword(ctx context.Context, name, password string) (bool, error) {
	agent, err := tenantService.GetTenantActiveAgentWithContext(ctx, name)
	if err != nil {
		return false, err
	}
	if agent == nil {
		return false, errors.Occur(errors.ErrObTenantNoActiveServer, name)
	}
	result := &bo.ObTenantPreCheckResult{}
	uri := constant.URI_TENANT_API_PREFIX + "/" + name + constant.URI_PRECHECK
	if err := sendPasswordRequest(ctx, agent, libhttp.GET, uri, param.TenantRootPasswordParam{RootPassword: &password}, result); err != nil {
		return false, err
	}
	return result.IsConnectable, nil
}

// Adapt peer errors to OBShell's response interface. Returning a bare ApiError
// would make BuildResponse replace its code with Common.Unexpected.
type passwordAPIError struct {
	*libhttp.ApiError
	status int
}

func (e *passwordAPIError) ErrorCode() errors.ErrorCode {
	return errors.ErrorCode{Code: e.ErrCode, Kind: e.status, OldCode: e.Code}
}

func (e *passwordAPIError) LocaleMessage(language.Tag) string {
	return e.Message
}

// Keep the existing encrypted protocol, including its authenticated forwarding
// marker, but bound each attempt and retain the remote error code for retries.
func sendPasswordRequest(ctx context.Context, agent meta.AgentInfoInterface, method, uri string, body interface{}, result interface{}) error {
	encrypted, key, iv, err := secure.BuildBody(agent, body)
	if err != nil {
		return err
	}
	url := fmt.Sprintf("%s://%s%s", global.Protocol, agent.String(), uri)
	return sendPreparedPasswordRequest(ctx, method, url, encrypted, secure.BuildHeaderForForward(agent, uri, key, iv), result)
}

func sendPreparedPasswordRequest(ctx context.Context, method, url string, body interface{}, headers map[string]string, result interface{}) error {
	var response libhttp.OcsAgentResponse
	response.Data = result
	// Tenant precheck already retries connection failures. Use the caller's
	// overall budget instead of timing out every three seconds and overlapping
	// still-running prechecks on the receiver.
	request := libhttp.NewClient().SetTimeout(passwordPersistTimeout).R().
		SetContext(ctx).SetHeaders(headers).
		SetHeader("Content-Type", "application/json").SetBody(body).
		SetResult(&response).SetError(&response)
	resp, err := request.Execute(method, url)
	if err != nil {
		return errors.Wrap(err, "tenant password request failed")
	}
	if response.Error != nil {
		return &passwordAPIError{ApiError: response.Error, status: resp.StatusCode()}
	}
	if resp.IsError() || !response.Successful {
		return errors.New("tenant password request returned an unsuccessful response")
	}
	return nil
}
