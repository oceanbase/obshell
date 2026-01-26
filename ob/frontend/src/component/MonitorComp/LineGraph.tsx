import { DATE_TIME_FORMAT } from '@/constant/datetime';
import { BASE_CHART_COLORS, EXTEND_CHART_COLORS, POINT_NUMBER } from '@/constant/monitor';
import { useColorMapping } from '@/hook/useColorMapping';
import useRequestOfMonitor from '@/hook/useRequestOfMonitor';
import { queryMetricsReq } from '@/service';
import { Line } from '@antv/g2plot';
import { formatNumber } from '@oceanbase/util';
import { useInViewport, useUpdateEffect } from 'ahooks';
import { Empty, Spin } from 'antd';
import dayjs from 'dayjs';
import React, { useEffect, useMemo, useRef, useState } from 'react';
import ChartLegend, { LegendDataProps } from './ChartLegend';

interface LineGraphProps {
  id: string;
  metrics: API.MetricMeta[];
  labels: Monitor.MetricsLabels;
  queryRange: API.QueryRange;
  groupLabels: Monitor.LabelKeys[];
  height?: number;
  isRefresh?: boolean;
  type?: Monitor.MonitorUseTarget;
  useFor: Monitor.MonitorUseFor;
  filterData?: any[];
  filterQueryMetric?: Monitor.MetricsLabels;
  activeDimension?: string;
}

