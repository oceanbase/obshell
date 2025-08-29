import { formatMessage } from '@/util/intl';
import tracert from '@/util/tracert';
import { useLocation } from 'umi';
import React from 'react';
import { Divider, theme } from '@oceanbase/design';

export interface PageDocLinkProps extends Omit<React.AnchorHTMLAttributes<HTMLElement>, 'title'> {
  divider?: true;
  title?: React.ReactNode;
}

// 页容器的文档链接入口，需要通过 PageContainer 的 subTitle 进行设置
const PageDocLink: React.FC<PageDocLinkProps> = ({
  divider = true,
  title = formatMessage({
    id: 'ocp-v2.src.component.PageDocLink.Features',
    defaultMessage: '功能介绍',
  }),
  style,
  ...restProps
}) => {
  const { token } = theme.useToken();
  const { pathname } = useLocation();

  return (
    <span>
      {divider && (
        <Divider
          type="vertical"
          style={{ borderColor: token.colorBorder, marginLeft: 0, marginRight: 12 }}
        />
      )}

      <a
        target="_blank"
        style={{
          fontSize: 14,
          ...style,
        }}
        {...restProps}
        data-aspm-click="ca29867.da14040"
        data-aspm-desc="页容器-文档跳转链接"
        data-aspm-expo
        // 扩展参数
        data-aspm-param={tracert.stringify({
          // 文档链接所处的页面路径
          docLinkPathname: pathname,
          // 文档链接的标题
          docLinkTitle: title,
          // 文档链接的 URL
          docLinkUrl: restProps.href,
        })}
      >
        {title}
      </a>
    </span>
  );
};

export default PageDocLink;
