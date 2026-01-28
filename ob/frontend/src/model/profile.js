import { history } from 'umi';
import { getEncryptLocalStorage, setEncryptLocalStorage } from '@/util';
import { SESSION_ID } from '@/constant/login';
import * as v1Service from '@/service/obshell/v1';

export const namespace = 'profile';

const model = {
  namespace,
  state: {
    password: '',
    // RSA 加密用的公钥
    publicKey: '',
  },

  effects: {
    *logout({}, { put, call }) {
      const res = yield call(v1Service.logout, {
        session_id: getEncryptLocalStorage(SESSION_ID),
      });
      if (res.successful) {
        setEncryptLocalStorage(SESSION_ID, '');
        yield put({
          type: 'update',
          payload: {
            password: null,
          },
        });

        setEncryptLocalStorage('login', 'false');
        history.push('/login');
      }
    },
  },

  reducers: {
    update(state, { payload }) {
      return { ...state, ...payload };
    },
  },
};

export default model;
