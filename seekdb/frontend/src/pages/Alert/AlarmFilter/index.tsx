import { LEVER_OPTIONS_ALARM, SEVERITY_MAP } from '@/constant/alert';
import { Alert } from '@/typing/env/alert';
import { formatMessage } from '@/util/intl';

import useSeekdbInfo from '@/hook/usSeekdbInfo';
import type { FormInstance } from '@oceanbase/design';
import { Button, Col, DatePicker, Form, Input, Row, Select, Space, Tag } from '@oceanbase/design';
import { useDebounceFn, useUpdateEffect } from 'ahooks';
import React, { useEffect, useState } from 'react';

interface AlarmFilterProps {
  form: FormInstance<any>;
  depend: (body: any) => void;
  type: 'event' | 'shield' | 'rules';
}

const DEFAULT_VISIBLE_CONFIG = {
  level: true,
  keyword: true,
  startTime: true,
  endTime: true,
};

const getInitialVisibleConfig = (type: 'event' | 'shield' | 'rules') => {
  if (type === 'event') {
    return {
      level: true,
      keyword: true,
      startTime: true,
      endTime: true,
    };
  }
  if (type === 'shield') {
    return {
      keyword: true,
      startTime: false,
      endTime: false,
      level: false,
    };
  }
  if (type === 'rules') {
    return {
      level: true,
      keyword: true,
      startTime: false,
      endTime: false,
    };
  }
  return DEFAULT_VISIBLE_CONFIG;
};

export default function AlarmFilter({ form, type, depend }: AlarmFilterProps) {
  const { seekdbInfoData } = useSeekdbInfo();
  const [visibleConfig, setVisibleConfig] = useState(() => getInitialVisibleConfig(type));

  const { run: debounceDepend } = useDebounceFn(depend, { wait: 500 });
  const formData: any = Form.useWatch([], form);

  // 初始化表单默认值
  useEffect(() => {
    if (seekdbInfoData?.cluster_name) {
      form.setFieldsValue({
        instance: {
          type: 'obcluster',
          obcluster: seekdbInfoData?.cluster_name,
        },
      });
    }
  }, [seekdbInfoData?.cluster_name]);

  useEffect(() => {
    setVisibleConfig(getInitialVisibleConfig(type));
  }, [type]);

  useUpdateEffect(() => {
    const filter: { [T: string]: any } = {};
    Object.keys(formData).forEach(key => {
      if (formData[key]) {
        if (typeof formData[key] === 'string') {
          filter[key] = formData[key];
        } else if (key === 'start_time' || key === 'end_time') {
          filter[key] = Math.ceil(formData[key].valueOf() / 1000);
        } else if (key === 'instance' && formData[key]?.type) {
          const temp: { [key: string]: any } = {};

          Object.keys(formData[key]).forEach(innerKey => {
            if (formData[key][innerKey]) {
              temp[innerKey] = formData[key][innerKey];
            }
          });
          filter[key] = temp;
        }
      }
    });
    if (filter.instance) {
      if (Object.keys(filter.instance).length === 1 && filter.instance?.type) {
        filter.instance_type = filter.instance.type;
        delete filter.instance;
      }
    }
    debounceDepend(filter);
  }, [formData]);

  return (
    <Form
      form={form}
      labelCol={{
        span: 9,
      }}
      wrapperCol={{
        span: 18,
      }}
    >
      {/* 隐藏字段：type 和 instance */}
      <Form.Item name={['instance', 'type']} hidden>
        <Input />
      </Form.Item>
      <Form.Item name={['instance', 'obcluster']} hidden>
        <Input />
      </Form.Item>

      <Row>
        {visibleConfig.level && (
          <Col span={6}>
            <Form.Item
              label={formatMessage({
                id: 'OBShell.Alert.AlarmFilter.AlarmLevel',
                defaultMessage: '告警等级',
              })}
              name={'severity'}
            >
              <Select
                allowClear
                placeholder={formatMessage({
                  id: 'OBShell.Alert.AlarmFilter.PleaseSelect',
                  defaultMessage: '请选择',
                })}
                options={LEVER_OPTIONS_ALARM!.map(item => ({
                  value: item.value,
                  label: (
                    <Tag color={SEVERITY_MAP[item.value as Alert.AlarmLevel]?.color}>
                      {item.label}
                    </Tag>
                  ),
                }))}
              />
            </Form.Item>
          </Col>
        )}

        {visibleConfig.keyword && (
          <Col span={6}>
            <Form.Item
              label={formatMessage({
                id: 'OBShell.Alert.AlarmFilter.Keywords',
                defaultMessage: '关键词',
              })}
              name={'keyword'}
            >
              <Input
                placeholder={formatMessage({
                  id: 'OBShell.Alert.AlarmFilter.PleaseEnterAKeyword',
                  defaultMessage: '请输入关键词',
                })}
              />
            </Form.Item>
          </Col>
        )}

        {visibleConfig.startTime && (
          <Col span={6}>
            <Form.Item
              label={formatMessage({
                id: 'OBShell.Alert.AlarmFilter.StartTime',
                defaultMessage: '开始时间',
              })}
              name={'start_time'}
            >
              <DatePicker
                style={{ width: '100%' }}
                placeholder={formatMessage({
                  id: 'OBShell.Alert.AlarmFilter.PleaseSelect',
                  defaultMessage: '请选择',
                })}
                showTime
              />
            </Form.Item>
          </Col>
        )}

        {visibleConfig.endTime && (
          <Col span={6}>
            <Form.Item
              name={'end_time'}
              label={formatMessage({
                id: 'OBShell.Alert.AlarmFilter.EndTime',
                defaultMessage: '结束时间',
              })}
            >
              <DatePicker
                placeholder={formatMessage({
                  id: 'OBShell.Alert.AlarmFilter.PleaseSelect',
                  defaultMessage: '请选择',
                })}
                style={{ width: '100%' }}
                showTime
              />
            </Form.Item>
          </Col>
        )}
      </Row>
      <Space style={{ float: 'right' }}>
        <Button type="link" onClick={() => form.resetFields()}>
          {formatMessage({ id: 'OBShell.Alert.AlarmFilter.Reset', defaultMessage: '重置' })}
        </Button>
      </Space>
    </Form>
  );
}
