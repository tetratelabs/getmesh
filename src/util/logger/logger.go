// Copyright 2021 Tetrate
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package logger

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"sync"
)

var l = &logger{w: os.Stdout, mux: &sync.Mutex{}}

type logger struct {
	w   io.Writer
	mux *sync.Mutex
}

func Infof(format string, v ...interface{}) {
	_, _ = l.w.Write([]byte(fmt.Sprintf(format, v...)))
}

func Warnf(format string, v ...interface{}) {
	base := fmt.Sprintf("[WARNING] %s", format)
	_, _ = l.w.Write([]byte(fmt.Sprintf(base, v...)))
}

func Errorf(format string, v ...interface{}) {
	base := fmt.Sprintf("[ERROR] %s", format)
	_, _ = l.w.Write([]byte(fmt.Sprintf(base, v...)))
}

func Lock() {
	l.mux.Lock()
}

func Unlock() {
	l.mux.Unlock()
}

func SetWriter(w io.Writer) {
	l.w = w
}

func GetWriter() io.Writer {
	return l.w
}

func ExecuteWithLock(f func()) *bytes.Buffer {
	Lock()
	defer Unlock()
	buf := new(bytes.Buffer)
	SetWriter(buf)
	f()
	return buf
}
