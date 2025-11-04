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

import { formatMessage } from '@/util/intl';
import React from 'react';
import { Alert, Form, Modal, message } from '@oceanbase/design';
import { useRequest } from 'ahooks';
import WhitelistInput from '@/component/WhitelistInput';
import { modifyWhitelist } from '@/service/obshell/observer';

const FormItem = Form.Item;

export interface ModifyWhitelistModalProps {
  obInfoData: API.ObserverInfo;
  onSuccess: () => void;
}

const ModifyWhitelistModal: React.FC<ModifyWhitelistModalProps> = ({
  obInfoData,
  onSuccess,
  ...restProps
}) => {
  const [form] = Form.useForm();
  const { validateFields } = form;

  const { run, loading } = useRequest(modifyWhitelist, {
    manual: true,
    onSuccess: res => {
      if (res.successful) {
        message.success(
          formatMessage({
            id: 'ocp-express.Detail.Component.ModifyWhitelistModal.IpAddressWhitelistModified',
            defaultMessage: 'IP 白名单修改成功',
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
      const { whitelist } = values;
      run({ whitelist });
    });
  };

  return (
    <Modal
      width={600}
      destroyOnClose={true}
      title={formatMessage({
        id: 'ocp-express.Detail.Component.ModifyWhitelistModal.ModifyTheWhitelist',
        defaultMessage: '修改白名单',
      })}
      onOk={handleSubmit}
      confirmLoading={loading}
      {...restProps}
    >
      <Alert
        type="info"
        showIcon={true}
        style={{ marginBottom: 24 }}
        message={formatMessage({
          id: 'ocp-express.Detail.Component.ModifyWhitelistModal.OnlyIpAddressesThatHave',
          defaultMessage: '只有已添加到白名单中的 IP 地址才可以访问该实例',
        })}
        description={
          <ul>
            <li>
              {formatMessage({
                id: 'ocp-express.Detail.Component.ModifyWhitelistModal.YouCanEnterAnIp',
                defaultMessage: '· 可以填写 IP 地址（如 127.0.0.1）或 IP 段（如 127.0.0.1/24）',
              })}
            </li>
            <li>
              {formatMessage({
                id: 'ocp-express.Detail.Component.ModifyWhitelistModal.MultipleIpAddressesNeedTo',
                defaultMessage: '· 多个 IP 需要用英文逗号隔开，如 127.0.0.1,127.0.0.1/24',
              })}
            </li>
            <li>
              {formatMessage({
                id: 'ocp-express.Detail.Component.ModifyWhitelistModal.IndicatesThatAccessToAnyIpAddressIs',
                defaultMessage: '127.0.0.1 表示禁止任何 IP 地址访问',
              })}
            </li>
          </ul>
        }
      />
      <Form form={form} preserve={false} layout="vertical" hideRequiredMark={true}>
        <FormItem
          name="whitelist"
          rules={[
            {
              validator: WhitelistInput.validate,
            },
          ]}
          initialValue={obInfoData.whitelist}
        >
          <WhitelistInput layout="vertical" />
        </FormItem>
      </Form>
    </Modal>
  );
};

export default ModifyWhitelistModal;
