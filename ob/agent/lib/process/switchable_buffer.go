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

package process

import (
	"bytes"
	"io"
	"os"
)

type switchableBuffer struct {
	memBuf   bytes.Buffer
	fileBuf  *os.File
	out      io.Writer
	memModel bool
}

func newSwitchableBuffer(fileBuffer *os.File, output io.Writer) (b *switchableBuffer, err error) {
	return &switchableBuffer{
		fileBuf:  fileBuffer,
		out:      output,
		memModel: true,
	}, nil
}

func (b *switchableBuffer) Write(p []byte) (n int, err error) {
	var buf io.Writer
	if b.memModel {
		buf = &b.memBuf
	} else {
		buf = b.fileBuf
		if stat, err := b.fileBuf.Stat(); err == nil {
			if stat.Size() > 128*1024*1024 {
				b.fileBuf.Truncate(0)
			}
		}
	}
	return buf.Write(p)
}

func (b *switchableBuffer) Flush() (err error) {
	str := b.memBuf.Bytes()
	b.memBuf.Reset()
	_, err = b.out.Write(str)
	return
}
