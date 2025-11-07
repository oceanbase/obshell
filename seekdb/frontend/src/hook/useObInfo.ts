import { useSelector } from 'umi';

/**
 * 获取 obInfoData 和加载状态的自定义 Hook
 * @returns {object} 包含 obInfoData 和 isObInfoDataLoaded 的对象
 */
const useObInfo = () => {
  const { obInfoData } = useSelector((state: DefaultRootState) => state.global);

  // 判断 obInfoData 是否已加载完成
  // 如果 obInfoData 为空对象或 undefined，说明数据还未加载
  const isObInfoDataLoaded = obInfoData && Object.keys(obInfoData).length > 0;

  return {
    obInfoData,
    isObInfoDataLoaded,
  };
};

export default useObInfo;
