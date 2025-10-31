import { POINT_NUMBER } from '@/constant/monitor';
import { DATE_TIME_FORMAT } from '@/constant/datetime';
import useRequestOfMonitor from '@/hook/useRequestOfMonitor';
import { Line } from '@antv/g2plot';
import { useInViewport, useUpdateEffect } from 'ahooks';
import { Empty, Spin } from 'antd';
import dayjs from 'dayjs';
import React, { useEffect, useRef, useState } from 'react';
import { queryMetricsReq } from '@/service';

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
}

export default function LineGraph({
  id,
  metrics,
  labels,
  queryRange,
  groupLabels,
  height = 186,
  isRefresh = false,
  type = 'DETAIL',
  useFor,
  filterData,
  filterQueryMetric,
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

  const lineInstanceRender = (metricsData: any) => {
    const values: number[] = [];
    for (const metric of metricsData) {
      values.push(metric.value);
    }

    const config = {
      data: metricsData,
      xField: 'date',
      yField: 'value',
      height: height,
      seriesField: 'name',
      xAxis: {
        nice: false,
        tickCount: POINT_NUMBER,
        label: {
          formatter: (text: number) => {
            const time = dayjs.unix(Math.ceil(text / 1000)).format('HH:mm');
            return time;
          },
        },
      },
      tooltip: {
        title: (value: number) => {
          return dayjs.unix(Math.ceil(value / 1000)).format(DATE_TIME_FORMAT);
        },
      },
      interactions: [{ type: 'marker-active' }],
    };

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
    onSuccess: metricsData => {
      if (!metricsData || !metricsData.length) {
        lineInstanceDestroy();
        return;
      }

      if (metricsData && metricsData.length > 0) {
        setIsEmpty(false);
        setIsloading(false);
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
    if (!isEmpty) {
      lineInstanceRender(metricsData);
    }
  }, [isEmpty]);

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
  }, [labels, queryRange, groupLabels, filterQueryMetric]);

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

  return (
    <div style={{ height: `${height}px` }}>
      <Spin spinning={isloading}>
        {isEmpty ? (
          <div ref={lineGraphRef}>
            <Empty />
          </div>
        ) : (
          <div id={id} ref={lineGraphRef}></div>
        )}
      </Spin>
    </div>
  );
}