export default function LineGraph({
  id,
  metrics,
  labels,
  queryRange,
  groupLabels,
  height = 110,
  isRefresh = false,
  type = 'DETAIL',
  useFor,
  filterData,
  filterQueryMetric,
  activeDimension,
}: LineGraphProps) {
  const [isEmpty, setIsEmpty] = useState<boolean>(true);
  const [isloading, setIsloading] = useState<boolean>(true);
  const lineGraphRef = useRef(null);
  const lineInstanceRef = useRef<Line | null>(null);
  const [inViewport] = useInViewport(lineGraphRef);
  // The number of times to enter the visible area,
  // only initiate a network request when entering the visible area for the first time
  const [inViewportCount, setInViewportCount] = useState<number>(0);
  const lastRequestRef = useRef<string>('');

  // 隐藏的图例列表
  const [hiddenLegends, setHiddenLegends] = useState<string[]>([]);
  // 图例数据
  const [legendData, setLegendData] = useState<Record<string, LegendDataProps>>({});
  // 原始图表数据（未过滤隐藏图例）
  const [originalChartData, setOriginalChartData] = useState<any[]>([]);

  // 获取所有指标名称用于颜色映射，使用 useMemo 避免每次渲染都创建新数组
  const metricsName = useMemo(
    () => Array.from(new Set(originalChartData.map(item => item.name))),
    [originalChartData]
  );
  const colorMapping = useColorMapping(metricsName);

  /**
   * The overview page only displays the first metrics
   * and passes empty labels to obtain all clusters/tenants.
   *
   * The details page displays all metrics and filters by labels.
   *
   * The tenant page in the cluster details needs labels to filter out which cluster the tenant is in.
   */
  const getQueryParams = () => {
    return {
      group_labels: groupLabels,
      labels: type === 'OVERVIEW' ? [] : labels, // If empty, query all clusters
      type,
      metrics,
      query_range: queryRange,
      useFor,
      filterData,
      filterQueryMetric,
      activeDimension,
    };
  };

  // 防重复请求的函数
  const safeQueryMetrics = () => {
    const params = getQueryParams();
    // 生成请求唯一标识，基于实际参数
    const requestKey = JSON.stringify({
      labels: params.labels,
      query_range: params.query_range,
      metrics: params.metrics,
      type: params.type,
      group_labels: params.group_labels,
      filterQueryMetric: params.filterQueryMetric,
      isRefresh,
      inViewport,
    });

    if (lastRequestRef.current === requestKey) {
      return;
    }

    lastRequestRef.current = requestKey;
    queryMetrics(params);
  };

  const lineInstanceRender = (metricsData: any, isLegendToggle = false) => {
    // 按时间戳排序数据,确保图表按时间顺序显示
    const sortedMetricsData = [...metricsData].sort((a, b) => a.date - b.date);
    const timeStampList: number[] = sortedMetricsData.map(item => item.date);
    const minTimestamp = Math.min(...timeStampList);
    const maxTimestamp = Math.max(...timeStampList);
    // 时间范围是否跨天
    const isCrossDay = dayjs(maxTimestamp).diff(dayjs(minTimestamp), 'day') > 0;

    const values: number[] = [];
    for (const metric of sortedMetricsData) {
      values.push(metric.value);
    }

    // 过滤掉隐藏的图例数据
    const filterLegendsData = sortedMetricsData.filter(item => !hiddenLegends.includes(item.name));

    // 统计不同的图例数量和名称
    const uniqueNames = Array.from(new Set(sortedMetricsData.map(item => item.name)));
    const legendCount = uniqueNames.length;
    // 图例超过5个时才启用滚动
    const needScroll = legendCount > 5;

    // 根据当前数据计算颜色映射（确保首次渲染就有正确的颜色）
    const chartColors = legendCount > 10 ? EXTEND_CHART_COLORS : BASE_CHART_COLORS;
    const currentColorMapping: Record<string, string> = uniqueNames.reduce((acc, cur, index) => {
      acc[cur] = chartColors[index % chartColors.length];
      return acc;
    }, {} as Record<string, string>);

    const config: any = {
      data: filterLegendsData,
      xField: 'date',
      yField: 'value',
      height: height,
      seriesField: 'name',
      color: (datum: any) => currentColorMapping[datum.name] || '#3D88F2',
      legend: false, // 始终隐藏 G2Plot 自带的图例
      xAxis: {
        nice: false,
        tickCount: POINT_NUMBER,
        label: {
          formatter: (text: string) => {
            const time = dayjs
              .unix(Math.ceil(Number(text) / 1000))
              .format(isCrossDay ? 'DD HH:mm' : 'HH:mm');
            return time;
          },
        },
      },
      tooltip: {
        title: (value: string) => {
          return dayjs.unix(Math.ceil(Number(value) / 1000)).format(DATE_TIME_FORMAT);
        },
        // 根据图例数量动态设置样式
        domStyles: needScroll
          ? {
              'g2-tooltip': {
                maxHeight: '100px',
                overflowY: 'auto',
                overflowX: 'hidden',
                marginTop: '12px',
              },
            }
          : undefined,
        // 设置tooltip偏移量,让它离鼠标远一点,方便左右滑动查看数据
        offset: needScroll ? 40 : 0,
        // 优化显示效果
        showCrosshairs: true,
        shared: true,
        // 图例超过5个时才允许鼠标进入tooltip
        enterable: needScroll,
      } as any,
      interactions: [{ type: 'marker-active' }],
      animation: isLegendToggle ? false : undefined, // 图例切换时禁用动画
    };

    // 如果是图例切换且图表实例已存在，使用 changeData 更新数据
    if (isLegendToggle && lineInstanceRef.current) {
      lineInstanceRef.current.changeData(filterLegendsData);
      return;
    }

    // 如果图表实例存在，先销毁它
    if (lineInstanceRef.current) {
      lineInstanceRef.current.destroy();
      lineInstanceRef.current = null;
    }

    // 创建新的图表实例
    lineInstanceRef.current = new Line(id, { ...config });
    lineInstanceRef.current.render();
  };

  const lineInstanceDestroy = () => {
    lineInstanceRef.current?.destroy();
    lineInstanceRef.current = null;
    if (isloading) setIsloading(false);
    if (!isEmpty) setIsEmpty(true);
  };

  // filter metricsData
  const { data: metricsData, run: queryMetrics } = useRequestOfMonitor(queryMetricsReq, {
    isRealTime: isRefresh,
    manual: true,
    onSuccess: (metricsData: any) => {
      if (!metricsData || !metricsData.length) {
        lineInstanceDestroy();
        return;
      }

      if (metricsData && metricsData.length > 0) {
        setIsEmpty(false);
        setIsloading(false);

        // 按时间戳排序数据
        const sortedMetricsData = [...metricsData].sort((a, b) => a.date - b.date);

        // 保存原始图表数据（只在数据请求成功时更新，避免在渲染函数中更新导致循环）
        setOriginalChartData(sortedMetricsData);

        // 计算图例数据
        const legendObject: Record<string, number[]> = {};
        sortedMetricsData.forEach(item => {
          if (!legendObject[item.name]) {
            legendObject[item.name] = [];
          }
          legendObject[item.name].push(item.value);
        });

        const newLegendData: Record<string, LegendDataProps> = Object.keys(legendObject).reduce(
          (pre, cur) => {
            const max = Math.max(...legendObject[cur]);
            const avg =
              legendObject[cur].reduce((acc, val) => acc + val, 0) / legendObject[cur].length;
            const currentItem = sortedMetricsData.filter(item => item.name === cur).pop();
            const current = currentItem?.value || 0;

            return {
              ...pre,
              [cur]: {
                name: cur,
                max: formatNumber(max),
                avg: formatNumber(avg),
                current: formatNumber(current),
              },
            };
          },
          {}
        );
        setLegendData(newLegendData);

        lineInstanceRender(metricsData);
      } else {
        lineInstanceDestroy();
      }
    },
    onError: () => {
      lineInstanceDestroy();
    },
  });

  useUpdateEffect(() => {
    if (!isEmpty && metricsData) {
      lineInstanceRender(metricsData, true); // 传入 true 表示是图例切换
    }
  }, [hiddenLegends]);

  useUpdateEffect(() => {
    if (inViewport) {
      setInViewportCount(inViewportCount + 1);
      // 图表进入可视区域时请求数据
      safeQueryMetrics();
    }
  }, [inViewport]);

  // 当ID变化时，重新检查视口状态
  useUpdateEffect(() => {
    if (inViewport && inViewportCount === 0) {
      setInViewportCount(1);
    }
  }, [id, inViewport]);

  // Process after turning on real-time mode
  useUpdateEffect(() => {
    if (!isRefresh) {
      if (inViewport) {
        safeQueryMetrics();
      } else if (inViewportCount >= 1) {
        setInViewportCount(0);
      }
    } else {
      // 自动刷新模式下，只有在可视区域内才发送请求
      if (inViewport) {
        safeQueryMetrics();
      }
    }
  }, [labels, queryRange, groupLabels, filterQueryMetric, metrics]);

  // 监听ID变化，当tab切换导致ID变化时，重新初始化图表
  useUpdateEffect(() => {
    if (lineInstanceRef.current) {
      lineInstanceDestroy();
    }
    // 重置inViewportCount，确保新tab能触发数据查询
    setInViewportCount(0);
    // 立即触发数据查询，不等待视口检测
    safeQueryMetrics();
  }, [id]);

  // 组件卸载时清理图表实例
  useEffect(() => {
    return () => {
      if (lineInstanceRef.current) {
        lineInstanceDestroy();
      }
      // 重置状态
      setInViewportCount(0);
      setIsEmpty(true);
      setIsloading(true);
    };
  }, []);

  // 点击图例
  const handleLegendClick = (name: string) => {
    setHiddenLegends(prev =>
      prev.includes(name) ? prev.filter(line => line !== name) : [...prev, name]
    );
  };

  return (
    <div>
      <Spin spinning={isloading}>
        {isEmpty ? (
          <div ref={lineGraphRef}>
            <Empty />
          </div>
        ) : (
          <>
            <div style={{ height: 'auto' }} id={id} ref={lineGraphRef}></div>
            {/* 自定义图例区域 */}
            <ChartLegend
              colorMapping={colorMapping}
              metricsName={metricsName}
              chartData={originalChartData}
              legendData={legendData}
              hiddenLegends={hiddenLegends}
              onClickLegend={handleLegendClick}
            />
          </>
        )}
      </Spin>
    </div>
  );
}
