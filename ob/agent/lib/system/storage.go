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

package system

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	log "github.com/sirupsen/logrus"
	"github.com/tencentyun/cos-go-sdk-v5"

	"github.com/oceanbase/obshell/ob/agent/constant"
	"github.com/oceanbase/obshell/ob/agent/errors"
	"github.com/oceanbase/obshell/ob/agent/meta"
)

const (
	host       = "host"
	accessID   = "access_id"
	accessKey  = "access_key"
	appID      = "appid"
	s3Region   = "s3_region"
	deleteMode = "delete_mode"
)

type StorageInterface interface {
	GenerateURI() string
	GenerateURIWithoutSecret() string
	GenerateURIWhitoutParams() string
	GenerateQueryParams() string
	GetResourceType() string
	CheckWritePermission() error
	NewWithObjectKey(string) StorageInterface
}

type OSSConfig struct {
	BaseConf
}

type BaseConf struct {
	BucketName string
	ObjectKey  string
	Host       string
	AccessID   string
	AccessKey  string
	DeleteMode string
}

func (c *OSSConfig) NewWithObjectKey(subpath string) StorageInterface {
	copy := new(OSSConfig)
	*copy = *c
	copy.ObjectKey = fmt.Sprintf("%s/%s", c.ObjectKey, subpath)
	return copy
}

func (c *OSSConfig) GenerateURI() (res string) {
	res = fmt.Sprintf("%s&%s=%s&%s=%s", c.GenerateURIWithoutSecret(), accessID, c.AccessID, accessKey, c.AccessKey)
	return
}

func (c *OSSConfig) GenerateURIWithoutSecret() (res string) {
	res = fmt.Sprintf("%s%s/%s?%s=%s", constant.PREFIX_OSS, c.BucketName, c.ObjectKey, host, c.Host)
	if c.DeleteMode != "" {
		res += fmt.Sprintf("&%s=%s", deleteMode, c.DeleteMode)
	}
	return
}

func (c *OSSConfig) GenerateURIWhitoutParams() string {
	return fmt.Sprintf("%s%s/%s", constant.PREFIX_OSS, c.BucketName, c.ObjectKey)
}

func (c *OSSConfig) GenerateQueryParams() string {
	return fmt.Sprintf("%s=%s&%s=%s&%s=%s", host, c.Host, accessID, c.AccessID, accessKey, c.AccessKey)
}

func (c *OSSConfig) GetResourceType() string {
	return constant.PROTOCOL_OSS
}

func (c *OSSConfig) CheckWritePermission() error {
	client, err := oss.New(c.Host, c.AccessID, c.AccessKey)
	if err != nil {
		return errors.Wrap(err, "create oss client")
	}
	log.Info("OSS client created")

	ossBucket, err := client.Bucket(c.BucketName)
	if err != nil {
		return errors.Wrap(err, "get oss bucket")
	}
	log.Infof("OSS bucket %s created: %#+v", c.BucketName, ossBucket)

	emptyContent := bytes.NewReader([]byte(""))
	testFile := path.Join(c.ObjectKey, meta.OCS_AGENT.GetIp(), fmt.Sprint(meta.OCS_AGENT.GetPort()))
	log.Infof("test file: %s", testFile)
	if err = ossBucket.PutObject(testFile, emptyContent); err != nil {
		return errors.Wrap(err, "put object")
	}

	if err = ossBucket.DeleteObject(testFile); err != nil {
		return errors.Wrap(err, "delete object")
	}

	return nil
}

type COSConfig struct {
	BaseConf
	AppID string
}

func (c *COSConfig) NewWithObjectKey(subpath string) StorageInterface {
	copy := new(COSConfig)
	*copy = *c
	copy.ObjectKey = fmt.Sprintf("%s/%s", c.ObjectKey, subpath)
	return copy
}

func (c *COSConfig) GetResourceType() string {
	return constant.PROTOCOL_COS
}

func (c *COSConfig) GenerateQueryParams() string {
	return fmt.Sprintf("%s=%s&%s=%s&%s=%s&%s=%s", host, c.Host, accessID, c.AccessID, accessKey, c.AccessKey, appID, c.AppID)
}

func (c *COSConfig) GenerateURIWhitoutParams() string {
	return fmt.Sprintf("%s%s/%s", constant.PREFIX_COS, c.BucketName, c.ObjectKey)
}

func (c *COSConfig) GenerateURI() (res string) {
	res = fmt.Sprintf("%s&%s=%s&%s=%s", c.GenerateURIWithoutSecret(), accessID, c.AccessID, accessKey, c.AccessKey)
	return
}

func (c *COSConfig) GenerateURIWithoutSecret() (res string) {
	res = fmt.Sprintf("%s%s/%s?%s=%s", constant.PREFIX_COS, c.BucketName, c.ObjectKey, host, c.Host)
	if c.AppID != "" {
		res += fmt.Sprintf("&%s=%s", appID, c.AppID)
	}
	if c.DeleteMode != "" {
		res += fmt.Sprintf("&%s=%s", deleteMode, c.DeleteMode)
	}
	return
}

