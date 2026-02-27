import PreText from '@/component/PreText';
import showDeleteConfirm from '@/component/customModal/showDeleteConfirm';
import { SHILED_STATUS_MAP } from '@/constant/alert';
import { DATE_TIME_FORMAT } from '@/constant/datetime';
import { deleteSilencer, listSilencers } from '@/service/obshell/alarm';
import { Alert } from '@/typing/env/alert';
import { formatMessage } from '@/util/intl';
import { Button, Card, Form, Space, Table, Tag, Tooltip, Typography } from '@oceanbase/design';
import { history, useLocation } from '@umijs/max';
import { useRequest } from 'ahooks';
import type { ColumnsType } from 'antd/es/table';
import dayjs from 'dayjs';
import React, { useState } from 'react';
import AlarmFilter from '../AlarmFilter';
import { sortAlarmShielding } from '../helper';
import ShieldDrawerForm from './ShieldDrawerForm';
const { Text } = Typography;

type InstancesRender = {
  [key: string]: string[];
};

export default function Shield() {
  const [form] = Form.useForm();
  const location = useLocation();
  const searchParams = new URLSearchParams(location.search);
  const [editShieldId, setEditShieldId] = useState<string>();
  const [drawerOpen, setDrawerOpen] = useState(Boolean(searchParams.get('instance')));
  const {
    data: listSilencersRes,
    refresh,
    run: listSilencersRun,
    loading,
  } = useRequest(listSilencers, {
    manual: true,
  });

  const { run: deleteSilencerRun } = useRequest(deleteSilencer, {
    onSuccess: ({ successful }) => {
      if (successful) {
        refresh();
      }
    },
  });

  const listSilencersData = sortAlarmShielding(listSilencersRes?.data?.contents || []);
  const drawerClose = () => {
    history.push(location.pathname);
    setEditShieldId(undefined);
    setDrawerOpen(false);
  };
  const editShield = (id: string) => {
    setEditShieldId(id);
    setDrawerOpen(true);
  };
  const columns: ColumnsType<API.SilencerResponse> = [
    {
      title: formatMessage({
        id: 'OBShell.Alert.Shield.MaskApplicationObjectType',
        defaultMessage: '屏蔽应用/对象类型',
      }),
      dataIndex: 'instances',
      key: 'type',
      fixed: true,
      render: (instances: API.OBInstance[]) => <Text>{instances?.[0]?.type || '-'}</Text>,
    },
    {
      title: formatMessage({ id: 'OBShell.Alert.Shield.MaskObject', defaultMessage: '屏蔽对象' }),
      dataIndex: 'instances',
      key: 'instances',
      width: 200,
      render: (instances: API.OBInstance[] = []) => {
        const temp: InstancesRender = {};
        for (const instance of instances) {
          Object.keys(instance).forEach((key: keyof API.OBInstance) => {
            if (temp[key]) {
              temp[key] = [...temp[key], instance[key]];
            } else {
              temp[key] = [instance[key]];
            }
          });
        }
        delete temp.type;

        const InstancesRender = () => (
          <div>
            {Object.keys(temp).map((key, index) => (
              <p key={index}>
                {key}：{temp[key]?.join(',') || '-'}
              </p>
            ))}
          </div>
        );

        return (
          <Tooltip title={<InstancesRender />}>
            <div>
              {Object.keys(temp).map(key => (
                <Text ellipsis style={{ width: 200 }}>
                  {key}：{temp[key]?.join(',') || '-'}
                </Text>
              ))}
            </div>
          </Tooltip>
        );
      },
    },
    {
      title: formatMessage({
        id: 'OBShell.Alert.Shield.MaskingAlarmRules',
        defaultMessage: '屏蔽告警规则',
      }),
      dataIndex: 'matchers',
      key: 'matchers',
      width: 300,
      render: rules => {
        return <PreText cols={7} value={rules} />;
      },
    },
    {
      title: formatMessage({
        id: 'OBShell.Alert.Shield.MaskEndTime',
        defaultMessage: '屏蔽结束时间',
      }),
      dataIndex: 'ends_at',
      key: 'ends_at',
      sorter: (preRecord, curRecord) => curRecord.starts_at - preRecord.starts_at,
      render: endsAt => <Text>{dayjs.unix(endsAt).format(DATE_TIME_FORMAT)}</Text>,
    },
    {
      title: formatMessage({ id: 'OBShell.Alert.Shield.Creator', defaultMessage: '创建人' }),
      dataIndex: 'created_by',
      key: 'created_by',
    },
    {
      title: formatMessage({ id: 'OBShell.Alert.Shield.Status', defaultMessage: '状态' }),
      dataIndex: 'status',
      key: 'status',
      sorter: (preRecord, curRecord) =>
        SHILED_STATUS_MAP[curRecord.status.state].weight -
        SHILED_STATUS_MAP[preRecord.status.state].weight,
      render: (status: API.Status) => (
        <Tag color={SHILED_STATUS_MAP[status.state].color}>
          {SHILED_STATUS_MAP[status.state]?.text || '-'}
        </Tag>
      ),
    },
    {
      title: formatMessage({ id: 'OBShell.Alert.Shield.CreationTime', defaultMessage: '创建时间' }),
      dataIndex: 'starts_at',
      key: 'starts_at',
      sorter: (preRecord, curRecord) => curRecord.starts_at - preRecord.starts_at,
      render: startsAt => <Text>{dayjs.unix(startsAt).format(DATE_TIME_FORMAT)}</Text>,
    },
    {
      title: formatMessage({ id: 'OBShell.Alert.Shield.Remarks', defaultMessage: '备注' }),
      dataIndex: 'comment',
      key: 'comment',
    },
    {
      title: formatMessage({ id: 'OBShell.Alert.Shield.Operation', defaultMessage: '操作' }),
      key: 'action',
      fixed: 'right',
      render: (_, record: API.SilencerResponse) => (
        <Space>
          <Button
            onClick={() => editShield(record.id)}
            style={{ paddingLeft: 0 }}
            disabled={record.status.state === 'expired'}
            type="link"
          >
            {formatMessage({ id: 'OBShell.Alert.Shield.Edit', defaultMessage: '编辑' })}
          </Button>
          <Button
            type="link"
            style={record.status.state !== 'expired' ? { color: '#ff4b4b' } : {}}
            disabled={record.status.state === 'expired'}
            onClick={() => {
              showDeleteConfirm({
                title: formatMessage({
                  id: 'OBShell.Alert.Shield.AreYouSureToRemove',
                  defaultMessage: '确定解除该告警屏蔽条件吗？',
                }),
                content: formatMessage({
                  id: 'OBShell.Alert.Shield.ItCannotBeRestoredAfter',
                  defaultMessage: '解除后不可恢复，请谨慎操作',
                }),
                okText: formatMessage({
                  id: 'OBShell.Alert.Shield.Disarm',
                  defaultMessage: '解除',
                }),
                onOk: () => {
                  deleteSilencerRun({ id: record.id });
                },
              });
            }}
          >
            {formatMessage({ id: 'OBShell.Alert.Shield.Unmask', defaultMessage: '解除屏蔽' })}
          </Button>
        </Space>
      ),
    },
  ];

  const formatInstanceParam = (instanceParam: Alert.InstanceParamType) => {
    const { obcluster, observer, obtenant, type } = instanceParam;
    const res: Alert.InstancesType = {
      type,
      obcluster: [obcluster!],
    };
    if (observer) res.observer = [observer];
    if (obtenant) res.obtenant = [obtenant];
    return res;
  };
  const initialValues: Alert.ShieldDrawerInitialValues = {};
  if (searchParams.get('instance')) {
    initialValues.instances = formatInstanceParam(JSON.parse(searchParams.get('instance')!));
  }
  if (searchParams.get('label')) {
    initialValues.matchers = JSON.parse(searchParams.get('label')!);
  }
  if (searchParams.get('rule')) {
    initialValues.rules = searchParams.get('rule') ? [searchParams.get('rule')!] : undefined;
  }
  return (
    <Space style={{ width: '100%' }} direction="vertical" size="large">
      <Card>
        <AlarmFilter depend={listSilencersRun} form={form} type="shield" />
      </Card>
      <Card
        title={
          <h2 style={{ marginBottom: 0 }}>
            {formatMessage({ id: 'OBShell.Alert.Shield.MaskList', defaultMessage: '屏蔽列表' })}
          </h2>
        }
        extra={
          <Button type="primary" onClick={() => setDrawerOpen(true)}>
            {formatMessage({ id: 'OBShell.Alert.Shield.NewShield', defaultMessage: '新建屏蔽' })}
          </Button>
        }
      >
        <Table
          columns={columns}
          dataSource={listSilencersData}
          rowKey="id"
          pagination={{ simple: true }}
          scroll={{ x: 1800 }}
          sticky
          loading={loading}
        />
      </Card>
      <ShieldDrawerForm
        width={880}
        initialValues={initialValues}
        onClose={drawerClose}
        submitCallback={refresh}
        open={drawerOpen}
        id={editShieldId}
      />
    </Space>
  );
}
