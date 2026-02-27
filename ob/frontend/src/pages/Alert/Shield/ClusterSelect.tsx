import { formatMessage } from '@/util/intl';
import { Col, Row, Select } from '@oceanbase/design';
import { useModel } from '@umijs/max';
import React, { useEffect, useRef, useState } from 'react';
import { getSelectList, isListSame, OptionsType } from '../helper';

type ClusterType = {
  obcluster?: string[];
};

interface ClusterSelectProps {
  onChange?: (value: ClusterType) => void;
  value?: ClusterType;
}

export default function ClusterSelect({ onChange, value = {} }: ClusterSelectProps) {
  const { clusterList = [] } = useModel('alarm');
  const [maxCount, setMaxCount] = useState<number>();
  const [clusterOptions, setClusterOptions] = useState<OptionsType[]>([]);
  const preOBClusterRef = useRef(value.obcluster);

  const getClusterDefaultOptions = () => {
    const list: string[] = getSelectList(clusterList, 'obcluster') as string[];
    const res = list?.map(clusterName => ({
      value: clusterName,
      label: clusterName,
    }));
    // if (res.length) {
    //   res.splice(0, 0, {
    //     value: 'allClusters',
    //     label: '全部集群',
    //   });
    // }
    return res;
  };
  const clusterChange = (selectedCluster: string[]) => {
    onChange?.({ ...value, obcluster: selectedCluster });
  };
  useEffect(() => {
    if (value?.obcluster?.length && value.obcluster.includes('allClusters')) {
      if (!isListSame(preOBClusterRef.current || [], value.obcluster)) {
        onChange?.({ ...value, obcluster: ['allClusters'] });
        setMaxCount(1);
      }
    } else {
      maxCount === 1 && setMaxCount(undefined);
      setClusterOptions(getClusterDefaultOptions());
      setMaxCount(undefined);
    }
    preOBClusterRef.current = value.obcluster;
  }, [value.obcluster]);

  // 设置默认值为第一个集群
  useEffect(() => {
    if (clusterList.length > 0 && (!value?.obcluster || value.obcluster.length === 0)) {
      const firstCluster = clusterList[0]?.cluster_name;
      if (firstCluster) {
        onChange?.({ ...value, obcluster: [firstCluster] });
      }
    }
  }, [clusterList, value?.obcluster]);
  return (
    <Row>
      <Col span={24}>
        <Select
          mode="multiple"
          maxCount={maxCount}
          value={value.obcluster}
          allowClear
          options={clusterOptions}
          onChange={clusterChange}
          placeholder={formatMessage({
            id: 'OBShell.Alert.Shield.ClusterSelect.PleaseSelectACluster',
            defaultMessage: '请选择集群',
          })}
        />
      </Col>
    </Row>
  );
}
