// @ts-ignore
/* eslint-disable */
import request from '@/util/request';

/** create tenant create tenant POST /api/v1/tenant */
export async function tenantCreate(body: API.CreateTenantParam, options?: { [key: string]: any }) {
  return request<API.OcsAgentResponse & { data?: API.DagDetailDTO }>('/api/v1/tenant', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    data: body,
    ...(options || {}),
  });
}

/** get tenant info get tenant info GET /api/v1/tenant/${param0} */
export async function getTenantInfo(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.getTenantInfoParams,
  options?: { [key: string]: any }
) {
  const { name: param0, ...queryParams } = params;
  return request<API.OcsAgentResponse & { data?: API.TenantInfo }>(`/api/v1/tenant/${param0}`, {
    method: 'GET',
    params: { ...queryParams },
    ...(options || {}),
  });
}

/** rename tenant rename tenant PUT /api/v1/tenant/${param0} */
export async function tenantRename(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.tenantRenameParams,
  body: API.RenameTenantParam,
  options?: { [key: string]: any }
) {
  const { name: param0, ...queryParams } = params;
  return request<API.OcsAgentResponse>(`/api/v1/tenant/${param0}`, {
    method: 'PUT',
    headers: {
      'Content-Type': 'application/json',
    },
    params: { ...queryParams },
    data: body,
    ...(options || {}),
  });
}

/** drop tenant drop tenant DELETE /api/v1/tenant/${param0} */
export async function tenantDrop(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.tenantDropParams,
  body: API.DropTenantParam,
  options?: { [key: string]: any }
) {
  const { name: param0, ...queryParams } = params;
  return request<API.OcsAgentResponse & { data?: API.DagDetailDTO }>(`/api/v1/tenant/${param0}`, {
    method: 'DELETE',
    headers: {
      'Content-Type': 'application/json',
    },
    params: { ...queryParams },
    data: body,
    ...(options || {}),
  });
}

/** trigger tenant major compaction trigger tenant major compaction POST /api/v1/tenant/${param0}/compact */
export async function tenantMajorCompaction(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.tenantMajorCompactionParams,
  options?: { [key: string]: any }
) {
  const { name: param0, ...queryParams } = params;
  return request<API.OcsAgentResponse>(`/api/v1/tenant/${param0}/compact`, {
    method: 'POST',
    params: { ...queryParams },
    ...(options || {}),
  });
}

/** get tenant major compaction info get tenant major compaction info GET /api/v1/tenant/${param0}/compaction */
export async function getTenantCompaction(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.getTenantCompactionParams,
  options?: { [key: string]: any }
) {
  const { name: param0, ...queryParams } = params;
  return request<API.OcsAgentResponse & { data?: API.TenantCompaction }>(
    `/api/v1/tenant/${param0}/compaction`,
    {
      method: 'GET',
      params: { ...queryParams },
      ...(options || {}),
    }
  );
}

/** clear tenant major compaction error clear tenant major compaction error DELETE /api/v1/tenant/${param0}/compaction-error */
export async function tenantClearCompactionError(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.tenantClearCompactionErrorParams,
  options?: { [key: string]: any }
) {
  const { name: param0, ...queryParams } = params;
  return request<API.OcsAgentResponse>(`/api/v1/tenant/${param0}/compaction-error`, {
    method: 'DELETE',
    params: { ...queryParams },
    ...(options || {}),
  });
}

/** list databases list databases from a tenant GET /api/v1/tenant/${param0}/databases */
export async function listDatabases(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.listDatabasesParams,
  options?: { [key: string]: any }
) {
  const { name: param0, ...queryParams } = params;
  return request<API.OcsAgentResponse & { data?: API.Database[] }>(
    `/api/v1/tenant/${param0}/databases`,
    {
      method: 'GET',
      params: { ...queryParams },
      ...(options || {}),
    }
  );
}

/** create database create database POST /api/v1/tenant/${param0}/databases */
export async function createDatabase(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.createDatabaseParams,
  body: API.CreateDatabaseParam,
  options?: { [key: string]: any }
) {
  const { name: param0, ...queryParams } = params;
  return request<API.OcsAgentResponse>(`/api/v1/tenant/${param0}/databases`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    params: { ...queryParams },
    data: body,
    ...(options || {}),
  });
}

