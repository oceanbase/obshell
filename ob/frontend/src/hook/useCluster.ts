import { useSelector } from '@umijs/max';

/**
 * 集群信息钩子
 * 用于获取集群数据并计算相关属性
 */
export function useCluster() {
  const { clusterData } = useSelector((state: DefaultRootState) => state.cluster);

  // 计算是否为单机版
  const isStandalone = clusterData?.is_standalone || false;

  // 获取是否为社区版
  const isCommunityEdition = clusterData?.is_community_edition || false;

  // 获取是否是分布式商业版
  const isDistributedBusiness = !isStandalone && !isCommunityEdition;

  return {
    clusterData,
    isStandalone,
    isDistributedBusiness,
    isCommunityEdition,
  };
}
