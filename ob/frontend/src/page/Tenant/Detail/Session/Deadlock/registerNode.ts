import { formatMessage } from '@/util/intl';
import { getFullPath } from '@/util/global';
import { fittingString } from '@/util/graph';
import G6 from '@antv/g6';
import { token } from '@oceanbase/design';
import { isNullValue } from '@oceanbase/util';

// 自定义死锁节点
G6.registerNode(
  'deadlockNode',
  {
    // options 不能为空，否则 edge 绘制有问题，原因待排查
    options: {},
    draw(cfg, group) {
      // 矩形
      const keyShape = group?.addShape('rect', {
        attrs: {
          width: 200,
          height: 60,
          fill: cfg?.roll_backed ? token.colorErrorBg : token.colorInfoBg,
          stroke: cfg?.roll_backed ? token.colorErrorBorder : token.colorInfoBorder,
          lineWidth: 1,
          radius: token.borderRadius,
        },
      });

      if (!isNullValue(cfg?.lockRequesterId)) {
        // 会话 ID
        group?.addShape('text', {
          attrs: {
            x: 40,
            y: 25,
            fill: token.colorTextSecondary,
            fontSize: 12,
            fontWeight: 500,
            text:
              cfg &&
              formatMessage(
                {
                  id: 'ocp-v2.Session.Deadlock.DeadlockGraph.SessionIdCfglockrequesterid',
                  defaultMessage: '会话 ID：{cfgLockRequesterId}',
                },
                { cfgLockRequesterId: cfg.lockRequesterId }
              ),
          },
        });
        // RowKey
        group?.addShape('text', {
          attrs: {
            x: 40,
            y: 45,
            fill: token.colorTextTertiary,
            fontSize: 12,
            // 长文本截断
            text:
              cfg &&
              fittingString(
                formatMessage(
                  {
                    id: 'ocp-v2.Session.Deadlock.DeadlockGraph.RowkeyCfgrowkey',
                    defaultMessage: 'RowKey：{cfgRowKey}',
                  },
                  { cfgRowKey: cfg.rowKey }
                ),
                160,
                12
              ),
          },
        });
      } else {
        // 编号 ID
        group?.addShape('text', {
          attrs: {
            x: 40,
            y: 25,
            fill: token.colorTextSecondary,
            text:
              cfg &&
              formatMessage(
                {
                  id: 'ocp-v2.Session.Deadlock.DeadlockGraph.IdCfgidx',
                  defaultMessage: '编号 ID：{cfgIdx}',
                },
                { cfgIdx: cfg.idx }
              ),
            fontSize: 12,
            fontWeight: 500,
          },
        });

        // 请求资源
        group?.addShape('text', {
          attrs: {
            x: 40,
            y: 45,
            fill: token.colorTextTertiary,
            fontSize: 12,
            // 长文本截断
            text:
              cfg &&
              fittingString(
                formatMessage(
                  {
                    id: 'ocp-v2.Session.Deadlock.DeadlockGraph.RequestResourceCfgresource',
                    defaultMessage: '请求资源：{cfgResource}',
                  },
                  { cfgResource: cfg.resource }
                ),
                160,
                12
              ),
          },
        });
      }
      // 锁 icon
      group?.addShape('image', {
        attrs: {
          x: 15,
          y: 12,
          width: 12,
          height: 14,
          img: getFullPath('/assets/tenant/deadlock.svg'),
        },
      });

      return keyShape;
    },
  },

  'rect'
);
