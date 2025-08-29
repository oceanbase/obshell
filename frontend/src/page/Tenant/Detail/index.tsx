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

import { formatMessage } from '@/util/intl';
import { history, useSelector, useDispatch } from 'umi';
import React, { useEffect, useState } from 'react';
import { Badge, Tooltip } from '@oceanbase/design';
import { findByValue, jsonParse } from '@oceanbase/util';
import { CaretDownFilled } from '@oceanbase/icons';
import { TENANT_STATUS_LIST } from '@/constant/tenant';
import { useBasicMenu, useTenantMenu } from '@/hook/useMenu';
import useDocumentTitle from '@/hook/useDocumentTitle';
import BasicLayout from '@/page/Layout/BasicLayout';
import TenantSelect from '@/component/common/TenantSelect';
import TenantAdminPasswordModal from '@/component/TenantAdminPasswordModal';
import useStyles from './index.style';
import ModifyTenantPasswordModal from '@/page/Tenant/Detail/Component/ModifyTenantPasswordModal';

interface DetailProps {
  location: {
    pathname: string;
  };
  match: {
    params: {
      tenantName: string;
    };
  };
  children: React.ReactNode;
}

const Detail: React.FC<DetailProps> = (props: DetailProps) => {
  const { styles } = useStyles();
  const {
    children,
    match: {
      params: { tenantName },
    },
    ...restProps
  } = props;

  const { tenantData, precheckResult } = useSelector((state: DefaultRootState) => state.tenant);

  const {
    location: { pathname },
  } = restProps;

  const {
    systemInfo: { monitorInfo: { collectInterval } = {} },
    showTenantAdminPasswordModal,
    tenantAdminPasswordErrorData,
  } = useSelector((state: DefaultRootState) => state.global);

  const dispatch = useDispatch();

  // 集群详情里面，会返回 oraclePrivilegeManagementSupprted，只对Oracle租户有效
  // 租户详情里面，Oracle租户有oraclePrivilegeManagementSupprted字段，MySQL租户没有这个字段
  const menus = useTenantMenu(
    tenantName,
    tenantData?.mode,
    tenantData?.oraclePrivilegeManagementSupported,
    tenantData?.obVersion
  );

  const subSideMenus = useBasicMenu();

  useEffect(() => {
    dispatch({
      type: 'tenant/getTenantData',
      payload: {
        name: tenantName,
      },
    });
    dispatch({
      type: 'cluster/getClusterData',
      payload: {},
    });

    return () => {
      dispatch({
        type: 'tenant/update',
        payload: {
          tenantData: {},
        },
      });
    };
  }, [tenantName]);

  useDocumentTitle(tenantData?.name && `${tenantData.name} | ${tenantData.clusterName}`);

  const statusItem = findByValue(TENANT_STATUS_LIST, tenantData.status);

  const ocpExpressEmptySuperUserPasswordTime = jsonParse(
    localStorage.getItem(`__OCP_EXPRESS_TENANT__${tenantName}_EMPTY_SUPER_USER_PASSWORD_TIME__`),
    []
  ) as any[];

  const [showTenantPasswordModal, setShowTenantPasswordModal] = useState(false);

  const timeDifference = new Date().getTime() - (ocpExpressEmptySuperUserPasswordTime || 0);

  const checkEmptyPasswordRootPath = [
    `/tenant/${tenantName}/overview`,
    `/tenant/${tenantName}/parameter`,
  ];

  const checkEmptyPasswordPath = [`/tenant/${tenantName}/database`, `/tenant/${tenantName}/user`];

  const isCheckEmptyPasswordRootPath = checkEmptyPasswordRootPath?.some(url =>
    pathname?.includes(url)
  );
  const isCheckEmptyPasswordPath = checkEmptyPasswordPath?.some(url => pathname?.includes(url));

  useEffect(() => {
    if (
      (!ocpExpressEmptySuperUserPasswordTime || timeDifference > 86400000) &&
      tenantName !== 'sys' &&
      Object.keys(precheckResult)?.length > 0 &&
      precheckResult?.is_empty_root_password &&
      isCheckEmptyPasswordRootPath &&
      precheckResult?.tenantName === tenantName
    ) {
      setShowTenantPasswordModal(true);
    } else {
      localStorage.removeItem(
        `__OCP_EXPRESS_TENANT__${tenantName}_EMPTY_SUPER_USER_PASSWORD_TIME__`
      );
    }
  }, [
    precheckResult?.tenantName,
    precheckResult?.is_empty_root_password,
    isCheckEmptyPasswordRootPath,
    pathname,
  ]);

  useEffect(() => {
    if (
      tenantName !== 'sys' &&
      Object.keys(precheckResult)?.length > 0 &&
      isCheckEmptyPasswordPath &&
      !precheckResult?.is_connectable &&
      precheckResult?.tenantName === tenantName
    ) {
      if (!precheckResult?.is_password_exists) {
        dispatch({
          type: 'global/update',
          payload: {
            showTenantAdminPasswordModal: true,
            tenantAdminPasswordErrorData: {
              type: 'ADD',
              errorMessage: formatMessage(
                {
                  id: 'ocp-v2.Tenant.Detail.ConnectionFailedTenantnameTheTenant',
                  defaultMessage: '连接失败，{tenantName} 租户无法连接，请输入此租户管理员密码',
                },
                { tenantName: tenantName }
              ),
              tenantName,
            },
          },
        });
      } else {
        dispatch({
          type: 'global/update',
          payload: {
            showTenantAdminPasswordModal: true,
            tenantAdminPasswordErrorData: {
              type: 'EDIT',
              errorMessage: formatMessage(
                {
                  id: 'ocp-v2.Tenant.Detail.TheConnectionFailedBecauseThe',
                  defaultMessage: '连接失败，{tenantName} 租户密码错误，请输入此租户的正确密码',
                },
                { tenantName: tenantName }
              ),
              tenantName,
            },
          },
        });
      }
    }
  }, [
    precheckResult?.is_password_exists,
    precheckResult?.is_connectable,
    precheckResult?.tenantName,
    isCheckEmptyPasswordPath,
    pathname,
  ]);

  const emptySuperUserPassword = !precheckResult?.is_empty_root_password || false;

  const handlePreCheck = () => {
    if (tenantName && tenantName !== 'sys') {
      dispatch({
        type: 'tenant/getTenantPreCheck',
        payload: {
          name: tenantName,
        },
      });
    }
  };

  useEffect(() => {
    if (tenantData?.locked !== 'YES') {
      handlePreCheck();
    }
  }, [tenantName, tenantData?.locked]);

  return (
    <BasicLayout
      menus={menus}
      subSideMenus={subSideMenus}
      subSideMenuProps={{ selectedKeys: ['/tenant'] }}
      sideHeader={
        <div className={styles.sideHeader}>
          <div className={styles.tenantWrapper}>
            <Tooltip title={statusItem.label}>
              <Badge
                status={statusItem.badgeStatus}
                style={{
                  cursor: 'pointer',
                  // 由于向右偏移了 3px，为了避免被 TenantSelect 覆盖无法选中，需要提升其 zIndex
                  zIndex: 1,
                  marginRight: -3,
                }}
              />
            </Tooltip>
            <Tooltip
              placement="right"
              title={
                <>
                  <div>
                    {formatMessage({
                      id: 'ocp-express.Tenant.Detail.TenantName',
                      defaultMessage: '租户名：',
                    })}
                    {tenantData?.tenant_name}
                  </div>
                  <div>
                    {formatMessage({
                      id: 'ocp-v2.Tenant.Detail.TenantId',
                      defaultMessage: '租户 ID：',
                    })}
                    {tenantData?.tenant_id}
                  </div>
                </>
              }
            >
              <TenantSelect
                valueProp="tenant_name"
                value={tenantName}
                bordered={false}
                suffixIcon={
                  <CaretDownFilled
                    className="ant-select-suffix"
                    style={{
                      fontSize: 12,
                      color: '#8592AD',
                    }}
                  />
                }
                onChange={value => {
                  history.push(`/tenant/${value}`);
                }}
                className={styles.tenantSelect}
              />
            </Tooltip>
          </div>
        </div>
      }
      {...restProps}
    >
      {children}
      {/* <TaskBubble tenantName={tenantName} /> */}

      <ModifyTenantPasswordModal
        visible={showTenantPasswordModal}
        onCancel={() => {
          if (!ocpExpressEmptySuperUserPasswordTime && emptySuperUserPassword) {
            localStorage.setItem(
              `__OCP_EXPRESS_TENANT__${tenantName}_EMPTY_SUPER_USER_PASSWORD_TIME__`,
              JSON.stringify(new Date().getTime())
            );
          }
          setShowTenantPasswordModal(false);
        }}
        onSuccess={() => {
          localStorage.removeItem(
            `__OCP_EXPRESS_TENANT__${tenantName}_EMPTY_SUPER_USER_PASSWORD_TIME__`
          );
          setShowTenantPasswordModal(false);

          handlePreCheck();
        }}
      />

      <TenantAdminPasswordModal
        visible={showTenantAdminPasswordModal}
        // visible={true}
        type={tenantAdminPasswordErrorData?.type}
        errorMessage={tenantAdminPasswordErrorData?.errorMessage}
        tenantName={tenantAdminPasswordErrorData?.tenantName}
        onCancel={() => {
          dispatch({
            type: 'global/update',
            payload: {
              showTenantAdminPasswordModal: false,
              tenantAdminPasswordErrorData: {},
            },
          });

          handlePreCheck();
        }}
        onSuccess={() => {
          dispatch({
            type: 'global/update',
            payload: {
              showTenantAdminPasswordModal: false,
              tenantAdminPasswordErrorData: {},
            },
          });

          handlePreCheck();
        }}
      />
    </BasicLayout>
  );
};

export default Detail;
