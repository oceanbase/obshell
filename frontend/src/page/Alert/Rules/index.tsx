import { formatMessage } from '@/util/intl';
import { listRules } from '@/service/obshell/alarm';
import { SEVERITY_MAP } from '@/constant/alert';
import { useRequest } from 'ahooks';
import { Card, Form, Space, Table, Tag, Typography } from '@oceanbase/design';
import type { ColumnsType } from 'antd/es/table';
import React from 'react';
import AlarmFilter from '../AlarmFilter';
import { formatDuration } from '../helper';

const { Text } = Typography;

export default function Rules() {
  const [form] = Form.useForm();
  const {
    data: listRulesRes,
    loading,
    run: getListRules,
  } = useRequest(listRules, {
    manual: true,
  });

  const listRulesData: API.RuleResponse[] =
    listRulesRes?.data?.contents?.sort(
      (pre, cur) => SEVERITY_MAP[cur?.severity]?.weight - SEVERITY_MAP[pre?.severity]?.weight
    ) || [];

  const columns: ColumnsType<API.RuleResponse> = [
    {
      title: formatMessage({
        id: 'OBShell.Alert.Rules.AlarmRuleName',
        defaultMessage: '告警规则名',
      }),
      dataIndex: 'name',
      fixed: true,
      key: 'name',
    },
    {
      title: formatMessage({ id: 'OBShell.Alert.Rules.TriggerRule', defaultMessage: '触发规则' }),
      dataIndex: 'query',
      width: '30%',
      key: 'query',
    },
    {
      title: formatMessage({ id: 'OBShell.Alert.Rules.Duration', defaultMessage: '持续时间' }),
      dataIndex: 'duration',
      key: 'duration',
      render: value => <Text>{formatDuration(value)}</Text>,
    },
    {
      title: formatMessage({ id: 'OBShell.Alert.Rules.ObjectType', defaultMessage: '对象类型' }),
      dataIndex: 'instance_type',
      key: 'instance_type',
    },
    {
      title: formatMessage({ id: 'OBShell.Alert.Rules.AlarmLevel', defaultMessage: '告警等级' }),
      dataIndex: 'severity',
      key: 'severity',
      sorter: (preRecord, curRecord) =>
        SEVERITY_MAP[curRecord.severity].weight - SEVERITY_MAP[preRecord.severity].weight,
      render: (severity: AlarmSeverity) => (
        <Tag color={SEVERITY_MAP[severity]?.color}>{SEVERITY_MAP[severity]?.label}</Tag>
      ),
    },
    {
      title: formatMessage({ id: 'OBShell.Alert.Rules.Type', defaultMessage: '类型' }),
      dataIndex: 'type',
      key: 'type',
      filters: [
        {
          text: formatMessage({ id: 'OBShell.Alert.Rules.Custom', defaultMessage: '自定义' }),
          value: 'customized',
        },
        {
          text: formatMessage({ id: 'OBShell.Alert.Rules.Default', defaultMessage: '默认' }),
          value: 'builtin',
        },
      ],

      onFilter: (value, record) => record.type === value,
      render: type => (
        <Text>
          {type === 'builtin'
            ? formatMessage({ id: 'OBShell.Alert.Rules.Default', defaultMessage: '默认' })
            : formatMessage({ id: 'OBShell.Alert.Rules.Custom', defaultMessage: '自定义' })}
        </Text>
      ),
    },
  ];

  return (
    <Space style={{ width: '100%' }} direction="vertical" size="large">
      <Card>
        <AlarmFilter depend={getListRules} form={form} type="rules" />
      </Card>
      <Card
        title={
          <h2 style={{ marginBottom: 0 }}>
            {formatMessage({ id: 'OBShell.Alert.Rules.RuleList', defaultMessage: '规则列表' })}
          </h2>
        }
      >
        <Table
          columns={columns}
          rowKey="name"
          dataSource={listRulesData}
          scroll={{ x: 1400 }}
          loading={loading}
        />
      </Card>
    </Space>
  );
}
