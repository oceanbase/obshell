import { formatMessage } from '@/util/intl';
import { listAlerts } from '@/service/obshell/alarm';
import { ALERT_STATE_MAP, SEVERITY_MAP } from '@/constant/alert';
import { DATE_TIME_FORMAT } from '@/constant/datetime';
import { history } from 'umi';
import { useRequest } from 'ahooks';
import { Button, Card, Form, Space, Table, Tag, Typography } from '@oceanbase/design';
import type { ColumnsType } from 'antd/es/table';
import dayjs from 'dayjs';
import React from 'react';
import AlarmFilter from '../AlarmFilter';

const { Text } = Typography;

export default function Event() {
  const [form] = Form.useForm();
  const {
    data: listAlertsRes,
    run: getListAlerts,
    loading,
  } = useRequest(listAlerts, {
    manual: true,
  });

  const listAlertsData: API.Alert[] = listAlertsRes?.data?.contents || [];

  const columns: ColumnsType<API.Alert> = [
    {
      title: formatMessage({ id: 'OBShell.Alert.Event.AlarmEvent', defaultMessage: '告警事件' }),
      dataIndex: 'summary',
      key: 'summary',
      width: '25%',
    },
    {
      title: formatMessage({ id: 'OBShell.Alert.Event.AlarmLevel', defaultMessage: '告警等级' }),
      dataIndex: 'severity',
      key: 'severity',
      sorter: (preRecord, curRecord) => {
        return SEVERITY_MAP[preRecord.severity].weight - SEVERITY_MAP[curRecord.severity].weight;
      },
      render: (severity: API.Severity) => (
        <Tag color={SEVERITY_MAP[severity]?.color}>{SEVERITY_MAP[severity]?.label}</Tag>
      ),
    },
    {
      title: formatMessage({ id: 'OBShell.Alert.Event.AlarmStatus', defaultMessage: '告警状态' }),
      dataIndex: 'status',
      key: 'status',
      sorter: (preRecord, curRecord) => {
        return (
          ALERT_STATE_MAP[preRecord.status.state].weight -
          ALERT_STATE_MAP[curRecord.status.state].weight
        );
      },
      render: (status: API.Status) => (
        <Tag color={ALERT_STATE_MAP[status?.state]?.color}>
          {ALERT_STATE_MAP[status?.state]?.text || '-'}
        </Tag>
      ),
    },
    {
      title: formatMessage({
        id: 'OBShell.Alert.Event.GenerationTime',
        defaultMessage: '产生时间',
      }),
      dataIndex: 'starts_at',
      key: 'starts_at',
      sorter: (preRecord, curRecord) => curRecord.starts_at - preRecord.starts_at,
      render: (startsAt: number) => <Text>{dayjs.unix(startsAt).format(DATE_TIME_FORMAT)}</Text>,
    },
    {
      title: formatMessage({ id: 'OBShell.Alert.Event.EndTime', defaultMessage: '结束时间' }),
      dataIndex: 'ends_at',
      key: 'ends_at',
      sorter: (preRecord, curRecord) => curRecord.ends_at - preRecord.ends_at,
      render: (endsAt: number) => <Text>{dayjs.unix(endsAt).format(DATE_TIME_FORMAT)}</Text>,
    },
    {
      title: formatMessage({ id: 'OBShell.Alert.Event.Operation', defaultMessage: '操作' }),
      key: 'action',
      render: (_, record) => (
        <Space>
          <Button
            disabled={record?.status?.state !== 'active'}
            // style={{ paddingLeft: 0 }}
            type="link"
            onClick={() => {
              history.push(
                `/alert/shield?instance=${JSON.stringify(record.instance)}&label=${JSON.stringify(
                  record.labels?.map((label: any) => ({
                    name: label.key,
                    value: label.value,
                  }))
                )}&rule=${record.rule}`
              );
            }}
          >
            {formatMessage({ id: 'OBShell.Alert.Event.Shielding', defaultMessage: '屏蔽' })}
          </Button>
        </Space>
      ),
    },
  ];

  return (
    <Space style={{ width: '100%' }} direction="vertical" size="large">
      <Card>
        <AlarmFilter depend={getListAlerts} form={form} type="event" />
      </Card>
      <Card
        title={
          <h2 style={{ marginBottom: 0 }}>
            {formatMessage({ id: 'OBShell.Alert.Event.EventList', defaultMessage: '事件列表' })}
          </h2>
        }
      >
        <Table
          loading={loading}
          columns={columns}
          dataSource={listAlertsData}
          rowKey="fingerprint"
          pagination={{ simple: true }}
        />
      </Card>
    </Space>
  );
}
