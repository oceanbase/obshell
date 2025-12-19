import { formatMessage } from '@/util/intl';
import { getMySQLDbUserNameValidateMessage } from './component';

export const OPERATION_LIST = [
  {
    value: 'edit',
    label: formatMessage({
      id: 'ocp-express.src.constant.credential.Edit',
      defaultMessage: '编辑',
    }),
  },

  {
    value: 'delete',
    label: formatMessage({
      id: 'ocp-express.src.constant.credential.Delete',
      defaultMessage: '删除',
    }),
  },
];

export const TAB_LIST = [
  {
    value: 'OB',
    label: formatMessage({
      id: 'ocp-express.src.constant.credential.ObCluster',
      defaultMessage: 'OB 集群',
    }),
  },

  {
    value: 'OB_PROXY',
    label: 'OBProxy',
  },

  {
    value: 'HOST',
    label: formatMessage({
      id: 'ocp-express.src.constant.credential.Host',
      defaultMessage: '主机',
    }),
  },
];

export const HOST_CONNECTION_RESULT = [
  {
    value: 'CONNECT_FAILED',
    label: formatMessage({
      id: 'ocp-express.src.constant.credential.ConnectionFailed',
      defaultMessage: '连接失败',
    }),
  },

  {
    value: 'SUDO_NOT_ALLOWED',
    label: formatMessage({
      id: 'ocp-express.src.constant.credential.NoSudoPermission',
      defaultMessage: '无 sudo 权限',
    }),
  },

  {
    value: 'HOST_NOT_EXISTS',
    label: formatMessage({
      id: 'ocp-express.src.constant.credential.TheHostDoesNotExist',
      defaultMessage: '主机不存在',
    }),
  },
];

// 凭据名 命名规则
export const CREDENTIAL_NAME_RULE = {
  pattern: /^[a-zA-Z]{1,1}[a-zA-Z0-9_]{0,62}[a-zA-Z0-9]{1,1}$/,
  message: getMySQLDbUserNameValidateMessage(),
};
