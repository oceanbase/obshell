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

package oracle

/*
#include <stdlib.h>

typedef struct st_mysql MYSQL;
typedef struct st_mysql_res MYSQL_RES;
typedef struct st_mysql_field {
    char *name;
    char *org_name;
    char *table;
    char *org_table;
    char *db;
    char *catalog;
    char *def;
    unsigned long length;
    unsigned long max_length;
    unsigned int name_length;
    unsigned int org_name_length;
    unsigned int table_length;
    unsigned int org_table_length;
    unsigned int db_length;
    unsigned int catalog_length;
    unsigned int def_length;
    unsigned int flags;
    unsigned int decimals;
    unsigned int charsetnr;
    int _type;
} MYSQL_FIELD;
typedef char **MYSQL_ROW;
typedef unsigned long long my_ulonglong;

// declare dynamic load functions
extern MYSQL* dynamic_mysql_init(MYSQL* mysql);
extern MYSQL* dynamic_mysql_real_connect(MYSQL* mysql, const char* host, const char* user, const char* passwd, const char* db, unsigned int port, const char* unix_socket, unsigned long clientflag);
extern int dynamic_mysql_query(MYSQL* mysql, const char* query);
extern MYSQL_RES* dynamic_mysql_store_result(MYSQL* mysql);
extern MYSQL_ROW dynamic_mysql_fetch_row(MYSQL_RES* result);
extern void dynamic_mysql_free_result(MYSQL_RES* result);
extern void dynamic_mysql_close(MYSQL* mysql);
extern const char* dynamic_mysql_error(MYSQL* mysql);
extern my_ulonglong dynamic_mysql_affected_rows(MYSQL* mysql);
extern unsigned int dynamic_mysql_num_fields(MYSQL_RES* result);
extern MYSQL_FIELD* dynamic_mysql_fetch_field_direct(MYSQL_RES* result, unsigned int fieldnr);
*/
import "C"

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"
	"unsafe"

	"github.com/oceanbase/obshell/agent/errors"
)

var driverRegistered sync.Once

func init() {
	driverRegistered.Do(func() {
		sql.Register("oracle", NewDriver())
	})
}

type Driver struct{}

func NewDriver() driver.Driver {
	return &Driver{}
}

// Open implements driver.Driver
func (d *Driver) Open(dsn string) (driver.Conn, error) {
	connector := NewConnector(dsn).(*Connector)
	return connector.Connect(nil)
}

type Connector struct {
	dsn    string
	handle *C.MYSQL
	cfg    *ConnConfig
}

func NewConnector(dsn string) driver.Connector {
	return &Connector{dsn: dsn}
}

func (c *Connector) ParseDsn(dsn string) error {
	cfg, err := ParseDSN(dsn)
	if err != nil {
		return err
	}
	c.cfg = cfg
	return nil
}

func (c *Connector) Connect(ctx context.Context) (driver.Conn, error) {
	if err := EnsureOracleLibraryLoaded(); err != nil {
		return nil, errors.Occur(errors.ErrOracleError, "failed to load Oracle client library: %s", err.Error())
	}

	if err := c.ParseDsn(c.dsn); err != nil {
		return nil, err
	}

	mysql := C.dynamic_mysql_init(nil)
	if mysql == nil {
		return nil, errors.Occur(errors.ErrOracleError, "failed to initialize connection")
	}

	var host *C.char = C.CString(c.cfg.Host)
	defer C.free(unsafe.Pointer(host))
	var port C.uint = C.uint(c.cfg.Port)
	var user *C.char = C.CString(c.cfg.User)
	defer C.free(unsafe.Pointer(user))
	var password *C.char = C.CString(c.cfg.Password)
	defer C.free(unsafe.Pointer(password))
	var database *C.char = C.CString(c.cfg.Database)
	defer C.free(unsafe.Pointer(database))

	c.handle = C.dynamic_mysql_real_connect(mysql, host, user, password, database, port, nil, 0)
	if c.handle == nil {
		errMsg := C.GoString(C.dynamic_mysql_error(mysql))
		return nil, errors.Occurf(errors.ErrOracleError, "failed to connect: %s", errMsg)
	}

	return &Conn{connector: c}, nil
}

