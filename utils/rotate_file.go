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

package utils

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"syscall"
	"time"
)

const (
	backupTimeFormat = "2006-01-02T15-04-05.000"
	defaultMaxSize   = 128

	megabyte = int64(1 << 20)
	oneDay   = time.Hour * 24
)

type RotateFile struct {
	fileName string
	dir      string
	ext      string
	prefix   string

	maxSize    int64 // in bytes
	maxAge     time.Duration
	maxBackups int

	size int64
	file *os.File
	mu   sync.Mutex

	cleanChan    chan bool
	startCleaner sync.Once
}

// NewRotateFile creates a new RotateFile.
// The maxSize is the maximum size in megabytes of the log file before it gets rotated. It defaults to 128 megabytes.
// The maxAge is the maximum number of days to retain old log files based on the timestamp encoded in their filename.
// If maxAge is 0, old files are not removed. It defaults to 0.
// The maxBackups is the maximum number of old log files to retain.
// If maxBackups is 0, all old log files are retained. It defaults to 0.
func NewRotateFile(fileName string, maxSize int64, maxAge, maxBackups int) *RotateFile {
	if maxSize == 0 {
		maxSize = defaultMaxSize
	}

	name := filepath.Base(fileName)
	ext := filepath.Ext(name)
	prefix := name[:len(name)-len(ext)] + "-"

	rf := &RotateFile{
		fileName:     fileName,
		dir:          filepath.Dir(fileName),
		ext:          ext,
		prefix:       prefix,
		maxSize:      maxSize * megabyte,
		maxAge:       oneDay * time.Duration(maxAge),
		maxBackups:   maxBackups,
		mu:           sync.Mutex{},
		startCleaner: sync.Once{},
	}
	return rf
}

func (rf *RotateFile) Write(p []byte) (n int, err error) {
	rf.mu.Lock()
	defer rf.mu.Unlock()

	writeLen := int64(len(p))
	if writeLen > rf.maxSize {
		return 0, fmt.Errorf(
			"write length %d exceeds maximum file size %d", writeLen, rf.maxSize,
		)
	}

	if rf.file == nil {
		if err = rf.openExistingOrNew(writeLen); err != nil {
			return
		}
	}

	if rf.size+writeLen > rf.maxSize {
		if err := rf.rotate(); err != nil {
			return 0, err
		}
	}

	n, err = rf.file.Write(p)
	rf.size += int64(n)
	return
}

// close closes the file if it is open.
func (rf *RotateFile) close() error {
	if rf.file == nil {
		return nil
	}
	err := rf.file.Close()
	rf.file = nil
	return err
}

func (rf *RotateFile) Close() error {
	rf.mu.Lock()
	defer rf.mu.Unlock()

	return rf.close()
}

func (rf *RotateFile) backupName() string {
	t := time.Now()
	timestamp := t.Format(backupTimeFormat)
	return filepath.Join(rf.dir, fmt.Sprintf("%s%s%s", rf.prefix, timestamp, rf.ext))
}

func (rf *RotateFile) chown(info os.FileInfo) error {
	f, err := os.OpenFile(rf.fileName, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, info.Mode())
	if err != nil {
		return err
	}
	f.Close()
	stat := info.Sys().(*syscall.Stat_t)
	return os.Chown(rf.fileName, int(stat.Uid), int(stat.Gid))
}

func (rf *RotateFile) openExistingOrNew(writeLen int64) error {
	rf.clean()

	info, err := os.Stat(rf.fileName)
	if err != nil {
		if os.IsNotExist(err) {
			return rf.openNew()
		}
		return fmt.Errorf("error getting log file info: %s", err)
	}

	if info.Size()+writeLen >= int64(rf.maxSize) {
		return rf.rotate()
	}

	file, err := os.OpenFile(rf.fileName, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		// if we fail to open the old log file for some reason, just ignore
		// it and open a new log file.
		return rf.openNew()
	}
	rf.file = file
	rf.size = info.Size()
	return nil
}

