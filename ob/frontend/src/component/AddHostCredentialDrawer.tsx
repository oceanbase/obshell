// AddHostCredentialDrawer（此组件） ：用于凭据管理和主机新增 HOST 凭据
import MyDrawer from '@/component/MyDrawer';
import MyInput from '@/component/MyInput';
import MySelect from '@/component/MySelect';
import { MAX_FORM_ITEM_LAYOUT, SELECT_TOKEN_SPEARATORS } from '@/constant';
import { USET_TYPE_LIST } from '@/constant/compute';
import { formatMessage } from '@/util/intl';
import type { HostCredential } from '@/page/Credential/Host';
import * as obService from '@/service/obshell/ob';
import * as CredentialController from '@/service/obshell/security';
import { getTextLengthRule, isEnglish } from '@/util';
import React, { useEffect, useState } from 'react';
import { find, pick, uniq } from 'lodash';
import { Form, InputNumber, message, Radio, Space, Tag, theme } from '@oceanbase/design';
import { useRequest } from 'ahooks';
import { CREDENTIAL_NAME_RULE } from '@/constant/credential';

const FormItem = Form.Item;
const { Option } = MySelect;

// 主机附带凭据信息
export interface HostAlignCredential {
  credential?: string;
  ip: string;
}

export interface AddHostCredentialDrawerProps {
  visible: boolean;
  onCancel?: () => void;
  onSuccess?: (credentia?: any) => void;
  // 编辑凭据   默认false,新增凭据
  isEdit?: boolean;
  credential?: HostCredential;
}

