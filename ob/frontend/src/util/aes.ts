import CryptoJS from 'crypto-js';
import cryptoRandomString from 'crypto-random-string';

export function aesEncrypt(content: string | object, generateKeysWhenEmpty = false) {
  if (content === undefined) {
    // 只有当 generateKeysWhenEmpty 为 true 时，才在 content 为空时生成密钥
    if (generateKeysWhenEmpty) {
      const secretKey = CryptoJS.enc.Hex.parse(cryptoRandomString({ length: 32 }));
      const iv = CryptoJS.enc.Hex.parse(cryptoRandomString({ length: 32 }));
      return { encryptedContent: undefined, secretKey, iv };
    }
    return { encryptedContent: undefined, secretKey: '', iv: '' };
  }

  const secretKey = CryptoJS.enc.Hex.parse(cryptoRandomString({ length: 32 }));
  const iv = CryptoJS.enc.Hex.parse(cryptoRandomString({ length: 32 }));

  try {
    let encryptedContent;
    if (typeof content === 'object') {
      encryptedContent = CryptoJS.AES.encrypt(JSON.stringify(content), secretKey, {
        iv,
        mode: CryptoJS.mode.CBC,
      }).toString();
    } else {
      encryptedContent = CryptoJS.AES.encrypt(content, secretKey, {
        iv,
        mode: CryptoJS.mode.CBC,
      }).toString();
    }

    return { encryptedContent, secretKey, iv };
  } catch (error) {
    console.log('aesEncrypt error: ', error);
    return { encryptedContent: undefined, secretKey: '', iv: '' };
  }
}

export function aesDecrypt(encryptedContent: string, secretKey: string, iv: string) {
  if (!encryptedContent || !secretKey || !iv) {
    return undefined;
  }
  try {
    const secretKeyWordArray = CryptoJS.enc.Hex.parse(secretKey);
    const ivWordArray = CryptoJS.enc.Hex.parse(iv);

    const decrypted = CryptoJS.AES.decrypt(encryptedContent, secretKeyWordArray, {
      iv: ivWordArray,
      mode: CryptoJS.mode.CBC,
    });
    const decryptedContent = decrypted.toString(CryptoJS.enc.Utf8);

    if (!decryptedContent) {
      return undefined;
    }
    return decryptedContent;
  } catch (error) {
    console.log('aesDecrypt error: ', error);
    return undefined;
  }
}

/**
 * 将 WordArray 转换为十六进制字符串
 */
export function wordArrayToHex(wordArray: CryptoJS.lib.WordArray) {
  if (!wordArray) return '';
  try {
    return CryptoJS.enc.Hex.stringify(wordArray);
  } catch (error) {
    console.error(error, 'wordArrayToHex error');
    return '';
  }
}
