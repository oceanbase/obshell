import { formatMessage } from '@/util/intl';
import { DeleteOutlined, PlusOutlined } from '@oceanbase/icons';
import { Button, Checkbox, Col, Input, Popconfirm, Row, Space } from '@oceanbase/design';
import React from 'react';

type Label = {
  [T: string]: string | boolean;
};

interface InputLabelCompPorps {
  value?: Label[];
  onChange?: (value: Label[]) => void;
  onBlur?: () => void;
  maxLength?: number;
  defaulLabelName?: string;
  regex?: boolean;
  disable?: boolean;
  allowDelete?: boolean;
}

export default function InputLabelComp(props: InputLabelCompPorps) {
  const {
    value: labels = [],
    onChange,
    maxLength,
    onBlur,
    defaulLabelName = 'key',
    regex,
    disable = false,
    allowDelete = true,
  } = props;

  const labelNameInput = (value: string, index: number) => {
    labels[index][defaulLabelName] = value;
    onChange?.([...labels]);
  };
  const labelValueInput = (value: string, index: number) => {
    labels[index].value = value;
    onChange?.([...labels]);
  };
  const regChange = (checked: boolean, index: number) => {
    labels[index].is_regex = checked;
    onChange?.([...labels]);
  };
  const add = () => {
    const temp: Label = {
      [defaulLabelName]: '',
      value: '',
    };
    if (regex) temp.is_regex = false;
    onChange?.([...labels, temp]);
  };
  const remove = (index: number) => {
    labels.splice(index, 1);
    onChange?.([...labels]);
  };
  return (
    <div>
      <Space style={{ width: '100%', marginBottom: 12 }} direction="vertical">
        {labels?.map((label, index) => (
          <Row gutter={[12, 12]} style={{ alignItems: 'center' }} key={index}>
            <Col span={regex ? 11 : 12}>
              <Input
                disabled={disable}
                value={label[defaulLabelName] as string}
                onBlur={onBlur}
                onChange={e => labelNameInput(e.target.value, index)}
                placeholder={formatMessage({
                  id: 'OBShell.component.InputLabelComp.PleaseEnterATagName',
                  defaultMessage: '请输入标签名',
                })}
              />
            </Col>
            <Col span={regex ? 10 : 11}>
              <Input
                disabled={disable}
                value={label.value as string}
                onBlur={onBlur}
                onChange={e => labelValueInput(e.target.value, index)}
                placeholder={formatMessage({
                  id: 'OBShell.component.InputLabelComp.PleaseEnterATagValue',
                  defaultMessage: '请输入标签值',
                })}
              />
            </Col>
            {regex && (
              <Col span={2}>
                <Checkbox
                  checked={label.is_regex as boolean}
                  onChange={e => regChange(e.target.checked, index)}
                />

                <span style={{ marginLeft: 8 }}>
                  {formatMessage({
                    id: 'OBShell.component.InputLabelComp.Regular',
                    defaultMessage: '正则',
                  })}
                </span>
              </Col>
            )}

            {labels.length > 1 && allowDelete && (
              <Popconfirm
                placement="left"
                title={formatMessage({
                  id: 'OBShell.component.InputLabelComp.AreYouSureYouWant',
                  defaultMessage: '确定要删除该配置项吗？',
                })}
                onConfirm={() => {
                  remove(index);
                }}
                okText={formatMessage({
                  id: 'OBShell.component.InputLabelComp.Delete',
                  defaultMessage: '删除',
                })}
                cancelText={formatMessage({
                  id: 'OBShell.component.InputLabelComp.Cancel',
                  defaultMessage: '取消',
                })}
                okButtonProps={{
                  danger: true,
                  ghost: true,
                }}
              >
                <DeleteOutlined style={{ color: 'rgba(0, 0, 0, .45)' }} />
              </Popconfirm>
            )}
          </Row>
        ))}
      </Space>

      {(!maxLength || labels.length < maxLength) && !disable ? (
        <Row>
          <Col span={24}>
            <Button type="dashed" block onClick={add} style={{ color: 'rgba(0,0,0,0.65)' }}>
              <PlusOutlined />
              {formatMessage({
                id: 'OBShell.component.InputLabelComp.Add',
                defaultMessage: '添加',
              })}
            </Button>
          </Col>
        </Row>
      ) : null}
    </div>
  );
}
