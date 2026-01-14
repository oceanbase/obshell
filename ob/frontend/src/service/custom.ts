/** 遥测上报API */
export async function telemetryReport(params: any): Promise<any> {
  const reportUrl =
    process.env.NODE_ENV === 'production'
      ? 'https://openwebapi.oceanbase.com'
      : 'http://openwebapi.test.alipay.net';

  return fetch(`${reportUrl}/api/web/oceanbase/report`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(params),
  }).then(response => response.json());
}