const AddHostCredentialDrawer: React.FC<AddHostCredentialDrawerProps> = ({
  onCancel = () => {},
  visible = false,
  onSuccess = () => {},
  isEdit = false,
  credential,
}) => {
  const { token } = theme.useToken();

  const [failItems, setFailItems] = useState<API.ValidationDetail[]>([]);
  const [form] = Form.useForm();
  const { resetFields, submit, getFieldValue } = form;
  const [hostAlignCredentialList, setHostAlignCredentialList] = useState<HostAlignCredential[]>([]);

  useEffect(() => {
    return () => {
      resetFields();
      setFailItems([]);
    };
  }, [visible]);

  // 获取 Host 类型的凭据
  const { data: credentialListData, loading: credentialsLoading } = useRequest(
    CredentialController.listCredentials,
    {
      ready: !!visible,
      defaultParams: [
        {
          target_type: 'HOST',
        },
      ],

      refreshDeps: [visible],
    }
  );

  const credentialList: API.Credential[] = credentialListData?.data?.contents || [];

  // 获取主机列表
  const { data: obInfoData, loading: listHostsLoading } = useRequest(obService.getObInfo, {
    ready: !!visible,
    defaultParams: [{}],
    refreshDeps: [visible],
  });
  const agentInfoList: API.AgentInstance[] = obInfoData?.data?.agent_info || [];

  const hostList: string[] = uniq(agentInfoList.map(item => item.ip));

  useEffect(() => {
    if (!credentialsLoading && !listHostsLoading) {
      // 实际上主机或者主机凭据列表可能没有数据
      setHostAlignCredentialList(hostAlignCredential());
    }
  }, [credentialsLoading, listHostsLoading]);

  // 创建主机凭据
  const { run: addCredential, loading: addCredentialLoading } = useRequest(
    CredentialController.createCredential,
    {
      manual: true,
      onSuccess: res => {
        if (res?.successful) {
          message.success('主机凭据创建成功');
          resetFields();
          onSuccess();
        }
      },
    }
  );

  // 更新主机凭据
  const { run: updateCredential, loading: updateCredentialLoading } = useRequest(
    CredentialController.updateCredential,
    {
      manual: true,
      onSuccess: res => {
        if (res?.successful) {
          message.success('主机凭据更新成功');
          resetFields();
          onSuccess();
        }
      },
    }
  );

  // 校验主机凭据
  const { runAsync: validateCredential, loading: validateCredentialLoading } = useRequest(
    CredentialController.validateCredential,
    {
      manual: true,
    }
  );

  const onFinish = (values: any) => {
    const {
      name,
      description = '',
      username = 'root',
      passphrase = '',
      targets = [],
      port = 22,
    } = values;

    const param: API.CreateCredentialParam = {
      target_type: 'HOST',
      name,
      description,
      ssh_credential_property: {
        type: 'PASSWORD',
        username,
        passphrase,
        targets: targets.map((target: string) => `${target}:${port}`),
      },
    };

    const validateParam = pick(param, ['ssh_credential_property', 'target_type']);

    if (isEdit) {
      updateCredential({ id: credential?.credential_id! }, param);
    } else {
      validateCredential(validateParam).then(res => {
        const { data = {} } = res;
        // 如果校验成功
        if (data.failed_count === 0) {
          addCredential(param);
        } else {
          const failedList =
            data?.details?.filter(
              (item: API.ValidationDetail) => item?.connection_result !== 'SUCCESS'
            ) || [];
          setFailItems(failedList);
        }
      });
    }
  };

  /**
   * 多选下拉定制化，1.展示 主机+凭据名，2.并根据校验结果做出颜色变化反馈
   * */
  const tagRender = (props: any) => {
    const { label, closable, onClose } = props;
    let color = 'default';
    const self = find(failItems, ({ target }) => label?.includes(target?.split(':')[0]));

    if (self) {
      // CONNECT_FAILED 显示 red
      color = self?.connection_result === 'CONNECT_FAILED' ? 'red' : 'default';
    }

    return (
      <Tag color={color} closable={closable} onClose={onClose}>
        {label}
      </Tag>
    );
  };

  /**
   * 为主机列表的每台主机附加上，所属主机的凭据信息（不同租户的不同主机有不同凭据）
   * */
  const hostAlignCredential: () => HostAlignCredential[] = () => {
    return hostList?.map(host => {
      const credentialItem = find(credentialList, item =>
        (item.ssh_secret?.targets?.map(target => target?.split(':')[0]) || [])?.includes(host)
      );

      return {
        ip: host,
        credential: credentialItem?.name,
      };
    });
  };

  return (
    <span>
      <MyDrawer
        title={
          isEdit
            ? formatMessage({
                id: 'ocp-v2.src.component.AddHostCredentialDrawer.ModifySecrets',
                defaultMessage: '编辑凭据',
              })
            : formatMessage({
                id: 'ocp-v2.src.component.AddHostCredentialDrawer.CreateSecret',
                defaultMessage: '新建凭据',
              })
        }
        width={520}
        visible={visible}
        destroyOnClose={true}
        onCancel={onCancel}
        onOk={() => {
          submit();
        }}
        confirmLoading={
          validateCredentialLoading || addCredentialLoading || updateCredentialLoading
        }
        maskClosable={false}
      >
        <Form
          form={form}
          layout="vertical"
          onFinish={onFinish}
          hideRequiredMark={true}
          {...MAX_FORM_ITEM_LAYOUT}
        >
          <FormItem
            initialValue={credential?.name}
            label={formatMessage({
              id: 'ocp-v2.src.component.AddHostCredentialDrawer.SecretName',
              defaultMessage: '凭据名',
            })}
            name="name"
            rules={[
              {
                required: true,
                message: formatMessage({
                  id: 'ocp-v2.src.component.AddHostCredentialDrawer.EnterASecretName',
                  defaultMessage: '请输入凭据名',
                }),
              },

              CREDENTIAL_NAME_RULE,
            ]}
          >
            <MyInput />
          </FormItem>

          <FormItem
            initialValue={credential?.username === 'root' ? 'root' : 'normal'}
            label={formatMessage({
              id: 'ocp-v2.src.component.AddHostCredentialDrawer.UserType',
              defaultMessage: '用户类型',
            })}
            name="userType"
            rules={[
              {
                required: true,
                message: formatMessage({
                  id: 'ocp-v2.src.component.AddHostCredentialDrawer.SelectAUserType',
                  defaultMessage: '请选择用户类型',
                }),
              },
            ]}
          >
            <Radio.Group>
              {USET_TYPE_LIST.map(item => (
                <Radio.Button key={item.value} value={item.value}>
                  {item.label}
                </Radio.Button>
              ))}
            </Radio.Group>
          </FormItem>

          <Form.Item
            noStyle
            shouldUpdate={(prevValues, currentValues) =>
              prevValues.userType !== currentValues.userType
            }
          >
            {() => {
              return getFieldValue('userType') === 'normal' ? (
                <FormItem
                  initialValue={credential?.username}
                  label={formatMessage({
                    id: 'ocp-v2.src.component.AddHostCredentialDrawer.UserName',
                    defaultMessage: '用户名',
                  })}
                  name="username"
                  rules={[
                    {
                      required: true,
                      message: formatMessage({
                        id: 'ocp-v2.src.component.AddHostCredentialDrawer.EnterAUsername',
                        defaultMessage: '请输入用户名',
                      }),
                    },
                  ]}
                >
                  <MyInput />
                </FormItem>
              ) : null;
            }}
          </Form.Item>

          <FormItem
            initialValue={credential?.passphrase}
            label={formatMessage({
              id: 'ocp-v2.src.component.AddHostCredentialDrawer.Password',
              defaultMessage: '密码',
            })}
            name="passphrase"
            rules={[
              {
                required: !isEdit,
                message: formatMessage({
                  id: 'ocp-v2.src.component.AddHostCredentialDrawer.EnterAPassword',
                  defaultMessage: '请输入密码',
                }),
              },
            ]}
            // 英文环境下需要加宽，否则 placeholder 无法展示完整
            {...(isEnglish() ? { wrapperCol: { span: 24 } } : {})}
          >
            <MyInput.Password
              autoComplete="new-password"
              placeholder={
                isEdit
                  ? formatMessage({
                      id: 'ocp-v2.src.component.AddHostCredentialDrawer.BlankForNotModifyPassword',
                      defaultMessage: '为空表示不修改密码',
                    })
                  : formatMessage({
                      id: 'ocp-v2.src.component.AddHostCredentialDrawer.EnterAPassword',
                      defaultMessage: '请输入密码',
                    })
              }
            />
          </FormItem>

          <FormItem
            label={formatMessage({
              id: 'ocp-v2.src.component.AddHostCredentialDrawer.ApplicationHost',
              defaultMessage: '应用主机',
            })}
            name="targets"
            rules={[
              {
                required: true,
                message: '请输入应用主机',
              },
            ]}
            initialValue={isEdit ? credential?.targets?.map(target => target?.split(':')[0]) : []}
            help={
              failItems.length > 0 ? (
                <Space
                  size={isEnglish() ? 0 : 'middle'}
                  style={{
                    color: token.colorTextDescription,
                    marginTop: '4px',
                  }}
                  direction={isEnglish() ? 'vertical' : 'horizontal'}
                >
                  {/* Tag padding: 0 重置自身样式,  display: 'flex', alignItems: 'center' 实现垂直居中 */}
                  <div style={{ display: 'flex', alignItems: 'center' }}>
                    <Tag color="red" style={{ width: 12, height: 12, padding: 0 }} />
                    {formatMessage({
                      id: 'ocp-v2.src.component.AddHostCredentialDrawer.TheSecretCannotBeConnected',
                      defaultMessage: '凭据无法连接',
                    })}
                  </div>
                </Space>
              ) : null
            }
          >
            {/* 获取所有可连接主机列表 */}
            <MySelect
              loading={listHostsLoading}
              mode="multiple"
              maxTagCount={5}
              tokenSeparators={SELECT_TOKEN_SPEARATORS}
              tagRender={tagRender}
              allowClear={true}
              optionLabelProp="label"
              className="ip-select"
            >
              {hostAlignCredentialList.map(host => (
                <Option key={host?.ip} value={host?.ip} label={host?.ip}>
                  <Space>
                    <>{host?.ip}</>
                    {host?.credential && <Tag style={{ borderRadius: 10 }}>{host?.credential}</Tag>}
                  </Space>
                </Option>
              ))}
            </MySelect>
          </FormItem>
          <FormItem
            label="ssh 端口"
            name="port"
            initialValue={credential?.port || 22}
            rules={[{ required: true, message: '请输入 ssh 端口' }]}
          >
            <InputNumber style={{ width: '100%' }} />
          </FormItem>

          <FormItem
            initialValue={credential?.description}
            label={
              <>
                {formatMessage({
                  id: 'ocp-v2.src.component.AddHostCredentialDrawer.Description',
                  defaultMessage: '说明',
                })}

                <span style={{ color: token.colorTextTertiary }}>
                  {formatMessage({
                    id: 'ocp-v2.src.component.AddHostCredentialDrawer.Optional',
                    defaultMessage: '（可选）',
                  })}
                </span>
              </>
            }
            name="description"
            rules={[getTextLengthRule(0, 256)]}
          >
            <MyInput.TextArea rows={3} />
          </FormItem>
        </Form>
      </MyDrawer>
    </span>
  );
};

export default AddHostCredentialDrawer;
