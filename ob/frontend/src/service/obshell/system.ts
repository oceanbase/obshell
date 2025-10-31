// @ts-ignore
/* eslint-disable */
import request from '@/util/request';

/** Get Alertmanager configuration Get Alertmanager configuration GET /api/v1/system/external/alertmanager */
export async function getSystemExternalAlertmanager(options?: { [key: string]: any }) {
  return request<API.OcsAgentResponse & { data?: API.AlertmanagerConfig }>(
    '/api/v1/system/external/alertmanager',
    {
      method: 'GET',
      ...(options || {}),
    }
  );
}

/** Set Alertmanager configuration Set Alertmanager configuration PUT /api/v1/system/external/alertmanager */
export async function putSystemExternalAlertmanager(
  body: API.AlertmanagerConfig,
  options?: { [key: string]: any }
) {
  return request<API.OcsAgentResponse>('/api/v1/system/external/alertmanager', {
    method: 'PUT',
    headers: {
      'Content-Type': 'application/json',
    },
    data: body,
    ...(options || {}),
  });
}

/** Get Prometheus configuration Get Prometheus configuration GET /api/v1/system/external/prometheus */
export async function getSystemExternalPrometheus(options?: { [key: string]: any }) {
  return request<API.OcsAgentResponse & { data?: API.PrometheusConfig }>(
    '/api/v1/system/external/prometheus',
    {
      method: 'GET',
      ...(options || {}),
    }
  );
}

/** Set Prometheus configuration Set Prometheus configuration PUT /api/v1/system/external/prometheus */
export async function putSystemExternalPrometheus(
  body: API.PrometheusConfig,
  options?: { [key: string]: any }
) {
  return request<API.OcsAgentResponse>('/api/v1/system/external/prometheus', {
    method: 'PUT',
    headers: {
      'Content-Type': 'application/json',
    },
    data: body,
    ...(options || {}),
  });
}