func (c *Connector) Driver() driver.Driver {
	return NewDriver()
}

// Conn implements driver.Conn with dynamic loading
type Conn struct {
	connector *Connector
}

// Close closes the connection
func (c *Conn) Close() error {
	if c.connector.handle != nil {
		C.dynamic_mysql_close(c.connector.handle)
		c.connector.handle = nil
	}
	return nil
}

// Begin starts a transaction
func (c *Conn) Begin() (driver.Tx, error) {
	return nil, errors.Occur(errors.ErrOracleError, "transactions not implemented")
}

// Prepare returns a prepared statement
func (c *Conn) Prepare(query string) (driver.Stmt, error) {
	return nil, errors.Occur(errors.ErrOracleError, "prepared statements not implemented")
}

// Exec implements driver.Execer
func (c *Conn) Exec(query string, args []driver.Value) (driver.Result, error) {
	// 预处理参数
	modifiedQuery := c.preprocessQuery(query, args)

	cQuery := C.CString(modifiedQuery)
	defer C.free(unsafe.Pointer(cQuery))

	result := C.dynamic_mysql_query(c.connector.handle, cQuery)
	if result != 0 {
		errMsg := C.GoString(C.dynamic_mysql_error(c.connector.handle))
		return nil, errors.Occur(errors.ErrOracleError, errMsg)
	}

	rowsAffected := C.dynamic_mysql_affected_rows(c.connector.handle)
	return driver.RowsAffected(rowsAffected), nil
}

// Query implements driver.Queryer
func (c *Conn) Query(query string, args []driver.Value) (driver.Rows, error) {
	// 预处理参数
	modifiedQuery := c.preprocessQuery(query, args)

	cQuery := C.CString(modifiedQuery)
	defer C.free(unsafe.Pointer(cQuery))

	result := C.dynamic_mysql_query(c.connector.handle, cQuery)
	if result != 0 {
		errMsg := C.GoString(C.dynamic_mysql_error(c.connector.handle))
		return nil, errors.Occur(errors.ErrOracleError, errMsg)
	}

	rows := C.dynamic_mysql_store_result(c.connector.handle)
	if rows == nil {
		errMsg := C.GoString(C.dynamic_mysql_error(c.connector.handle))
		return nil, errors.Occur(errors.ErrOracleError, errMsg)
	}

	numFields := C.dynamic_mysql_num_fields(rows)
	fields := make([]*C.MYSQL_FIELD, int(numFields))
	for i := 0; i < int(numFields); i++ {
		fields[i] = C.dynamic_mysql_fetch_field_direct(rows, C.uint(i))
	}

	return &Rows{rows: rows, fields: fields}, nil
}

// preprocessQuery 预处理查询语句，将参数替换到查询中
func (c *Conn) preprocessQuery(query string, args []driver.Value) string {
	if len(args) == 0 {
		return query
	}

	modifiedQuery := query
	for _, arg := range args {
		switch v := arg.(type) {
		case string:
			// 转义单引号
			escaped := strings.ReplaceAll(v, "'", "''")
			modifiedQuery = strings.Replace(modifiedQuery, "?", fmt.Sprintf("'%s'", escaped), 1)
		case int64, int32, int16, int8, int:
			modifiedQuery = strings.Replace(modifiedQuery, "?", fmt.Sprintf("%v", v), 1)
		case float64, float32:
			modifiedQuery = strings.Replace(modifiedQuery, "?", fmt.Sprintf("%v", v), 1)
		case bool:
			if v {
				modifiedQuery = strings.Replace(modifiedQuery, "?", "1", 1)
			} else {
				modifiedQuery = strings.Replace(modifiedQuery, "?", "0", 1)
			}
		case time.Time:
			timeStr := v.Format("2006-01-02 15:04:05")
			modifiedQuery = strings.Replace(modifiedQuery, "?", fmt.Sprintf("'%s'", timeStr), 1)
		case nil:
			modifiedQuery = strings.Replace(modifiedQuery, "?", "NULL", 1)
		default:
			modifiedQuery = strings.Replace(modifiedQuery, "?", fmt.Sprintf("%v", v), 1)
		}
	}

	return modifiedQuery
}

