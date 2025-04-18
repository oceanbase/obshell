// import * as SoftwarePackageController from '@/service/ocp-all-in-one/SoftwarePackageController';
import { uniqueId, find } from "lodash";
import React, { useState, useEffect } from "react";
import {
  Button,
  message,
  Drawer,
  Upload,
  Form,
  Table,
  Progress,
} from "@oceanbase/design";
import type { UploadFile } from "@oceanbase/design";
import { UploadOutlined } from "@oceanbase/icons";
import { useRequest } from "ahooks";
import CryptoJS from "crypto-js";
import { newPkgUpload } from "@/service/obshell/package";
import "./index.less";

interface UploadPackageDrawerProps {
  open: boolean;
  onCancel: () => void;
  onSuccess: () => void;
}

const formItemLayout = {
  wrapperCol: {
    span: 20,
  },
};

const SoftwarePackageController = {};
const UploadPackageDrawer: React.FC<UploadPackageDrawerProps> = ({
  open,
  onCancel,
  onSuccess,
  ...restProps
}) => {
  const [fileList, setFileList] = useState<UploadFile[]>([]);

  useEffect(() => {
    setFileList([]);
  }, [open]);

  // 用于中断 fetch 请求
  const [abortControl, setAbortControl] = useState(new AbortController());

  const { runAsync: upgradePkgUploadFn } = useRequest(newPkgUpload, {
    manual: true,
  });

  const normFile = (e: any) => {
    if (Array.isArray(e)) {
      return e;
    }
    return e && e.fileList;
  };

  const beforeUpload = (file) => {
    const isFileExists = fileList.some((f) => f.name === file.name);
    console.log(file, "file");
    if (isFileExists) {
      message.warning(`${file.name} 软件包已存在，请重新选择`);
      return false;
    }
    return true;
  };

  const handleUpload = (info: any) => {
    const currentFile = info?.file;

    const existedFile = find(
      fileList,
      (item) => item.name === currentFile.name
    );

    if (info?.file?.status === "uploading") {
      if (!existedFile || fileList.length === 0) {
        setFileList([
          ...fileList,
          {
            ...currentFile,
            key: uniqueId(),
          },
        ]);
      } else if (existedFile && existedFile.key) {
        const currentFileList = fileList.map((item) => {
          if (item.name === currentFile.name) {
            return {
              ...existedFile,
              percent: currentFile.percent,
            };
          } else {
            return item;
          }
        });
        setFileList(currentFileList);
      }
    }

    if (currentFile.status === "done") {
      if (existedFile) {
        const currentFileList = fileList.map((item) => {
          if (item.name === currentFile.name) {
            return {
              ...currentFile,
              key: existedFile.key,
              id: currentFile.response?.data?.id,
              md5: currentFile.response?.data?.md5,
              description: currentFile?.response?.error?.message,
              // hasSignature: currentFile.response?.data?.hasSignature,
              // signatureValid: currentFile.response?.data?.signatureValid,
            };
          } else {
            return item;
          }
        });
        setFileList(currentFileList);
      }
    } else if (info?.file?.status === "error") {
      if (existedFile) {
        const currentFileList = fileList.map((item) => {
          if (item.name === currentFile.name) {
            return {
              ...currentFile,
              key: existedFile.key,
              status: currentFile?.status,
              id: existedFile.response?.data?.id,
              md5: existedFile.response?.data?.md5,
              checkResult: "失败",
              description: currentFile?.response?.error?.message,
              percent: 100,
            };
          } else {
            return item;
          }
        });
        setFileList(currentFileList);
      }
    }
  };

  // const handleCancel = key => {
  //   if (key) {
  //     abortControl?.abort();
  //     setAbortControl(new AbortController());
  //     const currentFile = fileList.find(item => item.key === key);
  //     setFileList(fileList.filter(item => item.key !== key));
  //     if (currentFile.id) {
  //       deleteSoftwarePackage({ id: currentFile.id }).then(res => {
  //         if (res.successful) {
  //           message.success('软件包删除成功');
  //         }
  //       });
  //     } else {
  //       message.success('软件包删除成功');
  //     }
  //   } else if (onCancel) {
  //     fileList
  //       .filter(item => item.status === 'done')
  //       .forEach(item => {
  //         deleteSoftwarePackage({ id: item.id });
  //       });
  //     if (fileList.filter(item => item.status === 'uploading').length > 0) {
  //       abortControl?.abort();
  //       setAbortControl(new AbortController());
  //     }
  //     setFileList([]);
  //     onCancel();
  //   }
  // };

  const columns = [
    {
      title: "软件包名称",
      dataIndex: "name",
      ellipsis: true,
      render: (text: string, record) => (
        <div style={{ maxWidth: "460px" }}>
          <div>{text}</div>
          <div
            style={{ color: "#8592ad", fontSize: "12px", lineHeight: "20px" }}
          >
            md5: {record.md5}
          </div>
        </div>
      ),
    },

    {
      title: "上传进度",
      dataIndex: "percent",
      width: 110,
      render: (text, record) => {
        return !record.percent || record.percent === 0 ? (
          "-"
        ) : (
          <Progress
            percent={text > 0 ? parseFloat(text.toFixed()) : 0}
            {...(record?.status === "error" && { status: "exception" })}
          />
        );
      },
    },

    // {
    //   title: '说明',
    //   width: 110,
    //   dataIndex: 'checkResult',
    //   render: (text: string, record) => {
    //     if (record?.status === 'done') {
    //       // hasSignature = false ----> 当非正式包
    //       // hasSignature = true 加上 signatureValid = false ----> 报错签名校验不通过
    //       // hasSignature = true 加上 signatureValid = true ----> 没有提示
    //       if (record.response?.data?.hasSignature) {
    //         if (record.signatureValid) {
    //           return '通过';
    //         } else {
    //           return (
    //             <ContentWithIcon
    //               content={<span style={{ color: '#e68800' }}>{'存在风险'}</span>}
    //               tooltip={{
    //                 title: record?.description,
    //               }}
    //               suffixIcon={<InfoCircleOutlined style={{ color: '#5c6b8a' }} />}
    //             />
    //           );
    //         }
    //       } else {
    //         return (
    //           <ContentWithIcon
    //             content={<span style={{ color: '#e68800' }}>{'存在风险'}</span>}
    //             tooltip={{
    //               title: '此软件包为非正式包，存在一定风险，建议不要部署在正式集群环境',
    //             }}
    //             suffixIcon={<InfoCircleOutlined style={{ color: '#5c6b8a' }} />}
    //           />
    //         );
    //       }
    //     } else if (record?.status === 'error') {
    //       return (
    //         <ContentWithIcon
    //           content={text}
    //           tooltip={{
    //             title: record.description,
    //           }}
    //           color="#132039"
    //           suffixIcon={<InfoCircleOutlined style={{ color: '#5c6b8a' }} />}
    //         />
    //       );
    //     } else {
    //       return '-';
    //     }
    //   },
    // },

    // {
    //   title: '操作',
    //   width: 60,
    //   dataIndex: 'operation',
    //   render: (text, record) =>
    //     record?.status === 'done' && (
    //       <Popconfirm
    //         placement="topLeft"
    //         title={<div style={{ width: '235px' }}>确定要删除软件包 {record.name} 吗？</div>}
    //         okText={'删除'}
    //         cancelText={'取消'}
    //         okButtonProps={{
    //           danger: true,
    //         }}
    //         onConfirm={() => {
    //           new Promise(resolve => {
    //             setTimeout(() => {
    //               handleCancel(record.key);
    //               resolve(null);
    //             }, 2000);
    //           });
    //         }}
    //       >
    //         <a>{'删除'}</a>
    //       </Popconfirm>
    //     ),
    // },
  ];

  return (
    <Drawer
      title={"上传软件包"}
      open={open}
      width={800}
      destroyOnClose={true}
      maskClosable={false}
      okText={"上传"}
      onClose={() => {
        onCancel();
      }}
      afterOpenChange={() => {
        setFileList([]);
      }}
      footer={
        null
        // <Space>
        //   <Button
        //     onClick={() => {
        //       message.success('软件包上传成功');
        //       setFileList([]);
        //       if (onSuccess) {
        //         onSuccess();
        //       }
        //     }}
        //     type="primary"
        //     disabled={!fileList?.length || fileList.find(item => item.status === 'uploading')}
        //   >
        //     {'确定'}
        //   </Button>
        //   <Button
        //     onClick={() => {
        //       handleCancel(null);
        //     }}
        //   >
        //     {'取消'}
        //   </Button>
        //   {fileList?.length > 0 && fileList.find(item => item.status === 'uploading') && (
        //     <span style={{ color: '#8592ad' }}>{'仍有软件包在上传中，请耐心等待'}</span>
        //   )}
        // </Space>
      }
      {...restProps}
    >
      <Form hideRequiredMark={true} preserve={false} {...formItemLayout}>
        <Form.Item
          style={{
            marginBottom: 0,
          }}
          valuePropName="fileList"
          getValueFromEvent={normFile}
          rules={[
            {
              required: true,
              message: "请选择软件包",
            },
          ]}
          extra={"只支持后缀名为 .rpm 的文件"}
        >
          <Upload
            accept=".rpm"
            name="file"
            // action={`${window?.location?.hash}/api/v2/software-packages`}
            beforeUpload={beforeUpload}
            customRequest={async (options) => {
              const file = options.file;

              // 400 MB 等于 120 秒，一秒约等于 3 MB
              const totalMB = file.size / 1024 / 1024;
              let uploadedMb = 0;

              // umi 请求使用的 fetch 无法监听上传进度，改为 xhr 会绕过 interceptors 的加密逻辑，改为模拟进度条
              const timer = setInterval(() => {
                uploadedMb += 3;
                const percent = (uploadedMb / totalMB) * 100;
                console.log(totalMB, uploadedMb, "totalMB");

                if (percent >= 90) {
                  clearInterval(timer);
                  return;
                }

                options.onProgress({
                  percent,
                });
              }, 1000);

              // 1. 将 Blob 转换为 ArrayBuffer
              const arrayBuffer = await file.arrayBuffer();
              try {
                const wordArray = CryptoJS.lib.WordArray.create(arrayBuffer);
                const sha256 = CryptoJS.enc.Hex.stringify(
                  CryptoJS.SHA256(wordArray)
                );
                upgradePkgUploadFn({}, file, {
                  sha256sum: sha256,
                }).then((res) => {
                  if (res.successful) {
                    options.onSuccess(res.data);
                    message.success(`${file.name} 上传功能`);
                  } else {
                    setFileList(
                      fileList.filter((item) => {
                        item.uid !== file.uid;
                      })
                    );
                  }
                });
              } catch (error) {
                console.error("file error: ", error);
                message.warning("上传文件过大，请联系技术同学使用 SDK 上传");
              }
            }}
            withCredentials={true}
            showUploadList={false}
            multiple={10}
            onChange={handleUpload}
          >
            <Button>
              <UploadOutlined />
              {"选择软件包"}
            </Button>
          </Upload>
        </Form.Item>
      </Form>
      {fileList?.length > 0 && (
        <>
          <div
            style={{
              lineHeight: "22px",
              marginTop: "24px",
              marginBottom: "4px",
              color: "#132039",
            }}
          >
            {"软件包"}
            {`（${fileList?.length}）`}
          </div>
          <Table
            dataSource={fileList}
            columns={columns}
            rowKey="key"
            pagination={false}
          />
        </>
      )}
    </Drawer>
  );
};

export default UploadPackageDrawer;
