import { formatMessage } from '@/util/intl';
import React, { useState } from 'react';
import {
  Button,
  Checkbox,
  CheckboxChangeEvent,
  Col,
  Divider,
  Dropdown,
  Form,
  Row,
  Space,
} from '@oceanbase/design';
import { FilterOutlined } from '@oceanbase/icons';
import { ContentWithQuestion } from '@oceanbase/ui';
import styles from './index.less';

interface SelectMonitorMetrics {
  metricItems: API.ComOceanbaseOcpMonitorModelMetricGroup[];
  defaultSelectedKeys: string[];
  initialKeys: string[];
  onOk: (values: any) => void;
}

export default (props: SelectMonitorMetrics) => {
  const { initialKeys, metricItems, defaultSelectedKeys, onOk } = props;
  const [open, setOpen] = useState(false);
  const [form] = Form.useForm();
  const { setFieldsValue, getFieldValue, validateFields } = form;

  function handleOk() {
    validateFields().then(values => {
      onOk?.(values);
      setOpen(false);
    });
  }
  // 全选按钮的状态逻辑
  const onCheckAllChange = (e: CheckboxChangeEvent) => {
    setFieldsValue({
      selectedMetricGroupKeys: e.target.checked ? defaultSelectedKeys : [],
    });
  };

  // 清除按钮的逻辑
  const onClearAll = () => {
    setFieldsValue({
      selectedMetricGroupKeys: [],
    });
  };

  const dropdownContent = (
    <Form form={form} initialValues={{ selectedMetricGroupKeys: initialKeys }}>
      <Form.Item
        noStyle
        shouldUpdate={(prevValues, curValues) =>
          prevValues.selectedMetricGroupKeys !== curValues.selectedMetricGroupKeys
        }
      >
        {() => {
          const selectedItems = getFieldValue('selectedMetricGroupKeys') || [];

          const isAllChecked = selectedItems.length === defaultSelectedKeys.length;
          const isIndeterminate =
            selectedItems.length > 0 && selectedItems.length < defaultSelectedKeys.length;

          return (
            <div className={styles.dropdownContent}>
              {/* 头部 */}
              <div className={styles.dropdownHeader}>
                <Form.Item noStyle>
                  <Checkbox
                    indeterminate={isIndeterminate}
                    checked={isAllChecked}
                    onChange={onCheckAllChange}
                  >
                    {formatMessage({
                      id: 'OBShell.component.SelectMonitorMetrics.SelectAll',
                      defaultMessage: '全选',
                    })}
                  </Checkbox>
                </Form.Item>
                <Button style={{ paddingRight: 0 }} type="link" onClick={onClearAll}>
                  {formatMessage({
                    id: 'OBShell.component.SelectMonitorMetrics.ClearAll',
                    defaultMessage: '清除全部',
                  })}
                </Button>
              </div>
              <Divider style={{ margin: '6px 0 10px' }} />

              {/* 多列复选框 */}
              <div className={styles.metricGroup}>
                <Form.Item name="selectedMetricGroupKeys" noStyle>
                  <Checkbox.Group>
                    <Row gutter={[8, 16]}>
                      {metricItems?.map(({ key, name, metrics, description }) => (
                        <Col key={key} span={8}>
                          <Checkbox value={key}>
                            <ContentWithQuestion
                              content={name}
                              tooltip={{
                                placement: 'right',
                                title: metrics?.[0]?.unit
                                  ? formatMessage(
                                      {
                                        id: 'OBShell.component.SelectMonitorMetrics.DescriptionUnitMetricsunit',
                                        defaultMessage: '{description} （单位: {metricsUnit}）',
                                      },
                                      { description: description, metricsUnit: metrics[0].unit }
                                    )
                                  : description,
                              }}
                            />
                          </Checkbox>
                        </Col>
                      ))}
                    </Row>
                  </Checkbox.Group>
                </Form.Item>
              </div>
              {/* 底部按钮 */}
              <div className={styles.dropdownBottom}>
                <Space>
                  <Button onClick={() => handleVisibleChange(false)}>
                    {formatMessage({
                      id: 'OBShell.component.SelectMonitorMetrics.Cancel',
                      defaultMessage: '取消',
                    })}
                  </Button>
                  <Button type="primary" onClick={handleOk}>
                    {formatMessage({
                      id: 'OBShell.component.SelectMonitorMetrics.Application',
                      defaultMessage: '应用',
                    })}
                  </Button>
                </Space>
              </div>
            </div>
          );
        }}
      </Form.Item>
    </Form>
  );

  const handleVisibleChange = (visible: boolean) => {
    setOpen(visible);
    if (!visible) {
      setFieldsValue({ selectedMetricGroupKeys: initialKeys });
    }
  };

  return (
    <Dropdown
      overlay={dropdownContent}
      trigger={['hover']}
      visible={open}
      onVisibleChange={handleVisibleChange}
    >
      <Button icon={<FilterOutlined />} className={styles.selectButton}>
        {formatMessage({
          id: 'OBShell.component.SelectMonitorMetrics.SelectedChart',
          defaultMessage: '已选图表',
        })}

        <span className={styles.num}>{initialKeys.length}</span>
      </Button>
    </Dropdown>
  );
};
