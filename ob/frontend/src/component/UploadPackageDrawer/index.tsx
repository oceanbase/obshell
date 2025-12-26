import { formatMessage } from '@/util/intl'; // import * as SoftwarePackageController from '@/service/ocp-all-in-one/SoftwarePackageController';
import { uniqueId, find } from 'lodash';
import React, { useState, useEffect } from 'react';
import { Button, message, Drawer, Upload, Form, Table, Progress, Modal } from '@oceanbase/design';
import type { UploadFile } from '@oceanbase/design';
import { UploadOutlined } from '@oceanbase/icons';
import { newPkgUpload } from '@/service/obshell/package';
import { useCluster } from '@/hook/useCluster';
import { calculateFileSHA256 } from '@/util/package';
import styles from './index.less';

interface UploadPackageDrawerProps {
  open: boolean;
  onCancel: () => void;
  onSuccess: () => void;
}

// 扩展 UploadFile 类型，添加自定义属性
interface ExtendedUploadFile extends UploadFile {
  key?: string;
  id?: string;
  md5?: string;
  description?: string;
  checkResult?: string;
}

const formItemLayout = {
  wrapperCol: {
    span: 20,
  },
};

const UploadPackageDrawer: React.FC<UploadPackageDrawerProps> = ({
  open,
  onCancel,
  onSuccess,
  ...restProps
}) => {
  const { isStandalone, isDistributedBusiness } = useCluster();

  const [fileList, setFileList] = useState<ExtendedUploadFile[]>([]);

  useEffect(() => {
    setFileList([]);
  }, [open]);

  const normFile = (e: any) => {
    if (Array.isArray(e)) {
      return e;
    }
    return e && e.fileList;
  };

  const beforeUpload = (file: File) => {
    const isFileExists = fileList.some(f => f.name === file.name);
    if (isFileExists) {
      message.warning(
        formatMessage(
          {
            id: 'ocp-v2.component.UploadPackageDrawer.FilenamePackageAlreadyExistsPlease',
            defaultMessage: '{fileName} 软件包已存在，请重新选择',
          },
          { fileName: file.name }
        )
      );
      return false;
    }
    return true;
  };

  const handleUpload = (info: any) => {
    const currentFile = info?.file;

    const existedFile = find(fileList, item => item.name === currentFile.name);

    if (info?.file?.status === 'uploading') {
      if (!existedFile || fileList.length === 0) {
        setFileList(fList => [
          ...fList,
          {
            ...currentFile,
            key: uniqueId(),
          },
        ]);
      } else if (existedFile && existedFile.key) {
        setFileList(fList =>
          fList.map(item => {
            if (item.name === currentFile.name) {
              return {
                ...existedFile,
                percent: currentFile.percent,
              };
            } else {
              return item;
            }
          })
        );
      }
    }

    if (currentFile.status === 'done') {
      if (existedFile) {
        setFileList(fList =>
          fList.map(item => {
            if (item.name === currentFile.name) {
              return {
                ...currentFile,
                key: existedFile.uid,
                id: currentFile.response?.pkg_id,
                md5: currentFile.response?.md5,
                description: currentFile?.response?.error?.message,
              };
            } else {
              return item;
            }
          })
        );
      }
    } else if (info?.file?.status === 'error') {
      if (existedFile) {
        setFileList(fList =>
          fList.map(item => {
            if (item.name === currentFile.name) {
              return {
                ...currentFile,
                key: existedFile.key,
                status: currentFile?.status,
                id: existedFile.response?.data?.id,
                md5: existedFile.response?.data?.md5,
                checkResult: formatMessage({
                  id: 'ocp-v2.component.UploadPackageDrawer.Failed',
                  defaultMessage: '失败',
                }),
                description: currentFile?.response?.error?.message,
                percent: 100,
              };
            } else {
              return item;
            }
          })
        );
      }
    }
  };

  const columns = [
    {
      title: formatMessage({
        id: 'ocp-v2.component.UploadPackageDrawer.PackageName',
        defaultMessage: '软件包名称',
      }),
      dataIndex: 'name',
      ellipsis: true,
      render: (text: string, record: any) => (
        <div style={{ maxWidth: '460px' }}>
          <div>{text}</div>
          <div style={{ color: '#8592ad', fontSize: '12px', lineHeight: '20px' }}>
            md5: {record.md5}
          </div>
        </div>
      ),
    },

    {
      title: formatMessage({
        id: 'ocp-v2.component.UploadPackageDrawer.UploadProgress',
        defaultMessage: '上传进度',
      }),
      dataIndex: 'percent',
      width: 110,
      render: (text: any, record: any) => {
        return !record.percent || record.percent === 0 ? (
          '-'
        ) : (
          <Progress
            percent={text > 0 ? parseFloat(text.toFixed()) : 0}
            {...(record?.status === 'error' && { status: 'exception' })}
          />
        );
      },
    },
  ];

  return (
    <Drawer
      title={formatMessage({
        id: 'ocp-v2.component.UploadPackageDrawer.UploadSoftwarePackage',
        defaultMessage: '上传软件包',
      })}
      open={open}
      width={1000}
      destroyOnClose={true}
      maskClosable={false}
      okText={formatMessage({
        id: 'ocp-v2.component.UploadPackageDrawer.Upload',
        defaultMessage: '上传',
      })}
      onClose={() => {
        if (fileList.some(item => item.status !== 'done')) {
          Modal.confirm({
            title: formatMessage({
              id: 'ocp-v2.component.UploadPackageDrawer.TheCurrentFileIsBeing',
              defaultMessage: '当前文件上传中，是否确定退出？',
            }),
            content: formatMessage({
              id: 'ocp-v2.component.UploadPackageDrawer.AfterExitingTheFileWill',
              defaultMessage: '退出后文件会在后台继续上传，但是无法查看进度',
            }),

            okText: formatMessage({
              id: 'ocp-v2.component.UploadPackageDrawer.Exit',
              defaultMessage: '退出',
            }),
            onOk: () => {
              onCancel();
            },
          });
        } else {
          onCancel();
        }
      }}
      afterOpenChange={() => {
        setFileList([]);
      }}
      footer={null}
      {...restProps}
    >
      <Form hideRequiredMark={true} preserve={false} {...formItemLayout}>
        <Form.Item
          className={styles.uploadPackageFormItem}
          style={{
            marginBottom: 0,
          }}
          valuePropName="fileList"
          getValueFromEvent={normFile}
          rules={[
            {
              required: true,
              message: formatMessage({
                id: 'ocp-v2.component.UploadPackageDrawer.PleaseSelectASoftwarePackage',
                defaultMessage: '请选择软件包',
              }),
            },
          ]}
          extra={
            !isStandalone &&
            !isDistributedBusiness && (
              <a
                href="https://mirrors.aliyun.com/oceanbase/community/stable/el/7/x86_64/"
                target="_blank"
              >
                {formatMessage({
                  id: 'ocp-v2.component.UploadPackageDrawer.RpmPackageDownloadAddress',
                  defaultMessage: 'RPM 包下载地址',
                })}
              </a>
            )
          }
        >
          <Upload
            accept=".rpm"
            name="file"
            // action={`${window?.location?.hash}/api/v2/software-packages`}
            beforeUpload={beforeUpload}
            customRequest={async options => {
              const file = options.file as File;
              // 从原始文件中获取 uid（Ant Design Upload 会为每个文件生成 uid）
              const fileUid = (options.file as any).uid;

              // 计算文件大小
              const totalMB = file.size / 1024 / 1024;
              let uploadedMb = 2;

              // umi 请求使用的 fetch 无法监听上传进度，改为 xhr 会绕过 interceptors 的加密逻辑，改为模拟进度条
              const timer = setInterval(() => {
                uploadedMb += 2;
                const percent = (uploadedMb / totalMB) * 100;

                if (percent >= 95) {
                  clearInterval(timer);
                  return;
                }

                options.onProgress?.({
                  percent,
                });
              }, 1000);

              try {
                // 使用分片方式计算 SHA256，避免大文件内存溢出
                const sha256 = await calculateFileSHA256(file);

                // 上传文件
                newPkgUpload({}, file, {
                  sha256sum: sha256,
                }).then(res => {
                  if (res?.successful) {
                    // timer 中有更新 percent 的逻辑，避免和 options.onSuccess 影响，加一个 delay 处理
                    clearInterval(timer);
                    setTimeout(() => {
                      options.onSuccess?.(res.data);
                      message.success({
                        content: formatMessage(
                          {
                            id: 'ocp-v2.component.UploadPackageDrawer.FilenameUploadedSuccessfully',
                            defaultMessage: '{fileName} 上传成功',
                          },
                          { fileName: file.name }
                        ),
                        style: { marginTop: 30 },
                      });
                      onSuccess();
                    }, 500);
                  } else {
                    // 网络波动导致浏览器关闭请求，此处不会返回错误值
                    if (!res) {
                      message.error({
                        content: formatMessage({
                          id: 'ocp-v2.component.UploadPackageDrawer.NetworkExceptionPleaseTryAgain',
                          defaultMessage: '网络异常，请稍后再试',
                        }),
                        style: { marginTop: 30 },
                      });
                    }
                    clearInterval(timer);
                    setTimeout(() => {
                      setFileList(
                        fileList.filter(item => {
                          return item.uid !== fileUid;
                        })
                      );
                    }, 500);
                  }
                });
              } catch (error) {
                clearInterval(timer);
                console.error('File processing error:', error);

                // 根据错误类型显示不同的提示
                const errorMessage =
                  error instanceof RangeError
                    ? formatMessage({
                        id: 'OBShell.component.UploadPackageDrawer.TheFileIsTooLarge',
                        defaultMessage: '文件过大，浏览器无法处理，请联系技术同学使用 SDK 上传',
                      })
                    : formatMessage({
                        id: 'OBShell.component.UploadPackageDrawer.FileProcessingFailedPleaseTry',
                        defaultMessage: '文件处理失败，请重试或联系技术同学',
                      });

                message.error({
                  content: errorMessage,
                  style: { marginTop: 30 },
                });
              }
            }}
            withCredentials={true}
            showUploadList={false}
            multiple={10}
            onChange={handleUpload}
          >
            <Button>
              <UploadOutlined />
              {formatMessage({
                id: 'ocp-v2.component.UploadPackageDrawer.SelectSoftwarePackage',
                defaultMessage: '选择软件包',
              })}
            </Button>
          </Upload>
        </Form.Item>
      </Form>
      {fileList?.length > 0 && (
        <>
          <div
            style={{
              lineHeight: '22px',
              marginTop: '24px',
              marginBottom: '4px',
              color: '#132039',
            }}
          >
            {formatMessage({
              id: 'ocp-v2.component.UploadPackageDrawer.SoftwarePackage',
              defaultMessage: '软件包',
            })}
            {`（${fileList?.length}）`}
          </div>
          <Table dataSource={fileList} columns={columns} rowKey="key" pagination={false} />
        </>
      )}
    </Drawer>
  );
};

export default UploadPackageDrawer;
