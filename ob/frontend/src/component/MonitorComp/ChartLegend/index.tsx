import { Typography } from '@oceanbase/design';
import { CaretDownOutlined, CaretUpOutlined } from '@oceanbase/icons';
import React, { useMemo, useState } from 'react';
import styles from './index.less';

const { Text } = Typography;

export interface LegendDataProps {
  name: string;
  max: number | string;
  avg: number | string;
  current: number | string;
}

interface ChartLegendProps {
  colorMapping: Record<string, string>;
  metricsName: string[];
  chartData: any[];
  legendData: Record<string, LegendDataProps>;
  hiddenLegends: string[];
  onClickLegend: (name: string) => void;
  enableSort?: boolean; // 是否启用排序功能
}

const ChartLegend: React.FC<ChartLegendProps> = ({
  colorMapping, // 颜色映射
  metricsName, // 指标名称，用来对legendData进行排序
  chartData, // 图表数据
  legendData, // 图例数据
  hiddenLegends, // 隐藏的图例
  onClickLegend, // 点击图例回调
  enableSort = false, // 是否启用排序功能
}) => {
  const [sortField, setSortField] = useState<'max' | 'avg' | 'current' | null>(null);
  const [sortOrder, setSortOrder] = useState<'asc' | 'desc'>('desc');

  // 处理列头点击排序
  const handleSort = (field: 'max' | 'avg' | 'current') => {
    if (sortField === field) {
      // 如果点击的是当前排序列，切换排序顺序
      setSortOrder(sortOrder === 'asc' ? 'desc' : 'asc');
    } else {
      // 如果点击的是新列，设置新的排序列，默认降序
      setSortField(field);
      setSortOrder('desc');
    }
  };

  // 从 chartData 中计算原始数值（用于排序）
  const getRawValue = (name: string, field: 'max' | 'avg' | 'current'): number => {
    if (!chartData || chartData.length === 0) return 0;

    // 获取该指标的所有数值
    const values = chartData
      .filter(item => {
        // 根据 chartData 的结构，可能是 item.name
        const itemName = item.name || '';
        return itemName === name;
      })
      .map(item => item.value)
      .filter((val): val is number => typeof val === 'number' && !isNaN(val));

    if (values.length === 0) return 0;

    switch (field) {
      case 'max':
        return Math.max(...values);
      case 'avg': {
        const sum = values.reduce((acc, val) => acc + val, 0);
        return sum / values.length;
      }
      case 'current': {
        // 获取最新的值（按时间戳排序后的最后一个）
        const sortedByTime = [...chartData]
          .filter(item => {
            const itemName = item.name || '';
            return itemName === name;
          })
          .sort((a, b) => (a.date || 0) - (b.date || 0));
        const lastItem = sortedByTime[sortedByTime.length - 1];
        return lastItem?.value || 0;
      }
      default:
        return 0;
    }
  };

  // 缓存所有指标的原始值，只在 chartData 变化时重新计算
  const rawValueCache = useMemo<
    Record<string, { max: number; avg: number; current: number }>
  >(() => {
    const cache: Record<string, { max: number; avg: number; current: number }> = {};
    const keys = Object.keys(legendData);

    keys.forEach(name => {
      cache[name] = {
        max: getRawValue(name, 'max'),
        avg: getRawValue(name, 'avg'),
        current: getRawValue(name, 'current'),
      };
    });

    return cache;
  }, [chartData, legendData]);

  // 获取排序后的数据
  const getSortedKeys = () => {
    const keys = Object.keys(legendData);
    // 如果未启用排序功能，直接按 metricsName 顺序排序
    if (!enableSort) {
      return keys.sort((a, b) => {
        const indexA = metricsName.indexOf(a);
        const indexB = metricsName.indexOf(b);
        // 如果不在 metricsName 中，排在后面
        if (indexA === -1 && indexB === -1) return 0;
        if (indexA === -1) return 1;
        if (indexB === -1) return -1;
        return indexA - indexB;
      });
    }
    // 启用排序功能时
    if (!sortField) {
      // 如果没有选择排序列，按 metricsName 顺序排序
      return keys.sort((a, b) => {
        const indexA = metricsName.indexOf(a);
        const indexB = metricsName.indexOf(b);
        // 如果不在 metricsName 中，排在后面
        if (indexA === -1 && indexB === -1) return 0;
        if (indexA === -1) return 1;
        if (indexB === -1) return -1;
        return indexA - indexB;
      });
    }
    // 按选中的列排序，使用缓存的原始数值
    return keys.sort((a, b) => {
      const valueA = rawValueCache[a]?.[sortField] ?? 0;
      const valueB = rawValueCache[b]?.[sortField] ?? 0;
      const diff = valueA - valueB;
      return sortOrder === 'asc' ? diff : -diff;
    });
  };

  // 渲染排序图标
  const renderSortIcon = (field: 'max' | 'avg' | 'current') => {
    if (sortField !== field) {
      return null;
    }
    return sortOrder === 'asc' ? (
      <CaretUpOutlined className={styles.sortIcon} />
    ) : (
      <CaretDownOutlined className={styles.sortIcon} />
    );
  };

  return (
    <>
      {!chartData || chartData.length === 0 ? null : (
        <div className={styles.legendContainer}>
          {/* 将table的head和body分离，超出容器时，只会在body区域出现滚动条 */}
          <div className={styles.legendTableHead}>
            <table className={styles.legendTable}>
              <thead>
                <tr>
                  <th></th>
                  <th
                    style={{ width: 80 }}
                    className={enableSort ? styles.sortableHeader : ''}
                    onClick={enableSort ? () => handleSort('max') : undefined}
                  >
                    <span className={enableSort ? styles.headerContent : ''}>
                      max
                      {enableSort && renderSortIcon('max')}
                    </span>
                  </th>
                  <th
                    style={{ width: 80 }}
                    className={enableSort ? styles.sortableHeader : ''}
                    onClick={enableSort ? () => handleSort('avg') : undefined}
                  >
                    <span className={enableSort ? styles.headerContent : ''}>
                      avg
                      {enableSort && renderSortIcon('avg')}
                    </span>
                  </th>
                  <th
                    style={{ width: 80 }}
                    className={enableSort ? styles.sortableHeader : ''}
                    onClick={enableSort ? () => handleSort('current') : undefined}
                  >
                    <span className={enableSort ? styles.headerContent : ''}>
                      current
                      {enableSort && renderSortIcon('current')}
                    </span>
                  </th>
                </tr>
              </thead>
            </table>
          </div>
          <div className={styles.legendTableBody}>
            <table className={styles.legendTable}>
              <tbody>
                {getSortedKeys().map(item => {
                  const isHidden = hiddenLegends.includes(item);
                  return (
                    <tr key={item}>
                      {/* 图例 Name 列 */}
                      <td
                        style={{
                          color: isHidden ? '#bfbfbf' : 'rgba(0, 0, 0, 0.65)',
                        }}
                        onClick={() => onClickLegend(item)}
                      >
                        {/* 图例颜色条 */}
                        <div
                          className={styles.stick}
                          style={{
                            backgroundColor: isHidden ? '#bfbfbf' : colorMapping[item], // 颜色条，灰色表示隐藏状态
                          }}
                        />
                        <Text
                          style={{ fontSize: 12 }}
                          ellipsis={{
                            tooltip: {
                              overlayInnerStyle: {
                                whiteSpace: 'nowrap',
                              },
                              overlayStyle: {
                                maxWidth: 'none',
                                overflow: 'visible',
                              },
                            },
                          }}
                        >
                          {item}
                        </Text>
                      </td>

                      {/* 图例 Max 列 */}
                      <td
                        style={{
                          width: 80,
                          color: isHidden ? '#bfbfbf' : 'rgba(0, 0, 0, 0.85)',
                        }}
                      >
                        {legendData?.[item]?.max || 0}
                      </td>

                      {/* 图例 Avg 列 */}
                      <td
                        style={{
                          width: 80,
                          color: isHidden ? '#bfbfbf' : 'rgba(0, 0, 0, 0.85)',
                        }}
                      >
                        {legendData?.[item]?.avg || 0}
                      </td>

                      {/* 图例 Current 列 */}
                      <td
                        style={{
                          width: 80,
                          color: isHidden ? '#bfbfbf' : 'rgba(0, 0, 0, 0.85)',
                        }}
                      >
                        {legendData?.[item]?.current || 0}
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
        </div>
      )}
    </>
  );
};

export default ChartLegend;
