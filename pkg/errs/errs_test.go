// Copyright 2020 TiKV Project Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

package errs

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	. "github.com/pingcap/check"
	"github.com/pingcap/errors"
	"github.com/pingcap/log"
	"go.uber.org/zap"
)

// testingWriter is a WriteSyncer that writes to the the messages.
type testingWriter struct {
	messages []string
}

func newTestingWriter() *testingWriter {
	return &testingWriter{}
}

func (w *testingWriter) Write(p []byte) (n int, err error) {
	n = len(p)
	p = bytes.TrimRight(p, "\n")
	m := fmt.Sprintf("%s", p)
	w.messages = append(w.messages, m)
	return n, nil
}

func (w *testingWriter) Sync() error {
	return nil
}

type verifyLogger struct {
	*zap.Logger
	w *testingWriter
}

func (logger *verifyLogger) Message() string {
	if logger.w.messages == nil {
		return ""
	}
	return logger.w.messages[len(logger.w.messages)-1]
}

func newZapTestLogger(cfg *log.Config, opts ...zap.Option) verifyLogger {
	// TestingWriter is used to write to memory.
	// Used in the verify logger.
	writer := newTestingWriter()
	lg, _, _ := log.InitLoggerWithWriteSyncer(cfg, writer, opts...)

	return verifyLogger{
		Logger: lg,
		w:      writer,
	}
}

func Test(t *testing.T) {
	TestingT(t)
}

var _ = Suite(&testErrorSuite{})

type testErrorSuite struct{}

func (s *testErrorSuite) TestError(c *C) {
	conf := &log.Config{Level: "debug", File: log.FileLogConfig{}, DisableTimestamp: true}
	lg := newZapTestLogger(conf)
	log.ReplaceGlobals(lg.Logger, nil)

	rfc := `[error="[PD:tso:ErrInvalidTimestamp] invalid timestamp"]`
	log.Error("test", zap.Error(ErrInvalidTimestamp.FastGenByArgs()))
	c.Assert(strings.Contains(lg.Message(), rfc), IsTrue)
	err := errors.New("test error")
	log.Error("test", ZapError(ErrInvalidTimestamp, err))
	rfc = `[error="[PD:tso:ErrInvalidTimestamp] test error"]`
	fmt.Println(lg.Message())
	c.Assert(strings.Contains(lg.Message(), rfc), IsTrue)
}
