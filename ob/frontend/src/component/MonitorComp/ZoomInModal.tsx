import { DATE_TIME_FORMAT, DATE_TIME_FORMAT_DISPLAY } from '@/constant/datetime';
import {
  BASE_CHART_COLORS,
  DATE_RANGER_SELECTS,
  EXTEND_CHART_COLORS,
  POINT_NUMBER,
} from '@/constant/monitor';
import { useColorMapping } from '@/hook/useColorMapping';
import useRequestOfMonitor from '@/hook/useRequestOfMonitor';
import { queryMetricsReq } from '@/service';
import { formatMessage } from '@/util/intl';
import { Line } from '@antv/g2plot';
import { Empty, Modal, Spin } from '@oceanbase/design';
import { DateRanger } from '@oceanbase/ui';
import { formatNumber } from '@oceanbase/util';
import { useUpdateEffect } from 'ahooks';
import dayjs from 'dayjs';
import React, { useEffect, useMemo, useRef, useState } from 'react';
import ChartLegend, { LegendDataProps } from './ChartLegend';
import styles from './ZoomInModal.less';

interface ZoomInModalProps {
  visible: boolean;
  onCancel: () => void;
  title: string;
  metrics: API.MetricMeta[];
  labels: Monitor.MetricsLabels;
  queryRange: API.QueryRange;
  groupLabels: Monitor.LabelKeys[];
  type?: Monitor.MonitorUseTarget;
  useFor: Monitor.MonitorUseFor;
  filterData?: any[];
  filterQueryMetric?: Monitor.MetricsLabels;
  activeDimension?: string;
}

const chartHeight = 400;

