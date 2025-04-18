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
import React from 'react';
import { Checkbox, Button } from '@oceanbase/design';
import { DeleteOutlined } from '@oceanbase/icons';
import { DATABASE_PRIVILEGE_LIST } from '@/constant/tenant';
import styles from './index.less';

export interface PrivilegesCheckboxProps {
  dbPrivilegeParam: API.DbPrivilege;
  insideDb: boolean;
  deleteFn: (value: string) => void;
  updateDbPrivilege: (value: API.DbPrivilege) => void;
}

interface PrivilegesCheckboxState {
  checkedList: API.DbPrivType[];
  indeterminate: boolean;
  checkAll: boolean;
}

export default class PrivilegesCheckbox extends React.PureComponent<
  PrivilegesCheckboxProps,
  PrivilegesCheckboxState
> {
  constructor(props: PrivilegesCheckboxProps) {
    super(props);
    const privilegesLength = props?.dbPrivilegeParam?.privileges?.length;
    const isCheckAll =
      privilegesLength === (props.insideDb ? ['SELECT'] : DATABASE_PRIVILEGE_LIST).length;
    this.state = {
      checkAll: isCheckAll,
      indeterminate: privilegesLength !== 0 && !isCheckAll,
      checkedList: props?.dbPrivilegeParam?.privileges,
    };
  }

  onCheckAllChange = (e, dbName: string) => {
    const { insideDb, updateDbPrivilege } = this.props;
    const privilegeList = insideDb ? (['SELECT'] as API.DbPrivType[]) : DATABASE_PRIVILEGE_LIST;
    this.setState({
      checkedList: e.target.checked ? privilegeList : [],
      indeterminate: false,
      checkAll: e.target.checked,
    });

    updateDbPrivilege({
      dbName,
      // 由于 updateDbPrivilege 内部会对 privileges 做 splice 处理，因此需要解构赋值，避免影响原有的数组值
      // TODO: 待去掉 updateDbPrivilege 中的 splice 逻辑
      privileges: e.target.checked ? [...privilegeList] : [],
    });
  };

  collectCheckedParam = (checkedList: API.DbPrivType[], dbName: string) => {
    const { insideDb, updateDbPrivilege } = this.props;
    const privilegeList = insideDb ? (['SELECT'] as API.DbPrivType[]) : DATABASE_PRIVILEGE_LIST;
    this.setState({
      checkedList,
      indeterminate: !!checkedList.length && checkedList.length < privilegeList.length,
      checkAll: checkedList.length === privilegeList.length,
    });

    updateDbPrivilege({
      dbName,
      // 由于 updateDbPrivilege 内部会对 privileges 做 splice 处理，因此需要解构赋值，避免影响原有的数组值
      // TODO: 待去掉 updateDbPrivilege 中的 splice 逻辑
      privileges: [...checkedList],
    });
  };

  render() {
    const { dbPrivilegeParam, insideDb, deleteFn } = this.props;
    const { checkedList, indeterminate, checkAll } = this.state;
    return (
      <div className={styles.privilegedItem} key={dbPrivilegeParam.dbName}>
        <div className={styles.privilegedTitle}>
          <span style={{ fontWeight: 500 }}>{dbPrivilegeParam.dbName || '-'}</span>
          <div>
            <Checkbox
              indeterminate={indeterminate}
              onChange={(e) => this.onCheckAllChange(e, dbPrivilegeParam?.dbName)}
              checked={checkAll}
            >
              {formatMessage({
                id: 'ocp-express.component.DatabasePrivilegeTransfer.PrivilegesCheckbox.All',
                defaultMessage: '全部',
              })}
            </Checkbox>
            <Button
              className={styles.delete}
              onClick={() => {
                deleteFn(dbPrivilegeParam?.dbName);
              }}
            >
              <DeleteOutlined className={styles.deleteIcon} />
              {formatMessage({
                id: 'ocp-express.component.DatabasePrivilegeTransfer.PrivilegesCheckbox.Delete',
                defaultMessage: '删除',
              })}
            </Button>
          </div>
        </div>
        <Checkbox.Group
          style={{ display: 'inline-block' }}
          key={dbPrivilegeParam.dbName}
          value={checkedList}
          onChange={(value) =>
            this.collectCheckedParam(value as API.DbPrivType[], dbPrivilegeParam.dbName)
          }
        >
          {(insideDb ? ['SELECT'] : DATABASE_PRIVILEGE_LIST).map((node) => (
            <Checkbox className={styles.privilegeCheckbox} key={node} value={node}>
              {node}
            </Checkbox>
          ))}
        </Checkbox.Group>
      </div>
    );
  }
}
