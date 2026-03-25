import { saveSharedStorageKey, validateSharedStorageKey } from '@/service/obshell/obcluster';
import { formatMessage } from '@/util/intl';
import {
  Alert,
  Button,
  Col,
  Descriptions,
  DescriptionsProps,
  Form,
  Input,
  message,
  Modal,
  Row,
  Space,
  theme,
  Tooltip,
} from '@oceanbase/design';
import { isNullValue } from '@oceanbase/util';
import { useModel } from '@umijs/max';
import { useRequest } from 'ahooks';
import React, { useState } from 'react';

export interface RotatingKeyModalProps {
  visible: boolean;
  setSharedStorageVisible: (visible: boolean) => void;
  onCancel: () => void;
}
const RotatingKeyModal: React.FC<RotatingKeyModalProps> = ({
  visible,
  setSharedStorageVisible,
  onCancel,
}) => {
  const { clusterData, refreshClusterData } = useModel('cluster');
  const { token } = theme.useToken();
  const [form] = Form.useForm();
  const access_domain = clusterData?.shared_storage_info?.access_domain || '';

  const [validateFlag, setValidateFlag] = useState<boolean | undefined>(undefined);

  const rotatingKeyItemList: DescriptionsProps['items'] = [
    {
      label: formatMessage({
        id: 'OBShell.Overview.BaseInfo.RotatingKeyModal.Cluster',
        defaultMessage: '所属集群',
      }),
      children: clusterData?.cluster_name,
    },
    {
      label: formatMessage({
        id: 'OBShell.Overview.BaseInfo.RotatingKeyModal.AccessDomainName',
        defaultMessage: '访问域名',
      }),
      children: access_domain,
    },
    {
      label: formatMessage({
        id: 'OBShell.Overview.BaseInfo.RotatingKeyModal.Region',
        defaultMessage: '地域',
      }),
      children: clusterData?.shared_storage_info?.region,
    },
  ];

  const { runAsync: runValidateSharedStorageKey, loading: validateSharedStorageKeyLoading } =
    useRequest(validateSharedStorageKey, {
      manual: true,
      onSuccess: res => {
        setValidateFlag(res.successful);
      },
    });

  const { runAsync: runSaveSharedStorageKey, loading: saveSharedStorageKeyLoading } = useRequest(
    saveSharedStorageKey,
    {
      manual: true,
      onSuccess: res => {
        if (res.successful) {
          message.success(
            formatMessage({
              id: 'OBShell.Overview.BaseInfo.RotatingKeyModal.KeyRotationSucceeded',
              defaultMessage: '密钥轮转成功',
            })
          );
          handleCancel();
          setSharedStorageVisible(false);
          refreshClusterData();
        }
      },
    }
  );

  const handleValidate = () => {
    form.validateFields().then(values => {
      const { accessKey, secretKey } = values;
      runValidateSharedStorageKey(
        {
          access_key: accessKey,
          secret_key: secretKey,
          path: clusterData?.shared_storage_info?.storage_path || '',
          endpoint: access_domain,
        },
        {
          HIDE_ERROR_MESSAGE: true,
        }
      );
    });
  };

  const handleCancel = () => {
    setValidateFlag(undefined);
    form.resetFields();
    onCancel();
  };

  const handleSave = () => {
    form.validateFields().then(values => {
      const { accessKey, secretKey } = values;
      runSaveSharedStorageKey({
        access_key: accessKey,
        secret_key: secretKey,
        endpoint: access_domain ? `host=${access_domain}` : '',
        path: clusterData?.shared_storage_info?.storage_path || '',
      });
    });
  };

  return (
    <Modal
      visible={visible}
      onCancel={handleCancel}
      title={formatMessage({
        id: 'OBShell.Overview.BaseInfo.RotatingKeyModal.RotaryKey',
        defaultMessage: '轮转密钥',
      })}
      footer={
        <Space>
          <Button onClick={handleCancel}>
            {formatMessage({
              id: 'OBShell.Overview.BaseInfo.RotatingKeyModal.Cancel',
              defaultMessage: '取消',
            })}
          </Button>
          <Button onClick={handleValidate} loading={validateSharedStorageKeyLoading}>
            {formatMessage({
              id: 'OBShell.Overview.BaseInfo.RotatingKeyModal.Verify',
              defaultMessage: '验证',
            })}
          </Button>
          <Tooltip
            title={
              !validateFlag
                ? formatMessage({
                    id: 'OBShell.Overview.BaseInfo.RotatingKeyModal.PleaseVerifyConnectivityFirst',
                    defaultMessage: '请先验证联通性',
                  })
                : ''
            }
          >
            <Button
              type="primary"
              loading={saveSharedStorageKeyLoading}
              disabled={!validateFlag}
              onClick={handleSave}
            >
              {formatMessage({
                id: 'OBShell.Overview.BaseInfo.RotatingKeyModal.Save',
                defaultMessage: '保存',
              })}
            </Button>
          </Tooltip>
        </Space>
      }
    >
      <Descriptions
        column={1}
        size="small"
        items={rotatingKeyItemList}
        styles={{
          root: {
            backgroundColor: token.colorFillQuaternary,
            padding: 16,
            borderRadius: token.borderRadius,
            marginBottom: 24,
          },
        }}
      />

      <Form layout="vertical" form={form}>
        <Row gutter={16}>
          <Col span={12}>
            <Form.Item
              label="AK"
              name="accessKey"
              rules={[
                {
                  required: true,
                  message: formatMessage({
                    id: 'OBShell.Overview.BaseInfo.RotatingKeyModal.PleaseEnterAk',
                    defaultMessage: '请输入 AK',
                  }),
                },
              ]}
            >
              <Input />
            </Form.Item>
          </Col>
          <Col span={12}>
            <Form.Item
              label="SK"
              name="secretKey"
              rules={[
                {
                  required: true,
                  message: formatMessage({
                    id: 'OBShell.Overview.BaseInfo.RotatingKeyModal.PleaseEnterSk',
                    defaultMessage: '请输入 SK',
                  }),
                },
              ]}
            >
              <Input.Password autoComplete="new-password" />
            </Form.Item>
          </Col>
        </Row>
      </Form>
      {!isNullValue(validateFlag) && (
        <Alert
          showIcon
          message={
            validateFlag
              ? formatMessage({
                  id: 'OBShell.Overview.BaseInfo.RotatingKeyModal.ValidationSucceeded',
                  defaultMessage: '验证成功',
                })
              : formatMessage({
                  id: 'OBShell.Overview.BaseInfo.RotatingKeyModal.ValidationFailed',
                  defaultMessage: '验证失败',
                })
          }
          type={validateFlag ? 'success' : 'error'}
        />
      )}
    </Modal>
  );
};

export default RotatingKeyModal;
