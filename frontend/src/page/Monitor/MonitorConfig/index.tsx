import { formatMessage } from '@/util/intl';
import React, { useState } from 'react';
import { Form, Input, Button, message, Typography, Alert } from '@oceanbase/design';
import { useRequest } from 'ahooks';
import { putSystemExternalPrometheus } from '@/service/obshell/system';
import ConfigIcon from '@/asset/config.svg';
import styles from './index.less';
import { history } from 'umi';

const { Title, Paragraph } = Typography;

interface MonitorConfigProps {
  targetPath: string;
  onConfigSuccess?: () => void;
}

const MonitorConfig: React.FC<MonitorConfigProps> = ({ targetPath, onConfigSuccess }) => {
  const [form] = Form.useForm();
  const [errorMessage, setErrorMessage] = useState('');

  // 保存Prometheus配置
  const { run: savePrometheus, loading: savingPrometheus } = useRequest(
    putSystemExternalPrometheus,
    {
      manual: true,
      onSuccess: (res: API.OcsAgentResponse) => {
        if (res.successful && res.status === 200) {
          message.success(
            formatMessage({
              id: 'OBShell.Monitor.MonitorConfig.PrometheusConfigurationSuccessful',
              defaultMessage: 'Prometheus配置成功',
            })
          );
          // 通知父组件配置成功，让父组件更新状态
          onConfigSuccess?.();
          // 跳转到目标页面
          history.push(targetPath);
        } else {
          setErrorMessage(
            res.error?.message ||
              formatMessage({
                id: 'OBShell.Monitor.MonitorConfig.PrometheusConfigurationFailed',
                defaultMessage: 'Prometheus配置失败',
              })
          );
        }
      },
      onError: error => {
        message.error(
          formatMessage(
            {
              id: 'OBShell.Monitor.MonitorConfig.ConfigurationFailedErrormessage',
              defaultMessage: '配置失败: {errorMessage}',
            },
            { errorMessage: error.message }
          )
        );
      },
    }
  );

  // 处理表单提交
  const handleSubmit = async (values: any) => {
    const config: API.PrometheusConfig = {
      address: values.address,
      auth:
        values.username || values.password
          ? {
              username: values.username,
              password: values.password,
            }
          : undefined,
    };
    await savePrometheus(config);
  };

  return (
    <div className={styles.container}>
      <div className={styles.content}>
        {/* 图标放在表单左上角 */}
        <div className={styles.configIcon}>
          <img src={ConfigIcon} alt="config" />
        </div>
        {/* 标题和内容区域 */}
        <div style={{ marginBottom: 24 }}>
          <Title level={2} style={{ margin: 0, fontSize: 20, fontWeight: 600 }}>
            {formatMessage({
              id: 'OBShell.Monitor.MonitorConfig.WelcomeToPerformanceMonitoring',
              defaultMessage: '欢迎使用性能监控',
            })}
          </Title>

          <Paragraph className={styles.description}>
            {formatMessage({
              id: 'OBShell.Monitor.MonitorConfig.YouCanUseAnExisting',
              defaultMessage: '您可以使用已有的 Prometheus 地址开启性能监控；',
            })}

            <br />
            {formatMessage({
              id: 'OBShell.Monitor.MonitorConfig.IfThereIsNoAddress',
              defaultMessage: '若暂无地址，建议使用 OBD 安装部署。',
            })}
          </Paragraph>
        </div>

        {errorMessage && (
          <Alert
            closable
            message={errorMessage}
            onClose={() => setErrorMessage('')}
            type="error"
            style={{ marginBottom: 16, width: 400 }}
          />
        )}

        {/* 配置表单 */}
        <div style={{ width: 400, position: 'relative' }}>
          <Form
            form={form}
            hideRequiredMark={true}
            layout="vertical"
            onFinish={handleSubmit}
            size="large"
          >
            <Form.Item
              label={formatMessage({
                id: 'OBShell.Monitor.MonitorConfig.PrometheusAddress',
                defaultMessage: 'Prometheus 地址',
              })}
              name="address"
              rules={[
                {
                  required: true,
                  message: formatMessage({
                    id: 'OBShell.Monitor.MonitorConfig.PleaseEnterPrometheusAddress',
                    defaultMessage: '请输入Prometheus地址',
                  }),
                },
              ]}
            >
              <Input
                placeholder={formatMessage({
                  id: 'OBShell.Monitor.MonitorConfig.PleaseEnter',
                  defaultMessage: '请输入',
                })}
              />
            </Form.Item>

            <Form.Item
              label={formatMessage({
                id: 'OBShell.Monitor.MonitorConfig.UserName',
                defaultMessage: '用户名',
              })}
              name="username"
            >
              <Input
                placeholder={formatMessage({
                  id: 'OBShell.Monitor.MonitorConfig.PleaseEnter',
                  defaultMessage: '请输入',
                })}
              />
            </Form.Item>

            <Form.Item
              label={formatMessage({
                id: 'OBShell.Monitor.MonitorConfig.Password',
                defaultMessage: '密码',
              })}
              name="password"
            >
              <Input.Password
                placeholder={formatMessage({
                  id: 'OBShell.Monitor.MonitorConfig.PleaseEnterPassword',
                  defaultMessage: '请输入密码',
                })}
              />
            </Form.Item>

            <Form.Item style={{ marginTop: 32 }}>
              <Button type="primary" htmlType="submit" loading={savingPrometheus}>
                {formatMessage({
                  id: 'OBShell.Monitor.MonitorConfig.Open',
                  defaultMessage: '开启',
                })}
              </Button>
            </Form.Item>
          </Form>
        </div>
      </div>
    </div>
  );
};

export default MonitorConfig;
