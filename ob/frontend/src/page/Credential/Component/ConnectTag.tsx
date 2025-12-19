import { formatMessage } from '@/util/intl';
import { isEnglish } from '@/util';
import React from 'react';
import { Tag, Tooltip } from '@oceanbase/design';

export interface ConnectTagProps {
  status?: boolean;
  message: string;
}

const ConnectTag: React.FC<ConnectTagProps> = ({ status, message }) => {
  const tagStyle = {
    borderRadius: 10,
    whiteSpace: 'break-spaces',
    wordBreak: 'break-all',
    marginLeft: '8px',
  };

  const showTag = (
    <Tag color={status ? 'green' : 'red'} style={isEnglish() ? tagStyle : { borderRadius: 10 }}>
      {status
        ? formatMessage({
            id: 'ocp-v2.Credential.Component.ConnectTag.VerifiedSuccessfully',
            defaultMessage: '验证成功',
          })
        : formatMessage({
            id: 'ocp-v2.Credential.Component.ConnectTag.VerificationFailed',
            defaultMessage: '验证失败',
          })}
    </Tag>
  );

  return <>{!!message ? <Tooltip title={message}>{showTag}</Tooltip> : showTag}</>;
};

export default ConnectTag;
