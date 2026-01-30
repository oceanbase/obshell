import { formatMessage } from '@/util/intl';
import { EditOutlined, SearchOutlined } from '@oceanbase/icons';
import { useDeepCompareEffect } from 'ahooks';
import { Button, Checkbox, Col, Divider, Form, Input, Row, Space, theme, Tooltip } from 'antd';
import type { CheckboxChangeEvent } from 'antd/es/checkbox';
import { uniq } from 'lodash';
import React, { useEffect, useState } from 'react';
import MyDrawer from '../MyDrawer';

interface EditMetricsProps {
  groupMetrics: { name?: string; description?: string }[];
  initialKeys: string[];
  onOk?: (selectedMetrics: string[]) => void;
}

const EditMetrics: React.FC<EditMetricsProps> = ({ groupMetrics, initialKeys, onOk }) => {
  const { token } = theme.useToken();
  const [form] = Form.useForm();
  const { setFieldsValue } = form;

  const [editMetricsVisible, setEditMetricsVisible] = useState(false);
  const [search, setSearch] = useState(''); // 筛选框的输入内容

  const groupMetricsName = groupMetrics.map(item => item.name!);
  const [selectedMetricKeys, setSelectedMetricKeys] = useState<string[]>(initialKeys);

  useDeepCompareEffect(() => {
    setSelectedMetricKeys(initialKeys);
  }, [initialKeys]);

  // 筛选后的数据
  const filteredData = groupMetrics.filter(item => item.name!.includes(search));

  const isAllChecked =
    filteredData.length > 0 && filteredData.every(item => selectedMetricKeys.includes(item.name!));
  const isIndeterminate =
    !isAllChecked && filteredData.some(item => selectedMetricKeys.includes(item.name!));

  useEffect(() => {
    setFieldsValue({
      selectedMetrics: selectedMetricKeys,
    });
  }, [selectedMetricKeys, setFieldsValue]);

  // 全选/反选
  const onCheckAllChange = (checkAll: boolean) => {
    // 筛选后的指标名称
    const filteredDataNames: string[] = filteredData.map(item => item.name!);

    // 将已勾选的指标和筛选后的指标名称合并去重
    const uniqSelectedMetrics = uniq([...selectedMetricKeys, ...filteredDataNames]);

    let newSelectedList: string[] = [];
    if (checkAll) {
      newSelectedList = uniqSelectedMetrics;
    } else {
      newSelectedList = uniqSelectedMetrics.filter(item => !filteredDataNames.includes(item));
    }
    setSelectedMetricKeys(newSelectedList);
  };

  // 清除全部
  const onClearAll = () => {
    onCheckAllChange(false);
  };

  // 保存
  const handleSave = () => {
    const newCheckedList: string[] = [...selectedMetricKeys];
    // newCheckedList按指标组顺序排序
    newCheckedList.sort((a, b) => {
      const indexA = groupMetricsName.indexOf(a);
      const indexB = groupMetricsName.indexOf(b);
      return indexA - indexB;
    });
    onOk?.(newCheckedList);
    setEditMetricsVisible(false);
  };

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setSearch(e.target.value);
  };

  const handleClose = () => {
    setSelectedMetricKeys(initialKeys);
    setEditMetricsVisible(false);
  };

  return (
    <>
      <Tooltip
        title={formatMessage({
          id: 'OBShell.component.MonitorComp.EditMetrics.EditMonitoringMetrics',
          defaultMessage: '编辑监控指标',
        })}
      >
        <EditOutlined
          onClick={() => {
            setEditMetricsVisible(true);
          }}
          className="pointable"
          style={{ color: token.colorTextSecondary, cursor: 'pointer' }}
        />
      </Tooltip>
      <MyDrawer
        title={formatMessage({
          id: 'OBShell.component.MonitorComp.EditMetrics.EditMonitoringMetrics',
          defaultMessage: '编辑监控指标',
        })}
        width={350}
        open={editMetricsVisible}
        onClose={handleClose}
        footer={
          <Space>
            <Button type="primary" onClick={handleSave}>
              {formatMessage({
                id: 'OBShell.component.MonitorComp.EditMetrics.Save',
                defaultMessage: '保存',
              })}
            </Button>
            <Button onClick={handleClose}>
              {formatMessage({
                id: 'OBShell.component.MonitorComp.EditMetrics.Cancel',
                defaultMessage: '取消',
              })}
            </Button>
          </Space>
        }
        maskClosable={false}
      >
        <Input
          prefix={<SearchOutlined />}
          placeholder={formatMessage({
            id: 'OBShell.component.MonitorComp.EditMetrics.PleaseEnter',
            defaultMessage: '请输入',
          })}
          value={search}
          onChange={handleInputChange}
          style={{ marginBottom: 24 }}
        />

        <Form form={form} preserve={false} initialValues={{ selectedMetrics: initialKeys }}>
          <div
            style={{
              display: 'flex',
              justifyContent: 'space-between',
              alignItems: 'center',
            }}
          >
            <Checkbox
              checked={isAllChecked}
              indeterminate={isIndeterminate}
              onChange={(e: CheckboxChangeEvent) => onCheckAllChange(e.target.checked)}
            >
              {formatMessage({
                id: 'OBShell.component.MonitorComp.EditMetrics.SelectAll',
                defaultMessage: '全选',
              })}
            </Checkbox>
            <Button type="link" onClick={onClearAll}>
              {formatMessage({
                id: 'OBShell.component.MonitorComp.EditMetrics.ClearAll',
                defaultMessage: '清除全部',
              })}
            </Button>
          </div>
          <Divider style={{ margin: '6px 0 12px 0' }} />
          <Form.Item name="selectedMetrics">
            <Checkbox.Group>
              <Row gutter={[0, 12]}>
                {filteredData.map(({ name, description }) => {
                  return (
                    <Col span={24} key={name}>
                      <Checkbox
                        value={name}
                        onChange={(e: CheckboxChangeEvent) => {
                          if (e.target.checked) {
                            setSelectedMetricKeys([...selectedMetricKeys, name!]);
                          } else {
                            setSelectedMetricKeys(selectedMetricKeys.filter(item => item !== name));
                          }
                        }}
                      >
                        <div>{name}</div>
                        {description && (
                          <div style={{ color: token.colorTextTertiary, fontSize: 12 }}>
                            {description}
                          </div>
                        )}
                      </Checkbox>
                    </Col>
                  );
                })}
              </Row>
            </Checkbox.Group>
          </Form.Item>
        </Form>
      </MyDrawer>
    </>
  );
};

export default EditMetrics;
