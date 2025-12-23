import BatchOperationBar from '@/component/BatchOperationBar';
import { formatMessage } from '@/util/intl';
import * as CredentialController from '@/service/obshell/security';
import React from 'react';
import { Button, message, Modal } from '@oceanbase/design';
import { useRequest } from 'ahooks';
import { HostCredential } from '../Host';

export interface FilterBarProps {
  selectedRowKeys?: HostCredential[];
  setSelectedRowKeys?: (value: HostCredential[]) => void;
  isFilter: boolean;
  setIsFilter: (value: boolean) => void;
  isClick: boolean;
  setIsClick: (value: boolean) => void;
  batchValidateCredential: (value: { credential_id_list: number[] }) => void;
  onCancel?: () => void;
  onSuccess?: () => void;
  batchValidateCredentialLoading: boolean;
}

const FilterBar: React.FC<FilterBarProps> = ({
  selectedRowKeys,
  setSelectedRowKeys,
  isFilter,
  setIsFilter,
  isClick,
  batchValidateCredential,
  batchValidateCredentialLoading,
  onCancel = () => {},
  onSuccess = () => {},
  setIsClick,
}) => {
  // 批量删除
  const { runAsync: batchDeleteCredentials } = useRequest(
    CredentialController.batchDeleteCredentials,
    {
      manual: true,
      onSuccess: res => {
        if (res.successful) {
          message.success('凭据删除成功');
          if (setSelectedRowKeys) {
            setSelectedRowKeys([]);
          }
          onSuccess();
        }
      },
    }
  );

  return (
    <BatchOperationBar
      description={
        <Button
          type="link"
          onClick={() => setIsFilter(!isFilter)}
          style={{ margin: '0 4px' }}
          disabled={batchValidateCredentialLoading && isClick}
        >
          {isFilter
            ? formatMessage({
                id: 'ocp-v2.Credential.Component.FilterBar.ReturnAll',
                defaultMessage: '返回全部',
              })
            : formatMessage({
                id: 'ocp-v2.Credential.Component.FilterBar.OnlySelected',
                defaultMessage: '仅看已选',
              })}
        </Button>
      }
      style={{ marginBottom: 16 }}
      selectedCount={selectedRowKeys?.length || 0}
      onCancel={() => {
        onCancel();
        if (setSelectedRowKeys) {
          setSelectedRowKeys([]);
        }
      }}
      actions={[
        <>
          <Button
            type="link"
            style={{ margin: '0 16px' }}
            loading={batchValidateCredentialLoading && isClick}
            onClick={() => {
              setIsClick(true);
              batchValidateCredential({
                credential_id_list: selectedRowKeys?.map(item => item?.credential_id!) || [],
              });
            }}
          >
            {formatMessage({
              id: 'ocp-v2.Credential.Component.FilterBar.BatchVerification',
              defaultMessage: '批量验证',
            })}
          </Button>
        </>,
        <Button
          type="link"
          disabled={batchValidateCredentialLoading && isClick}
          onClick={() => {
            Modal.confirm({
              title:
                selectedRowKeys &&
                formatMessage(
                  {
                    id: 'ocp-v2.Credential.Component.FilterBar.AreYouSureYouWantToDeleteThese',
                    defaultMessage: '确定要删除这 {selectedRowKeysLength} 个吗？',
                  },

                  { selectedRowKeysLength: selectedRowKeys.length }
                ),

              okText: formatMessage({
                id: 'ocp-v2.Credential.Component.FilterBar.Delete',
                defaultMessage: '删除',
              }),

              okButtonProps: { danger: true, ghost: true },
              onOk: () => {
                return batchDeleteCredentials({
                  credential_id_list:
                    selectedRowKeys?.map(selected => selected?.credential_id!) || [],
                });
              },
              onCancel: () => {},
            });
          }}
        >
          {formatMessage({
            id: 'ocp-v2.Credential.Component.FilterBar.BatchDeletion',
            defaultMessage: '批量删除',
          })}
        </Button>,
      ]}
    />
  );
};

export default FilterBar;