func (c *COSConfig) CheckWritePermission() error {
	cosURL := fmt.Sprintf("https://%s.%s", c.BucketName, c.Host)
	u, err := url.Parse(cosURL)
	if err != nil {
		return errors.Wrap(err, "parse cos uri")
	}

	b := &cos.BaseURL{
		BucketURL: u,
	}
	client := cos.NewClient(b, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  c.AccessID,
			SecretKey: c.AccessKey,
		},
	})
	log.Info("COS client created")

	emptyContent := bytes.NewReader([]byte(""))
	testFile := path.Join(c.ObjectKey, meta.OCS_AGENT.GetIp(), fmt.Sprint(meta.OCS_AGENT.GetPort()))
	_, err = client.Object.Put(context.Background(), testFile, emptyContent, nil)
	if err != nil {
		return errors.Wrap(err, "put object")
	}

	_, err = client.Object.Delete(context.Background(), testFile)
	if err != nil {
		return errors.Wrap(err, "delete cos object")
	}

	return nil
}

type S3Config struct {
	BaseConf
	S3Region string
}

func (c *S3Config) NewWithObjectKey(subpath string) StorageInterface {
	copy := new(S3Config)
	*copy = *c
	copy.ObjectKey = fmt.Sprintf("%s/%s", c.ObjectKey, subpath)
	return copy
}

func (c *S3Config) GetResourceType() string {
	return constant.PROTOCOL_S3
}

func (c *S3Config) GenerateURI() (res string) {
	res = fmt.Sprintf("%s&%s=%s&%s=%s", c.GenerateURIWithoutSecret(), accessID, c.AccessID, accessKey, c.AccessKey)
	return
}

func (c *S3Config) GenerateURIWithoutSecret() (res string) {
	res = fmt.Sprintf("%s%s/%s?%s=%s", constant.PREFIX_S3, c.BucketName, c.ObjectKey, host, c.Host)
	if c.S3Region != "" {
		res += fmt.Sprintf("&%s=%s", s3Region, c.S3Region)
	}
	if c.DeleteMode != "" {
		res += fmt.Sprintf("&%s=%s", deleteMode, c.DeleteMode)
	}
	return
}

func (c *S3Config) GenerateURIWhitoutParams() string {
	return fmt.Sprintf("%s%s/%s", constant.PREFIX_S3, c.BucketName, c.ObjectKey)
}

func (c *S3Config) GenerateQueryParams() string {
	return fmt.Sprintf("%s=%s&%s=%s&%s=%s", host, c.Host, accessID, c.AccessID, accessKey, c.AccessKey)
}

func (c *S3Config) CheckWritePermission() (err error) {
	var sess *session.Session
	if c.S3Region != "" {
		sess, err = session.NewSession(&aws.Config{
			Region:      aws.String(c.S3Region),
			Credentials: credentials.NewStaticCredentials(c.AccessID, c.AccessKey, ""),
		})
	} else {
		sess, err = session.NewSession(&aws.Config{
			Region:      aws.String("auto"),
			Endpoint:    aws.String(c.Host),
			Credentials: credentials.NewStaticCredentials(c.AccessID, c.AccessKey, ""),
		})
	}
	if err != nil {
		return errors.Wrap(err, "create s3 session")
	}

	svc := s3.New(sess)
	log.Info("S3 client created")

	emptyContent := bytes.NewReader([]byte(""))
	testFile := path.Join(c.ObjectKey, meta.OCS_AGENT.GetIp(), fmt.Sprint(meta.OCS_AGENT.GetPort()))
	_, err = svc.PutObject(
		&s3.PutObjectInput{
			Bucket: aws.String(c.BucketName),
			Key:    aws.String(testFile),
			Body:   emptyContent,
		},
	)
	if err != nil {
		return errors.Wrap(err, "put s3 object")
	}
	log.Infof("put s3 object %s", testFile)

	_, err = svc.DeleteObject(
		&s3.DeleteObjectInput{
			Bucket: aws.String(c.BucketName),
			Key:    aws.String(testFile),
		},
	)
	if err != nil {
		return errors.Wrap(err, "delete s3 object")
	}

	return nil
}

type NFSConfig struct {
	Path string
}

func (c *NFSConfig) NewWithObjectKey(subpath string) StorageInterface {
	copy := new(NFSConfig)
	*copy = *c
	copy.Path = fmt.Sprintf("%s/%s", c.Path, subpath)
	return copy
}

func (c *NFSConfig) GetResourceType() string {
	return constant.PROTOCOL_FILE
}

func (c *NFSConfig) GenerateQueryParams() string {
	return ""
}

func (c *NFSConfig) GenerateURIWhitoutParams() string {
	return c.GenerateURIWithoutSecret()
}

func (c *NFSConfig) GenerateURI() (res string) {
	return c.GenerateURIWithoutSecret()
}

func (c *NFSConfig) GenerateURIWithoutSecret() (res string) {
	return fmt.Sprintf("%s%s", constant.PREFIX_FILE, c.Path)
}

