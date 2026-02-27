import { formatMessage } from '@/util/intl';
import React, { useEffect } from 'react';
import G6 from '@antv/g6';
import { Space, theme } from '@oceanbase/design';
import { isNullValue } from '@oceanbase/util';
import './registerNode';
import styles from './DeadlockGraph.less';

export interface DeadlockGraphProps {
  id?: string;
  nodeList?: API.DeadLockNode[];
  relationList?: API.ObSessionLockWaitRelation[];
}

const DeadlockGraph: React.FC<DeadlockGraphProps> = ({ id, nodeList, relationList }) => {
  const { token } = theme.useToken();
  const containerId = `ocp-deadlock-graph-${id}`;

  let nodes = [];
  let edges = [];

  // 根据传入的不同数据源，分别构造图节点和边
  if (nodeList) {
    nodes = nodeList?.map(item => ({
      ...item,
      id: `${item.idx}`,
      label: `${item.idx}`,
    }));

    nodes.forEach((item, index) => {
      // 终点 -> 起点
      if (index === nodes.length - 1) {
        edges.push({
          source: `${item.idx}`,
          target: `${nodes?.[0].idx}`,
        });
      } else {
        // 上一个点 -> 下一个点
        edges.push({
          source: `${item.idx}`,
          target: `${nodes?.[index + 1].idx}`,
        });
      }
    });
  } else if (relationList) {
    nodes = relationList?.map(item => ({
      ...item,
      id: `${item.lockRequesterId}`,
      label: `${item.lockRequesterId}`,
    }));

    nodes.forEach(item => {
      // 请求会话 ID -> 持有会话 ID
      edges.push({
        source: `${item.lockRequesterId}`,
        target: `${item.lockOwnerId}`,
      });
    });
  }

  // 环形布局 + 直线边，两个节点默认不会成环，需要手动设置连接点让其成环
  if (nodes.length === 2) {
    nodes = nodes.map((node, index) => {
      return {
        ...node,
        anchorPoints:
          // 设置节点的连接点
          index === 0
            ? [
                [0, 0.25],
                [0, 0.75],
              ]
            : [
                [1, 0.25],
                [1, 0.75],
              ],
      };
    });
    edges = edges.map((edge, index) => {
      return {
        ...edge,
        // 顺时针成环
        sourceAnchor: index === 0 ? 1 : 0,
        targetAnchor: index === 0 ? 1 : 0,
      };
    });
  }

  const data = {
    nodes,
    edges,
  };

  useEffect(() => {
    const container = document.getElementById(containerId);
    const width = container?.clientWidth || 0;
    const height = container?.clientHeight || 0;

    const graph = new G6.Graph({
      container: containerId,
      width,
      height,
      fitView: true,
      layout: {
        // 环形布局
        type: 'circular',
        // 圆半径: 节点数为 2，则圆半径为 250；节点数 > 2，则圆半径随节点数增加而扩大
        // 目的都是为了避免节点挤压到一起，影响环图展示
        radius: nodes.length === 2 ? 250 : nodes.length * 40,
      },

      defaultNode: {
        // 使用自定义节点
        type: 'deadlockNode',
        style: {
          fill: token.colorInfoBg,
          stroke: token.colorInfoBorder,
          lineWidth: 1,
          radius: token.borderRadius,
        },
      },

      defaultEdge: {
        // 使用内置直线边
        type: 'line',
        style: {
          stroke: '#5F95FF',
          lineWidth: 2,
          // 设置箭头
          endArrow: {
            path: 'M 0,0 L 8,3 L 8,-3 Z',
            fill: '#5F95FF',
          },
        },
      },

      nodeStateStyles: {
        // hover 样式
        hover: {
          fill: token.colorBgContainer,
          fillOpacity: 0.05,
          lineWidth: 2,
          shadowColor: 'rgba(95,149,255,0.2)',
          shadowBlur: 10,
        },

        // 带有回滚状态的 hover 样式
        rollBackedHover: {
          fill: token.colorBgContainer,
          lineWidth: 2,
          shadowColor: 'rgba(245,34,45,0.25)',
          shadowBlur: 10,
        },
      },

      modes: {
        default: [
          {
            type: 'tooltip',
            formatText: (model: any) => {
              const lockRequesterIdLabel = formatMessage(
                {
                  id: 'OBShell.Session.Deadlock.DeadlockGraph.SessionIdModellockrequesterid',
                  defaultMessage: '会话 ID：{modelLockRequesterId}',
                },
                { modelLockRequesterId: model.lockRequesterId }
              );

              const rowKeyLabel = `RowKey：${model.rowKey}`;

              const idxLabel = formatMessage(
                {
                  id: 'OBShell.Session.Deadlock.DeadlockGraph.NumberIdModelidx',
                  defaultMessage: '编号 ID：{modelIdx}',
                },
                { modelIdx: model.idx }
              );

              const resourceLabel = formatMessage(
                {
                  id: 'OBShell.Session.Deadlock.DeadlockGraph.RequestResourcesModelresource',
                  defaultMessage: '请求资源：{modelResource}',
                },
                { modelResource: model.resource }
              );

              return !isNullValue(model.lockRequesterId)
                ? `
              <div>
                <div>${lockRequesterIdLabel}</div>
                <div>${rowKeyLabel}</div>
              </div>`
                : `
                <div>
                  <div>${idxLabel}</div>
                  <div>${resourceLabel}</div>
                </div>`;
            },
          },
        ],
      },
    });

    graph.data(data);
    graph.render();

    // 修改 tooltip 的样式，确保层级足够高（只设置 z-index，保持原有的 position）
    setTimeout(() => {
      const tooltipElements = document.querySelectorAll('.g6-component-tooltip, .g6-tooltip');
      tooltipElements.forEach((el: Element) => {
        (el as HTMLElement).style.zIndex = '100000';
      });
    }, 100);

    // 进入节点
    graph.on('node:mouseenter', (e: any) => {
      const item = e.item;
      const model = item.getModel();
      graph.setItemState(e.item, model.roll_backed ? 'rollBackedHover' : 'hover', true);

      // 每次 tooltip 显示时，确保其层级正确（只设置 z-index）
      setTimeout(() => {
        const tooltipElements = document.querySelectorAll('.g6-component-tooltip, .g6-tooltip');
        tooltipElements.forEach((el: Element) => {
          (el as HTMLElement).style.zIndex = '100000';
        });
      }, 10);
    });
    // 离开节点
    graph.on('node:mouseleave', (e: any) => {
      const item = e.item;
      const model = item.getModel();
      graph.setItemState(e.item, model.roll_backed ? 'rollBackedHover' : 'hover', false);
    });
  }, []);

  return (
    <>
      {/* 仅死锁历史存在中断节点，才需要展示图例 */}
      {nodeList && (
        <Space size={16}>
          <div
            style={{
              display: 'flex',
              justifyContent: 'center',
              alignItems: 'center',
            }}
          >
            <span
              style={{
                width: 16,
                height: 16,
                marginRight: 4,
                display: 'inline-block',
                background: token.colorInfoBg,
                border: `1px solid ${token.colorInfoBorder}`,
                borderRadius: token.borderRadiusSM,
              }}
            />

            <span style={{ fontSize: '12px' }}>
              {formatMessage({
                id: 'OBShell.Session.Deadlock.DeadlockGraph.Uninterrupted',
                defaultMessage: '未中断',
              })}
            </span>
          </div>
          <div
            style={{
              display: 'flex',
              justifyContent: 'center',
              alignItems: 'center',
            }}
          >
            <span
              style={{
                width: 16,
                height: 16,
                marginRight: 4,
                display: 'inline-block',
                background: token.colorErrorBg,
                border: `1px solid ${token.colorErrorBorder}`,
                borderRadius: token.borderRadiusSM,
              }}
            />

            <span style={{ fontSize: '12px' }}>
              {formatMessage({
                id: 'OBShell.Session.Deadlock.DeadlockGraph.Interrupted',
                defaultMessage: '已中断',
              })}
            </span>
          </div>
        </Space>
      )}
      <div id={containerId} style={{ height: '140%' }} className={styles.graphContainer} />
    </>
  );
};

export default DeadlockGraph;
