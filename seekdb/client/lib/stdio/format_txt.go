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

package stdio

// Define ANSI color codes.
const (
	RESET  = "\033[0m"
	RED    = "\033[31m"
	GREEN  = "\033[32m"
	YELLOW = "\033[33m"
	BLUE   = "\033[34m"
)

// FormattedText is a struct that holds text and its associated color.
type FormattedText struct {
	text      string
	colorText string
}

// NewFormattedText creates a new FormattedText with the given text and color.
func NewFormattedText(text, color string) *FormattedText {
	return &FormattedText{
		text:      text,
		colorText: color + text + RESET,
	}
}

// Format decides whether to return colored text or plain text based on istty.
func (ft *FormattedText) Format(istty bool) string {
	if istty {
		return ft.colorText
	}
	return ft.text
}

// String method to comply with the Stringer interface.
func (ft *FormattedText) String() string {
	return ft.Format(true) // Default to true for demonstration purposes
}

// Helper functions to create FormattedText with different colors.
func info(text string) *FormattedText {
	return NewFormattedText(text, BLUE)
}

func success(text string) *FormattedText {
	return NewFormattedText(text, GREEN)
}

func warn(text string) *FormattedText {
	return NewFormattedText(text, YELLOW)
}

func err(text string) *FormattedText {
	return NewFormattedText(text, RED)
}
