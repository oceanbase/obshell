import AlertDrawer from '@/component/AlertDrawer';
import IconTip from '@/component/IconTip';
import InputLabelComp from '@/component/InputLabelComp';
import InputTimeComp from '@/component/InputTimeComp';
import {
  LABELNAME_REG,
  LABEL_NAME_RULE,
  LEVER_OPTIONS_ALARM,
  SEVERITY_MAP,
  VALIDATE_DEBOUNCE,
} from '@/constant/alert';
import { listRules } from '@/service/obshell/alarm';
import { formatMessage } from '@/util/intl';
import type { DrawerProps } from '@oceanbase/design';
import { Col, Form, Input, Radio, Row, Select, Tag, message } from '@oceanbase/design';
import { useRequest } from 'ahooks';
import React, { useEffect } from 'react';
import { validateLabelValues } from '../helper';

type AlertRuleDrawerProps = {
  ruleName?: string;
  onClose: () => void;
  submitCallback?: () => void;
} & DrawerProps;
const { TextArea } = Input;
export default function RuleDrawerForm({
  ruleName,
  submitCallback,
  onClose,
  ...props
}: AlertRuleDrawerProps) {
  const [form] = Form.useForm();
  const { data: rulesRes } = useRequest(listRules);
  const rules = rulesRes?.data;
  const isEdit = !!ruleName;
  const initialValues = {
    labels: [
      {
        key: '',
        value: '',
      },
    ],

    instanceType: 'obcluster',
  };
  const submit = (values: RuleRule) => {
    if (!values.labels) values.labels = [];
    values.labels = values.labels.filter(label => label.key && label.value);
    alert.createOrUpdateRule(values).then(({ successful }) => {
      if (successful) {
        message.success(
          formatMessage({
            id: 'OBShell.Alert.Rules.RuleDrawerForm.OperationSuccessful',
            defaultMessage: '操作成功！',
          })
        );
        onClose();
        submitCallback && submitCallback();
      }
    });
  };

  useEffect(() => {
    if (ruleName) {
      alert.getRule(ruleName).then(({ data, successful }) => {
        if (successful) {
          form.setFieldsValue({ ...data });
        }
      });
    }
  }, [ruleName]);

  return (
    <AlertDrawer
      destroyOnClose={true}
      onSubmit={() => form.submit()}
      title={formatMessage({
        id: 'OBShell.Alert.Rules.RuleDrawerForm.AlarmRuleConfiguration',
        defaultMessage: '告警规则配置',
      })}
      onClose={onClose}
      {...props}
    >
      <Form
        initialValues={initialValues}
        preserve={false}
        style={{ marginBottom: 64 }}
        layout="vertical"
        onFinish={submit}
        form={form}
      >
        <Row gutter={[24, 0]}>
          <Col span={24}>
            <Form.Item
              rules={[
                {
                  required: true,
                  message: formatMessage({
                    id: 'OBShell.Alert.Rules.RuleDrawerForm.PleaseSelect',
                    defaultMessage: '请选择',
                  }),
                },
              ]}
              name={'instanceType'}
              label={formatMessage({
                id: 'OBShell.Alert.Rules.RuleDrawerForm.ObjectType',
                defaultMessage: '对象类型',
              })}
            >
              <Radio.Group>
                <Radio value="obcluster">
                  {formatMessage({
                    id: 'OBShell.Alert.Rules.RuleDrawerForm.Cluster',
                    defaultMessage: '集群',
                  })}
                </Radio>
                <Radio value="obtenant">
                  {formatMessage({
                    id: 'OBShell.Alert.Rules.RuleDrawerForm.Tenant',
                    defaultMessage: '租户',
                  })}
                </Radio>
                <Radio value="observer"> OBServer </Radio>
              </Radio.Group>
            </Form.Item>
          </Col>

          <Col span={16}>
            <Form.Item
              name={'name'}
              validateFirst
              rules={
                !isEdit
                  ? [
                      {
                        required: true,
                        message: formatMessage({
                          id: 'OBShell.Alert.Rules.RuleDrawerForm.PleaseEnter',
                          defaultMessage: '请输入',
                        }),
                      },
                      {
                        validator: (_, value) => {
                          if (!LABELNAME_REG.test(value)) {
                            return Promise.reject(
                              formatMessage({
                                id: 'OBShell.Alert.Rules.RuleDrawerForm.TheAlarmRuleNameMust',
                                defaultMessage:
                                  '告警规则名需满足：以字母或下划线开头，包含字母，数字，下划线',
                              })
                            );
                          }
                          return Promise.resolve();
                        },
                      },
                      {
                        validator: async (_, value) => {
                          if (rules) {
                            for (const rule of rules) {
                              if (rule.name === value) {
                                return Promise.reject(
                                  new Error(
                                    formatMessage({
                                      id: 'OBShell.Alert.Rules.RuleDrawerForm.AlarmRuleAlreadyExistsPlease',
                                      defaultMessage: '告警规则已存在，请重新输入',
                                    })
                                  )
                                );
                              }
                            }
                          }
                          return Promise.resolve();
                        },
                      },
                    ]
                  : []
              }
              label={formatMessage({
                id: 'OBShell.Alert.Rules.RuleDrawerForm.AlarmRuleName',
                defaultMessage: '告警规则名',
              })}
            >
              <Input
                disabled={isEdit}
                placeholder={formatMessage({
                  id: 'OBShell.Alert.Rules.RuleDrawerForm.PleaseEnter',
                  defaultMessage: '请输入',
                })}
              />
            </Form.Item>
          </Col>
          <Col span={7}>
            <Form.Item
              rules={[
                {
                  required: true,
                  message: formatMessage({
                    id: 'OBShell.Alert.Rules.RuleDrawerForm.PleaseEnter',
                    defaultMessage: '请输入',
                  }),
                },
              ]}
              name={'severity'}
              label={formatMessage({
                id: 'OBShell.Alert.Rules.RuleDrawerForm.AlarmLevel',
                defaultMessage: '告警级别',
              })}
            >
              <Select
                options={LEVER_OPTIONS_ALARM?.map(item => ({
                  value: item.value,
                  label: <Tag color={SEVERITY_MAP[item?.value]?.color}>{item.label}</Tag>,
                }))}
                placeholder={formatMessage({
                  id: 'OBShell.Alert.Rules.RuleDrawerForm.PleaseSelect',
                  defaultMessage: '请选择',
                })}
              />
            </Form.Item>
          </Col>
          <Col span={16}>
            <Form.Item
              name={'query'}
              rules={[
                {
                  required: true,
                  message: formatMessage({
                    id: 'OBShell.Alert.Rules.RuleDrawerForm.PleaseEnter',
                    defaultMessage: '请输入',
                  }),
                },
              ]}
              label={
                <IconTip
                  tip={formatMessage({
                    id: 'OBShell.Alert.Rules.RuleDrawerForm.PromqlExpressionForDecisionAlarm',
                    defaultMessage: '判定告警的 PromQL 表达式',
                  })}
                  content={formatMessage({
                    id: 'OBShell.Alert.Rules.RuleDrawerForm.IndicatorCalculationExpression',
                    defaultMessage: '指标计算表达式',
                  })}
                />
              }
            >
              <Input
                placeholder={formatMessage({
                  id: 'OBShell.Alert.Rules.RuleDrawerForm.PleaseEnter',
                  defaultMessage: '请输入',
                })}
              />
            </Form.Item>
          </Col>
          <Col span={7}>
            <Form.Item
              rules={[
                {
                  required: true,
                  message: formatMessage({
                    id: 'OBShell.Alert.Rules.RuleDrawerForm.PleaseEnter',
                    defaultMessage: '请输入',
                  }),
                },
              ]}
              label={formatMessage({
                id: 'OBShell.Alert.Rules.RuleDrawerForm.Duration',
                defaultMessage: '持续时间',
              })}
              name={'duration'}
            >
              <InputTimeComp />
            </Form.Item>
          </Col>
          <Col span={24}>
            <Form.Item
              name={'summary'}
              rules={[
                {
                  required: true,
                  message: formatMessage({
                    id: 'OBShell.Alert.Rules.RuleDrawerForm.PleaseEnter',
                    defaultMessage: '请输入',
                  }),
                },
              ]}
              label={
                <IconTip
                  tip={formatMessage({
                    id: 'OBShell.Alert.Rules.RuleDrawerForm.TheSummaryInformationTemplateOf',
                    defaultMessage: '告警事件的摘要信息模版，可以使用 {{ }} 来标记需要替换的值',
                  })}
                  content={formatMessage({
                    id: 'OBShell.Alert.Rules.RuleDrawerForm.SummaryInformation',
                    defaultMessage: 'summary 信息',
                  })}
                />
              }
            >
              <TextArea
                rows={4}
                placeholder={formatMessage({
                  id: 'OBShell.Alert.Rules.RuleDrawerForm.PleaseEnter',
                  defaultMessage: '请输入',
                })}
              />
            </Form.Item>
          </Col>
          <Col span={24}>
            <Form.Item
              name={'description'}
              rules={[
                {
                  required: true,
                  message: formatMessage({
                    id: 'OBShell.Alert.Rules.RuleDrawerForm.PleaseEnter',
                    defaultMessage: '请输入',
                  }),
                },
              ]}
              label={
                <IconTip
                  tip={formatMessage({
                    id: 'OBShell.Alert.Rules.RuleDrawerForm.TheDetailsTemplateOfThe',
                    defaultMessage: '告警事件的详情信息模版，可以使用 {{ }} 来标记需要替换的值',
                  })}
                  content={formatMessage({
                    id: 'OBShell.Alert.Rules.RuleDrawerForm.AlarmDetails',
                    defaultMessage: '告警详情信息',
                  })}
                />
              }
            >
              <TextArea
                rows={4}
                placeholder={formatMessage({
                  id: 'OBShell.Alert.Rules.RuleDrawerForm.PleaseEnter',
                  defaultMessage: '请输入',
                })}
              />
            </Form.Item>
          </Col>
          <Col span={24}>
            <Form.Item
              label={
                <IconTip
                  tip={formatMessage({
                    id: 'OBShell.Alert.Rules.RuleDrawerForm.TagsAddedToAlarmEvents',
                    defaultMessage: '添加到告警事件的标签',
                  })}
                  content={formatMessage({
                    id: 'OBShell.Alert.Rules.RuleDrawerForm.Label',
                    defaultMessage: '标签',
                  })}
                />
              }
              validateFirst
              validateDebounce={VALIDATE_DEBOUNCE}
              rules={[
                {
                  validator: (_, value: CommonKVPair[]) => {
                    if (!validateLabelValues(value)) {
                      return Promise.reject(
                        formatMessage({
                          id: 'OBShell.Alert.Rules.RuleDrawerForm.PleaseCheckWhetherTheLabel',
                          defaultMessage: '请检查标签是否完整输入',
                        })
                      );
                    }
                    return Promise.resolve();
                  },
                },
                LABEL_NAME_RULE,
              ]}
              name="labels"
            >
              <InputLabelComp />
            </Form.Item>
          </Col>
        </Row>
      </Form>
    </AlertDrawer>
  );
}
