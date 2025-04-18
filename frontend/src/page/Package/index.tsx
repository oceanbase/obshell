/*
 * Copyright (c) 2023 OceanBase
 * OCP Express is licensed under Mulan PSL v2.
 * You can use this software according to the terms and conditions of the Mulan PSL v2.
 * You may obtain a copy of Mulan PSL v2 at:
 *          http://license.coscl.org.cn/MulanPSL2
 * THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
 * EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
 * MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
 * See the Mulan PSL v2 for more details.
 */

import { formatMessage } from '@/util/intl';
import { connect, useDispatch } from 'umi';
import React, { useState, useEffect } from 'react';
import { Card, Table, Tooltip, Typography, Form, Tag, Modal } from '@oceanbase/design';
import { PageContainer } from '@oceanbase/ui';
import { sortByString, sortByMoment, byte2MB } from '@oceanbase/util';
import { PAGINATION_OPTION_10 } from '@/constant';
import useDocumentTitle from '@/hook/useDocumentTitle';
import { isEnglish } from '@/util';
import { formatTime } from '@/util/datetime';
import MyInput from '@/component/MyInput';
import ContentWithReload from '@/component/ContentWithReload';
import ContentWithQuestion from '@/component/ContentWithQuestion';
import { useRequest } from 'ahooks';
import { upgradePkgInfo, upgradePkgRoute } from '@/service/obshell/upgrade';
import { Button } from 'antd';

const FormItem = Form.Item;

const { Paragraph } = Typography;

export interface PackageProps {
  location: {
    query: {
      keyword?: string;
    };
  };
  submitLoading: boolean;
}

const PackagePage: React.FC<PackageProps> = ({
  location: {
    query: { keyword: defaultKeyword },
  },
  submitLoading,
}) => {
  useDocumentTitle(
    formatMessage({ id: 'ocp-express.page.Package.SystemParameters', defaultMessage: '系统配置' })
  );

  const { data, loading, refresh } = useRequest(upgradePkgInfo);
  const packageList = data?.data?.contents || [];
  console.log(packageList, 'packageList');

  // architecture?: string;
  // chunkCount?: number;
  // distribution?: string;
  // gmtModify?: string;
  // md5?: string;
  // name?: string;
  // payloadSize?: number;
  // pkgId?: number;
  // release?: string;
  // releaseDistribution?: string;
  // size?: number;
  // upgradeDepYaml?: string;
  // version?: string;

  // 更新 Package 要求具有 Package:*:UPDATE 的权限，而不仅仅是单个 Package 的更新权限
  const [form] = Form.useForm();
  const [keyword, setKeyword] = useState(defaultKeyword);
  const [visible, setVisible] = useState(false);
  const [currentRecord, setCurrentRecord] = useState<API.PackageMeta | null>(null);

  const columns = [
    {
      title: formatMessage({ id: 'ocp-v2.page.Package.PackageName', defaultMessage: '软件包名' }),
      dataIndex: 'name',
      width: isEnglish() ? 200 : 220,
      sorter: true,
      render: (text: string) => <div style={{ wordBreak: 'break-all' }}>{text}</div>,
    },

    {
      title: formatMessage({ id: 'ocp-v2.page.Package.Type', defaultMessage: '类型' }),
      dataIndex: 'type',
      // filters: PACKAGE_TYPE_LIST.map(item => ({
      //   text: item.label,
      //   value: item.value,
      // })),

      // render: (text: string) => findByValue(PACKAGE_TYPE_LIST, text)?.label || '-',
    },

    {
      title: formatMessage({ id: 'ocp-v2.page.Package.Version', defaultMessage: '版本' }),
      dataIndex: 'version',
      filters: (data?.data?.versions || []).map(item => ({
        text: item,
        value: item,
      })),
    },

    {
      title: formatMessage({ id: 'ocp-v2.page.Package.System', defaultMessage: '系统' }),
      dataIndex: 'operatingSystem',
      filters: (data?.data?.operatingSystems || []).map(item => ({
        text: item,
        value: item,
      })),
    },

    {
      title: formatMessage({
        id: 'ocp-v2.page.Package.HardwareArchitecture',
        defaultMessage: '硬件架构',
      }),

      dataIndex: 'architecture',
      ...(isEnglish()
        ? {
            width: 120,
          }
        : {}),
      filters: (data?.data?.architectures || []).map(item => ({
        text: item,
        value: item,
      })),
    },

    {
      title: formatMessage({ id: 'ocp-v2.page.Package.SizeMb', defaultMessage: '大小（MB）' }),
      dataIndex: 'size',
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
      dataIndex: 'gmtModify',
      sorter: true,
      width: 150,
      render: (text: string) => formatTime(text),
    },
  ];

  return (
    <PageContainer
      ghost={true}
      header={{
        title: <ContentWithReload content={'软件包'} spin={loading} onClick={refresh} />,
      }}
    >
      <Card
        bordered={false}
        title={
          <MyInput.Search
            data-aspm-click="c304243.d308723"
            data-aspm-desc="系统参数列表-搜索参数"
            data-aspm-param={``}
            data-aspm-expo
            value={keyword}
            allowClear={true}
            onChange={e => {
              setKeyword(e.target.value);
            }}
            placeholder={'搜索软件包名称'}
            className="search-input"
          />
        }
        extra={<Button type="primary">上传软件包</Button>}
        className="card-without-padding"
      >
        <Table
          data-aspm="c304243"
          data-aspm-desc="系统参数列表"
          data-aspm-param={``}
          data-aspm-expo
          loading={loading}
          dataSource={packageList.filter(
            item => !keyword || (item.key && item.key.includes(keyword))
          )}
          columns={columns}
          rowKey={record => record.id}
          pagination={PAGINATION_OPTION_10}
        />
      </Card>
    </PageContainer>
  );
};

export default PackagePage;
