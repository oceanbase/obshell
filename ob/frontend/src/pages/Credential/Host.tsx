// import AddHostCredentialDrawer from '@/component/AddHostCredentialDrawer';
import MyInput from '@/component/MyInput';
import TagRenderer from '@/component/TagRenderer';
import { PAGINATION_OPTION_10 } from '@/constant';
import { AUTHORIZE_TYPE_LIST } from '@/constant/compute';
import { OPERATION_LIST } from '@/constant/credential';
import { useListCredentials } from '@/hook/useListCredentials';
import { formatMessage } from '@/util/intl';
import * as CredentialController from '@/service/obshell/security';
import { isEnglish } from '@/util';
import { getOperationComponent } from '@/util/component';
import React, { forwardRef, useEffect, useImperativeHandle, useState } from 'react';
import { find, isEmpty, orderBy } from 'lodash';
import {
  Button,
  Card,
  Col,
  message,
  Modal,
  Row,
  Space,
  Table,
  TableProps,
  Tooltip,
} from '@oceanbase/design';
import { findByValue } from '@oceanbase/util';
import { useRequest } from 'ahooks';
import ConnectTag from './Component/ConnectTag';
import FilterBar from './Component/FilterBar';
import AddHostCredentialDrawer from '@/component/AddHostCredentialDrawer';

export interface HostCredential {
  credential_id: number;
  name?: string;
  targets?: API.Target[];
  type?: string;
  username?: string;
  passphrase?: string;
  description?: string;
  target_type?: 'HOST';
  failed_count?: number;
  details?: API.ValidationDetail[];
}

export interface HostRef {
  refresh: () => void;
  loading: boolean;
}
interface HostProps {
  onLoadingChange?: (loading: boolean) => void;
}

