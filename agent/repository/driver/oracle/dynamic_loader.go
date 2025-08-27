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
#include <dlfcn.h>
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

typedef MYSQL* (*mysql_init_func)(MYSQL* mysql);
typedef MYSQL* (*mysql_real_connect_func)(MYSQL* mysql, const char* host, const char* user, const char* passwd, const char* db, unsigned int port, const char* unix_socket, unsigned long clientflag);
typedef int (*mysql_query_func)(MYSQL* mysql, const char* query);
typedef MYSQL_RES* (*mysql_store_result_func)(MYSQL* mysql);
typedef MYSQL_ROW (*mysql_fetch_row_func)(MYSQL_RES* result);
typedef void (*mysql_free_result_func)(MYSQL_RES* result);
typedef void (*mysql_close_func)(MYSQL* mysql);
typedef const char* (*mysql_error_func)(MYSQL* mysql);
typedef my_ulonglong (*mysql_affected_rows_func)(MYSQL* mysql);
typedef unsigned int (*mysql_num_fields_func)(MYSQL_RES* result);
typedef MYSQL_FIELD* (*mysql_fetch_field_direct_func)(MYSQL_RES* result, unsigned int fieldnr);

static void* lib_handle = NULL;
static mysql_init_func p_mysql_init = NULL;
static mysql_real_connect_func p_mysql_real_connect = NULL;
static mysql_query_func p_mysql_query = NULL;
static mysql_store_result_func p_mysql_store_result = NULL;
static mysql_fetch_row_func p_mysql_fetch_row = NULL;
static mysql_free_result_func p_mysql_free_result = NULL;
static mysql_close_func p_mysql_close = NULL;
static mysql_error_func p_mysql_error = NULL;
static mysql_affected_rows_func p_mysql_affected_rows = NULL;
static mysql_num_fields_func p_mysql_num_fields = NULL;
static mysql_fetch_field_direct_func p_mysql_fetch_field_direct = NULL;

int load_oracle_lib(const char* lib_path) {
    if (lib_handle != NULL) {
        return 0;
    }

    lib_handle = dlopen(lib_path, RTLD_LAZY);
    if (!lib_handle) {
        return -1;
    }

    p_mysql_init = (mysql_init_func)dlsym(lib_handle, "mysql_init");
    if (!p_mysql_init) goto error;

    p_mysql_real_connect = (mysql_real_connect_func)dlsym(lib_handle, "mysql_real_connect");
    if (!p_mysql_real_connect) goto error;

    p_mysql_query = (mysql_query_func)dlsym(lib_handle, "mysql_query");
    if (!p_mysql_query) goto error;

    p_mysql_store_result = (mysql_store_result_func)dlsym(lib_handle, "mysql_store_result");
    if (!p_mysql_store_result) goto error;

    p_mysql_fetch_row = (mysql_fetch_row_func)dlsym(lib_handle, "mysql_fetch_row");
    if (!p_mysql_fetch_row) goto error;

    p_mysql_free_result = (mysql_free_result_func)dlsym(lib_handle, "mysql_free_result");
    if (!p_mysql_free_result) goto error;

    p_mysql_close = (mysql_close_func)dlsym(lib_handle, "mysql_close");
    if (!p_mysql_close) goto error;

    p_mysql_error = (mysql_error_func)dlsym(lib_handle, "mysql_error");
    if (!p_mysql_error) goto error;

    p_mysql_affected_rows = (mysql_affected_rows_func)dlsym(lib_handle, "mysql_affected_rows");
    if (!p_mysql_affected_rows) goto error;

    p_mysql_num_fields = (mysql_num_fields_func)dlsym(lib_handle, "mysql_num_fields");
    if (!p_mysql_num_fields) goto error;

    p_mysql_fetch_field_direct = (mysql_fetch_field_direct_func)dlsym(lib_handle, "mysql_fetch_field_direct");
    if (!p_mysql_fetch_field_direct) goto error;

    return 0;

error:
    if (lib_handle) {
        dlclose(lib_handle);
        lib_handle = NULL;
    }
    return -1;
}

