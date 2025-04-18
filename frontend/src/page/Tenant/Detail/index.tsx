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
import React, { useEffect } from 'react';
import { Badge, Tooltip } from '@oceanbase/design';
import { toNumber } from 'lodash';
import { findByValue } from '@oceanbase/util';
import { CaretDownFilled } from '@oceanbase/icons';
import { TENANT_STATUS_LIST } from '@/constant/tenant';
import { useBasicMenu, useTenantMenu } from '@/hook/useMenu';
import useDocumentTitle from '@/hook/useDocumentTitle';
import BasicLayout from '@/page/Layout/BasicLayout';
import TaskBubble from '@/component/TaskBubble';
import TenantSelect from '@/component/common/TenantSelect';
import useStyles from './index.style';

interface DetailProps {
  location: {
    pathname: string;
  };
  match: {
    params: {
      tenantName: number;
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
  const { tenantData } = useSelector((state: DefaultRootState) => state.tenant);
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

  console.log(statusItem, tenantData, 'statusItem');
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
                    {formatMessage(
                      {
                        id: 'ocp-express.Tenant.Detail.tenantNametenantName',
                        defaultMessage: '租户 ID：{tenantId}',
                      },
                      { tenantId: tenantData?.tenant_id }
                    )}
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
      <TaskBubble tenantName={tenantName} />
    </BasicLayout>
  );
};

export default Detail;