/** get database get database from a tenant GET /api/v1/tenant/${param0}/databases/${param1} */
export async function getDatabase(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.getDatabaseParams,
  options?: { [key: string]: any }
) {
  const { name: param0, database: param1, ...queryParams } = params;
  return request<API.OcsAgentResponse & { data?: API.Database }>(
    `/api/v1/tenant/${param0}/databases/${param1}`,
    {
      method: 'GET',
      params: { ...queryParams },
      ...(options || {}),
    }
  );
}

/** update database update database PUT /api/v1/tenant/${param0}/databases/${param1} */
export async function updateDatabase(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.updateDatabaseParams,
  body: API.ModifyDatabaseParam,
  options?: { [key: string]: any }
) {
  const { name: param0, database: param1, ...queryParams } = params;
  return request<API.OcsAgentResponse>(`/api/v1/tenant/${param0}/databases/${param1}`, {
    method: 'PUT',
    headers: {
      'Content-Type': 'application/json',
    },
    params: { ...queryParams },
    data: body,
    ...(options || {}),
  });
}

/** delete database delete database DELETE /api/v1/tenant/${param0}/databases/${param1} */
export async function deleteDatabase(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.deleteDatabaseParams,
  options?: { [key: string]: any }
) {
  const { name: param0, database: param1, ...queryParams } = params;
  return request<API.OcsAgentResponse>(`/api/v1/tenant/${param0}/databases/${param1}`, {
    method: 'DELETE',
    params: { ...queryParams },
    ...(options || {}),
  });
}

/** list tenant deadlocks list tenant deadlocks GET /api/v1/tenant/${param0}/deadlocks */
export async function listTenantDeadlocks(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.listTenantDeadlocksParams,
  options?: { [key: string]: any }
) {
  const { name: param0, ...queryParams } = params;
  return request<API.OcsAgentResponse & { data?: API.PaginatedDeadLocks }>(
    `/api/v1/tenant/${param0}/deadlocks`,
    {
      method: 'GET',
      params: {
        ...queryParams,
      },
      ...(options || {}),
    }
  );
}

/** lock tenant lock tenant POST /api/v1/tenant/${param0}/lock */
export async function tenantLock(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.tenantLockParams,
  options?: { [key: string]: any }
) {
  const { name: param0, ...queryParams } = params;
  return request<API.OcsAgentResponse>(`/api/v1/tenant/${param0}/lock`, {
    method: 'POST',
    params: { ...queryParams },
    ...(options || {}),
  });
}

/** unlock tenant unlock tenant DELETE /api/v1/tenant/${param0}/lock */
export async function tenantUnlock(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.tenantUnlockParams,
  options?: { [key: string]: any }
) {
  const { name: param0, ...queryParams } = params;
  return request<API.OcsAgentResponse>(`/api/v1/tenant/${param0}/lock`, {
    method: 'DELETE',
    params: { ...queryParams },
    ...(options || {}),
  });
}

/** list objects list objects from a tenant GET /api/v1/tenant/${param0}/objects */
export async function listObjects(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.listObjectsParams,
  body: API.TenantRootPasswordParam,
  options?: { [key: string]: any }
) {
  const { name: param0, ...queryParams } = params;
  return request<API.OcsAgentResponse & { data?: API.DbaObjectBo[] }>(
    `/api/v1/tenant/${param0}/objects`,
    {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
      },
      params: { ...queryParams },
      data: body,
      ...(options || {}),
    }
  );
}

/** get tenant parameter get tenant parameter GET /api/v1/tenant/${param0}/parameter/${param1} */
export async function getTenantParameter(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.getTenantParameterParams,
  options?: { [key: string]: any }
) {
  const { name: param0, para: param1, ...queryParams } = params;
  return request<API.OcsAgentResponse & { data?: API.GvObParameter }>(
    `/api/v1/tenant/${param0}/parameter/${param1}`,
    {
      method: 'GET',
      params: { ...queryParams },
      ...(options || {}),
    }
  );
}

/** get tenant parameters get tenant parameters GET /api/v1/tenant/${param0}/parameters */
export async function getTenantParameters(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.getTenantParametersParams,
  options?: { [key: string]: any }
) {
  const { name: param0, ...queryParams } = params;
  return request<API.OcsAgentResponse & { data?: API.GvObParameter[] }>(
    `/api/v1/tenant/${param0}/parameters`,
    {
      method: 'GET',
      params: {
        ...queryParams,
      },
      ...(options || {}),
    }
  );
}

