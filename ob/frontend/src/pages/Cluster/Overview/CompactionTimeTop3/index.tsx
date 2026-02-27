import { formatMessage } from '@/util/intl';
import React from 'react';
import { Empty, theme as designTheme } from '@oceanbase/design';
import { orderBy } from 'lodash';
import { useRequest } from 'ahooks';
import MyCard from '@/component/MyCard';
import ContentWithQuestion from '@/component/ContentWithQuestion';
import { formatDuration } from '@/util';
import { getTenantTopCompaction } from '@/service/obshell/tenant';
import MyProgress from '@/component/MyProgress';
import styles from './index.less';

const CompactionTimeTop3: React.FC = () => {
  const { token } = designTheme.useToken();

  // 获取合并时间 Top3 的租户合并数据
  const { data: topCompactionListData, loading } = useRequest(() =>
    getTenantTopCompaction(
      {
        limit: '3',
      },
      {
        HIDE_ERROR_MESSAGE: true,
      }
    )
  );

  const topCompactionList: API.TenantCompactionHistory[] =
    orderBy(topCompactionListData?.data?.contents || [], ['cost_time'], ['desc']) || [];

  return (
    <MyCard
      loading={loading}
      title={
        <ContentWithQuestion
          content={formatMessage({
            id: 'ocp-express.Component.CompactionTimeTop3.TenantMergeTimeTop',
            defaultMessage: '租户合并时间 Top3',
          })}
          tooltip={{
            title: (
              <div>
                <div>
                  {formatMessage({
                    id: 'ocp-express.Component.CompactionTimeTop3.SortByLastMergeTime',
                    defaultMessage: '按最近 1 次合并时间排序',
                  })}
                </div>
              </div>
            ),
          }}
        />
      }
      style={{
        height: 156,
      }}
    >
      {topCompactionList.length > 0 ? (
        <div>
          {topCompactionList.map(item => {
            const durationData = formatDuration(item?.cost_time, 1);

            return (
              <MyProgress
                wrapperClassName={styles.progressWrapper}
                showInfo={false}
                prefix={item.tenant_name}
                prefixWidth={80}
                strokeColor={token.colorTextSecondary}
                affix={
                  <>
                    <span style={{ fontSize: 14, color: token.colorText, marginRight: 2 }}>
                      {durationData.value}
                    </span>
                    <span style={{ color: token.colorTextTertiary }}>{durationData.unit}</span>
                  </>
                }
                affixWidth={50}
                percent={durationData.value}
                key={item.tenant_name}
              />
            );
          })}
        </div>
      ) : (
        <Empty imageStyle={{ height: 70 }} description="" />
      )}
    </MyCard>
  );
};

export default CompactionTimeTop3;
