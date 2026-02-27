import { useSelector } from '@umijs/max';

/**
 * 获取 seekdbInfoData 和加载状态的自定义 Hook
 * @returns {object} 包含 seekdbInfoData 和 isSeekdbInfoDataLoaded 的对象
 */
const useSeekdbInfo = () => {
  const { seekdbInfoData } = useSelector((state: DefaultRootState) => state.global);

  // 判断 seekdbInfoData 是否已加载完成
  // 如果 seekdbInfoData 为空对象或 undefined，说明数据还未加载
  const isSeekdbInfoDataLoaded = seekdbInfoData && Object.keys(seekdbInfoData).length > 0;

  return {
    seekdbInfoData,
    isSeekdbInfoDataLoaded,
  };
};

export default useSeekdbInfo;
