import * as CredentialController from '@/service/obshell/security';
import { useState } from 'react';
import { useRequest } from 'ahooks';

export const useListCredentials = (targetType: string) => {
  const [keyword, setKeyword] = useState('');
  const { data, loading, refresh } = useRequest(
    () =>
      CredentialController.listCredentials({
        target_type: targetType,
        key_word: keyword,
      }),
    {
      refreshDeps: [keyword],
    }
  );

  const credentialList: API.Credential[] = data?.data?.contents || [];

  return {
    loading,
    credentialList,
    refresh,
    setKeyword,
  };
};