/** set tenant parameters set tenant parameters PUT /api/v1/tenant/${param0}/parameters */
export async function tenantSetParameters(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.tenantSetParametersParams,
  body: API.SetTenantParametersParam,
  options?: { [key: string]: any }
) {
  const { name: param0, ...queryParams } = params;
  return request<API.OcsAgentResponse>(`/api/v1/tenant/${param0}/parameters`, {
    method: 'PUT',
    headers: {
      'Content-Type': 'application/json',
    },
    params: { ...queryParams },
    data: body,
    ...(options || {}),
  });
}

/** modify tenant root password modify tenant root password PUT /api/v1/tenant/${param0}/password */
export async function tenantModifyPassword(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.tenantModifyPasswordParams,
  body: API.ModifyTenantRootPasswordParam,
  options?: { [key: string]: any }
) {
  const { name: param0, ...queryParams } = params;
  return request<API.OcsAgentResponse>(`/api/v1/tenant/${param0}/password`, {
    method: 'PUT',
    headers: {
      'Content-Type': 'application/json',
    },
    params: { ...queryParams },
    data: body,
    ...(options || {}),
  });
}

/** persist tenant root password persist tenant root password POST /api/v1/tenant/${param0}/password/persist */
export async function persistTenantRootPassword(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.persistTenantRootPasswordParams,
  body: API.PersistTenantRootPasswordParam,
  options?: { [key: string]: any }
) {
  const { name: param0, ...queryParams } = params;
  return request<API.OcsAgentResponse>(`/api/v1/tenant/${param0}/password/persist`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    params: { ...queryParams },
    data: body,
    ...(options || {}),
  });
}

/** check tenant accessibility check tenant accessibility GET /api/v1/tenant/${param0}/precheck */
export async function tenantPreCheck(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.tenantPreCheckParams,
  options?: { [key: string]: any }
) {
  const { name: param0, ...queryParams } = params;
  return request<API.OcsAgentResponse & { data?: API.ObTenantPreCheckResult }>(
    `/api/v1/tenant/${param0}/precheck`,
    {
      method: 'GET',
      params: { ...queryParams },
      ...(options || {}),
    }
  );
}

/** modify tenant primary zone modify tenant primary zone PUT /api/v1/tenant/${param0}/primary-zone */
export async function tenantModifyPrimaryZone(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.tenantModifyPrimaryZoneParams,
  body: API.ModifyTenantPrimaryZoneParam,
  options?: { [key: string]: any }
) {
  const { name: param0, ...queryParams } = params;
  return request<API.OcsAgentResponse & { data?: API.DagDetailDTO }>(
    `/api/v1/tenant/${param0}/primary-zone`,
    {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json',
      },
      params: { ...queryParams },
      data: body,
      ...(options || {}),
    }
  );
}

/** add replicas to tenant add replicas to tenant POST /api/v1/tenant/${param0}/replicas */
export async function tenantAddReplicas(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.tenantAddReplicasParams,
  body: API.ScaleOutTenantReplicasParam,
  options?: { [key: string]: any }
) {
  const { name: param0, ...queryParams } = params;
  return request<API.OcsAgentResponse & { data?: API.DagDetailDTO }>(
    `/api/v1/tenant/${param0}/replicas`,
    {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      params: { ...queryParams },
      data: body,
      ...(options || {}),
    }
  );
}

/** remove replicas from tenant remove replicas from tenant DELETE /api/v1/tenant/${param0}/replicas */
export async function tenantRemoveReplicas(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.tenantRemoveReplicasParams,
  body: API.ScaleInTenantReplicasParam,
  options?: { [key: string]: any }
) {
  const { name: param0, ...queryParams } = params;
  return request<API.OcsAgentResponse & { data?: API.DagDetailDTO }>(
    `/api/v1/tenant/${param0}/replicas`,
    {
      method: 'DELETE',
      headers: {
        'Content-Type': 'application/json',
      },
      params: { ...queryParams },
      data: body,
      ...(options || {}),
    }
  );
}

