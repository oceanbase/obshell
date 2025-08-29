import { formatMessage } from '@/util/intl';
import { createOrUpdateSilencer, getSilencer, listRules } from '@/service/obshell/alarm';
import AlertDrawer from '@/component/AlertDrawer';
import IconTip from '@/component/IconTip';
import InputLabelComp from '@/component/InputLabelComp';
import { VALIDATE_DEBOUNCE, LABEL_NAME_RULE, OBJECT_OPTIONS_ALARM } from '@/constant/alert';
import { DATE_TIME_FORMAT } from '@/constant/datetime';
import { Alert } from '@/typing/env/alert';
import { useModel } from 'umi';
import { useRequest } from 'ahooks';
import type { DrawerProps } from 'antd';
import {
  Button,
  Col,
  DatePicker,
  Form,
  Input,
  Radio,
  Row,
  Select,
  Space,
  message,
} from '@oceanbase/design';
import dayjs from 'dayjs';
import React, { useEffect } from 'react';
import {
  formatShieldSubmitData,
  getInstancesFromRes,
  getSelectList,
  validateLabelValues,
} from '../helper';
import ClusterSelect from './ClusterSelect';
import ServerSelect from './ServerSelect';
import TenantSelect from './TenantSelect';

interface ShieldDrawerProps extends DrawerProps {
  id?: string;
  initialValues?: Alert.ShieldDrawerInitialValues;
  onClose: () => void;
  submitCallback?: () => void;
}

const { TextArea } = Input;

