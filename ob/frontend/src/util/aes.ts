import CryptoJS from 'crypto-js';
import cryptoRandomString from 'crypto-random-string';

export function aesEncrypt(content: string | object) {
  if (content === undefined) {
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