const Host = forwardRef<HostRef, HostProps>(({ onLoadingChange }, ref) => {
  const [visible, setVisible] = useState(false);
  const [currentRecord, setCurrentRecord] = useState<HostCredential>();
  const [isEdit, setIsEdit] = useState(false);
  const [selectedRowKeys, setSelectedRowKeys] = useState<HostCredential[]>([]);
  const [isFilter, setIsFilter] = useState(false);
  const [credentialId, setCredentialId] = useState<number>();

  // 当前是否为单个连接测试
  const [singleValidate, setSingleValidate] = useState(false);

  // 获取主机的密码箱列表
  const { loading, credentialList, refresh, setKeyword } = useListCredentials('HOST');

  // 单个测试 loading 展示
  const [polling, setPolling] = useState({
    [-1]: false,
  });

  // 当前点击批量测试状态
  const [isClick, setIsClick] = useState(false);

  // 监听 loading 变化，通知父组件
  useEffect(() => {
    onLoadingChange?.(loading);
  }, [loading, onLoadingChange]);

  useImperativeHandle(ref, () => ({
    refresh: () => {
      refresh();
    },
    loading: loading,
  }));

  // 验证全部凭据
  const {
    run: batchValidateCredential,
    data: batchValidateCredentialData,
    loading: batchValidateCredentialLoading,
  } = useRequest(CredentialController.batchValidateCredentials, {
    manual: true,
    onSuccess: res => {
      if (res.successful) {
        setPolling({
          [credentialId]: false,
        });
        setSelectedRowKeys([]);
        setSingleValidate(false);
        setIsClick(false);
        // failed_count > 0 表示失败了，0表示成功
        const failure = res?.data?.contents?.filter(
          (item: API.ValidationResult) => item.failed_count && item.failed_count > 0
        );
        if (isEmpty(failure)) {
          message.success(
            singleValidate
              ? formatMessage({
                  id: 'OBShell.page.Credential.Host.ValidationSucceeded',
                  defaultMessage: '验证成功',
                })
              : formatMessage({
                  id: 'OBShell.page.Credential.Host.VerificationComplete',
                  defaultMessage: '验证完成',
                })
          );
        } else {
          if (singleValidate) {
            message.error(
              formatMessage({
                id: 'OBShell.page.Credential.Host.AuthenticationFailedPleaseModifyCredentials',
                defaultMessage: '验证失败，请修改凭据',
              })
            );
          } else {
            message.warning(
              formatMessage(
                {
                  id: 'OBShell.page.Credential.Host.ValidationCompletedThereFailurelengthAre',
                  defaultMessage: "'验证已完成，存在 {failureLength} 个失败凭据",
                },
                { failureLength: failure.length }
              )
            );
          }
        }
      }
    },
    onError: () => {
      setPolling({
        [credentialId]: false,
      });
      setSelectedRowKeys([]);
      setSingleValidate(false);
      setIsClick(false);
    },
  });

  const batchValidateCredentialList: API.ValidationResult[] =
    batchValidateCredentialData?.data?.contents || [];

  // 删除凭证
  const { run: deleteCredential } = useRequest(CredentialController.deleteCredential, {
    manual: true,
    onSuccess: (res, params) => {
      setSelectedRowKeys(prev => prev.filter(item => item.credential_id !== params?.[0]?.id));
      message.success(
        formatMessage({
          id: 'OBShell.page.Credential.Host.CredentialsDeletedSuccessfully',
          defaultMessage: '凭据删除成功',
        })
      );
      refresh();
    },
  });
  /**
   * 删除 HOST 凭据
   */
  const handleDelete = (credential: HostCredential) => {
    Modal.confirm({
      title: formatMessage(
        {
          id: 'OBShell.page.Credential.Host.AreYouSureYouWant',
          defaultMessage: '确定要删除凭据 {credentialName} 吗',
        },
        { credentialName: credential.name }
      ),

      okText: formatMessage({ id: 'OBShell.page.Credential.Host.Delete', defaultMessage: '删除' }),

      okButtonProps: { danger: true, ghost: true },
      onOk: () => {
        deleteCredential({ id: credential.credential_id! });
      },
    });
  };
  /**
   * 前端划分的 HOST 数据源
   */
  const dataSource = credentialList.map((credential: API.Credential) => {
    const { targets = [], type = '', username = '' } = credential?.ssh_secret || {};
    const validateCredential = find(
      batchValidateCredentialList,
      item => item.credential_id === credential?.credential_id
    );

    return credential
      ? {
          credential_id: credential?.credential_id,
          name: credential?.name,
          targets,
          type,
          username,
          passphrase: '',
          description: credential?.description,
          targetType: credential?.target_type || 'HOST',
          failed_count: validateCredential?.failed_count,
          details: validateCredential?.details,
        }
      : {};
  });

  function handleOperation(key: string, record: HostCredential) {
    if (key === 'edit') {
      setCurrentRecord(record);
      setIsEdit(true);
      setVisible(true);
    } else if (key === 'delete') {
      handleDelete(record);
    }
  }

  const columns: TableProps<HostCredential>['columns'] = [
    {
      title: formatMessage({
        id: 'OBShell.page.Credential.Host.CredentialName',
        defaultMessage: '凭据名',
      }),

      dataIndex: 'name',
      render: (text: string, record) => {
        const errorMge = record?.details?.find(e => e.message);
        return (
          <Space style={{ wordWrap: 'break-word', wordBreak: 'break-word' }}>
            <Tooltip placement="topLeft" title={text}>
              <span>{text}</span>
            </Tooltip>
            {/*
            验证失败:1，验证成功:0
            主机凭据没有主机 IP，则不显示 tag
            */}
            {!isEmpty(record?.details) && record?.failed_count !== undefined && (
              <ConnectTag
                status={record?.failed_count === 0 ? true : false}
                message={errorMge?.message!}
              />
            )}
          </Space>
        );
      },
    },

    {
      title: formatMessage({
        id: 'OBShell.page.Credential.Host.ApplicationHost',
        defaultMessage: '应用主机',
      }),

      dataIndex: 'targets',
      ellipsis: true,
      width: '20%',
      render: (targets: API.Target[], record) => {
        return targets?.length > 0 ? (
          <TagRenderer
            tags={
              // 错误信息的靠前排列
              orderBy(
                targets?.map(target => {
                  const host = target.ip;
                  const errorTarget = record?.details?.find(
                    item1 => item1.target?.ip === target?.ip && item1.target?.port === target?.port
                  );
                  // 查找当前主机的错误信息
                  const errorMge =
                    !isEmpty(errorTarget) && errorTarget?.connection_result !== 'SUCCESS';

                  const color = errorMge ? 'red' : 'default';
                  return {
                    text: host!,
                    color,
                  };
                }),
                'color',
                'desc'
              )
            }
          />
        ) : (
          '-'
        );
      },
    },

    {
      title: formatMessage({
        id: 'OBShell.page.Credential.Host.AuthorizationType',
        defaultMessage: '授权类型',
      }),

      dataIndex: 'type',

      render: (text: string) => <span>{findByValue(AUTHORIZE_TYPE_LIST, text).label}</span>,
    },

    {
      title: formatMessage({
        id: 'OBShell.page.Credential.Host.UserName',
        defaultMessage: '用户名',
      }),

      dataIndex: 'username',
    },

    {
      title: formatMessage({
        id: 'OBShell.page.Credential.Host.Description',
        defaultMessage: '说明',
      }),

      ellipsis: true,
      dataIndex: 'description',
      render: (text: string) => {
        return text ? (
          <Tooltip placement="topLeft" title={text}>
            <span>{text}</span>
          </Tooltip>
        ) : (
          <span>-</span>
        );
      },
    },
    {
      title: formatMessage({
        id: 'OBShell.page.Credential.Host.Operation',
        defaultMessage: '操作',
      }),

      dataIndex: 'operation',
      render: (text: string, record: HostCredential) => {
        return (
          <Space size={16}>
            <Button
              size="small"
              key={record.credential_id}
              loading={polling[record.credential_id]}
              onClick={() => {
                setPolling({
                  ...polling,
                  [record.credential_id]: true,
                });
                setSingleValidate(true);
                setCredentialId(record.credential_id);
                batchValidateCredential({ credential_id_list: [record.credential_id] });
              }}
            >
              {formatMessage({
                id: 'OBShell.page.Credential.Host.Validation',
                defaultMessage: '验证',
              })}
            </Button>
            {getOperationComponent({
              operations: OPERATION_LIST,
              handleOperation,
              record,
              displayCount: 0,
            })}
          </Space>
        );
      },
    },
  ];

  return (
    <>
      <Row gutter={[16, 16]}>
        <Col span={24}>
          <Card
            divided={true}
            bordered={false}
            title={formatMessage({
              id: 'OBShell.page.Credential.Host.CredentialList',
              defaultMessage: '凭据列表',
            })}
            extra={
              <Space size={16}>
                <span>
                  <MyInput.Search
                    allowClear={true}
                    onSearch={value => {
                      setKeyword(value);
                    }}
                    placeholder={formatMessage({
                      id: 'OBShell.page.Credential.Host.PleaseEnterCredentialNameHost',
                      defaultMessage: '请输入凭据名、主机 IP',
                    })}
                    className={isEnglish() ? 'search-input-large' : 'search-input'}
                  />
                </span>
              </Space>
            }
            className="card-without-padding"
          >
            {selectedRowKeys.length > 0 && (
              <FilterBar
                selectedRowKeys={selectedRowKeys}
                setSelectedRowKeys={setSelectedRowKeys}
                isFilter={isFilter}
                setIsFilter={setIsFilter}
                onCancel={() => {
                  setIsFilter(false);
                  setIsClick(false);
                }}
                onSuccess={() => {
                  refresh();
                }}
                isClick={isClick}
                setIsClick={setIsClick}
                batchValidateCredential={batchValidateCredential}
                batchValidateCredentialLoading={batchValidateCredentialLoading}
              />
            )}

            <Table
              loading={loading}
              pagination={PAGINATION_OPTION_10}
              dataSource={
                isFilter && selectedRowKeys.length > 0
                  ? dataSource.filter(item =>
                      selectedRowKeys
                        .map(selected => selected?.credential_id)
                        .includes(item.credential_id!)
                    )
                  : dataSource
              }
              columns={columns}
              rowKey={record => record.credential_id}
              rowSelection={{
                selectedRowKeys: selectedRowKeys.map(item => item.credential_id!),
                onChange: (keys, rows) => {
                  setSelectedRowKeys(rows);
                },
                getCheckboxProps: () => ({
                  disabled: batchValidateCredentialLoading && isClick,
                }),
              }}
              hiddenCancelBtn
              toolAlertRender={false}
            />
          </Card>
        </Col>
      </Row>

      <AddHostCredentialDrawer
        visible={visible}
        onCancel={() => {
          setVisible(false);
          setCurrentRecord(undefined);
        }}
        onSuccess={() => {
          setVisible(false);
          refresh();
          setCurrentRecord(undefined);
        }}
        isEdit={isEdit}
        credential={currentRecord}
      />
    </>
  );
});

export default Host;