/** modify tenant replicas modify tenant replicas PATCH /api/v1/tenant/${param0}/replicas */
export async function tenantModifyReplicas(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.tenantModifyReplicasParams,
  body: API.ModifyReplicasParam,
  options?: { [key: string]: any }
) {
  const { name: param0, ...queryParams } = params;
  return request<API.OcsAgentResponse & { data?: API.DagDetailDTO }>(
    `/api/v1/tenant/${param0}/replicas`,
    {
      method: 'PATCH',
      headers: {
        'Content-Type': 'application/json',
      },
      params: { ...queryParams },
      data: body,
      ...(options || {}),
    }
  );
}

/** create role create role POST /api/v1/tenant/${param0}/role */
export async function createRole(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.createRoleParams,
  body: API.CreateRoleParam,
  options?: { [key: string]: any }
) {
  const { name: param0, ...queryParams } = params;
  return request<API.OcsAgentResponse>(`/api/v1/tenant/${param0}/role`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    params: { ...queryParams },
    data: body,
    ...(options || {}),
  });
}

/** get role get role GET /api/v1/tenant/${param0}/role/${param1} */
export async function getRole(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.getRoleParams,
  options?: { [key: string]: any }
) {
  const { name: param0, role: param1, ...queryParams } = params;
  return request<API.OcsAgentResponse & { data?: API.ObRole }>(
    `/api/v1/tenant/${param0}/role/${param1}`,
    {
      method: 'GET',
      params: { ...queryParams },
      ...(options || {}),
    }
  );
}

/** drop role drop role DELETE /api/v1/tenant/${param0}/role/${param1} */
export async function dropRole(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.dropRoleParams,
  body: API.DropRoleParam,
  options?: { [key: string]: any }
) {
  const { name: param0, role: param1, ...queryParams } = params;
  return request<API.OcsAgentResponse>(`/api/v1/tenant/${param0}/role/${param1}`, {
    method: 'DELETE',
    headers: {
      'Content-Type': 'application/json',
    },
    params: { ...queryParams },
    data: body,
    ...(options || {}),
  });
}

/** modify role global privilege modify role global privilege PUT /api/v1/tenant/${param0}/role/${param1}/global-privileges */
export async function modifyRoleGlobalPrivilege(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.modifyRoleGlobalPrivilegeParams,
  body: API.ModifyRoleGlobalPrivilegeParam,
  options?: { [key: string]: any }
) {
  const { name: param0, role: param1, ...queryParams } = params;
  return request<API.OcsAgentResponse>(
    `/api/v1/tenant/${param0}/role/${param1}/global-privileges`,
    {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json',
      },
      params: { ...queryParams },
      data: body,
      ...(options || {}),
    }
  );
}

/** modify role object privilege modify role object privilege PUT /api/v1/tenant/${param0}/role/${param1}/object-privileges */
export async function modifyRoleObjectPrivilege(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.modifyRoleObjectPrivilegeParams,
  body: API.ModifyObjectPrivilegeParam,
  options?: { [key: string]: any }
) {
  const { name: param0, role: param1, ...queryParams } = params;
  return request<API.OcsAgentResponse>(
    `/api/v1/tenant/${param0}/role/${param1}/object-privileges`,
    {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json',
      },
      params: { ...queryParams },
      data: body,
      ...(options || {}),
    }
  );
}

/** grant role object privilege grant role object privilege POST /api/v1/tenant/${param0}/role/${param1}/object-privileges */
export async function grantRoleObjectPrivilege(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.grantRoleObjectPrivilegeParams,
  body: API.GrantObjectPrivilegeParam,
  options?: { [key: string]: any }
) {
  const { name: param0, role: param1, ...queryParams } = params;
  return request<API.OcsAgentResponse>(
    `/api/v1/tenant/${param0}/role/${param1}/object-privileges`,
    {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      params: { ...queryParams },
      data: body,
      ...(options || {}),
    }
  );
}

/** revoke role object privilege revoke role object privilege DELETE /api/v1/tenant/${param0}/role/${param1}/object-privileges */
export async function revokeRoleObjectPrivilege(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.revokeRoleObjectPrivilegeParams,
  body: API.RevokeObjectPrivilegeParam,
  options?: { [key: string]: any }
) {
  const { name: param0, role: param1, ...queryParams } = params;
  return request<API.OcsAgentResponse>(
    `/api/v1/tenant/${param0}/role/${param1}/object-privileges`,
    {
      method: 'DELETE',
      headers: {
        'Content-Type': 'application/json',
      },
      params: { ...queryParams },
      data: body,
      ...(options || {}),
    }
  );
}