void unload_oracle_lib() {
    if (lib_handle) {
        dlclose(lib_handle);
        lib_handle = NULL;
    }
}

int is_oracle_lib_loaded() {
    return lib_handle != NULL ? 1 : 0;
}

MYSQL* dynamic_mysql_init(MYSQL* mysql) {
    if (!p_mysql_init) return NULL;
    return p_mysql_init(mysql);
}

MYSQL* dynamic_mysql_real_connect(MYSQL* mysql, const char* host, const char* user, const char* passwd, const char* db, unsigned int port, const char* unix_socket, unsigned long clientflag) {
    if (!p_mysql_real_connect) return NULL;
    return p_mysql_real_connect(mysql, host, user, passwd, db, port, unix_socket, clientflag);
}

int dynamic_mysql_query(MYSQL* mysql, const char* query) {
    if (!p_mysql_query) return -1;
    return p_mysql_query(mysql, query);
}

MYSQL_RES* dynamic_mysql_store_result(MYSQL* mysql) {
    if (!p_mysql_store_result) return NULL;
    return p_mysql_store_result(mysql);
}

MYSQL_ROW dynamic_mysql_fetch_row(MYSQL_RES* result) {
    if (!p_mysql_fetch_row) return NULL;
    return p_mysql_fetch_row(result);
}

void dynamic_mysql_free_result(MYSQL_RES* result) {
    if (p_mysql_free_result) {
        p_mysql_free_result(result);
    }
}

void dynamic_mysql_close(MYSQL* mysql) {
    if (p_mysql_close) {
        p_mysql_close(mysql);
    }
}

const char* dynamic_mysql_error(MYSQL* mysql) {
    if (!p_mysql_error) return "Library not loaded";
    return p_mysql_error(mysql);
}

my_ulonglong dynamic_mysql_affected_rows(MYSQL* mysql) {
    if (!p_mysql_affected_rows) return 0;
    return p_mysql_affected_rows(mysql);
}

unsigned int dynamic_mysql_num_fields(MYSQL_RES* result) {
    if (!p_mysql_num_fields) return 0;
    return p_mysql_num_fields(result);
}

MYSQL_FIELD* dynamic_mysql_fetch_field_direct(MYSQL_RES* result, unsigned int fieldnr) {
    if (!p_mysql_fetch_field_direct) return NULL;
    return p_mysql_fetch_field_direct(result, fieldnr);
}
*/
import "C"

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"unsafe"
)

var (
	libraryMutex sync.Mutex
	isLoaded     bool
)

const (
	DefaultLibPath = "/usr/lib64/libobclnt.so"
	LibPathEnvVar  = "OBCLNT_LIB_PATH"
)

func LoadOracleLibrary() error {
	libraryMutex.Lock()
	defer libraryMutex.Unlock()

	if isLoaded {
		return nil
	}

	libPath := os.Getenv(LibPathEnvVar)
	if libPath == "" {
		libPath = DefaultLibPath
	}

	if _, err := os.Stat(libPath); os.IsNotExist(err) {
		return fmt.Errorf("Oracle client library not found at %s", libPath)
	}

	absPath, err := filepath.Abs(libPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %v", err)
	}

	cPath := C.CString(absPath)
	defer C.free(unsafe.Pointer(cPath))

	result := C.load_oracle_lib(cPath)
	if result != 0 {
		return fmt.Errorf("failed to load Oracle client library from %s", absPath)
	}

	isLoaded = true
	return nil
}

func UnloadOracleLibrary() {
	libraryMutex.Lock()
	defer libraryMutex.Unlock()

	if isLoaded {
		C.unload_oracle_lib()
		isLoaded = false
	}
}

func IsOracleLibraryLoaded() bool {
	libraryMutex.Lock()
	defer libraryMutex.Unlock()

	return isLoaded && C.is_oracle_lib_loaded() == 1
}

func EnsureOracleLibraryLoaded() error {
	if IsOracleLibraryLoaded() {
		return nil
	}
	return LoadOracleLibrary()
}
