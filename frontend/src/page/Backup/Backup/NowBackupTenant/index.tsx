import MyCard from '@/component/MyCard';
import type { MyDrawerProps } from '@/component/MyDrawer';
import { formatMessage } from '@/util/intl';
import * as ObTenantBackupController from '@/service/obshell/backup';
import { obclusterInfo } from '@/service/obshell/obcluster';
import { taskSuccess } from '@/util/task';
import { breadcrumbItemRender } from '@/util/component';
import { DATA_BACKUP_MODE_LIST } from '@/constant/backup';
import { history, useSelector } from 'umi';
import React, { useEffect, useState } from 'react';
import type { Route } from '@oceanbase/design/es/breadcrumb/Breadcrumb';
import {
  Button,
  Col,
  Card,
  Popconfirm,
  Descriptions,
  Row,
  Form,
  Radio,
  Modal,
  message,
} from '@oceanbase/design';
import { PageContainer } from '@oceanbase/ui';
import { useRequest } from 'ahooks';
import StorageConfigCard from '../../Component/StorageConfigCard';

export interface BackupNowDrawerProps extends MyDrawerProps {
  match: {
    params: {
      tenantName?: string;
    };
  };
}

const NowBackupTenant: React.FC<BackupNowDrawerProps> = ({
  match: {
    params: { tenantName },
  },
}) => {
  // 是否测试过
  const [checkedStatus, setCheckedStatus] = useState<string>('unchecked');
  // 目录是否被修改
  const [isDirectoryModified, setIsDirectoryModified] = useState<boolean>(false);
  // 日志存储目录是否被修改
  const [isArchiveLogModified, setIsArchiveLogModified] = useState<boolean>(false);

  const { tenantData } = useSelector((state: DefaultRootState) => state.tenant);
  const { data: clusterInfo } = useRequest(obclusterInfo, {
    defaultParams: [{}],
  });

  const clusterData = clusterInfo?.data || {};

  const [form] = Form.useForm();
  const { validateFields } = form;

  const { data: tenantBackupConfigData, run } = useRequest(
    ObTenantBackupController.getTenantBackupConfig,
    {
      manual: true,
      defaultParams: [
        {
          name: tenantName,
        },
      ],

      onSuccess: res => {
        if (res?.successful) {
          const { archive_base_uri, data_base_uri } = res?.data;
          form.setFieldsValue({
            data_backup_uri: data_base_uri.split('://')[1],
            archive_log_uri: archive_base_uri.split('://')[1],
          });
          setCheckedStatus('checked');
        }
      },
    }
  );

  useEffect(() => {
    run({
      name: tenantName,
    });
  }, []);

  const tenantBackupConfigInfo = tenantBackupConfigData?.data || {};

  // 备份配置
  const { runAsync: tenantBackupConfig, loading: backupNowLoading } = useRequest(
    ObTenantBackupController.tenantBackupConfig,
    {
      manual: true,
    }
  );

  // 立即备份
  const { runAsync: tenantStartBackup, loading: tenantStartBackupLoading } = useRequest(
    ObTenantBackupController.tenantStartBackup,
    {
      manual: true,
    }
  );

  const handleCheck = () => {
    validateFields().then(values => {
      const { archive_log_uri, data_backup_uri } = values;

      // 如果日志存储目录被修改，显示确认弹框
      if (isArchiveLogModified) {
        Modal.confirm({
          title: formatMessage({
            id: 'OBShell.Backup.NowBackupTenant.LogStoreDirectoryChangeConfirmation',
            defaultMessage: '日志存储目录变更确认',
          }),
          content: formatMessage({
            id: 'OBShell.Backup.NowBackupTenant.LogArchivingIsAutomaticallyStopped',
            defaultMessage: '变更日志存储目录时，系统会自动停止日志归档。确定要继续吗？',
          }),
          okText: formatMessage({
            id: 'OBShell.Backup.NowBackupTenant.Determine',
            defaultMessage: '确定',
          }),
          cancelText: formatMessage({
            id: 'OBShell.Backup.NowBackupTenant.Cancel',
            defaultMessage: '取消',
          }),
          onOk: () => {
            // 用户确认后执行测试配置
            executeBackupConfig(archive_log_uri, data_backup_uri);
          },
        });
      } else {
        // 直接执行测试配置
        executeBackupConfig(archive_log_uri, data_backup_uri);
      }
    });
  };

  const executeBackupConfig = (archive_log_uri: string, data_backup_uri: string) => {
    tenantBackupConfig(
      {
        name: tenantData?.tenant_name,
      },
      {
        archive_base_uri: 'file://' + archive_log_uri,
        data_base_uri: 'file://' + data_backup_uri,
      }
    ).then(res => {
      if (res.successful) {
        const taskId = res.data?.id;
        taskSuccess({
          taskId,
          message: formatMessage({
            id: 'OBShell.Backup.NowBackupTenant.TestedAndConfiguredTasksAre',
            defaultMessage: '测试并配置的任务提交成功',
          }),
        });
        // 测试成功后重置修改状态
        setIsDirectoryModified(false);
        setIsArchiveLogModified(false);
        // run({
        //   name: tenantName,
        // });
      }
    });
  };

  const handleSubmit = () => {
    if (checkedStatus === 'unchecked') {
      return message.error(
        formatMessage({
          id: 'OBShell.Backup.NowBackupTenant.TestTheStorageDirectoryFirst',
          defaultMessage: '请先测试存储目录',
        })
      );
    }
    validateFields().then(values => {
      const { mode } = values;

      tenantStartBackup(
        {
          name: tenantData?.tenant_name,
        },

        {
          mode,
        }
      ).then(res => {
        if (res.successful) {
          const taskId = res.data?.id;
          taskSuccess({
            taskId,
            message: formatMessage({
              id: 'OBShell.Backup.NowBackupTenant.TheTaskOfImmediateBackup',
              defaultMessage: '立即备份的任务提交成功',
            }),
          });
        }
      });
    });
  };

  const routes: Route[] = [
    {
      path: `/tenant/${tenantName}/backup`,
      breadcrumbName: formatMessage({
        id: 'OBShell.Backup.NowBackupTenant.BackupRecovery',
        defaultMessage: '备份恢复',
      }),
    },

    {
      breadcrumbName: formatMessage({
        id: 'OBShell.Backup.NowBackupTenant.BackupNow',
        defaultMessage: '立即备份',
      }),
    },
  ];

  return (
    <PageContainer
      ghost={true}
      header={{
        title: formatMessage({
          id: 'OBShell.Backup.NowBackupTenant.BackupNow',
          defaultMessage: '立即备份',
        }),
        breadcrumb: { routes, itemRender: breadcrumbItemRender },
        onBack: () => history.goBack(),
      }}
      footer={[
        <Popconfirm
          placement="topRight"
          title={formatMessage({
            id: 'OBShell.Backup.NowBackupTenant.AreYouSureToBack',
            defaultMessage: '确定立即备份吗？',
          })}
          onConfirm={() => {
            handleSubmit();
          }}
          okText={formatMessage({
            id: 'ocp-v2.component.Compaction.DumpPolicyDrawer.Ok',
            defaultMessage: '确定',
          })}
          cancelText={formatMessage({
            id: 'ocp-v2.component.Compaction.DumpPolicyDrawer.Cancel',
            defaultMessage: '取消',
          })}
        >
          <Button loading={tenantStartBackupLoading} type="primary" disabled={isDirectoryModified}>
            {formatMessage({
              id: 'ocp-v2.Backup.Component.BackupNowDrawer.BackUpNow',
              defaultMessage: '立即备份',
            })}
          </Button>
        </Popconfirm>,
        <Button
          key="check"
          loading={backupNowLoading}
          disabled={!isDirectoryModified}
          onClick={() => {
            handleCheck();
          }}
        >
          {formatMessage({
            id: 'OBShell.Backup.NowBackupTenant.TestAndConfigure',
            defaultMessage: '测试并配置',
          })}
        </Button>,
        <Button
          key="Cancel"
          onClick={() => {
            history.goBack();
          }}
        >
          {formatMessage({
            id: 'ocp-v2.Backup.Component.BackupNowDrawer.Cancel',
            defaultMessage: '取消',
          })}
        </Button>,
      ]}
    >
      <Row gutter={[0, 16]}>
        <Col span={24}>
          <MyCard
            title={formatMessage({
              id: 'OBShell.Backup.NowBackupTenant.BackupObject',
              defaultMessage: '备份对象',
            })}
          >
            <Descriptions layout="vertical">
              <Descriptions.Item
                label={formatMessage({
                  id: 'OBShell.Backup.NowBackupTenant.BackupTenant',
                  defaultMessage: '备份租户',
                })}
              >
                {`${clusterData?.cluster_name}/${tenantData?.tenant_name}`}
              </Descriptions.Item>
              <Descriptions.Item
                label={formatMessage({
                  id: 'OBShell.Backup.NowBackupTenant.OceanbaseVersion',
                  defaultMessage: 'OceanBase 版本',
                })}
              >
                {clusterData?.ob_version}
              </Descriptions.Item>
              <Descriptions.Item
                label={formatMessage({
                  id: 'OBShell.Backup.NowBackupTenant.BackupMethod',
                  defaultMessage: '备份方式',
                })}
              >
                {formatMessage({
                  id: 'OBShell.Backup.NowBackupTenant.PhysicalBackup',
                  defaultMessage: '物理备份',
                })}
              </Descriptions.Item>
            </Descriptions>
          </MyCard>
        </Col>
        <Col span={24}>
          <Form
            form={form}
            layout="vertical"
            colon={false}
            hideRequiredMark={true}
            onValuesChange={(changedValues, allValues) => {
              setCheckedStatus('unchecked');
            }}
          >
            <StorageConfigCard
              form={form}
              obVersion={clusterData?.ob_version}
              backupConfig={tenantBackupConfigInfo}
              onDirectoryChange={modified => {
                setIsDirectoryModified(modified);
                // 同时检查日志存储目录是否被修改
                if (modified) {
                  const allValues = form.getFieldsValue();
                  const currentArchiveUri = allValues.archive_log_uri || '';
                  const originalArchiveUri =
                    tenantBackupConfigInfo?.archive_base_uri?.split('://')[1] || '';
                  const isArchiveLogModified =
                    currentArchiveUri && currentArchiveUri !== originalArchiveUri;
                  setIsArchiveLogModified(isArchiveLogModified);
                } else {
                  setIsArchiveLogModified(false);
                }
              }}
            />
          </Form>
        </Col>
        <Col span={24}>
          <Card
            title={formatMessage({
              id: 'ocp-v2.Backup.Component.BackupNowDrawer.DataBackupMethod',
              defaultMessage: '数据备份方式',
            })}
          >
            <Form form={form} layout="vertical" colon={false} hideRequiredMark={true}>
              <Form.Item
                label={formatMessage({
                  id: 'ocp-v2.Backup.Component.BackupNowDrawer.DataBackupMethod',
                  defaultMessage: '数据备份方式',
                })}
                style={{
                  marginBottom: 0,
                }}
                extra={
                  <>
                    <div>
                      {formatMessage({
                        id: 'ocp-v2.Backup.NowBackupTenant.ForPhysicalBackupAFull',
                        defaultMessage: '物理备份下，增量备份之前必须有一次全量的数据备份',
                      })}
                    </div>
                  </>
                }
                name="mode"
                initialValue={'full'}
                rules={[
                  {
                    required: true,
                    message: formatMessage({
                      id: 'ocp-v2.Backup.Component.BackupNowDrawer.SelectADataBackupMethod',
                      defaultMessage: '请选择数据备份方式',
                    }),
                  },
                ]}
              >
                <Radio.Group>
                  {DATA_BACKUP_MODE_LIST.map(item => (
                    <Radio key={item.value} value={item.value}>
                      {item.fullLabel}
                    </Radio>
                  ))}
                </Radio.Group>
              </Form.Item>
            </Form>
          </Card>
        </Col>
      </Row>
    </PageContainer>
  );
};

export default NowBackupTenant;
