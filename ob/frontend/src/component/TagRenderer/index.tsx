import type { CSSProperties } from 'react';
import React, { useLayoutEffect, useMemo, useRef, useState } from 'react';
import { Tag, Tooltip } from '@oceanbase/design';
import { useSize } from 'ahooks';
import styles from './index.less';

export interface ColorfulTag {
  color?: string;
  text: string;
}
export type AllTag = ColorfulTag | string;
export interface TagProps {
  width?: number;
  tags: AllTag[];
  maxTagTextLength?: number;
  style?: CSSProperties;
  className?: string;
  tooltipVisible?: boolean;
  ellipsisTooltipVisble?: boolean;
}

const isColorfulTag = (tag: AllTag) => !!(tag as ColorfulTag).text;
export default (props: TagProps) => {
  const { className, style, tooltipVisible = true, ellipsisTooltipVisble = true } = props;
  const [limit, setLimit] = useState(0);
  const [done, setDone] = useState(false);
  const tagsRef = useRef<HTMLDivElement>(null);
  const parentRef = useRef<HTMLDivElement>(null);
  const parentSize = useSize(parentRef);
  useLayoutEffect(() => {
    // tags 或者 width 改变，重新开始计算
    setDone(false);
    setLimit(0);
  }, [props.tags, props.width, parentSize?.width, parentSize?.height]);
  // 使用 layoutEffect 阻塞渲染
  useLayoutEffect(() => {
    // 如果未结束
    if (!done) {
      // @ts-ignore
      const tagLength = tagsRef.current.offsetWidth;
      // @ts-ignore
      const parentLength = parentRef.current.offsetWidth;
      if (tagLength <= parentLength) {
        setLimit(l => (l < props.tags.length ? l + 1 : l));
      } else {
        // 已达到最大数量，停止计算
        setDone(true);
        setLimit(l => (l > 0 ? l - 1 : l));
      }
    }
  }, [limit, done]);
  // 设置 tag 超长字数，默认 20
  const maxCount = props.maxTagTextLength || 20;
  const tags = useMemo(() => {
    return props.tags.map(ele => {
      if (!isColorfulTag(ele)) {
        return {
          text: ele as string,
        } as ColorfulTag;
      }
      return ele as ColorfulTag;
    });
  }, [props.tags]);
  return (
    <div
      ref={parentRef}
      className={className}
      style={{ maxWidth: props.width, overflow: 'hidden', whiteSpace: 'nowrap', ...style }}
    >
      <div style={{ width: 'max-content' }} ref={tagsRef}>
        {tags.map((tag, index) => {
          const overLength = tag.text.length > maxCount;
          return (
            index <= limit &&
            (overLength && tooltipVisible ? (
              <Tooltip
                key={tag.text}
                overlayClassName={styles.techTagRendererPopup}
                title={
                  <Tag color={tag.color} key={tag.text}>
                    {tag.text}
                  </Tag>
                }
              >
                <Tag color={tag.color} key={tag.text}>
                  {overLength ? `${tag.text.slice(0, maxCount)}...` : tag.text}
                </Tag>
              </Tooltip>
            ) : (
              <Tag color={tag.color} key={tag.text}>
                {overLength ? `${tag.text.slice(0, maxCount)}...` : tag.text}
              </Tag>
            ))
          );
        })}
        {limit < tags.length &&
          (ellipsisTooltipVisble ? (
            <Tooltip
              overlayClassName={styles.techTagRendererPopup}
              title={tags.slice(limit + 1).map(ele => (
                <Tag color={ele.color} key={ele.text}>
                  {ele.text}
                </Tag>
              ))}
            >
              <Tag style={{ userSelect: 'none', cursor: 'default' }} color={tags[0]?.color}>
                ...
              </Tag>
            </Tooltip>
          ) : (
            <Tag style={{ userSelect: 'none', cursor: 'default' }} color={tags[0]?.color}>
              ...
            </Tag>
          ))}
      </div>
    </div>
  );
};
