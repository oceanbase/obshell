import { getRanges, VALID_NAME_RULE } from '@/constant';
import { DATE_TIME_FORMAT_DISPLAY, RFC3339_DATE_TIME_FORMAT } from '@/constant/datetime';
import { formatMessage } from '@/util/intl';
import * as ObAshController from '@/service/obshell/tenant';
import { getTextLengthRule } from '@/util';
import { taskSuccess } from '@/util/task';
import { useDispatch } from 'umi';
import React from 'react';
import moment from 'moment';
import { DatePicker, Form, Input, Modal } from '@oceanbase/design';
import type { ModalProps } from '@oceanbase/design/es/modal';
import { useRequest } from 'ahooks';
const { RangePicker } = DatePicker;

export interface GenerateReportModalProps extends ModalProps {
  onSuccess: () => void;
  tenantData: API.TenantInfo;
}

const GenerateReportModal: React.FC<GenerateReportModalProps> = ({
  onSuccess,
  tenantData,
  ...restProps
}) => {
  const dispatch = useDispatch();
  const [form] = Form.useForm();
  const { setFieldsValue, validateFields } = form;

  // 生成报告
  const { run: generateAshReport, loading: generateAshReportLoading } = useRequest(
    ObAshController.generateAshReport,
    {
      manual: true,
      onSuccess: res => {
        if (res.successful) {
          taskSuccess({
            taskId: res?.data?.taskInstanceId,
            message: formatMessage({
              id: 'ocp-v2.Session.Report.GenerateReportModal.TheTaskForGeneratingThe',
              defaultMessage: '生成活跃会话历史报告的任务提交成功',
            }),
          });

          dispatch({
            type: 'task/update',
            payload: {
              runningTaskListDataRefreshDep: res?.data?.taskInstanceId,
            },
          });

          if (onSuccess) {
            onSuccess();
          }
        }
      },
    }
  );

  const handleSubmit = () => {
    validateFields().then(values => {
      const { name, range } = values;
      generateAshReport({
        id: tenantData.clusterId,
        tenantId: tenantData.id,
        name,
        startTime: range && range[0] && range[0].format(RFC3339_DATE_TIME_FORMAT),
        endTime: range && range[1] && range[1].format(RFC3339_DATE_TIME_FORMAT),
      });
    });
  };

  return (
    <Modal
      title={formatMessage({
        id: 'ocp-v2.Session.Report.GenerateReportModal.GenerateActiveSessionHistoryReports',
        defaultMessage: '生成活跃会话历史报告',
      })}
      destroyOnClose={true}
      confirmLoading={generateAshReportLoading}
      onOk={handleSubmit}
      {...restProps}
    >
      <Form layout="vertical" form={form} preserve={false} hideRequiredMark={true}>
        <div style={{ marginBottom: 24 }}>
          {formatMessage(
            {
              id: 'ocp-v2.Session.Report.GenerateReportModal.TenantTenantdataname',
              defaultMessage: '租户：{tenantDataName}',
            },
            { tenantDataName: tenantData.name }
          )}
        </div>
        <Form.Item
          name="range"
          label={formatMessage({
            id: 'ocp-v2.Session.Report.GenerateReportModal.AnalysisTime',
            defaultMessage: '分析时间',
          })}
          extra={
            <div>
              <div>
                {formatMessage({
                  id: 'ocp-v2.Session.Report.GenerateReportModal.YouCanSelectOnlyThe',
                  defaultMessage: '只支持选择近 8 天的时间',
                })}
              </div>
              <div>
                {formatMessage({
                  id: 'ocp-v2.Session.Report.GenerateReportModal.TheTimeSpanBetweenThe',
                  defaultMessage: '起始分析时间与结束分析时间的时间跨度仅支持 60 分钟以内',
                })}
              </div>
            </div>
          }
          rules={[
            {
              required: true,
              message: formatMessage({
                id: 'ocp-v2.Session.Report.GenerateReportModal.SelectAnAnalysisTime',
                defaultMessage: '请选择分析时间',
              }),
            },
          ]}
        >
          <RangePicker
            showTime={true}
            format={DATE_TIME_FORMAT_DISPLAY}
            ranges={getRanges()}
            disabledDate={current => {
              // 只支持选择近 8 天的时间
              return current && (current < moment().subtract(8, 'day') || current > moment());
            }}
            onChange={value => {
              if (value && value.length > 0) {
                const startTime = value && value[0] && value[0].format('YYYYMMDDHHmmss');
                const endTime = value && value[1] && value[1].format('YYYYMMDDHHmmss');
                setFieldsValue({
                  name: `OBASH_${tenantData.clusterName}_${tenantData.obClusterId}_${tenantData.name}_${startTime}_${endTime}`,
                });
              } else {
                setFieldsValue({
                  name: undefined,
                });
              }
            }}
          />
        </Form.Item>
        <Form.Item
          name="name"
          label={formatMessage({
            id: 'ocp-v2.Report.Component.GenerateReportModal.ReportName',
            defaultMessage: '报告名',
          })}
          rules={[
            {
              required: true,
              message: formatMessage({
                id: 'ocp-v2.Session.Report.GenerateReportModal.EnterAReportName',
                defaultMessage: '请输入报告名',
              }),
            },

            // 长度限制
            getTextLengthRule(0, 128),
            // 字符限制
            VALID_NAME_RULE,
          ]}
        >
          <Input
            placeholder={formatMessage({
              id: 'ocp-v2.Session.Report.GenerateReportModal.EnterAReportName',
              defaultMessage: '请输入报告名',
            })}
          />
        </Form.Item>
      </Form>
    </Modal>
  );
};

export default GenerateReportModal;