type Rows struct {
	rows   *C.MYSQL_RES
	fields []*C.MYSQL_FIELD
}

func (r *Rows) Close() error {
	C.dynamic_mysql_free_result(r.rows)
	return nil
}

func (r *Rows) Columns() []string {
	numFields := C.dynamic_mysql_num_fields(r.rows)
	fields := make([]string, int(numFields))

	for i := 0; i < int(numFields); i++ {
		field := C.dynamic_mysql_fetch_field_direct(r.rows, C.uint(i))
		fields[i] = C.GoString(field.name)
	}

	return fields
}

func (r *Rows) isDateTimeField(fieldIndex int) bool {
	if fieldIndex >= len(r.fields) {
		return false
	}
	field := r.fields[fieldIndex]
	const (
		MYSQL_TYPE_DATE      = 10
		MYSQL_TYPE_TIME      = 11
		MYSQL_TYPE_DATETIME  = 12
		MYSQL_TYPE_TIMESTAMP = 7
	)

	fieldType := int(field._type)
	return fieldType == MYSQL_TYPE_DATE ||
		fieldType == MYSQL_TYPE_TIME ||
		fieldType == MYSQL_TYPE_DATETIME ||
		fieldType == MYSQL_TYPE_TIMESTAMP
}

func (r *Rows) parseDateTime(dateStr string) (time.Time, error) {
	formats := []string{
		"2006-01-02 15:04:05",      // YYYY-MM-DD HH:MM:SS
		"2006-01-02",               // YYYY-MM-DD
		"02-Jan-06",                // DD-MON-YY (Oracle standard)
		"02-Jan-2006",              // DD-MON-YYYY
		"02-Jan-06 15.04.05",       // DD-MON-YY HH.MM.SS
		"02-Jan-2006 15.04.05",     // DD-MON-YYYY HH.MM.SS
		"02-Jan-06 15:04:05.000",   // DD-MON-YY HH.MM.SS.fff
		"02-Jan-2006 15:04:05.000", // DD-MON-YYYY HH.MM.SS.fff
		time.RFC3339,               // ISO 8601
		time.RFC3339Nano,           // ISO 8601 with nanoseconds
	}

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t, nil
		}
	}

	return time.Time{}, errors.Occur(errors.ErrOracleError, "unable to parse date: "+dateStr)
}

func (r *Rows) Next(dest []driver.Value) error {
	row := C.dynamic_mysql_fetch_row(r.rows)
	if row == nil {
		return io.EOF
	}

	numFields := C.dynamic_mysql_num_fields(r.rows)
	for i := range dest {
		if i >= int(numFields) {
			break
		}

		rowPtr := unsafe.Pointer(uintptr(unsafe.Pointer(row)) + uintptr(i)*unsafe.Sizeof(unsafe.Pointer(nil)))
		cellPtr := *(*unsafe.Pointer)(rowPtr)

		if cellPtr == nil {
			dest[i] = nil
		} else {
			str := C.GoString((*C.char)(cellPtr))
			if r.isDateTimeField(i) {
				if parsedTime, err := r.parseDateTime(str); err == nil {
					dest[i] = parsedTime
				} else {
					dest[i] = str
				}
			} else {
				dest[i] = str
			}
		}
	}
	return nil
}
