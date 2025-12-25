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
import React, { useState } from 'react';
import { Card, message, Modal, Table, TableProps, Tooltip, Button } from '@oceanbase/design';
import { PageContainer } from '@oceanbase/ui';
import { sortByString, sortByMoment, byte2MB } from '@oceanbase/util';
import { PAGINATION_OPTION_10 } from '@/constant';
import useDocumentTitle from '@/hook/useDocumentTitle';
import { isEnglish } from '@/util';
import { formatTime } from '@/util/datetime';
import MyInput from '@/component/MyInput';
import ContentWithReload from '@/component/ContentWithReload';
import { useRequest } from 'ahooks';
import { upgradePkgInfo } from '@/service/obshell/upgrade';
import UploadPackageDrawer from '@/component/UploadPackageDrawer';
import { uniq } from 'lodash';
import { deletePackage } from '@/service/obshell/package';

export interface PackageProps {
  location: {
    query: {
      keyword?: string;
    };
  };
}

const PackagePage: React.FC<PackageProps> = ({
  location: {
    query: { keyword: defaultKeyword },
  },
}) => {
  useDocumentTitle(
    formatMessage({
      id: 'ocp-express.src.util.menu.SystemManagement',
      defaultMessage: '系统管理',
    })
  );

  const { data, loading, refresh } = useRequest(upgradePkgInfo);
  const packageList: API.UpgradePkgInfo[] = data?.data?.contents || [];

  const [keyword, setKeyword] = useState(defaultKeyword);
  const [visible, setVisible] = useState(false);

  const handleDeletePackage = (record: API.UpgradePkgInfo) => {
    Modal.confirm({
      title: formatMessage({
        id: 'OBShell.page.Package.ConfirmSoftwarePackageDeletion',
        defaultMessage: '确认删除软件包',
      }),
      content: formatMessage(
        {
          id: 'OBShell.page.Package.AreYouSureYouWant',
          defaultMessage: '确认删除软件包 {recordName} 吗？此操作不可撤销。',
        },
        { recordName: record.name }
      ),
      okText: formatMessage({ id: 'OBShell.page.Package.Confirm', defaultMessage: '确认' }),
      cancelText: formatMessage({ id: 'OBShell.page.Package.Cancel', defaultMessage: '取消' }),
      okType: 'danger',
      onOk: async () => {
        try {
          const res = await deletePackage({ pkg_id: record.pkg_id! });
          if (res.successful) {
            message.success(
              formatMessage({
                id: 'OBShell.page.Package.SoftwarePackageDeletedSuccessfully',
                defaultMessage: '删除软件包成功',
              })
            );
          } else {
            message.error(
              formatMessage({
                id: 'OBShell.page.Package.FailedToDeletePackage',
                defaultMessage: '删除软件包失败',
              })
            );
          }
          refresh();
        } catch (error) {
          message.error(
            formatMessage({
              id: 'OBShell.page.Package.FailedToDeletePackage',
              defaultMessage: '删除软件包失败',
            })
          );
        }
      },
    });
  };

  const columns: TableProps<API.UpgradePkgInfo>['columns'] = [
    {
      title: formatMessage({
        id: 'ocp-express.page.Package.PackageName',
        defaultMessage: '软件包名',
      }),
      dataIndex: 'name',
      width: isEnglish() ? 200 : 220,
      sorter: (a, b) => sortByString(a, b, 'name'),
      render: (text: string) => <div style={{ wordBreak: 'break-all' }}>{text}</div>,
    },

    {
      title: formatMessage({
        id: 'ocp-express.page.Package.Version',
        defaultMessage: '版本',
      }),
      dataIndex: 'version',
      width: 270,
      filters: uniq(packageList.map(item => item.version || '')).map(item => ({
        text: item,
        value: item,
      })),
      onFilter: (value: string, record: API.UpgradePkgInfo) => record.version === value,
      render: (text, record: API.UpgradePkgInfo) => {
        return `${record?.version}-${record.release_distribution}`;
      },
    },

    {
      title: formatMessage({
        id: 'ocp-express.page.Package.HardwareArchitecture',
        defaultMessage: '硬件架构',
      }),

      dataIndex: 'architecture',
      filters: uniq(packageList.map(item => item.architecture) || []).map(item => ({
        text: item,
        value: item,
      })),
      onFilter: (value: string, record: API.UpgradePkgInfo) => record.architecture === value,
    },

    {
      title: formatMessage({ id: 'ocp-v2.page.Package.SizeMb', defaultMessage: '大小（MB）' }),
      dataIndex: 'payload_size',
      // 后端返回的是 bit
      render: (text: number) => <span>{byte2MB(text)}</span>,
    },

    {
      title: 'MD5',
      dataIndex: 'md5',
      ellipsis: true,
      render: (text: string) => (
        <Tooltip placement="topLeft" title={text}>
          <span>{text}</span>
        </Tooltip>
      ),
    },

    {
      title: formatMessage({ id: 'ocp-v2.page.Package.UploadTime', defaultMessage: '上传时间' }),
      dataIndex: 'gmt_modify',
      width: 250,
      sorter: (a, b) => sortByMoment(a, b, 'gmt_modify'),
      render: (text: string) => formatTime(text),
    },
    {
      title: formatMessage({ id: 'OBShell.page.Package.Operation', defaultMessage: '操作' }),
      width: 100,
      render: (text: string, record: API.UpgradePkgInfo) => (
        <Button type="link" onClick={() => handleDeletePackage(record)}>
          {formatMessage({ id: 'OBShell.page.Package.Delete', defaultMessage: '删除' })}
        </Button>
      ),
    },
  ];

  return (
    <PageContainer
      ghost={true}
      header={{
        title: (
          <ContentWithReload
            content={formatMessage({
              id: 'ocp-express.page.Package.SoftwarePackage',
              defaultMessage: '软件包',
            })}
            spin={loading}
            onClick={refresh}
          />
        ),
      }}
    >
      <Card
        bordered={false}
        title={
          <MyInput.Search
            value={keyword}
            allowClear={true}
            onChange={e => {
              setKeyword(e.target.value);
            }}
            placeholder={formatMessage({
              id: 'ocp-v2.page.Package.SearchPackageName',
              defaultMessage: '搜索软件包名称',
            })}
            className="search-input"
          />
        }
        extra={
          <Button
            type="primary"
            onClick={() => {
              setVisible(true);
            }}
          >
            {formatMessage({
              id: 'ocp-v2.page.Package.UploadSoftwarePackage',
              defaultMessage: '上传软件包',
            })}
          </Button>
        }
        className="card-without-padding"
      >
        <Table
          loading={loading}
          dataSource={packageList.filter(
            item => !keyword || (item.name && item.name.includes(keyword))
          )}
          columns={columns}
          rowKey={record => record.pkg_id}
          pagination={PAGINATION_OPTION_10}
        />
      </Card>

      <UploadPackageDrawer
        open={visible}
        onCancel={() => {
          setVisible(false);
        }}
        onSuccess={() => {
          refresh();
        }}
      />
    </PageContainer>
  );
};

export default PackagePage;