const ZoomInModal: React.FC<ZoomInModalProps> = ({
  visible,
  onCancel,
  title,
  metrics,
  labels,
  queryRange,
  groupLabels,
  type = 'DETAIL',
  useFor,
  filterData,
  filterQueryMetric,
  activeDimension,
}) => {
  const currentMoment = useRef(dayjs());
  const [isEmpty, setIsEmpty] = useState<boolean>(true);
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const lineGraphRef = useRef<HTMLDivElement>(null);
  const lineInstanceRef = useRef<Line | null>(null);
  const isInitialOpenRef = useRef<boolean>(false);

  // 隐藏的图例列表
  const [hiddenLegends, setHiddenLegends] = useState<string[]>([]);
  // 图例数据
  const [legendData, setLegendData] = useState<Record<string, LegendDataProps>>({});
  // 原始图表数据（未过滤隐藏图例）
  const [originalChartData, setOriginalChartData] = useState<any[]>([]);

  // 弹窗内部的时间范围状态
  const [modalQueryRange, setModalQueryRange] = useState<API.QueryRange>(queryRange);

  // 用于 DateRanger 的受控值
  const dateRangerValue = useMemo(() => {
    if (modalQueryRange.start_timestamp && modalQueryRange.end_timestamp) {
      // queryRange 中的时间戳是秒级的，需要转换为毫秒级
      return [
        dayjs(modalQueryRange.start_timestamp * 1000),
        dayjs(modalQueryRange.end_timestamp * 1000),
      ];
    }
    return undefined;
  }, [modalQueryRange.start_timestamp, modalQueryRange.end_timestamp]);

  // 获取所有指标名称用于颜色映射，使用 useMemo 避免每次渲染都创建新数组
  const metricsName = useMemo(
    () => Array.from(new Set(originalChartData.map(item => item.name))),
    [originalChartData]
  );

  // 使用 hook 生成颜色映射，基于实际数据的指标
  const internalColorMapping = useColorMapping(metricsName);

  const getQueryParams = () => {
    return {
      group_labels: groupLabels,
      labels: type === 'OVERVIEW' ? [] : labels,
      type,
      metrics,
      query_range: modalQueryRange, // 使用弹窗内部的时间范围
      useFor,
      filterData,
      filterQueryMetric,
      activeDimension,
    };
  };

  const lineInstanceRender = (metricsData: any, isLegendToggle = false) => {
    // 按时间戳排序数据,确保图表按时间顺序显示
    const sortedMetricsData = [...metricsData].sort((a, b) => a.date - b.date);
    const timeStampList: number[] = sortedMetricsData.map(item => item.date);
    const minTimestamp = Math.min(...timeStampList);
    const maxTimestamp = Math.max(...timeStampList);
    // 时间范围是否跨天
    const isCrossDay = dayjs(maxTimestamp).diff(dayjs(minTimestamp), 'day') > 0;

    // 过滤掉隐藏的图例数据
    const filterLegendsData = sortedMetricsData.filter(item => !hiddenLegends.includes(item.name));

    // 统计不同的图例数量和名称
    const uniqueNames = Array.from(new Set(sortedMetricsData.map(item => item.name)));
    const legendCount = uniqueNames.length;
    // 图例超过5个时才启用滚动
    const needScroll = legendCount > 17;

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
      height: chartHeight,
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
                maxHeight: '450px',
                overflowY: 'auto',
                overflowX: 'hidden',
                marginTop: '12px',
              },
            }
          : undefined,
        // 设置tooltip偏移量
        offset: needScroll ? 40 : 0,
        // 优化显示效果
        showCrosshairs: true,
        shared: true,
        // 图例超过5个时才允许鼠标进入tooltip
        enterable: needScroll,
      } as any,
      interactions: [{ type: 'marker-active' }],
      animation: isLegendToggle ? false : undefined,
    };

    // 确保 DOM 元素已挂载
    if (!lineGraphRef.current) {
      return;
    }

    // 如果是图例切换且图表实例已存在，仅更新数据
    if (isLegendToggle && lineInstanceRef.current) {
      lineInstanceRef.current.changeData(filterLegendsData);
      return;
    }

    // 销毁旧的图表实例
    if (lineInstanceRef.current) {
      lineInstanceRef.current.destroy();
      lineInstanceRef.current = null;
    }

    // 创建新的图表实例
    lineInstanceRef.current = new Line(lineGraphRef.current, { ...config });
    lineInstanceRef.current.render();
  };

  const lineInstanceDestroy = () => {
    lineInstanceRef.current?.destroy();
    lineInstanceRef.current = null;
    if (isLoading) setIsLoading(false);
    if (!isEmpty) setIsEmpty(true);
  };

  // 获取监控数据
  const { data: metricsData, run: queryMetrics } = useRequestOfMonitor(queryMetricsReq, {
    isRealTime: false,
    manual: true,
    onSuccess: (metricsData: any) => {
      if (!metricsData || !metricsData.length) {
        lineInstanceDestroy();
        return;
      }

      if (metricsData && metricsData.length > 0) {
        setIsEmpty(false);
        setIsLoading(false);

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
      } else {
        lineInstanceDestroy();
      }
    },
    onError: () => {
      lineInstanceDestroy();
    },
  });

  // 当数据准备好且 DOM 已挂载时，渲染图表
  useEffect(() => {
    if (originalChartData.length > 0 && lineGraphRef.current && !isEmpty) {
      lineInstanceRender(originalChartData);
    }
  }, [originalChartData, isEmpty]);

  useUpdateEffect(() => {
    if (!isEmpty && metricsData) {
      lineInstanceRender(metricsData, true);
    }
  }, [hiddenLegends]);

  // 当弹窗打开时请求数据
  useEffect(() => {
    if (visible) {
      // 弹窗打开时，重置为外部传入的时间范围
      isInitialOpenRef.current = true;
      setIsLoading(true);
      setIsEmpty(true);
      setModalQueryRange(queryRange);
    } else {
      // 弹窗关闭时重置状态
      setHiddenLegends([]);
      setOriginalChartData([]);
      lineInstanceDestroy();
    }
  }, [visible]);

  // 当弹窗打开且 modalQueryRange 变化时，请求数据
  useUpdateEffect(() => {
    if (visible) {
      // 只有非初次打开时才需要重新查询（初次打开由下面的 effect 处理）
      if (!isInitialOpenRef.current) {
        setIsLoading(true);
      }
      queryMetrics(getQueryParams());
      isInitialOpenRef.current = false;
    }
  }, [modalQueryRange]);

  // 组件卸载时清理
  useEffect(() => {
    return () => {
      lineInstanceDestroy();
    };
  }, []);

  // 点击图例
  const handleLegendClick = (name: string) => {
    setHiddenLegends(prev =>
      prev.includes(name) ? prev.filter(line => line !== name) : [...prev, name]
    );
  };

  // DateRanger 变化时更新时间范围
  const handleDateRangerChange = (dates: any) => {
    if (dates && Array.isArray(dates) && dates.length === 2) {
      const [start, end] = dates;
      if (start && end) {
        setModalQueryRange({
          start_timestamp: Math.floor(start.valueOf() / 1000), // 转换为秒级时间戳
          end_timestamp: Math.floor(end.valueOf() / 1000), // 转换为秒级时间戳
          step: modalQueryRange.step,
        });
      }
    }
  };

  return (
    <Modal
      width="85vw"
      title={<div className={styles.modalTitle}>{title}</div>}
      open={visible}
      destroyOnClose={true}
      footer={null}
      onCancel={() => {
        onCancel?.();
      }}
    >
      <Spin spinning={isLoading}>
        <div
          style={{
            display: 'flex',
            justifyContent: 'flex-end',
            alignItems: 'center',
            marginBottom: 16,
          }}
        >
          <div className={styles.dateRangerWrapper}>
            <DateRanger
              isMoment={false}
              selects={DATE_RANGER_SELECTS}
              hasRewind={true}
              hasForward={true}
              hasSync={true}
              allowClear={false}
              value={dateRangerValue as any}
              disabledDate={date => date.isAfter(currentMoment?.current)}
              format={DATE_TIME_FORMAT_DISPLAY}
              onChange={handleDateRangerChange}
            />
          </div>
        </div>
        {isEmpty ? (
          <Empty
            description={formatMessage({
              id: 'OBShell.component.MonitorComp.ZoomInModal.NoData',
              defaultMessage: '暂无数据',
            })}
            style={{
              height: chartHeight + 110,
              display: 'flex',
              flexDirection: 'column',
              alignItems: 'center',
              justifyContent: 'center',
            }}
          />
        ) : (
          <>
            <div style={{ height: chartHeight, minHeight: chartHeight }} ref={lineGraphRef}></div>
            {/* 自定义图例区域 */}
            <ChartLegend
              colorMapping={internalColorMapping}
              metricsName={metricsName}
              chartData={originalChartData}
              legendData={legendData}
              hiddenLegends={hiddenLegends}
              onClickLegend={handleLegendClick}
            />
          </>
        )}
      </Spin>
    </Modal>
  );
};

export default ZoomInModal;
