import { LEVER_OPTIONS_ALARM, OBJECT_OPTIONS_ALARM, SEVERITY_MAP } from '@/constant/alert';
import { Alert } from '@/typing/env/alert';
import { formatMessage } from '@/util/intl';

import { Space } from '@oceanbase/design';
import { DownOutlined, UpOutlined } from '@oceanbase/icons';
import { useModel } from '@umijs/max';
import { useDebounceFn, useUpdateEffect } from 'ahooks';
import type { FormInstance } from 'antd';
import { Button, Col, DatePicker, Form, Input, Row, Select, Tag } from 'antd';
import { flatten } from 'lodash';
import React, { useEffect, useState } from 'react';
import { getSelectList } from '../helper';

interface AlarmFilterProps {
  form: FormInstance<any>;
  depend: (body: any) => void;
  type: 'event' | 'shield' | 'rules';
}

const DEFAULT_VISIBLE_CONFIG = {
  objectType: true,
  object: true,
  level: true,
  keyword: true,
  startTime: true,
  endTime: true,
};

const getInitialVisibleConfig = (type: 'event' | 'shield' | 'rules') => {
  if (type === 'event') {
    return {
      objectType: true,
      object: true,
      level: true,
      keyword: true,
      startTime: true,
      endTime: true,
    };
  }
  if (type === 'shield') {
    return {
      objectType: true,
      object: true,
      keyword: true,
      startTime: false,
      endTime: false,
      level: false,
    };
  }
  if (type === 'rules') {
    return {
      objectType: true,
      level: true,
      keyword: true,
      startTime: false,
      endTime: false,
      object: false,
    };
  }
  return DEFAULT_VISIBLE_CONFIG;
};

export default function AlarmFilter({ form, type, depend }: AlarmFilterProps) {
  const { clusterList, tenantList } = useModel('alarm');
  const [isExpand, setIsExpand] = useState(true);
  const [visibleConfig, setVisibleConfig] = useState(() => getInitialVisibleConfig(type));

  const getOptionsFromType = (type: 'obcluster' | 'obtenant' | 'observer') => {
    if (!type || !clusterList || (type === 'obtenant' && !tenantList)) return [];
    const list = getSelectList(clusterList, type, tenantList);
    if (type === 'obcluster') {
      return list?.map(clusterName => ({
        value: clusterName,
        label: clusterName,
      }));
    }
    if (type === 'obtenant') {
      return flatten(
        (list as Alert.TenantsList[]).map(cluster =>
          cluster?.tenants?.map(item => ({
            value: item,
            label: item,
          }))
        )
      );
    }
    if (type === 'observer') {
      return flatten(
        (list as Alert.ServersList[]).map(cluster =>
          cluster?.servers?.map(item => ({
            value: item,
            label: item,
          }))
        )
      );
    }
  };

  const { run: debounceDepend } = useDebounceFn(depend, { wait: 500 });
  const formData: any = Form.useWatch([], form);

  // 当 type 变化时，重置为对应类型的初始配置
  useEffect(() => {
    setVisibleConfig(getInitialVisibleConfig(type));
    setIsExpand(true); // 重置展开状态
  }, [type]);

  useUpdateEffect(() => {
    if (isExpand) {
      setVisibleConfig({
        objectType: true,
        object: true,
        level: true,
        keyword: true,
        startTime: true,
        endTime: true,
      });
    } else {
      setVisibleConfig({
        objectType: true,
        object: true,
        level: true,
        keyword: false,
        startTime: false,
        endTime: false,
      });
    }
  }, [isExpand]);

  useUpdateEffect(() => {
    const filter: { [T: string]: unknown } = {};
    Object.keys(formData).forEach(key => {
      if (formData[key]) {
        if (typeof formData[key] === 'string') {
          filter[key] = formData[key];
        } else if (key === 'start_time' || key === 'end_time') {
          filter[key] = Math.ceil(formData[key].valueOf() / 1000);
        } else if (key === 'instance' && formData[key]?.type) {
          const temp = {};
          if (formData[key]?.obtenant) {
            formData[key].obcluster = clusterList?.[0]?.cluster_name;
          }
          if (formData[key]?.observer) {
            formData[key].obcluster = clusterList[0]?.cluster_name;
          }
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
      <Row>
        {visibleConfig.objectType && (
          <Col span={6}>
            <Form.Item
              label={formatMessage({
                id: 'OBShell.Alert.AlarmFilter.ObjectType',
                defaultMessage: '对象类型',
              })}
              name={['instance', 'type']}
            >
              <Select
                allowClear
                placeholder={formatMessage({
                  id: 'OBShell.Alert.AlarmFilter.PleaseSelect',
                  defaultMessage: '请选择',
                })}
                options={OBJECT_OPTIONS_ALARM}
              />
            </Form.Item>
          </Col>
        )}

        {visibleConfig.object && (
          <Col span={6}>
            <Form.Item noStyle dependencies={[['instance', 'type']]}>
              {({ getFieldValue }) => {
                return (
                  <Form.Item
                    name={['instance', getFieldValue(['instance', 'type'])]}
                    label={
                      type === 'event'
                        ? formatMessage({
                            id: 'OBShell.Alert.AlarmFilter.AlarmObject',
                            defaultMessage: '告警对象',
                          })
                        : formatMessage({
                            id: 'OBShell.Alert.AlarmFilter.MaskObject',
                            defaultMessage: '屏蔽对象',
                          })
                    }
                  >
                    <Select
                      style={{ width: '100%' }}
                      allowClear
                      options={getOptionsFromType(getFieldValue(['instance', 'type']))}
                      placeholder={formatMessage({
                        id: 'OBShell.Alert.AlarmFilter.PleaseSelect',
                        defaultMessage: '请选择',
                      })}
                    />
                  </Form.Item>
                );
              }}
            </Form.Item>
          </Col>
        )}

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
        <Col span={6} offset={type === 'event' && isExpand ? 6 : 0}>
          <Space style={{ float: 'right' }}>
            <Button type="link" onClick={() => form.resetFields()}>
              {formatMessage({ id: 'OBShell.Alert.AlarmFilter.Reset', defaultMessage: '重置' })}
            </Button>
            {type === 'event' &&
              (isExpand ? (
                <Button onClick={() => setIsExpand(false)} type="link">
                  {formatMessage({ id: 'OBShell.Alert.AlarmFilter.Stow', defaultMessage: '收起' })}

                  <UpOutlined />
                </Button>
              ) : (
                <Button onClick={() => setIsExpand(true)} type="link">
                  {formatMessage({
                    id: 'OBShell.Alert.AlarmFilter.Expand',
                    defaultMessage: '展开',
                  })}

                  <DownOutlined />
                </Button>
              ))}
          </Space>
        </Col>
      </Row>
    </Form>
  );
}
