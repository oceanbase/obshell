/*
 * Copyright (c) 2024 OceanBase.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

import { formatMessage } from '@/util/intl';
import React, { useState } from 'react';
import { Tag, Input } from '@oceanbase/design';
import { uniq, findIndex } from 'lodash';

/**
 * 说明
 * Input Tag
 *  */
interface BatchAddObjectTagsProps {}

const BatchAddObjectTags: React.FC<BatchAddObjectTagsProps> = () => {
  const [tags, setTags] = useState<string[]>([]);
  const [inputValue, setInputValue] = useState('');

  const handleInputChange = e => {
    if (e.target.value.replace(/\s*/g, '') !== '') {
      setInputValue(e.target.value);
    }
  };

  const handleInputConfirm = () => {
    if (inputValue.indexOf(',') !== -1) {
      setTags(uniq([...tags, ...inputValue.split(',').map(item => item.replace(/\s*/g, ''))]));
    } else if (inputValue !== '' && findIndex(tags, item => item === inputValue) === -1) {
      setTags(uniq([...tags, inputValue]));
    }
    setInputValue('');
  };

  const paddingValue = tags?.length > 0 ? 5 : 0;
  return (
    <>
      <div
        style={{
          border: '1px solid #d9d9d9',
          maxHeight: 560,
          padding: paddingValue,
          overflow: 'auto',
        }}
      >
        {tags.map(
          item =>
            item !== '' && (
              <Tag key={item} closable={true}>
                {item}
              </Tag>
            )
        )}

        <Input
          type="text"
          placeholder={formatMessage({
            id: 'ocp-express.Oracle.Component.BatchAddObjectTags.Enter',
            defaultMessage: '请输入',
          })}
          value={inputValue}
          bordered={false}
          onChange={handleInputChange}
          onBlur={handleInputConfirm}
          onPressEnter={handleInputConfirm}
        />
      </div>
    </>
  );
};
export default BatchAddObjectTags;