/** patch role object privilege patch role object privilege PATCH /api/v1/tenant/${param0}/role/${param1}/object-privileges */
export async function patchRoleObjectPrivilege(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.patchRoleObjectPrivilegeParams,
  body: API.ModifyObjectPrivilegeParam,
  options?: { [key: string]: any }
) {
  const { name: param0, role: param1, ...queryParams } = params;
  return request<API.OcsAgentResponse>(
    `/api/v1/tenant/${param0}/role/${param1}/object-privileges`,
    {
      method: 'PATCH',
      headers: {
        'Content-Type': 'application/json',
      },
      params: { ...queryParams },
      data: body,
      ...(options || {}),
    }
  );
}

/** modify role modify role PUT /api/v1/tenant/${param0}/role/${param1}/roles */
export async function modifyRole(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.modifyRoleParams,
  body: API.ModifyRoleParam,
  options?: { [key: string]: any }
) {
  const { name: param0, role: param1, ...queryParams } = params;
  return request<API.OcsAgentResponse>(`/api/v1/tenant/${param0}/role/${param1}/roles`, {
    method: 'PUT',
    headers: {
      'Content-Type': 'application/json',
    },
    params: { ...queryParams },
    data: body,
    ...(options || {}),
  });
}

/** list roles list roles only for oracle tenant GET /api/v1/tenant/${param0}/roles */
export async function listRoles(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.listRolesParams,
  options?: { [key: string]: any }
) {
  const { name: param0, ...queryParams } = params;
  return request<API.OcsAgentResponse & { data?: API.ObRole[] }>(`/api/v1/tenant/${param0}/roles`, {
    method: 'GET',
    params: { ...queryParams },
    ...(options || {}),
  });
}

/** get tenant sessions get tenant sessions GET /api/v1/tenant/${param0}/sessions */
export async function getTenantSessions(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.getTenantSessionsParams,
  options?: { [key: string]: any }
) {
  const { name: param0, ...queryParams } = params;
  return request<API.OcsAgentResponse & { data?: API.PaginatedTenantSessions }>(
    `/api/v1/tenant/${param0}/sessions`,
    {
      method: 'GET',
      params: {
        ...queryParams,
      },
      ...(options || {}),
    }
  );
}

/** kill tenant sessions kill tenant sessions DELETE /api/v1/tenant/${param0}/sessions */
export async function killTenantSessions(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.killTenantSessionsParams,
  body: API.KillTenantSessionsParam,
  options?: { [key: string]: any }
) {
  const { name: param0, ...queryParams } = params;
  return request<API.OcsAgentResponse>(`/api/v1/tenant/${param0}/sessions`, {
    method: 'DELETE',
    headers: {
      'Content-Type': 'application/json',
    },
    params: { ...queryParams },
    data: body,
    ...(options || {}),
  });
}

/** get tenant session get tenant session GET /api/v1/tenant/${param0}/sessions/${param1} */
export async function getTenantSession(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.getTenantSessionParams,
  options?: { [key: string]: any }
) {
  const { name: param0, sessionId: param1, ...queryParams } = params;
  return request<API.OcsAgentResponse & { data?: API.TenantSession }>(
    `/api/v1/tenant/${param0}/sessions/${param1}`,
    {
      method: 'GET',
      params: { ...queryParams },
      ...(options || {}),
    }
  );
}

/** kill tenant session query kill tenant session query DELETE /api/v1/tenant/${param0}/sessions/queries */
export async function killTenantSessionQuery(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.killTenantSessionQueryParams,
  body: API.KillTenantSessionQueryParam,
  options?: { [key: string]: any }
) {
  const { name: param0, ...queryParams } = params;
  return request<API.OcsAgentResponse>(`/api/v1/tenant/${param0}/sessions/queries`, {
    method: 'DELETE',
    headers: {
      'Content-Type': 'application/json',
    },
    params: { ...queryParams },
    data: body,
    ...(options || {}),
  });
}

