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

import React from 'react';
import { formatMessage } from '@/util/intl';
import { connect } from 'umi';
import { Descriptions, Modal } from '@oceanbase/design';

export interface DeleteUnitSpecModalProps {
  dispatch: any;
  visible: boolean;
  unitSpec: API.UnitSpec;
  onSuccess?: Function;
  onCancel?: (e: any) => void;
}

const DeleteUnitSpecModal: React.FC<DeleteUnitSpecModalProps> = ({
  dispatch,
  visible,
  unitSpec = {},
  onSuccess,
  ...restProps
}) => {
  const handleDelete = () => {
    dispatch({
      type: 'tenant/deleteUnitSpec',
      payload: {
        specId: unitSpec.id,
      },

      name: unitSpec.name,
      onSuccess: () => {
        if (typeof onSuccess === 'function') {
          onSuccess();
        }
      },
    });
  };

  return (
    <Modal
      title={formatMessage({
        id: 'ocp-express.src.component.DeleteUnitSpecModal.DeleteAUnitSpecification',
        defaultMessage: '删除 Unit 规格',
      })}
      visible={visible}
      onOk={handleDelete}
      okText={formatMessage({
        id: 'ocp-express.src.component.DeleteUnitSpecModal.Delete',
        defaultMessage: '删除',
      })}
      okButtonProps={{ danger: true, ghost: true }}
      destroyOnClose={true}
      {...restProps}
    >
      <Descriptions column={1}>
        <Descriptions.Item
          label={formatMessage({
            id: 'ocp-express.src.component.AddUnitSpecModal.SpecificationName',
            defaultMessage: '规格名称',
          })}
        >
          {unitSpec?.name}
        </Descriptions.Item>
        <Descriptions.Item
          label={formatMessage({
            id: 'ocp-express.src.component.DeleteUnitSpecModal.CpuCores',
            defaultMessage: 'CPU 核数',
          })}
        >
          {`${unitSpec?.minCpuCoreCount} ~ ${unitSpec?.maxCpuCoreCount}`}
        </Descriptions.Item>
        <Descriptions.Item
          label={formatMessage({
            id: 'ocp-express.src.component.DeleteUnitSpecModal.Memory',
            defaultMessage: '内存',
          })}
          style={{ paddingBottom: 0 }}
        >
          {`${unitSpec?.minMemorySize || 0} ~ ${unitSpec?.maxMemorySize || 0} GB`}
        </Descriptions.Item>
      </Descriptions>
    </Modal>
  );
};

export default connect(() => ({}))(DeleteUnitSpecModal);
