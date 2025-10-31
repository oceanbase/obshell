import ClusterLabel from '@/component/ClusterLabel';
import ContentWithIcon from '@/component/ContentWithIcon';
import MySelect from '@/component/MySelect';
import BinlogAssociationMsg from '@/component/BinlogAssociationMsg';
import { ALL, MODAL_FORM_ITEM_LAYOUT } from '@/constant';
// import * as ObClusterController from '@/service/ocp-all-in-one/ObClusterController';
import { handleMultipleSelectChangeWithAll } from '@/util';
import { taskSuccess } from '@/util/task';
import { formatMessage, useDispatch } from 'umi';
import React from 'react';
import { isEqual } from 'lodash';
import { Checkbox, Form, Modal, theme } from '@oceanbase/design';
import type { ModalProps } from '@oceanbase/design/es/modal';
import { ExclamationCircleFilled } from '@oceanbase/icons';
import { useRequest } from 'ahooks';
import { ContentWithIcon as OBUIContentWithIcon } from '@oceanbase/ui';

const { Option } = MySelect;

export interface RestartModalProps extends ModalProps {
  obClusterRelatedBinlogService?: API.TenantBinlogService;
  isStandAloneCluster?: boolean;
  clusterData: API.ClusterInfo;
  onSuccess: () => void;
}

