// @ts-ignore
/* eslint-disable */
import request from '@/util/request';

/** get dag detail by id get dag detail by id GET /api/v1/task/dag/${param0} */
export async function getDagDetail(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.getDagDetailParams,
  options?: { [key: string]: any }
) {
  const { id: param0, ...queryParams } = params;
  return request<API.OcsAgentResponse & { data?: API.DagDetailDTO }>(`/api/v1/task/dag/${param0}`, {
    method: 'GET',
    params: {
      ...queryParams,
    },
    ...(options || {}),
  });
}

/** operate dag operate dag POST /api/v1/task/dag/${param0} */
export async function dagHandler(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.dagHandlerParams,
  body: API.DagOperator,
  options?: { [key: string]: any }
) {
  const { id: param0, ...queryParams } = params;
  return request<API.OcsAgentResponse>(`/api/v1/task/dag/${param0}`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    params: { ...queryParams },
    data: body,
    ...(options || {}),
  });
}

/** get agent unfinished dags get agent unfinished dags GET /api/v1/task/dag/agent/unfinish */
export async function getAgentUnfinishDags(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.GetAgentUnfinishDagsParams,
  options?: { [key: string]: any }
) {
  return request<API.OcsAgentResponse & { data?: API.DagDetailDTO[] }>(
    '/api/v1/task/dag/agent/unfinish',
    {
      method: 'GET',
      params: {
        ...params,
      },
      ...(options || {}),
    }
  );
}

/** get agent last maintenance dag get agent last maintenance dag GET /api/v1/task/dag/maintain/agent */
export async function getAgentLastMaintenanceDag(options?: { [key: string]: any }) {
  return request<API.OcsAgentResponse & { data?: API.DagDetailDTO }>(
    '/api/v1/task/dag/maintain/agent',
    {
      method: 'GET',
      ...(options || {}),
    }
  );
}

/** get ob last maintenance dag get ob last maintenance dag GET /api/v1/task/dag/maintain/ob */
export async function getObLastMaintenanceDag(options?: { [key: string]: any }) {
  return request<API.OcsAgentResponse & { data?: API.DagDetailDTO }>(
    '/api/v1/task/dag/maintain/ob',
    {
      method: 'GET',
      ...(options || {}),
    }
  );
}

/** get cluster unfinished dags get cluster unfinished dags GET /api/v1/task/dag/ob/unfinish */
export async function getClusterUnfinishDags(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.GetClusterUnfinishDagsParams,
  options?: { [key: string]: any }
) {
  return request<API.OcsAgentResponse & { data?: API.DagDetailDTO[] }>(
    '/api/v1/task/dag/ob/unfinish',
    {
      method: 'GET',
      params: {
        ...params,
      },
      ...(options || {}),
    }
  );
}

/** get unfinished dags get unfinished dags GET /api/v1/task/dag/unfinish */
export async function getUnfinishedDags(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.GetUnfinishedDagsParams,
  options?: { [key: string]: any }
) {
  return request<API.OcsAgentResponse & { data?: API.DagDetailDTO[] }>(
    '/api/v1/task/dag/unfinish',
    {
      method: 'GET',
      params: {
        ...params,
      },
      ...(options || {}),
    }
  );
}

/** get all dags in the agent get all dags in the agent GET /api/v1/task/dags/agent */
export async function getAllAgentDags(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.getAllAgentDagsParams,
  options?: { [key: string]: any }
) {
  return request<API.OcsAgentResponse & { data?: API.DagDetailDTO[] }>('/api/v1/task/dags/agent', {
    method: 'GET',
    params: {
      ...params,
    },
    ...(options || {}),
  });
}

/** get all dags in cluster get all dags in cluster GET /api/v1/task/dags/ob */
export async function getAllClusterDags(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.getAllClusterDagsParams,
  options?: { [key: string]: any }
) {
  return request<API.OcsAgentResponse & { data?: API.DagDetailDTO[] }>('/api/v1/task/dags/ob', {
    method: 'GET',
    params: {
      ...params,
    },
    ...(options || {}),
  });
}

/** get node detail by node_id get node detail by node_id GET /api/v1/task/node/${param0} */
export async function getNodeDetail(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.getNodeDetailParams,
  options?: { [key: string]: any }
) {
  const { id: param0, ...queryParams } = params;
  return request<API.OcsAgentResponse & { data?: API.NodeDetailDTO }>(
    `/api/v1/task/node/${param0}`,
    {
      method: 'GET',
      params: {
        ...queryParams,
      },
      ...(options || {}),
    }
  );
}

/** operate node operate node POST /api/v1/task/node/${param0} */
export async function nodeHandler(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.nodeHandlerParams,
  body: API.NodeOperator,
  options?: { [key: string]: any }
) {
  const { id: param0, ...queryParams } = params;
  return request<API.OcsAgentResponse>(`/api/v1/task/node/${param0}`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    params: { ...queryParams },
    data: body,
    ...(options || {}),
  });
}

/** get sub_task detail by sub_task_id get sub_task detail by sub_task_id GET /api/v1/task/sub_task/${param0} */
export async function getSubTaskDetail(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.getSubTaskDetailParams,
  options?: { [key: string]: any }
) {
  const { id: param0, ...queryParams } = params;
  return request<API.OcsAgentResponse & { data?: API.TaskDetailDTO }>(
    `/api/v1/task/sub_task/${param0}`,
    {
      method: 'GET',
      params: {
        ...queryParams,
      },
      ...(options || {}),
    }
  );
}

/** get full sub_task logs by sub_task_id get full sub_task logs by sub_task_id GET /api/v1/task/sub_task/${param0}/logs */
export async function getFullSubTaskLogs(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.getFullSubTaskLogsParams,
  options?: { [key: string]: any }
) {
  const { id: param0, ...queryParams } = params;
  return request<API.OcsAgentResponse & { data?: API.TaskDetailDTO }>(
    `/api/v1/task/sub_task/${param0}/logs`,
    {
      method: 'GET',
      params: { ...queryParams },
      ...(options || {}),
    }
  );
}