func (rf *RotateFile) rotate() error {
	if err := rf.close(); err != nil {
		return err
	}

	if err := rf.backup(); err != nil {
		return err
	}

	if err := rf.openNew(); err != nil {
		return err
	}
	return nil
}

func (rf *RotateFile) backup() error {
	info, err := os.Stat(rf.fileName)
	if err == nil {
		newname := rf.backupName()
		if err := os.Rename(rf.fileName, newname); err != nil {
			return fmt.Errorf("can't rename log file: %s", err)
		}

		if err := rf.chown(info); err != nil {
			return err
		}
		rf.clean()
	}
	return nil
}

func (rf *RotateFile) openNew() error {
	err := os.MkdirAll(rf.dir, 0755)
	if err != nil {
		return fmt.Errorf("can't make directories for new logfile: %s", err)
	}

	mode := os.FileMode(0644)
	info, err := os.Stat(rf.fileName)
	if err == nil {
		mode = info.Mode()
	}

	f, err := os.OpenFile(rf.fileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, mode)
	if err != nil {
		return fmt.Errorf("can't open new logfile: %s", err)
	}
	rf.file = f
	rf.size = 0
	return nil
}

func (rf *RotateFile) timeFromName(filename string) (time.Time, error) {
	if !strings.HasPrefix(filename, rf.prefix) {
		return time.Time{}, errors.New("mismatched prefix")
	}
	if !strings.HasSuffix(filename, rf.ext) {
		return time.Time{}, errors.New("mismatched extension")
	}
	ts := filename[len(rf.prefix) : len(filename)-len(rf.ext)]
	return time.ParseInLocation(backupTimeFormat, ts, time.Local)
}

// getRotatedLogFiles returns the list of rotated log files.
// ascending order of the creation time.
func (rf *RotateFile) getRotatedLogFiles() ([]logFileInfo, error) {
	files, err := os.ReadDir(rf.dir)
	if err != nil {
		return nil, err
	}

	logFiles := make([]logFileInfo, 0, len(files))
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		if t, err := rf.timeFromName(f.Name()); err == nil {
			logFiles = append(logFiles, logFileInfo{file: f, modTime: t})
		}
	}

	sortLogFiles(logFiles)
	return logFiles, nil
}

func (rf *RotateFile) clean() {
	rf.startCleaner.Do(func() {
		rf.cleanChan = make(chan bool, 1)
		go rf.cleanLoop()
	})
	select {
	case rf.cleanChan <- true:
	default:
	}
}

func (rf *RotateFile) cleanLoop() {
	for range rf.cleanChan {
		rf.doClean()
	}
}

func (rf *RotateFile) doClean() error {
	if rf.maxAge == 0 && rf.maxBackups == 0 {
		return nil
	}

	files, err := rf.getRotatedLogFiles()
	if err != nil {
		return err
	}

	now := time.Now()
	cutoff := now.Add(-1 * rf.maxAge)

	var needRemove []logFileInfo
	if rf.maxBackups > 0 && len(files) > rf.maxBackups-1 {
		needRemove = files[:len(files)-rf.maxBackups+1]
		files = files[len(files)-rf.maxBackups+1:]
	}

	if rf.maxAge > 0 {
		for _, fi := range files {
			if fi.modTime.Before(cutoff) {
				needRemove = append(needRemove, fi)
			} else {
				break
			}
		}
	}

	for _, fi := range needRemove {
		if err1 := os.Remove(filepath.Join(rf.dir, fi.file.Name())); err1 != nil {
			err = err1
		}
	}

	return err
}

type logFileInfo struct {
	file    fs.DirEntry
	modTime time.Time
}

func sortLogFiles(files []logFileInfo) {
	sort.Slice(files, func(i, j int) bool {
		return files[i].modTime.Before(files[j].modTime)
	})
}
