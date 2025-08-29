import { formatMessage } from '@/util/intl';
import { Tag, token } from '@oceanbase/design';
import { isNullValue } from '@oceanbase/util';
import React from 'react';

export const removeReturnChar = (str: string) => {
  if (!str?.length) return '';
  let newStr = str;
  ['\t', '\r', '\f', '\n'].forEach(char => (newStr = newStr?.replaceAll(char, '')));
  return newStr;
};

export function convertTagArray(array: API.Tag[]) {
  const result = [];

  const map = new Map();
  for (const item of array) {
    const { key, value, id } = item;

    if (!map.has(key)) {
      map.set(key, { label: key, value: key, children: [] });
    }

    const parent = map.get(key);
    parent.children.push({ label: value, value: id });
  }
  for (const item of map.values()) {
    result.push(item);
  }
  return result;
}

export function sortTagList(list: API.Tag[]) {
  const grouped = {};
  const sortedList = [];
  list.forEach(obj => {
    const key = obj.key;
    if (!grouped[key]) {
      grouped[key] = [];
    }
    grouped[key].push(obj);
  });
  for (const key in grouped) {
    if (grouped.hasOwnProperty(key)) {
      sortedList.push(...grouped[key]);
    }
  }
  return sortedList;
}

export function getTagLabel(tag: API.Tag) {
  return !isNullValue(tag?.value) ? `${tag?.key}：${tag?.value}` : tag?.key;
}

export function getTagRender(max: number, useTag: boolean, tags: API.Tag[]) {
  const firstTag = getTagLabel(tags?.[0]);
  // 第一个标签可能会过长，为保证窄屏显示，因此进行截取显示
  const firstTagLabel = firstTag?.slice(0, max);
  const firstContent =
    tags?.length > 1
      ? formatMessage(
          {
            id: 'OBShell.src.util.tag.FirsttaglabelAndOtherTagslengthItems',
            defaultMessage: '{firstTagLabel}...等{tagsLength}项',
          },
          { firstTagLabel: firstTagLabel, tagsLength: tags.length }
        )
      : (firstTag?.length || 0) > max + 1
      ? `${firstTagLabel}...`
      : firstTagLabel;
  const firstTagRender = useTag ? <Tag ellipsis={false}>{firstContent}</Tag> : firstContent;
  const tagTooltip = tags?.map(item => {
    return <Tag key={item?.id}>{getTagLabel(item)}</Tag>;
  });
  return { tagTooltip, firstTagRender };
}

export function getTagCascaderOption(options: any, currentValue?: number[], isFilter?: boolean) {
  const labelStyle = isFilter
    ? { marginBottom: '-5px' }
    : {
        maxWidth: '400px',
        minWidth: '130px',
        lineHeight: 'inherit',
        display: 'block',
      };
  return options?.map(option => {
    const selected = currentValue
      ?.filter(item => item?.[0] === option?.value)
      ?.map(item => item?.[1]);
    option?.children?.map(child => {
      const disabled = !!selected?.length && !selected?.includes(child?.value);
      return Object.assign(child, {
        disabled,
        label: isNullValue(child?.label) ? (
          <span
            style={{
              color: disabled ? token.colorTextDisabled : token.colorTextTertiary,
              ...labelStyle,
            }}
          >
            {formatMessage({ id: 'OBShell.src.util.tag.NullValue', defaultMessage: '空值' })}
          </span>
        ) : (
          <span className="ellipsis" style={labelStyle} title={child?.label}>
            {child?.label}
          </span>
        ),
      });
    });
    option.label = (
      <span className="ellipsis" style={labelStyle} title={option?.label}>
        {option?.label}
      </span>
    );

    return option;
  });
}

export function isEqualWithEmpty(value1, value2) {
  if (isNullValue(value1) && isNullValue(value2)) {
    return true;
  }
  if (isNullValue(value1) || isNullValue(value2)) {
    return false;
  }
  return value1 === value2;
}
