import { getEncryptLocalStorage, setEncryptLocalStorage } from '@/util';
import { LOGIN_IV, LOGIN_SECRET_KEY, SESSION_ID } from '@/constant/login';
import { aesDecrypt } from '@/util/aes';

/**
 * 处理登录成功后的响应数据
 * 如果响应数据是加密的字符串，进行解密并保存到 SESSION_ID
 * @param data 登录接口的响应数据
 */
export function handleLoginSuccess(data: string) {
  // 获取保存的 secretKey 和 iv 进行解密
  const secretKeyHex = getEncryptLocalStorage(LOGIN_SECRET_KEY);
  const ivHex = getEncryptLocalStorage(LOGIN_IV);

  if (secretKeyHex && ivHex) {
    try {
      // 如果响应数据是加密的字符串，进行解密
      if (typeof data === 'string') {
        const decryptedData = aesDecrypt(data, secretKeyHex, ivHex);
        if (decryptedData !== undefined) {
          setEncryptLocalStorage(SESSION_ID, decryptedData);
        }
      }
    } catch (error) {
      console.error('Failed to decrypt login response:', error);
    }
  }
}
