import { formatMessage } from '@/util/intl';
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

import { Card, Table, Tag } from '@oceanbase/design';
import React from 'react';
import SyncStatusBadge from './SyncStatusBadge';

export interface StandbyRelationCardProps {
  standbyLocal?: API.LocalStandbyStatus;
  standbyPeers: API.PeerStandbyStatus[];
  standbyRole?: string;
  instanceName?: string;
}

const StandbyRelationCard: React.FC<StandbyRelationCardProps> = ({
  standbyLocal,
  standbyPeers,
  standbyRole,
  instanceName,
}) => {
  const localName = standbyLocal?.instance_name || instanceName || '-';

  const currentInstanceTag = (name: string) => (
    <span>
      {name}
      <Tag style={{ marginLeft: 8 }}>
        {formatMessage({
          id: 'seekdb.Overview.components.StandbyRelationCard.CurrentInstance',
          defaultMessage: '当前实例',
        })}
      </Tag>
    </span>
  );

  const peerLink = (record: API.PeerStandbyStatus) => {
    const peerName =
      record.instance_name || `${record.peer_host || '-'}:${record.peer_obshell_port || '-'}`;
    return peerName;
  };

  /** 主实例列：连续相同展示内容的行合并（如 PRIMARY 下多行均为当前实例） */
  const primaryRowSpans = React.useMemo(() => {
    const peers = standbyPeers;
    const getMergeKey = (r: API.PeerStandbyStatus) => {
      if (standbyRole === 'PRIMARY') {
        return `primary:${localName}`;
      }
      if (r.direction === 'UPSTREAM') {
        return `upstream:${r.instance_name || ''}:${r.peer_host || ''}:${
          r.peer_obshell_port ?? ''
        }`;
      }
      return `local:${localName}`;
    };
    const spans: number[] = [];
    for (let i = 0; i < peers.length; i++) {
      const key = getMergeKey(peers[i]);
      if (i > 0 && getMergeKey(peers[i - 1]) === key) {
        spans.push(0);
        continue;
      }
      let span = 1;
      for (let j = i + 1; j < peers.length; j++) {
        if (getMergeKey(peers[j]) === key) {
          span += 1;
        } else {
          break;
        }
      }
      spans.push(span);
    }
    return spans;
  }, [standbyPeers, standbyRole, localName]);

  const columns = [
    {
      title: formatMessage({
        id: 'seekdb.Overview.components.StandbyRelationCard.PrimaryInstance',
        defaultMessage: '主实例',
      }),
      dataIndex: 'primaryName',
      onCell: (_: API.PeerStandbyStatus, index?: number) => ({
        rowSpan: primaryRowSpans[index ?? 0] ?? 1,
      }),
      render: (_: any, record: API.PeerStandbyStatus) => {
        if (standbyRole === 'PRIMARY') {
          return currentInstanceTag(localName);
        }
        if (record.direction === 'UPSTREAM') {
          return peerLink(record);
        }
        return currentInstanceTag(localName);
      },
    },
    {
      title: formatMessage({
        id: 'seekdb.Overview.components.StandbyRelationCard.Instance',
        defaultMessage: '备实例',
      }),
      dataIndex: 'standbyName',
      // rowSpan 合并主租户列后，后续行第一个 td 是本列；避免命中 card-without-padding 的 td:first-child 24px
      onCell: () => ({ className: 'table-subrow-first-cell' }),
      render: (_: any, record: API.PeerStandbyStatus) => {
        if (standbyRole === 'PRIMARY') {
          return peerLink(record);
        }
        if (record.direction === 'UPSTREAM') {
          return currentInstanceTag(localName);
        }
        return peerLink(record);
      },
    },
    {
      title: formatMessage({
        id: 'seekdb.Overview.components.StandbyRelationCard.Status',
        defaultMessage: '状态',
      }),
      dataIndex: 'sync_status',
      width: 150,
      render: (value: string, record: API.PeerStandbyStatus) => {
        const displayStatus = record.direction === 'UPSTREAM' ? standbyLocal?.sync_status : value;
        return <SyncStatusBadge status={displayStatus} />;
      },
    },
    {
      title: formatMessage({
        id: 'seekdb.Overview.components.StandbyRelationCard.LogLatency',
        defaultMessage: '日志延迟',
      }),
      dataIndex: 'lag_seconds',
      width: 140,
      render: (value: number, record: API.PeerStandbyStatus) => {
        const displayStatus = record.direction === 'UPSTREAM' ? standbyLocal?.sync_status : record.sync_status;
        if (displayStatus === 'UNKNOWN') {
          return '-';
        }
        const lagValue =
          record.direction === 'DOWNSTREAM'
            ? value ?? '-'
            : standbyLocal?.sync_scn != null
            ? value ?? '-'
            : '-';
        if (lagValue === '-') return '-';
        return formatMessage(
          {
            id: 'seekdb.Overview.components.StandbyRelationCard.LagvalueSeconds',
            defaultMessage: '{lagValue} 秒',
          },
          { lagValue: lagValue }
        );
      },
    },
  ];

  return (
    <Card
      className="card-without-padding"
      bordered={false}
      title={formatMessage({
        id: 'seekdb.Overview.components.StandbyRelationCard.PrimaryStandbyRelationship',
        defaultMessage: '主备关系',
      })}
    >
      <Table
        bordered
        className="table-without-outside-border primary-backup-tenant-list-table"
        rowKey={(record: API.PeerStandbyStatus) =>
          `${record.peer_host || '-'}-${record.peer_obshell_port || '-'}`
        }
        columns={columns}
        dataSource={standbyPeers}
        pagination={false}
      />
    </Card>
  );
};

export default StandbyRelationCard;
