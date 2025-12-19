import AddHostCredentialDrawer from '@/component/AddHostCredentialDrawer';
import useDocumentTitle from '@/hook/useDocumentTitle';
import { formatMessage } from '@/util/intl';
import React, { useRef, useState } from 'react';
import { Button } from '@oceanbase/design';
import { PageContainer } from '@oceanbase/ui';
import Host, { HostRef } from './Host';

import styles from './index.less';
import ContentWithReload from '@/component/ContentWithReload';

const Credential: React.FC = () => {
  useDocumentTitle(
    formatMessage({ id: 'ocp-v2.page.Credential.CredentialManagement', defaultMessage: '凭据管理' })
  );

  const [hostVisible, setHostVisible] = useState(false);
  const [loading, setLoading] = useState(false);
  const hostRef = useRef<HostRef>(null);

  return (
    <PageContainer
      ghost={true}
      className={styles.credentialContainer}
      header={{
        title: (
          <ContentWithReload
            content="凭据管理"
            onClick={() => {
              hostRef.current?.refresh();
            }}
            spin={loading}
          />
        ),

        extra: (
          <Button
            type="primary"
            onClick={() => {
              setHostVisible(true);
            }}
          >
            {formatMessage({
              id: 'ocp-v2.page.Credential.CreateCredentials',
              defaultMessage: '新建凭据',
            })}
          </Button>
        ),
      }}
    >
      <Host ref={hostRef} onLoadingChange={setLoading} />

      <AddHostCredentialDrawer
        visible={hostVisible}
        onCancel={() => {
          setHostVisible(false);
        }}
        onSuccess={() => {
          setHostVisible(false);
          hostRef.current?.refresh();
        }}
      />
    </PageContainer>
  );
};

export default Credential;