/** get tenant sessions stats get tenant sessions stats GET /api/v1/tenant/${param0}/sessions/stats */
export async function getTenantSessionsStats(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.getTenantSessionsStatsParams,
  options?: { [key: string]: any }
) {
  const { name: param0, ...queryParams } = params;
  return request<API.OcsAgentResponse & { data?: API.TenantSessionStats }>(
    `/api/v1/tenant/${param0}/sessions/stats`,
    {
      method: 'GET',
      params: { ...queryParams },
      ...(options || {}),
    }
  );
}

/** list users list users from a tenant GET /api/v1/tenant/${param0}/user */
export async function listUsers(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.listUsersParams,
  options?: { [key: string]: any }
) {
  const { name: param0, ...queryParams } = params;
  return request<API.OcsAgentResponse & { data?: API.ObUser[] }>(`/api/v1/tenant/${param0}/user`, {
    method: 'GET',
    params: { ...queryParams },
    ...(options || {}),
  });
}

/** create user create user POST /api/v1/tenant/${param0}/user */
export async function createUser(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.createUserParams,
  body: API.CreateUserParam,
  options?: { [key: string]: any }
) {
  const { name: param0, ...queryParams } = params;
  return request<API.OcsAgentResponse>(`/api/v1/tenant/${param0}/user`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    params: { ...queryParams },
    data: body,
    ...(options || {}),
  });
}

/** get user get user from a tenant GET /api/v1/tenant/${param0}/user/${param1} */
export async function getUser(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.getUserParams,
  options?: { [key: string]: any }
) {
  const { name: param0, user: param1, ...queryParams } = params;
  return request<API.OcsAgentResponse & { data?: API.ObUser }>(
    `/api/v1/tenant/${param0}/user/${param1}`,
    {
      method: 'GET',
      params: { ...queryParams },
      ...(options || {}),
    }
  );
}

/** drop user drop user DELETE /api/v1/tenant/${param0}/user/${param1} */
export async function dropUser(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.dropUserParams,
  body: API.DropUserParam,
  options?: { [key: string]: any }
) {
  const { name: param0, user: param1, ...queryParams } = params;
  return request<API.OcsAgentResponse>(`/api/v1/tenant/${param0}/user/${param1}`, {
    method: 'DELETE',
    headers: {
      'Content-Type': 'application/json',
    },
    params: { ...queryParams },
    data: body,
    ...(options || {}),
  });
}

/** modify db privilege of a user modify db privilege of a user PUT /api/v1/tenant/${param0}/user/${param1}/db-privilege */
export async function modifyDbPrivilege(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.modifyDbPrivilegeParams,
  body: API.ModifyUserDbPrivilegeParam,
  options?: { [key: string]: any }
) {
  const { name: param0, user: param1, ...queryParams } = params;
  return request<API.OcsAgentResponse>(`/api/v1/tenant/${param0}/user/${param1}/db-privilege`, {
    method: 'PUT',
    headers: {
      'Content-Type': 'application/json',
    },
    params: { ...queryParams },
    data: body,
    ...(options || {}),
  });
}

/** modify global privilege of a user modify global privilege of a user PUT /api/v1/tenant/${param0}/user/${param1}/global-privilege */
export async function modifyGlobalPrivilege(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.modifyGlobalPrivilegeParams,
  body: API.ModifyUserGlobalPrivilegeParam,
  options?: { [key: string]: any }
) {
  const { name: param0, user: param1, ...queryParams } = params;
  return request<API.OcsAgentResponse>(`/api/v1/tenant/${param0}/user/${param1}/global-privilege`, {
    method: 'PUT',
    headers: {
      'Content-Type': 'application/json',
    },
    params: { ...queryParams },
    data: body,
    ...(options || {}),
  });
}

/** lock user lock user PUT /api/v1/tenant/${param0}/user/${param1}/lock */
export async function lockUser(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.lockUserParams,
  options?: { [key: string]: any }
) {
  const { name: param0, user: param1, ...queryParams } = params;
  return request<API.OcsAgentResponse>(`/api/v1/tenant/${param0}/user/${param1}/lock`, {
    method: 'PUT',
    params: { ...queryParams },
    ...(options || {}),
  });
}

