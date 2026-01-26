/*
 * Copyright (c) 2024 OceanBase.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

import { BASE_CHART_COLORS, EXTEND_CHART_COLORS } from '@/constant/monitor';
import { useMemo } from 'react';

/**
 * 根据指标名称数组生成颜色映射
 * @param defaultMetricsName 指标名字数组
 * @returns 指标名称到颜色的映射对象
 */
export const useColorMapping = (defaultMetricsName: string[]) => {
  const colorMapping: Record<string, string> = useMemo(() => {
    // 图例数量 <= 10 个时，使用基础色板，否则使用扩展色板
    const chartColors = defaultMetricsName.length > 10 ? EXTEND_CHART_COLORS : BASE_CHART_COLORS;

    return defaultMetricsName.reduce((acc, cur, index) => {
      acc[cur] = chartColors[index % chartColors.length];
      return acc;
    }, {} as Record<string, string>);
  }, [defaultMetricsName]);
  return colorMapping;
};
