// @ts-ignore
/* eslint-disable */
import request from '@/util/request';

/** create credential create a new credential with SSH validation and encryption POST /api/v1/security/credential */
export async function createCredential(
  body: API.CreateCredentialParam,
  options?: { [key: string]: any }
) {
  return request<API.OcsAgentResponse & { data?: API.Credential }>('/api/v1/security/credential', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    data: body,
    ...(options || {}),
  });
}

/** get credential get credential by ID GET /api/v1/security/credential/${param0} */
export async function getCredential(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.getCredentialParams,
  options?: { [key: string]: any }
) {
  const { id: param0, ...queryParams } = params;
  return request<API.OcsAgentResponse & { data?: API.Credential }>(
    `/api/v1/security/credential/${param0}`,
    {
      method: 'GET',
      params: { ...queryParams },
      ...(options || {}),
    }
  );
}

/** delete credential delete credential by ID (hard delete) DELETE /api/v1/security/credential/${param0} */
export async function deleteCredential(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.deleteCredentialParams,
  options?: { [key: string]: any }
) {
  const { id: param0, ...queryParams } = params;
  return request<API.OcsAgentResponse>(`/api/v1/security/credential/${param0}`, {
    method: 'DELETE',
    params: { ...queryParams },
    ...(options || {}),
  });
}

/** update credential update credential with SSH validation and encryption PATCH /api/v1/security/credential/${param0} */
export async function updateCredential(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.updateCredentialParams,
  body: API.UpdateCredentialParam,
  options?: { [key: string]: any }
) {
  const { id: param0, ...queryParams } = params;
  return request<API.OcsAgentResponse & { data?: API.Credential }>(
    `/api/v1/security/credential/${param0}`,
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

/** update AES key for credential encryption update AES key and re-encrypt all stored credential passphrases PUT /api/v1/security/credential/encrypt-secret-key */
export async function updateCredentialEncryptSecretKey(
  body: API.UpdateCredentialEncryptSecretKeyParam,
  options?: { [key: string]: any }
) {
  return request<API.OcsAgentResponse>('/api/v1/security/credential/encrypt-secret-key', {
    method: 'PUT',
    headers: {
      'Content-Type': 'application/json',
    },
    data: body,
    ...(options || {}),
  });
}

/** validate credential validate credential without storing it POST /api/v1/security/credential/validate */
export async function validateCredential(
  body: API.ValidateCredentialParam,
  options?: { [key: string]: any }
) {
  return request<API.OcsAgentResponse & { data?: API.ValidationResult }>(
    '/api/v1/security/credential/validate',
    {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      data: body,
      ...(options || {}),
    }
  );
}

/** list credentials list credentials with filtering, pagination, and sorting GET /api/v1/security/credentials */
export async function listCredentials(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.listCredentialsParams,
  options?: { [key: string]: any }
) {
  return request<API.OcsAgentResponse & { data?: API.PaginatedCredentialResponse }>(
    '/api/v1/security/credentials',
    {
      method: 'GET',
      params: {
        // page has a default value: 1
        page: '1',
        // page_size has a default value: 10
        page_size: '10',

        ...params,
      },
      ...(options || {}),
    }
  );
}

/** batch delete credentials delete multiple credentials by ID list (hard delete) DELETE /api/v1/security/credentials */
export async function batchDeleteCredentials(
  body: API.BatchDeleteCredentialParam,
  options?: { [key: string]: any }
) {
  return request<API.OcsAgentResponse>('/api/v1/security/credentials', {
    method: 'DELETE',
    headers: {
      'Content-Type': 'application/json',
    },
    data: body,
    ...(options || {}),
  });
}

/** batch validate credentials validate multiple credentials by ID list POST /api/v1/security/credentials/validate */
export async function batchValidateCredentials(
  body: API.BatchValidateCredentialParam,
  options?: { [key: string]: any }
) {
  return request<API.OcsAgentResponse & { data?: API.ValidationResult[] }>(
    '/api/v1/security/credentials/validate',
    {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      data: body,
      ...(options || {}),
    }
  );
}
