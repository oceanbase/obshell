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

package constant

var (
	val_true  = true
	val_false = false
	nil_str   = ""
	nil_int   = 0

	PTR_TRUE    = &val_true
	PTR_FALSE   = &val_false
	PTR_NIL_STR = &nil_str
	PTR_NIL_INT = &nil_int
)