/** unlock user unlock user DELETE /api/v1/tenant/${param0}/user/${param1}/lock */
export async function unlockUser(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.unlockUserParams,
  options?: { [key: string]: any }
) {
  const { name: param0, user: param1, ...queryParams } = params;
  return request<API.OcsAgentResponse>(`/api/v1/tenant/${param0}/user/${param1}/lock`, {
    method: 'DELETE',
    params: { ...queryParams },
    ...(options || {}),
  });
}

/** modify user object privilege modify user object privilege PUT /api/v1/tenant/${param0}/user/${param1}/object-privileges */
export async function modifyUserObjectPrivilege(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.modifyUserObjectPrivilegeParams,
  body: API.ModifyObjectPrivilegeParam,
  options?: { [key: string]: any }
) {
  const { name: param0, user: param1, ...queryParams } = params;
  return request<API.OcsAgentResponse>(
    `/api/v1/tenant/${param0}/user/${param1}/object-privileges`,
    {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json',
      },
      params: { ...queryParams },
      data: body,
      ...(options || {}),
    }
  );
}

/** grant user object privilege grant user object privilege POST /api/v1/tenant/${param0}/user/${param1}/object-privileges */
export async function grantUserObjectPrivilege(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.grantUserObjectPrivilegeParams,
  body: API.GrantObjectPrivilegeParam,
  options?: { [key: string]: any }
) {
  const { name: param0, user: param1, ...queryParams } = params;
  return request<API.OcsAgentResponse>(
    `/api/v1/tenant/${param0}/user/${param1}/object-privileges`,
    {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      params: { ...queryParams },
      data: body,
      ...(options || {}),
    }
  );
}

/** revoke user object privilege revoke user object privilege DELETE /api/v1/tenant/${param0}/user/${param1}/object-privileges */
export async function revokeUserObjectPrivilege(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.revokeUserObjectPrivilegeParams,
  body: API.RevokeObjectPrivilegeParam,
  options?: { [key: string]: any }
) {
  const { name: param0, user: param1, ...queryParams } = params;
  return request<API.OcsAgentResponse>(
    `/api/v1/tenant/${param0}/user/${param1}/object-privileges`,
    {
      method: 'DELETE',
      headers: {
        'Content-Type': 'application/json',
      },
      params: { ...queryParams },
      data: body,
      ...(options || {}),
    }
  );
}

/** patch user object privilege patch user object privilege PATCH /api/v1/tenant/${param0}/user/${param1}/object-privileges */
export async function patchUserObjectPrivilege(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.patchUserObjectPrivilegeParams,
  body: API.ModifyObjectPrivilegeParam,
  options?: { [key: string]: any }
) {
  const { name: param0, user: param1, ...queryParams } = params;
  return request<API.OcsAgentResponse>(
    `/api/v1/tenant/${param0}/user/${param1}/object-privileges`,
    {
      method: 'PATCH',
      headers: {
        'Content-Type': 'application/json',
      },
      params: { ...queryParams },
      data: body,
      ...(options || {}),
    }
  );
}

/** change user password change user password PUT /api/v1/tenant/${param0}/user/${param1}/password */
export async function changePassword(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.changePasswordParams,
  body: API.ChangeUserPasswordParam,
  options?: { [key: string]: any }
) {
  const { name: param0, user: param1, ...queryParams } = params;
  return request<API.OcsAgentResponse>(`/api/v1/tenant/${param0}/user/${param1}/password`, {
    method: 'PUT',
    headers: {
      'Content-Type': 'application/json',
    },
    params: { ...queryParams },
    data: body,
    ...(options || {}),
  });
}

/** modify user role modify user role PUT /api/v1/tenant/${param0}/user/${param1}/roles */
export async function modifyUserRoles(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.modifyUserRolesParams,
  body: API.ModifyRoleParam,
  options?: { [key: string]: any }
) {
  const { name: param0, user: param1, ...queryParams } = params;
  return request<API.OcsAgentResponse>(`/api/v1/tenant/${param0}/user/${param1}/roles`, {
    method: 'PUT',
    headers: {
      'Content-Type': 'application/json',
    },
    params: { ...queryParams },
    data: body,
    ...(options || {}),
  });
}

/** get user stats get user stats GET /api/v1/tenant/${param0}/user/${param1}/stats */
export async function getStats(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.getStatsParams,
  options?: { [key: string]: any }
) {
  const { name: param0, user: param1, ...queryParams } = params;
  return request<API.OcsAgentResponse & { data?: API.ObUserStats }>(
    `/api/v1/tenant/${param0}/user/${param1}/stats`,
    {
      method: 'GET',
      params: { ...queryParams },
      ...(options || {}),
    }
  );
}