export default function ShieldDrawerForm({
  id,
  onClose,
  initialValues,
  submitCallback,
  ...props
}: ShieldDrawerProps) {
  const [form] = Form.useForm<Alert.ShieldDrawerForm>();
  const { clusterList, tenantList } = useModel('alarm');
  const isEdit = !!id;

  const newInitialValues = {
    matchers: [
      {
        name: '',
        value: '',
        isRegex: false,
      },
    ],

    instances: {
      type: 'obcluster',
    },
    ...initialValues,
  };
  const { data: listRulesRes } = useRequest(listRules);
  const listRulesData = listRulesRes?.data?.contents?.map((rule: API.RuleResponse) => ({
    label: rule.name,
    value: rule.name,
  }));
  const fieldEndTimeChange = (time: number | Date) => {
    if (typeof time === 'number') {
      form.setFieldValue('ends_at', dayjs(new Date().valueOf() + time));
    } else {
      form.setFieldValue('ends_at', dayjs(time));
    }
    form.validateFields(['ends_at']);
  };

  const submit = (values: Alert.ShieldDrawerForm) => {
    if (!values.matchers) values.matchers = [];
    values.matchers = values.matchers.filter(matcher => matcher.name && matcher.value);
    const _clusterList = getSelectList(clusterList!, values.instances.type, tenantList);
    if (isEdit) values.id = id;
    createOrUpdateSilencer(formatShieldSubmitData(values, _clusterList)).then(({ successful }) => {
      if (successful) {
        message.success(
          formatMessage({
            id: 'OBShell.Alert.Shield.ShieldDrawerForm.OperationSuccessful',
            defaultMessage: '操作成功!',
          })
        );
        submitCallback && submitCallback();
        onClose();
      }
    });
  };

  useEffect(() => {
    if (isEdit) {
      getSilencer({ id }).then(({ successful, data }) => {
        if (successful) {
          form.setFieldsValue({
            comment: data.comment,
            matchers: data.matchers,
            ends_at: dayjs(data.ends_at * 1000),
            instances: getInstancesFromRes(data.instances),
            rules: data.rules,
          });
        }
      });
    }
  }, [id]);

  return (
    <AlertDrawer
      onClose={() => {
        onClose();
      }}
      destroyOnClose={true}
      onSubmit={() => form.submit()}
      title={formatMessage({
        id: 'OBShell.Alert.Shield.ShieldDrawerForm.ShieldingConditions',
        defaultMessage: '屏蔽条件',
      })}
      {...props}
    >
      <Form
        form={form}
        preserve={false}
        onFinish={submit}
        layout="vertical"
        initialValues={newInitialValues}
      >
        <Form.Item
          rules={[
            {
              required: true,
              message: formatMessage({
                id: 'OBShell.Alert.Shield.ShieldDrawerForm.PleaseSelect',
                defaultMessage: '请选择',
              }),
            },
          ]}
          name={['instances', 'type']}
          label={formatMessage({
            id: 'OBShell.Alert.Shield.ShieldDrawerForm.MaskObjectType',
            defaultMessage: '屏蔽对象类型',
          })}
        >
          <Radio.Group
            onChange={() => {
              form.setFieldsValue({
                instances: {
                  obcluster: undefined,
                  obtenant: undefined,
                  observer: undefined,
                },
              });
            }}
          >
            {OBJECT_OPTIONS_ALARM.map(item => (
              <Radio key={item.value} value={item.value}>
                {item.label}
              </Radio>
            ))}
          </Radio.Group>
        </Form.Item>
        <Form.Item noStyle dependencies={[['instances', 'type']]}>
          {({ getFieldValue }) => {
            const type = getFieldValue(['instances', 'type']);
            return (
              <Form.Item
                label={formatMessage({
                  id: 'OBShell.Alert.Shield.ShieldDrawerForm.MaskObject',
                  defaultMessage: '屏蔽对象',
                })}
                name={['instances']}
                required
                rules={[
                  {
                    validator: (_, instances: Alert.InstancesType) => {
                      if (
                        !instances.obcluster ||
                        !instances.obcluster.length ||
                        !instances[instances.type] ||
                        !instances[instances.type]?.length
                      ) {
                        return Promise.reject(
                          new Error(
                            formatMessage({
                              id: 'OBShell.Alert.Shield.ShieldDrawerForm.PleaseSelect',
                              defaultMessage: '请选择',
                            })
                          )
                        );
                      }
                      return Promise.resolve();
                    },
                  },
                ]}
              >
                {type === 'obcluster' && <ClusterSelect key="cluster" />}
                {type === 'obtenant' && <TenantSelect key="tenant" />}
                {type === 'observer' && <ServerSelect key="server" />}
              </Form.Item>
            );
          }}
        </Form.Item>

        <Form.Item
          rules={[
            {
              required: true,
              message: formatMessage({
                id: 'OBShell.Alert.Shield.ShieldDrawerForm.PleaseSelect',
                defaultMessage: '请选择',
              }),
            },
          ]}
          name={'rules'}
          label={formatMessage({
            id: 'OBShell.Alert.Shield.ShieldDrawerForm.MaskingAlarmRules',
            defaultMessage: '屏蔽告警规则',
          })}
        >
          <Select
            mode="multiple"
            allowClear
            style={{ width: '100%' }}
            placeholder={formatMessage({
              id: 'OBShell.Alert.Shield.ShieldDrawerForm.PleaseSelect',
              defaultMessage: '请选择',
            })}
            options={listRulesData}
          />
        </Form.Item>
        <Form.Item
          name={'matchers'}
          validateFirst
          validateDebounce={VALIDATE_DEBOUNCE}
          rules={[
            {
              validator: (_, value: API.Matcher[]) => {
                if (!validateLabelValues(value)) {
                  return Promise.reject(
                    formatMessage({
                      id: 'OBShell.Alert.Shield.ShieldDrawerForm.PleaseCheckWhetherTheLabel',
                      defaultMessage: '请检查标签是否完整输入',
                    })
                  );
                }
                return Promise.resolve();
              },
            },
            LABEL_NAME_RULE,
          ]}
          label={
            <IconTip
              tip={formatMessage({
                id: 'OBShell.Alert.Shield.ShieldDrawerForm.MasksAlarmsBasedOnTag',
                defaultMessage:
                  '按照标签匹配条件屏蔽告警，支持值匹配或者正则表达式，当所有条件都满足时告警才会被屏蔽',
              })}
              content={formatMessage({
                id: 'OBShell.Alert.Shield.ShieldDrawerForm.Label',
                defaultMessage: '标签',
              })}
            />
          }
        >
          <InputLabelComp regex={true} maxLength={8} defaulLabelName="name" />
        </Form.Item>
        <Row style={{ alignItems: 'center' }}>
          <Col>
            <Form.Item
              rules={[
                {
                  required: true,
                  message: formatMessage({
                    id: 'OBShell.Alert.Shield.ShieldDrawerForm.PleaseSelect',
                    defaultMessage: '请选择',
                  }),
                },
              ]}
              name="ends_at"
              label={formatMessage({
                id: 'OBShell.Alert.Shield.ShieldDrawerForm.MaskEndTime',
                defaultMessage: '屏蔽结束时间',
              })}
            >
              <DatePicker showTime format={DATE_TIME_FORMAT} />
            </Form.Item>
          </Col>
          <Col>
            <Space>
              <Button type="link" onClick={() => fieldEndTimeChange(6 * 3600 * 1000)}>
                {formatMessage({
                  id: 'OBShell.Alert.Shield.ShieldDrawerForm.Hours',
                  defaultMessage: '6小时',
                })}
              </Button>
              <Button type="link" onClick={() => fieldEndTimeChange(12 * 3600 * 1000)}>
                {formatMessage({
                  id: 'OBShell.Alert.Shield.ShieldDrawerForm.Hours.1',
                  defaultMessage: '12小时',
                })}
              </Button>
              <Button type="link" onClick={() => fieldEndTimeChange(24 * 3600 * 1000)}>
                {formatMessage({
                  id: 'OBShell.Alert.Shield.ShieldDrawerForm.Day',
                  defaultMessage: '1天',
                })}
              </Button>
              <Button
                onClick={() => fieldEndTimeChange(new Date('2099-12-31 23:59:59'))}
                type="link"
              >
                {formatMessage({
                  id: 'OBShell.Alert.Shield.ShieldDrawerForm.Permanent',
                  defaultMessage: '永久',
                })}
              </Button>
            </Space>
          </Col>
        </Row>
        <Form.Item
          rules={[
            {
              required: true,
              message: formatMessage({
                id: 'OBShell.Alert.Shield.ShieldDrawerForm.PleaseEnter',
                defaultMessage: '请输入',
              }),
            },
          ]}
          name={'comment'}
          label={formatMessage({
            id: 'OBShell.Alert.Shield.ShieldDrawerForm.NoteInformation',
            defaultMessage: '备注信息',
          })}
        >
          <TextArea
            rows={4}
            placeholder={formatMessage({
              id: 'OBShell.Alert.Shield.ShieldDrawerForm.PleaseEnter',
              defaultMessage: '请输入',
            })}
          />
        </Form.Item>
      </Form>
    </AlertDrawer>
  );
}
