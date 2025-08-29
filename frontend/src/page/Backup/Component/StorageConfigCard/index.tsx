import MyInput from '@/component/MyInput';
import EditText from '@/component/EditText';
import { formatMessage } from '@/util/intl';
import React, { useEffect, useState } from 'react';
import { Alert, Form, Card, Col, Row, Spin, message } from '@oceanbase/design';
import type { FormInstance } from '@oceanbase/design/es/form';
import type { CardProps } from '@oceanbase/design/es/card';

const FormItem = Form.Item;

/**
 * 该组件目前的应用场景有四种，不同的场景下展示形式均有不同，因此需要扩展多个参数来满足需求:
 * 1.安装备份恢复服务
 * 2.新建备份策略
 * 3.立即备份
 * 4.发起恢复
 */
export interface StorageConfigCardProps extends CardProps {
  form: FormInstance;
  displayMode?: 'page' | 'component';
  useType?: 'backup' | 'restore';
  obVersion?: string;
  loading?: boolean;
  className?: string;
  backupConfig?: any;
  checkedStatus?: string;
  style?: React.CSSProperties;
  recoveryCheckResult?: any;
  onDirectoryChange?: (modified: boolean) => void;
}

const StorageConfigCard: React.FC<StorageConfigCardProps> = ({
  form,
  loading,
  useType = 'backup',
  displayMode = 'page',
  obVersion,
  backupConfig,
  checkedStatus,
  recoveryCheckResult,
  onDirectoryChange,
  ...restProps
}) => {
  // 测试 alert 是否展示
  const [alertVisible, setAlertVisible] = useState(checkedStatus === 'passed');
  // 跟踪目录是否被修改
  const [isDirectoryModified, setIsDirectoryModified] = useState(false);

  useEffect(() => {
    setAlertVisible(checkedStatus === 'passed');
  }, [checkedStatus]);

  // 检查目录是否被修改的方法
  const checkDirectoryModified = (changedValue: string, isDataBackup: boolean) => {
    const allValues = form.getFieldsValue();
    const [, originalDataUri = ''] = backupConfig?.data_base_uri?.split('://') || [];
    const [, originalArchiveUri = ''] = backupConfig?.archive_base_uri?.split('://') || [];

    // 检查是否有初始配置
    const hasInitialConfig = !!(backupConfig?.data_base_uri || backupConfig?.archive_base_uri);

    const isModified = hasInitialConfig
      ? (isDataBackup ? changedValue : allValues.data_backup_uri) !== originalDataUri ||
        (isDataBackup ? allValues.archive_log_uri : changedValue) !== originalArchiveUri
      : changedValue || (isDataBackup ? allValues.archive_log_uri : allValues.data_backup_uri); // 没有初始配置时，只要设置了值就算修改

    setIsDirectoryModified(isModified);
    onDirectoryChange?.(isModified);
  };

  return (
    <Card
      bordered={false}
      title={formatMessage({
        id: 'ocp-v2.Component.StorageConfigCard.StorageConfiguration',
        defaultMessage: '存储配置',
      })}
      {...restProps}
    >
      <Row gutter={24}>
        <Col span={useType === 'backup' ? 14 : 12}>
          <FormItem
            label={formatMessage({
              id: 'OBShell.Component.StorageConfigCard.DataStorageDirectory',
              defaultMessage: '数据存储目录',
            })}
            name="data_backup_uri"
            rules={[
              {
                required: true,
                message: formatMessage({
                  id: 'OBShell.Component.StorageConfigCard.PleaseEnterADataStorage',
                  defaultMessage: '请输入数据存储目录',
                }),
              },
            ]}
            initialValue={backupConfig?.data_base_uri?.split('://')[1]}
          >
            {useType === 'backup' && backupConfig?.data_base_uri ? (
              <EditText
                placeholder="obbackup_bucket/obbackup_dest"
                addonBefore="file://"
                address={backupConfig?.data_base_uri?.split('://')[1]}
                // tipTitle="测试并配置成功后才会生效。"
                onOk={value => {
                  form.setFieldsValue({ data_backup_uri: value });
                  checkDirectoryModified(value, true);
                }}
              />
            ) : (
              <MyInput
                placeholder="obbackup_bucket/obbackup_dest"
                addonBefore="file://"
                disabled={loading}
                onChange={e => {
                  form.setFieldsValue({ data_backup_uri: e.target.value });
                  checkDirectoryModified(e.target.value, true);
                }}
              />
            )}
          </FormItem>
        </Col>
      </Row>
      <Row gutter={24}>
        <Col span={useType === 'backup' ? 14 : 12}>
          <FormItem
            label={formatMessage({
              id: 'OBShell.Component.StorageConfigCard.LogStorageDirectory',
              defaultMessage: '日志存储目录',
            })}
            name="archive_log_uri"
            rules={[
              {
                required: true,
                message: formatMessage({
                  id: 'OBShell.Component.StorageConfigCard.PleaseEnterLogStorageDirectory',
                  defaultMessage: '请输入日志存储目录',
                }),
              },
            ]}
            initialValue={backupConfig?.archive_base_uri?.split('://')[1]}
            style={{ marginBottom: 0 }}
          >
            {useType === 'backup' && backupConfig?.archive_base_uri ? (
              <EditText
                placeholder="obbackup_bucket/obachive_dest"
                addonBefore="file://"
                address={backupConfig?.archive_base_uri?.split('://')[1]}
                // tipTitle="测试并配置成功后才会生效。"
                onOk={value => {
                  form.setFieldsValue({ archive_log_uri: value });
                  checkDirectoryModified(value, false);
                }}
              />
            ) : (
              <MyInput
                placeholder="obbackup_bucket/obachive_dest"
                addonBefore="file://"
                disabled={loading}
                onChange={e => {
                  form.setFieldsValue({ archive_log_uri: e.target.value });
                  checkDirectoryModified(e.target.value, false);
                }}
              />
            )}
          </FormItem>
        </Col>
      </Row>
      {alertVisible && (
        <Alert
          type={checkedStatus === 'failed' ? 'error' : 'success'}
          showIcon={true}
          closable={true}
          message={
            (checkedStatus === 'failed' && `测试失败，${recoveryCheckResult?.message}`) ||
            (checkedStatus === 'passed' &&
              formatMessage({
                id: 'OBShell.Component.StorageConfigCard.TestSuccess',
                defaultMessage: '测试成功',
              }))
          }
          onClose={() => {
            setAlertVisible(false);
          }}
          style={{ marginTop: 24 }}
        />
      )}
      {isDirectoryModified && useType === 'backup' && (
        <Alert
          type="warning"
          showIcon={true}
          message={formatMessage({
            id: 'OBShell.Component.StorageConfigCard.ThereAreChangesThatHave',
            defaultMessage: '存在未生效的更改，请点击【测试并配置】按钮进行配置。',
          })}
          style={{ marginTop: 24 }}
        />
      )}
    </Card>
  );
};

export default StorageConfigCard;