/** get tenant variable get tenant variable GET /api/v1/tenant/${param0}/variable/${param1} */
export async function getTenantVariable(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.getTenantVariableParams,
  options?: { [key: string]: any }
) {
  const { name: param0, var: param1, ...queryParams } = params;
  return request<API.OcsAgentResponse & { data?: API.CdbObSysVariable }>(
    `/api/v1/tenant/${param0}/variable/${param1}`,
    {
      method: 'GET',
      params: { ...queryParams },
      ...(options || {}),
    }
  );
}

/** get tenant variables get tenant variables GET /api/v1/tenant/${param0}/variables */
export async function getTenantVariables(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.getTenantVariablesParams,
  options?: { [key: string]: any }
) {
  const { name: param0, ...queryParams } = params;
  return request<API.OcsAgentResponse & { data?: API.CdbObSysVariable[] }>(
    `/api/v1/tenant/${param0}/variables`,
    {
      method: 'GET',
      params: {
        ...queryParams,
      },
      ...(options || {}),
    }
  );
}

/** set tenant variables set tenant variables PUT /api/v1/tenant/${param0}/variables */
export async function tenantSetVariable(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.tenantSetVariableParams,
  body: API.SetTenantVariablesParam,
  options?: { [key: string]: any }
) {
  const { name: param0, ...queryParams } = params;
  return request<API.OcsAgentResponse>(`/api/v1/tenant/${param0}/variables`, {
    method: 'PUT',
    headers: {
      'Content-Type': 'application/json',
    },
    params: { ...queryParams },
    data: body,
    ...(options || {}),
  });
}

/** modify tenant whitelist modify tenant whitelist PUT /api/v1/tenant/${param0}/whitelist */
export async function tenantModifyWhitelist(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.tenantModifyWhitelistParams,
  body: API.ModifyTenantWhitelistParam,
  options?: { [key: string]: any }
) {
  const { name: param0, ...queryParams } = params;
  return request<API.OcsAgentResponse>(`/api/v1/tenant/${param0}/whitelist`, {
    method: 'PUT',
    headers: {
      'Content-Type': 'application/json',
    },
    params: { ...queryParams },
    data: body,
    ...(options || {}),
  });
}

/** list support parameter templates list support parameter templates GET /api/v1/tenant/support-templates */
export async function listSupportParameterTemplates(options?: { [key: string]: any }) {
  return request<API.OcsAgentResponse & { data?: API.ParameterTemplate[] }>(
    '/api/v1/tenant/support-templates',
    {
      method: 'GET',
      ...(options || {}),
    }
  );
}

/** query tenant information ranked by the cost of major compaction. query tenant information ranked by the cost of major compaction, limited to the top n. GET /api/v1/tenant/top-compactions */
export async function getTenantTopCompaction(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.getTenantTopCompactionParams,
  options?: { [key: string]: any }
) {
  return request<API.OcsAgentResponse & { data?: API.TenantCompactionHistory[] }>(
    '/api/v1/tenant/top-compactions',
    {
      method: 'GET',
      params: {
        ...params,
      },
      ...(options || {}),
    }
  );
}

/** query tenant information ranked by the number of slow SQL statements. query tenant information ranked by the number of slow SQL statements, limited to the top n. GET /api/v1/tenant/top-slow-sqls */
export async function getTenantTopSlowSqlRank(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.getTenantTopSlowSqlRankParams,
  options?: { [key: string]: any }
) {
  return request<API.OcsAgentResponse & { data?: API.TenantSlowSqlCount[] }>(
    '/api/v1/tenant/top-slow-sqls',
    {
      method: 'GET',
      params: {
        ...params,
      },
      ...(options || {}),
    }
  );
}

/** get tenant overview get tenant overview GET /api/v1/tenants/overview */
export async function getTenantOverView(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.getTenantOverViewParams,
  options?: { [key: string]: any }
) {
  return request<API.OcsAgentResponse & { data?: API.DbaObTenant[] }>('/api/v1/tenants/overview', {
    method: 'GET',
    params: {
      ...params,
    },
    ...(options || {}),
  });
}
