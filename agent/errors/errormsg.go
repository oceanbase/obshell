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

package errors

import (
	"encoding/json"
	"fmt"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"

	"github.com/oceanbase/obshell/agent/bindata"
)

const errorsI18nEnResourceFile = "agent/assets/i18n/error/en.json"

var defaultLanguage = language.English
var bundle *i18n.Bundle

func init() {
	bundle = i18n.NewBundle(defaultLanguage)
	bundle.RegisterUnmarshalFunc("json", json.Unmarshal)
	loadBundleMessage(errorsI18nEnResourceFile)
}

func loadBundleMessage(assetName string) {
	asset, _ := bindata.Asset(assetName)
	bundle.MustParseMessageFileBytes(asset, assetName)
}

// GetMessage will get message from i18n bundle and format it with args
func GetMessage(lang language.Tag, errorCode ErrorCode, args []interface{}) string {
	localizer := i18n.NewLocalizer(bundle, lang.String())
	message, err := localizer.Localize(&i18n.LocalizeConfig{
		MessageID: errorCode.key,
	})
	if err != nil {
		return errorCode.key
	}
	return fmt.Sprintf(message, args...)
}
