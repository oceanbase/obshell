import FrameBox from '@/component/FrameBox';
import PageLoading from '@/component/PageLoading';
import useDocumentTitle from '@/hook/useDocumentTitle';
import { formatMessage } from '@/util/intl';
import * as ObAshController from '@/service/obshell/tenant';
import * as TenantService from '@/service/obshell/tenant';
import React, { useEffect } from 'react';
import { useRequest } from 'ahooks';

interface DetailProps {
  match: {
    params: {
      clusterId: number;
      tenantId: number;
      reportId: number;
    };
  };
}

const Detail: React.FC<DetailProps> = ({
  match: {
    params: { clusterId, tenantId, reportId },
  },
}) => {
  const { data, loading: getTenantLoading } = useRequest(TenantService.getTenant, {
    defaultParams: [
      {
        id: clusterId,
        tenantId,
      },
    ],
  });

  const currentTenant = data?.data || {};

  const tenantLabel = `${currentTenant?.name || '-'}:${currentTenant?.obTenantId || '-'}`;

  useDocumentTitle(
    tenantLabel &&
      formatMessage(
        {
          id: 'ocp-v2.Session.Report.Detail.ActiveSessionHistoryOfTenantlabel',
          defaultMessage: '{tenantLabel} 的活跃会话历史报告',
        },
        { tenantLabel }
      )
  );
  /**
   * 获取报告信息
   */ const { loading, data: htmlData } = useRequest(ObAshController.getAshReport, {
    defaultParams: [
      {
        id: clusterId,
        tenantId,
        reportId,
      },
    ],
  });

  /**
   * 监听 message ，同步子 iframe 的路由到 OCP
   * @param event 事件对象
   */
  const handleReceiveMessage = (event: any) => {
    if (event !== undefined) {
      if (typeof event?.data === 'string') {
        window.history.replaceState({}, '', event.data);
      }
    }
  };

  useEffect(() => {
    // 监听message事件
    window.addEventListener('message', handleReceiveMessage, false);
    return () => {
      window.removeEventListener('message', handleReceiveMessage, false);
    };
  }, []);

  return (
    <>
      {loading || getTenantLoading ? (
        <PageLoading />
      ) : (
        <FrameBox
          title={formatMessage({
            id: 'ocp-v2.Session.Report.Detail.ActiveSessionHistoryReport',
            defaultMessage: '活跃会话历史报告',
          })}
          data={htmlData}
        />
      )}
    </>
  );
};

export default Detail;
