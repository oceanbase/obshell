import { formatMessage } from '@/util/intl';
import { ExclamationCircleFilled } from '@oceanbase/icons';
import type { ModalFuncProps } from '@oceanbase/design';
import { Modal } from '@oceanbase/design';
import { debounce } from 'lodash';
import React from 'react';
import styles from './index.less';

const { confirm } = Modal;
function showDeleteConfirm(props: ModalFuncProps) {
  confirm({
    icon: <ExclamationCircleFilled />,
    okText: formatMessage({
      id: 'OBShell.customModal.showDeleteConfirm.Yes',
      defaultMessage: '是的',
    }),
    okType: 'danger',
    cancelText: formatMessage({
      id: 'OBShell.customModal.showDeleteConfirm.Cancel',
      defaultMessage: '取消',
    }),
    className: styles.deleteContainer,
    ...props,
  });
}
export default debounce(showDeleteConfirm, 500, {
  leading: true,
  trailing: false,
});