const RestartModal: React.FC<RestartModalProps> = ({
  visible,
  obClusterRelatedBinlogService,
  isStandAloneCluster,
  clusterData,
  onSuccess,
  ...restProps
}) => {
  const { token } = theme.useToken();
  const [form] = Form.useForm();
  const { validateFields } = form;
  const dispatch = useDispatch();

  const ObClusterController = {};
  // 重启集群
  const { runAsync: restartObCluster, loading } = useRequest(ObClusterController.restartObCluster, {
    manual: true,
    onSuccess: res => {
      if (res.successful) {
        const taskId = res.data && res.data.id;
        taskSuccess({
          taskId,
          message: formatMessage({
            id: 'ocp-v2.src.model.cluster.TheClusterRestartTaskWas',
            defaultMessage: '重启集群的任务提交成功',
          }),
        });

        if (onSuccess) {
          onSuccess();
        }
        dispatch({
          type: 'cluster/getClusterData',
          payload: {
            id: clusterData.id,
          },
        });

        dispatch({
          type: 'task/update',
          payload: {
            runningTaskListDataRefreshDep: taskId,
          },
        });
      }
    },
  });

  const handleSubmit = () => {
    validateFields().then(values => {
      const { zoneNames, restartMode, freezeServer } = values;

      const body = {
        restartMode,
        ...(isEqual(zoneNames, [ALL]) ? {} : { zoneNames }),
        freezeServer,
      };

      if (restartMode === 'FORCE') {
        Modal.confirm({
          title: formatMessage({
            id: 'ocp-v2.Detail.Component.RestartModal.AreYouSureYouWant',
            defaultMessage: '确定要强制重启吗？',
          }),
          content: formatMessage({
            id: 'ocp-v2.Detail.Component.RestartModal.ForcedRestartInterruptsTheService',
            defaultMessage: '强制重启会中断业务对数据库的访问，对业务造成不可用的影响',
          }),
          onOk: () => {
            return restartObCluster(
              {
                id: clusterData.id,
              },

              body
            );
          },
        });
      } else {
        restartObCluster(
          {
            id: clusterData.id,
          },

          body
        );
      }
    });
  };

  return (
    <Modal
      title={formatMessage({
        id: 'ocp-v2.Detail.Component.RestartModal.RestartTheCluster',
        defaultMessage: '重启集群',
      })}
      visible={visible}
      destroyOnClose={true}
      confirmLoading={loading}
      onOk={handleSubmit}
      okText={formatMessage({
        id: 'ocp-v2.Detail.Component.RestartModal.Restart',
        defaultMessage: '重启',
      })}
      okButtonProps={{
        ghost: true,
        danger: true,
      }}
      {...restProps}
    >
      <BinlogAssociationMsg
        clusterType="OceanBase"
        type="component"
        binlogService={obClusterRelatedBinlogService}
      />

      <Form
        form={form}
        preserve={false}
        hideRequiredMark={true}
        layout="horizontal"
        className="form-with-small-margin"
      >
        <Form.Item
          label={formatMessage({
            id: 'ocp-v2.Detail.Component.RestartModal.Cluster',
            defaultMessage: '集群',
          })}
          style={{ marginBottom: '8px' }}
        >
          <ClusterLabel cluster={clusterData} />
        </Form.Item>
        <Form.Item
          {...MODAL_FORM_ITEM_LAYOUT}
          label={formatMessage({
            id: 'ocp-v2.Detail.Component.RestartModal.RestartRange',
            defaultMessage: '重启范围',
          })}
          name="zoneNames"
          // 默认选中全部
          initialValue={[ALL]}
          rules={[
            {
              required: true,
              message: formatMessage({
                id: 'ocp-v2.Detail.Component.RestartModal.PleaseSelectTheRestartRange',
                defaultMessage: '请选择重启范围',
              }),
            },
          ]}
        >
          <MySelect
            mode="multiple"
            showSearch={true}
            onChange={(value: string[]) => {
              handleMultipleSelectChangeWithAll(form, 'zoneNames', value);
            }}
          >
            <Option key={ALL} value={ALL}>
              {formatMessage({
                id: 'ocp-v2.Detail.Component.RestartModal.All',
                defaultMessage: '全部',
              })}
            </Option>
            {(clusterData.zones || []).map(item => (
              <Option key={item.name} value={item.name}>
                {item.name}
              </Option>
            ))}
          </MySelect>
        </Form.Item>
        <Form.Item shouldUpdate={true} noStyle={true}>
          {({ getFieldValue }) => (
            <Form.Item
              {...MODAL_FORM_ITEM_LAYOUT}
              label={
                <OBUIContentWithIcon
                  iconType="question"
                  content={formatMessage({
                    id: 'ocp-v2.Detail.Component.RestartModal.RestartMode',
                    defaultMessage: '重启方式',
                  })}
                  tooltip={{
                    placement: 'right',
                    title: (
                      <div>
                        <div>
                          {formatMessage({
                            id: 'ocp-v2.Detail.Component.RestartModal.RestartDoesNotInterruptService',
                            defaultMessage: '重启：不会中断业务对数据库的访问',
                          })}
                        </div>
                        <div>
                          {formatMessage({
                            id: 'ocp-v2.Detail.Component.RestartModal.ForcedRestartInterruptsServiceAccess',
                            defaultMessage:
                              '强制重启：会中断业务对数据库的访问，对业务造成不可用的影响',
                          })}
                        </div>
                      </div>
                    ),
                  }}
                />
              }
              name="restartMode"
              initialValue={isStandAloneCluster ? 'FORCE' : 'ROLLING'}
              rules={[
                {
                  required: true,
                  message: formatMessage({
                    id: 'ocp-v2.Detail.Component.RestartModal.SelectARestartMethod',
                    defaultMessage: '请选择重启方式',
                  }),
                },
              ]}
              extra={
                getFieldValue('restartMode') === 'FORCE' && (
                  <ContentWithIcon
                    prefixIcon={{
                      component: ExclamationCircleFilled,
                      style: {
                        color: token.colorWarning,
                      },
                    }}
                    content={formatMessage({
                      id: 'ocp-v2.Detail.Component.RestartModal.ForcedRestartInterruptsTheService',
                      defaultMessage: '强制重启会中断业务对数据库的访问，对业务造成不可用的影响',
                    })}
                    style={{
                      color: token.colorWarning,
                      // 与 icon 的 marginTop 配合，可以实现 icon 和文本的上对齐
                      alignItems: 'flex-start',
                    }}
                  />
                )
              }
            >
              <MySelect disabled={isStandAloneCluster}>
                {!isStandAloneCluster && (
                  <Option value="ROLLING">
                    {formatMessage({
                      id: 'ocp-v2.Detail.Component.RestartModal.RestartRecommended',
                      defaultMessage: '重启（推荐）',
                    })}
                  </Option>
                )}

                <Option value="FORCE">
                  {formatMessage({
                    id: 'ocp-v2.Detail.Component.RestartModal.ForceRestart',
                    defaultMessage: '强制重启',
                  })}
                </Option>
              </MySelect>
            </Form.Item>
          )}
        </Form.Item>
        <Form.Item
          label={
            <OBUIContentWithIcon
              iconType="question"
              content={formatMessage({
                id: 'ocp-v2.Detail.Component.RestartModal.PerformADumpOperationBeforeStoppingTheProcess',
                defaultMessage: '停止进程前执行转储操作',
              })}
              tooltip={{
                title: formatMessage({
                  id: 'ocp-v2.Detail.Component.RestartModal.PerformingThisActionWillProlongTheResponseTime',
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
      </Form>
    </Modal>
  );
};

export default RestartModal;
