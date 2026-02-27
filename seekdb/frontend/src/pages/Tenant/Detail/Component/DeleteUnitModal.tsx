/*
 * Copyright (c) 2024 OceanBase.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

import { MODAL_HORIZONTAL_FORM_ITEM_LAYOUT } from '@/constant';
import { formatMessage } from '@/util/intl';
import { Form } from '@ant-design/compatible';
import '@ant-design/compatible/assets/index.css';
import { Modal } from '@oceanbase/design';
import { connect } from '@umijs/max';
import React from 'react';

const FormItem = Form.Item;

export interface DeleteUnitModalProps {
  dispatch: any;
  tenantData: API.TenantInfo;
  unit: API.Unit;
  onSuccess: () => void;
}

class DeleteUnitModal extends React.PureComponent<DeleteUnitModalProps> {
  public handleSubmit = () => {
    const { dispatch, tenantData, unit, onSuccess } = this.props;
    dispatch({
      type: 'tenant/deleteUnit',
      payload: {
        id: tenantData.clusterId,
        tenantId: tenantData.id,
        unitId: unit.id,
      },

      onSuccess: () => {
        if (onSuccess) {
          onSuccess();
        }
      },
    });
  };

  public render() {
    const { tenantData, unit, ...restProps } = this.props;
    return (
      <Modal
        title={formatMessage({
          id: 'ocp-express.Detail.Component.DeleteUnitModal.AreYouSureYouWant',
          defaultMessage: '确定要删除 Unit 吗？',
        })}
        destroyOnClose={true}
        onOk={this.handleSubmit}
        okText={formatMessage({
          id: 'ocp-express.Detail.Component.DeleteUnitModal.Delete',
          defaultMessage: '删除',
        })}
        okButtonProps={{
          ghost: true,
          danger: true,
        }}
        {...restProps}
      >
        <Form
          className="form-with-small-margin"
          layout="horizontal"
          hideRequiredMark={true}
          {...MODAL_HORIZONTAL_FORM_ITEM_LAYOUT}
        >
          <FormItem
            label={formatMessage({
              id: 'ocp-express.Detail.Component.DeleteUnitModal.Tenant',
              defaultMessage: '所属租户',
            })}
          >
            {tenantData.name}
          </FormItem>
          <FormItem
            label={formatMessage({
              id: 'ocp-express.Detail.Component.DeleteUnitModal.BelongsZone',
              defaultMessage: '所属 Zone',
            })}
          >
            {unit.zoneName}
          </FormItem>
          <FormItem label="Unit ID">{unit.id}</FormItem>
        </Form>
      </Modal>
    );
  }
}

export default connect(() => ({}))(DeleteUnitModal);
