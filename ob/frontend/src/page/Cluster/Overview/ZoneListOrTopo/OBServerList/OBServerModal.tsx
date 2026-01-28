import { obStart, obStop } from '@/service/obshell/ob';
import { formatMessage } from '@/util/intl';
import { taskSuccess } from '@/util/task';
import {
  Alert,
  Button,
  Checkbox,
  Descriptions,
  Form,
  Modal,
  Space,
  theme,
} from '@oceanbase/design';
import type { ModalProps } from '@oceanbase/design/es/modal';
import { ContentWithIcon } from '@oceanbase/ui';
import { useRequest } from 'ahooks';
import React from 'react';
import { useModel } from 'umi';
import type { Operation } from '../ZoneList';

export interface OBServerModalProps extends ModalProps {
  operation: Operation;
  server: API.Observer & { zoneName?: string };
  onCancel: () => void;
  onSuccess: () => void;
  failedTasksInfo: { content: string; forcePassDag: { id: any[] } } | null;
}

const OBServerModal: React.FC<OBServerModalProps> = ({
  server,
  operation,
  onCancel,
  onSuccess,
  failedTasksInfo,
  ...restProps
}) => {
  const [form] = Form.useForm();
  const { token } = theme.useToken();
  const { setPendingOperation } = useModel('topologyInfo');

  const { run: startOBServer, loading: startLoading } = useRequest(obStart, {
    manual: true,
    onSuccess: (res, params) => {
      if (res.successful && res.data?.id) {
        const serverKey = server.ip && server.obshell_port ? `${server.ip}:${server.obshell_port}` : undefined;
        if (serverKey) {
          setPendingOperation(serverKey, res.data.id);
        }
        taskSuccess({
          taskId: res.data.id,
          message: formatMessage({
            id: 'OBShell.ZoneListOrTopo.OBServerList.OBServerModal.StartupObserverTaskSubmittedSuccessfully',
            defaultMessage: '启动 OBServer 的任务提交成功',
          }),
        });
        onSuccess();
      }
    },
  });
  const { run: stopOBServer, loading: stopLoading } = useRequest(obStop, {
    manual: true,
    onSuccess: (res, parameters) => {
      if (res.successful && res.data?.id) {
        const serverKey = server.ip && server.obshell_port ? `${server.ip}:${server.obshell_port}` : undefined;
        if (serverKey) {
          setPendingOperation(serverKey, res.data.id);
        }
        const isStopService =
          parameters?.[0]?.force === false && parameters?.[0]?.terminate === false;
        taskSuccess({
          taskId: res.data.id,
          message: `停止 OBServer ${isStopService ? '服务' : '进程'}的任务提交成功`,
        });
        onSuccess();
      }
    },
  });

  const handleSubmit = () => {
    const baseParams = {
      scope: {
        type: 'SERVER' as const,
        target: [`${server.ip}:${server.obshell_port}`],
      },
      forcePassDag: failedTasksInfo?.forcePassDag || { id: [] },
    };

    const operationHandlers: Record<string, () => void> = {
      start: () => {
        startOBServer(baseParams);
      },
      stopProcess: () => {
        const freezeServer = form.getFieldValue('freezeServer');
        stopOBServer({
          ...baseParams,
          ...(freezeServer ? { terminate: true } : { force: true }),
        });
      },
      stopService: () => {
        stopOBServer({
          ...baseParams,
          force: false,
          terminate: false,
        });
      },
    };

    const handler = operationHandlers[operation.value];
    if (handler) {
      handler();
    }
  };

  const okButtonProps = operation.isDanger
    ? {
      danger: true,
      ghost: true,
    }
    : {};

  return (
    <>
      <Modal
        title={operation.modalTitle || `${operation.label} OBServer`}
        destroyOnClose={true}
        onCancel={() => {
          onCancel();
        }}
        okButtonProps={okButtonProps}
        {...restProps}
        footer={
          <Space>
            <Button
              onClick={() => {
                onCancel();
              }}
            >
              {formatMessage({
                id: 'OBShell.ZoneListOrTopo.OBServerList.OBServerModal.Cancel',
                defaultMessage: '取消',
              })}
            </Button>
            <Button
              {...okButtonProps}
              loading={startLoading || stopLoading}
              type="primary"
              onClick={() => {
                handleSubmit();
              }}
            >
              {operation.label}
            </Button>
          </Space>
        }
      >
        {operation.value === 'stopService' && (
          <Alert
            type="warning"
            message={formatMessage({
              id: 'OBShell.ZoneListOrTopo.OBServerList.OBServerModal.StoppingTheObserverServiceWill',
              defaultMessage:
                '停止 OBServer 服务会将所有流量和主副本的角色切到其他 OBServer，请谨慎操作',
            })}
            showIcon={true}
            style={{ marginBottom: 16 }}
          />
        )}

        {operation.value === 'stopProcess' && (
          <Alert
            type="warning"
            message={formatMessage({
              id: 'OBShell.ZoneListOrTopo.OBServerList.OBServerModal.StoppingTheObserverProcessWill',
              defaultMessage: '停止 OBServer 进程会导致 OBServer 不可用，无法提供服务，请谨慎操作',
            })}
            showIcon={true}
            style={{ marginBottom: 16 }}
          />
        )}

        <Descriptions column={1} style={{ marginBottom: 24 }} className="description-title-margin">
          <Descriptions.Item
            label={formatMessage({
              id: 'OBShell.ZoneListOrTopo.OBServerList.OBServerModal.HostIp',
              defaultMessage: '主机 IP',
            })}
          >
            {server.ip}
          </Descriptions.Item>
          <Descriptions.Item
            label={formatMessage({
              id: 'OBShell.ZoneListOrTopo.OBServerList.OBServerModal.RpcPort',
              defaultMessage: 'RPC 端口',
            })}
          >
            {server.svr_port}
          </Descriptions.Item>
          <Descriptions.Item
            label={formatMessage({
              id: 'OBShell.ZoneListOrTopo.OBServerList.OBServerModal.Zone',
              defaultMessage: '所属 Zone',
            })}
          >
            {server.zoneName}
          </Descriptions.Item>
          <Descriptions.Item
            label={formatMessage({
              id: 'OBShell.ZoneListOrTopo.OBServerList.OBServerModal.HardwareArchitecture',
              defaultMessage: '硬件架构',
            })}
          >
            {server.architecture}
          </Descriptions.Item>
        </Descriptions>

        <Form
          form={form}
          layout="vertical"
          preserve={false}
          hideRequiredMark={true}
          className="form-without-padding"
        >
          {operation.value === 'stopProcess' && (
            <Form.Item
              layout="horizontal"
              label={
                <ContentWithIcon
                  iconType="question"
                  content={formatMessage({
                    id: 'OBShell.ZoneListOrTopo.OBServerList.OBServerModal.PerformADumpOperationBefore',
                    defaultMessage: '停止进程前执行转储操作',
                  })}
                  tooltip={{
                    title: formatMessage({
                      id: 'OBShell.ZoneListOrTopo.OBServerList.OBServerModal.PerformingThisActionWillIncrease',
                      defaultMessage:
                        '执行本动作会延长停止进程的响应时间，但可以显著缩短 OBServer 恢复时间。',
                    }),
                  }}
                />
              }
              initialValue={true}
              valuePropName="checked"
              name="freezeServer"
            >
              <Checkbox />
            </Form.Item>
          )}
          <div style={{ color: token.colorTextTertiary }}>{failedTasksInfo?.content || ''}</div>
        </Form>
      </Modal>
    </>
  );
};

export default OBServerModal;
