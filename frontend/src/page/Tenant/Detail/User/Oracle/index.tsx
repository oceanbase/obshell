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
import { history, useSelector } from 'umi';
import React, { useState, useEffect } from 'react';
import { Button, Space } from '@oceanbase/design';
import { PageContainer } from '@oceanbase/ui';
import { useRequest } from 'ahooks';
import { listUsers, listRoles } from '@/service/obshell/tenant';
import Empty from '@/component/Empty';
import MyInput from '@/component/MyInput';
import MyCard from '@/component/MyCard';
import ContentWithInfo from '@/component/ContentWithInfo';
import UserList from './UserList';
import RoleList from './RoleList';
import AddOracleUserOrRoleDrawer from './Component/AddOracleUserOrRoleDrawer';
import useStyles from './index.style';
import ContentWithReload from '@/component/ContentWithReload';

export interface IndexProps {
  pathname?: string;
  tenantName: string;
}

const Index: React.FC<IndexProps> = ({ pathname, tenantId, tenantName }) => {
  const { styles } = useStyles();
  const { tenantData, precheckResult } = useSelector((state: DefaultRootState) => state.tenant);

  const [keyword, setKeyword] = useState('');
  const pathnameList: string[] = pathname?.split('/') || [];
  const [tab, setTab] = useState(pathnameList[pathnameList?.length - 1]);

  useEffect(() => {
    if (tab === 'user') {
      getDbUserList({
        name: tenantName,
      });
    } else if (tab === 'role') {
      getDbRoleList({
        name: tenantName,
      });
    }
  }, [tab]);

  const ready =
    tenantData?.tenant_name === 'sys' ||
    (Object.keys(precheckResult)?.length > 0 && precheckResult?.is_connectable);

  // 新增用户 / 角色
  const [addUserVisible, setAddUserVisible] = useState(false);
  //  新增的是用户还是角色
  const [addIsUser, setAddIsUser] = useState('user');

  // 获取用户列表
  const {
    run: getDbUserList,
    data: dbUserListData,
    refresh: refreshDbUser,
    loading: dbUserLoading,
  } = useRequest(listUsers, {
    manual: true,
    ready,
    defaultParams: [
      {
        name: tenantName,
      },
    ],
    refreshDeps: [ready],
  });

  const dbUserList: API.ObUser[] = dbUserListData?.data?.contents || [];
  // 获取角色列表
  const {
    run: getDbRoleList,
    data: dbRoleListData,
    refresh: refreshDbRole,
    loading: dbRoleLoading,
  } = useRequest(listRoles, {
    manual: true,
    ready,
    defaultParams: [
      {
        name: tenantName,
      },
    ],
    refreshDeps: [ready],
  });

  const dbRoleList: API.ObRole[] = dbRoleListData?.data?.contents || [];

  return tenantData.status === 'CREATING' ? (
    <Empty
      title={formatMessage({ id: 'ocp-express.User.Oracle.NoData', defaultMessage: '暂无数据' })}
      description={formatMessage({
        id: 'ocp-express.User.Oracle.TheTenantIsBeingCreated',
        defaultMessage: '租户正在创建中，请等待租户创建完成',
      })}
    >
      <Button
        type="primary"
        onClick={() => {
          history.push(`/tenant/${tenantName}`);
        }}
      >
        {formatMessage({
          id: 'ocp-express.User.Oracle.AccessTheOverviewPage',
          defaultMessage: '访问总览页',
        })}
      </Button>
    </Empty>
  ) : (
    <PageContainer
      className={styles.container}
      ghost={true}
      header={{
        title: (
          <Space>
            <ContentWithReload
              spin={dbUserLoading || dbRoleLoading}
              content={formatMessage({
                id: 'ocp-express.User.Oracle.UserManagement',
                defaultMessage: '用户管理',
              })}
              onClick={() => {
                if (tab === 'user') {
                  refreshDbUser();
                } else if (tab === 'role') {
                  refreshDbRole();
                }
              }}
            />
            <ContentWithInfo
              content={
                <span
                  style={{
                    marginLeft: 8,
                  }}
                >
                  {formatMessage({
                    id: 'ocp-express.User.Oracle.SystemPermissionsRolesAndAccessibleObjectsCanBe',
                    defaultMessage: '系统权限、拥有角色和可访问对象可进入用户和角色详情页进行修改',
                  })}
                </span>
              }
              size={0}
            />
          </Space>
        ),

        extra: (
          <Space>
            <Button
              data-aspm-click="c304260.d308767"
              data-aspm-desc="Oracle 角色列表-新建角色"
              data-aspm-param={``}
              data-aspm-expo
              onClick={() => {
                setAddUserVisible(true);
                setAddIsUser('role');
              }}
            >
              {formatMessage({
                id: 'ocp-express.User.Oracle.CreateARole',
                defaultMessage: '新建角色',
              })}
            </Button>
            <Button
              data-aspm-click="c304262.d308773"
              data-aspm-desc="Oracle 用户列表-新建用户"
              data-aspm-param={``}
              data-aspm-expo
              type="primary"
              onClick={() => {
                setAddUserVisible(true);
                setAddIsUser('user');
              }}
            >
              {formatMessage({
                id: 'ocp-express.User.Oracle.CreateUser',
                defaultMessage: '新建用户',
              })}
            </Button>
          </Space>
        ),
      }}
    >
      <MyCard
        bordered={false}
        className={`card-without-padding ${styles.card}`}
        activeTabKey={tab}
        onTabChange={key => {
          setTab(key);
          history.push({
            pathname: `/tenant/${tenantName}/user${key === 'role' ? '/role' : ''}`,
          });
        }}
        tabList={[
          {
            key: 'user',
            tab: formatMessage({
              id: 'ocp-express.User.Oracle.UserList',
              defaultMessage: '用户列表',
            }),
          },

          {
            key: 'role',
            tab: formatMessage({
              id: 'ocp-express.User.Oracle.RoleList',
              defaultMessage: '角色列表',
            }),
          },
        ]}
        tabBarExtraContent={
          <MyInput.Search
            allowClear={true}
            onSearch={(value: string) => setKeyword(value)}
            onChange={e => setKeyword(e.target.value)}
            placeholder={formatMessage({
              id: 'ocp-express.User.Oracle.EnterAUsernameAndRole',
              defaultMessage: '请输入用户名、角色名',
            })}
            className="search-input"
          />
        }
      >
        {tab === 'user' && (
          <UserList
            dbUserLoading={dbUserLoading}
            refreshDbUser={refreshDbUser}
            dbUserList={dbUserList?.filter(
              item =>
                !keyword ||
                (item.user_name && item.user_name.includes(keyword)) ||
                (item.granted_roles &&
                  item.granted_roles.filter(o => o.includes(keyword)).length > 0)
            )}
          />
        )}

        {tab === 'role' && (
          <RoleList
            dbRoleLoading={dbRoleLoading}
            refreshDbRole={refreshDbRole}
            dbRoleList={dbRoleList?.filter(
              item =>
                !keyword ||
                (item.role && item.role.includes(keyword)) ||
                (item.granted_roles &&
                  item.granted_roles.filter(o => o.includes(keyword)).length > 0)
            )}
          />
        )}
      </MyCard>
      <AddOracleUserOrRoleDrawer
        visible={addUserVisible}
        addIsUser={addIsUser}
        onCancel={() => {
          setAddUserVisible(false);
        }}
        onSuccess={() => {
          setAddUserVisible(false);
          if (tab === 'user') {
            refreshDbUser();
          } else if (tab === 'role') {
            refreshDbRole();
          }
        }}
      />
    </PageContainer>
  );
};

export default Index;