func (c *NFSConfig) CheckWritePermission() error {
	if err := os.MkdirAll(c.Path, 0755); err != nil {
		return err
	}

	if _, err := os.Open(c.Path); err != nil {
		return errors.Wrap(err, "open nfs path")
	}

	testFile := path.Join(c.Path, meta.OCS_AGENT.String())
	f, err := os.Create(testFile)
	if err != nil {
		return errors.Wrap(err, "create test file")
	}
	defer f.Close()
	if err := os.Remove(testFile); err != nil {
		return errors.Wrap(err, "remove test file")
	}
	return nil
}

func GetStorageInterfaceByURI(uri string) (StorageInterface, error) {
	if strings.HasPrefix(uri, constant.PREFIX_OSS) {
		return GetOSSStorage(uri)
	} else if strings.HasPrefix(uri, constant.PREFIX_COS) {
		return GetCOSStorage(uri)
	} else if strings.HasPrefix(uri, constant.PREFIX_S3) {
		return GetS3Storage(uri)
	} else if strings.HasPrefix(uri, constant.PREFIX_FILE) {
		return GetNFSStorage(uri)
	} else {
		return nil, errors.Occur(errors.ErrObStorageURIInvalid, "invalid uri protocol")
	}
}

func GetResourceType(uri string) (t string, err error) {
	if strings.HasPrefix(uri, constant.PREFIX_OSS) {
		t = constant.PROTOCOL_OSS
	} else if strings.HasPrefix(uri, constant.PREFIX_COS) {
		t = constant.PROTOCOL_COS
	} else if strings.HasPrefix(uri, constant.PREFIX_S3) {
		t = constant.PROTOCOL_S3
	} else if strings.HasPrefix(uri, constant.PREFIX_FILE) {
		t = constant.PROTOCOL_FILE
	} else {
		err = errors.Occur(errors.ErrObStorageURIInvalid, "invalid path type")
	}
	return
}

func GetOSSStorage(url string) (StorageInterface, error) {
	conf := &OSSConfig{}
	urlWithoutScheme := strings.TrimPrefix(url, constant.PREFIX_OSS)
	if _, err := conf.parseParams(urlWithoutScheme); err != nil {
		return nil, errors.Wrap(err, "parse oss config")
	}
	return conf, nil
}

func GetCOSStorage(url string) (StorageInterface, error) {
	conf := &COSConfig{}
	urlWithoutScheme := strings.TrimPrefix(url, constant.PREFIX_COS)
	params, err := conf.parseParams(urlWithoutScheme)
	if err != nil {
		return nil, errors.Wrap(err, "parse cos config")
	}
	conf.AppID = params.Get(appID)
	if conf.AppID == "" {
		return nil, errors.Occur(errors.ErrObStorageURIInvalid, "cos appid is required")
	}
	return conf, nil
}

func GetS3Storage(url string) (StorageInterface, error) {
	conf := &S3Config{}
	urlWithoutScheme := strings.TrimPrefix(url, constant.PREFIX_S3)
	params, err := conf.parseParams(urlWithoutScheme)
	if err != nil {
		return nil, errors.Wrap(err, "parse s3 config")
	}
	conf.S3Region = params.Get(s3Region)
	return conf, nil
}

func GetNFSStorage(url string) (StorageInterface, error) {
	conf := &NFSConfig{}
	conf.Path = strings.TrimPrefix(url, constant.PREFIX_FILE)
	if strings.ContainsAny(conf.Path, "?") {
		return nil, errors.Occur(errors.ErrObStorageURIInvalid, "invalid file path, contains invalid character '?'")
	}
	if !strings.HasPrefix(conf.Path, "/") {
		return nil, errors.Occur(errors.ErrObStorageURIInvalid, "invalid file path, must start with '/'")
	}
	return conf, nil
}

func (c *BaseConf) parseParams(urlWithoutScheme string) (url.Values, error) {
	parts := strings.SplitN(urlWithoutScheme, "/", 2)
	if len(parts) < 2 {
		return nil, errors.Occur(errors.ErrObStorageURIInvalid, "invalid url format")
	}

	c.BucketName = parts[0]
	log.Info("c.BucketName is ", c.BucketName)

	rest := parts[1]
	queryParamsStart := strings.Index(rest, "?")
	if queryParamsStart == -1 || queryParamsStart == len(rest)-1 {
		return nil, errors.Occur(errors.ErrObStorageURIInvalid, "invalid url format, missing query params")
	}

	c.ObjectKey = rest[:queryParamsStart]
	c.ObjectKey = strings.TrimRight(c.ObjectKey, "/")
	log.Info("c.ObjectKey is ", c.ObjectKey)

	queryParams := rest[queryParamsStart+1:]
	fixedQueryParams := strings.Replace(queryParams, "+", "%2B", -1)
	params, err := url.ParseQuery(fixedQueryParams)
	if err != nil {
		return nil, err
	}

	c.Host = params.Get(host)
	c.AccessID = params.Get(accessID)
	c.AccessKey = params.Get(accessKey)
	c.DeleteMode = params.Get(deleteMode)

	log.Info("c.Host is ", c.Host)
	return params, nil
}
