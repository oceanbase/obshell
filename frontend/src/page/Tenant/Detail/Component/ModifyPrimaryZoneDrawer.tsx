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

import { tenantModifyPrimaryZone } from '@/service/obshell/tenant';
import { formatMessage } from '@/util/intl';
import { useDispatch } from 'umi';
import React from 'react';
import { Form, message } from '@oceanbase/design';
import * as ObTenantController from '@/service/ocp-express/ObTenantController';
import { useRequest } from 'ahooks';
import MyDrawer from '@/component/MyDrawer';
import FormPrimaryZone from '@/component/FormPrimaryZone';
import { taskSuccess } from '@/util/task';

const FormItem = Form.Item;

export interface ModifyPrimaryZoneDrawerProps {
  tenantData?: API.TenantInfo;
  onSuccess?: () => void;
}

const ModifyPrimaryZoneDrawer: React.FC<ModifyPrimaryZoneDrawerProps> = ({
  tenantData,
  onSuccess,
  zones,
  ...restProps
}) => {
  const [form] = Form.useForm();
  const { validateFields } = form;

  const { run, loading } = useRequest(tenantModifyPrimaryZone, {
    manual: true,
    onSuccess: res => {
      if (res.successful) {
        taskSuccess(res.data.id);
        message.success(
          formatMessage({
            id: 'ocp-express.Detail.Component.ModifyPrimaryZoneDrawer.TheZonePriorityHasBeen',
            defaultMessage: 'Zone 优先级修改成功',
          })
        );
        if (onSuccess) {
          onSuccess();
        }
      }
    },
  });

  const handleSubmit = () => {
    validateFields().then(values => {
      const { primaryZone } = values;
      run(
        {
          name: tenantData?.name,
        },

        { primary_zone: primaryZone }
      );
    });
  };

  return (
    <MyDrawer
      width={1184}
      destroyOnClose={true}
      title={formatMessage({
        id: 'ocp-express.Detail.Component.ModifyPrimaryZoneDrawer.ModifyZonePriority',
        defaultMessage: '修改 Zone 优先级',
      })}
      onOk={handleSubmit}
      confirmLoading={loading}
      {...restProps}
    >
      <Form layout="inline" hideRequiredMark={true}>
        <FormItem
          label={formatMessage({
            id: 'ocp-express.Detail.Component.ModifyPrimaryZoneDrawer.TenantName',
            defaultMessage: '租户名',
          })}
          style={{ marginBottom: 24 }}
        >
          {tenantData.name}
        </FormItem>
      </Form>
      <Form form={form} preserve={false} layout="vertical" hideRequiredMark={true}>
        <FormItem
          label={formatMessage({
            id: 'ocp-express.Detail.Component.ModifyPrimaryZoneDrawer.ZonePrioritySorting',
            defaultMessage: 'Zone 优先级排序',
          })}
          name="primaryZone"
          initialValue={tenantData.primaryZone}
        >
          <FormPrimaryZone zoneList={(zones || []).map(item => item.name)} />
        </FormItem>
      </Form>
    </MyDrawer>
  );
};

export default ModifyPrimaryZoneDrawer;
