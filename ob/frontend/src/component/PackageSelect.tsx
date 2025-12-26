import MySelect from '@/component/MySelect';
import UploadPackageDrawer from '@/component/UploadPackageDrawer';
import SelectDropdownRender from '@/component/SelectDropdownRender';
import { buildVersionCompare, versionCompare } from '@/util/package';
import { formatMessage } from '@/util/intl';
import React, { useImperativeHandle, useState } from 'react';
import { uniqBy } from 'lodash';
import { Space, Tooltip } from '@oceanbase/design';
import type { SelectProps } from '@oceanbase/design/es/select';
import { useRequest } from 'ahooks';
import { upgradePkgInfo } from '@/service/obshell/upgrade';
import { useCluster } from '@/hook/useCluster';

const { Option } = MySelect;

export interface PackageSelectProps extends SelectProps<string> {
  // 接口支持的筛选参数
  keyword?: string;
  version?: string[];
  operatingSystem?: string[];
  architecture?: string[];
  /* 是否包含安装信息 */
  withInstallationData?: boolean;
  // 自定义参数
  /* 用于筛选的构建版本 */
  buildVersion?: string;
  /* 最小语义化版本 (包含)，常用于限制小于该版本的软件包不能被选中 (置灰处理) */
  minVersionForDisabled?: string;
  /* 最大语义化版本 (不包含)，常用于限制大于等于该版本的软件包不能被选中 (置灰处理) */
  maxVersionForDisabled?: string;
  /* 当前的构建版本，常用于升级场景中，过滤掉小于等于该构建版本的软件包 */
  currentBuildVersionForUpgrade?: string;
  /* 支持同版本降级时，传入当前ob 版本，不支持时传 null*/
  currentObVersion?: string;
  // 用途
  useType?: string;
  /* 是否是异构升级*/
  isHeterogeneousUpgrade?: boolean;
  isObproxyUpgrade?: boolean;
  /* 是否把每个选项的 label 包装到 value 中*/
  labelInValue?: boolean;
  // 当前 OBProxy 已有的软件包版本
  obproxyClusterOBVersions?: string[];

  onSuccess?: (packageList: API.UpgradePkgInfo[]) => void;
}

export interface PackageSelectRef {
  refresh: () => void;
}

const PackageSelect = React.forwardRef<PackageSelectRef, PackageSelectProps>(
  (
    {
      keyword,
      version,
      useType,
      operatingSystem,
      architecture,
      withInstallationData = true,
      buildVersion,
      minVersionForDisabled,
      maxVersionForDisabled,
      currentBuildVersionForUpgrade,
      obproxyClusterOBVersions,
      isHeterogeneousUpgrade,
      currentObVersion,
      labelInValue,
      isObproxyUpgrade,
      onSuccess,
      ...restProps
    },

    ref
  ) => {
    const { isStandalone, isDistributedBusiness } = useCluster();

    const [visible, setVisible] = useState(false);

    const { data, loading, refresh } = useRequest(upgradePkgInfo, {
      onSuccess: res => {
        if (res.successful) {
          onSuccess?.(res?.data?.contents || []);
        }
      },
    });

    const packageData: API.UpgradePkgInfo[] = data?.data?.contents || [];

    let packageList = packageData.filter(item => {
      if (useType === 'UPGRADE_CLUSTER') {
        if (isStandalone) {
          return item.name === 'oceanbase-standalone';
        }
        if (isDistributedBusiness) {
          return item.name === 'oceanbase';
        }
        return item.name === 'oceanbase-ce';
      }
      if (useType === 'UPGRADE_AGENT') {
        return item.name === 'obshell';
      }
      return true;
    });

    if (currentBuildVersionForUpgrade) {
      packageList = packageList.filter(item => {
        return (
          !currentBuildVersionForUpgrade ||
          buildVersionCompare(
            `${item.version}-${item.release}`,
            currentBuildVersionForUpgrade,
            'gt'
          )
        );
      });
    } else if (currentObVersion) {
      packageList = packageList.filter(item =>
        versionCompare(item.version, currentObVersion, 'gte')
      );
    }

    // 向组件外部暴露 refresh 属性函数，可通过 ref 引用
    useImperativeHandle(ref, () => ({
      refresh: () => {
        refresh();
      },
    }));

    return (
      <span>
        <MySelect
          loading={loading}
          showSearch={true}
          optionLabelProp="label"
          labelInValue={labelInValue}
          dropdownMatchSelectWidth={false}
          dropdownRender={menu => (
            <SelectDropdownRender
              menu={menu}
              text={formatMessage({
                id: 'ocp-v2.component.UploadPackageDrawer.UploadSoftwarePackage',
                defaultMessage: '上传软件包',
              })}
              onClick={() => {
                setVisible(true);
              }}
            />
          )}
          {...restProps}
        >
          {(isHeterogeneousUpgrade
            ? uniqBy(
                packageList,
                item =>
                  `${item.name}-${item.version}-${item.release_distribution}-${item.distribution}`
              )
            : isObproxyUpgrade
            ? uniqBy(
                packageList,
                item => `${item.name}-${item.version}-${item.release_distribution}`
              )
            : packageList
          )
            // 对文件名按照 ASCII 码进行从小到大排序，方便用户选择
            // TODO 升级排序方式
            .sort((a, b) => {
              if (
                buildVersionCompare(
                  `${a?.version}-${a?.release_distribution}`,
                  `${b?.version}-${b?.release_distribution}`,
                  'gt'
                )
              ) {
                return 1;
              }
              if (
                buildVersionCompare(
                  `${a?.version}-${a?.release_distribution}`,
                  `${b?.version}-${b?.release_distribution}`,
                  'lt'
                )
              ) {
                return -1;
              }
              return 0;
            })

            // .filter(
            //   item =>
            //     !buildVersion ||
            //     buildVersionCompare(`${item.version}-${item.buildNumber}`, buildVersion, 'eq')
            // )

            .map(item => {
              const label = `${item.name}-${item.version}-${item.release_distribution}`;
              // https://aone.alipay.com/issue/101249594
              // 升级增加名字去重
              // 此处在使用时，需要将 name部分去除，再传给接口
              const value = item.pkg_id;

              return (
                <Option key={item.id} value={labelInValue ? value : label} label={label}>
                  <Tooltip title={label}>
                    <Space size={4} style={{ width: '100%' }}>
                      <span>{label}</span>
                    </Space>
                  </Tooltip>
                </Option>
              );
            })}
        </MySelect>
        <UploadPackageDrawer
          open={visible}
          onCancel={() => {
            setVisible(false);
          }}
          onSuccess={() => {
            refresh();
          }}
        />
      </span>
    );
  }
);

export default PackageSelect;
