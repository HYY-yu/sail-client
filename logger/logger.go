// The MIT License (MIT)
//
//Copyright (c) 2014 Steve Francia
//
//Permission is hereby granted, free of charge, to any person obtaining a copy
//of this software and associated documentation files (the "Software"), to deal
//in the Software without restriction, including without limitation the rights
//to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
//copies of the Software, and to permit persons to whom the Software is
//furnished to do so, subject to the following conditions:
//
//The above copyright notice and this permission notice shall be included in all
//copies or substantial portions of the Software.
//
//THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
//IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
//FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
//AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
//LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
//OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
//SOFTWARE.

package logger

import (
	"fmt"

	jww "github.com/spf13/jwalterweatherman"
)

// Logger is a unified interface for various logging use cases and practices, including:
//   - leveled logging
//   - structured logging
type Logger interface {
	// Trace logs a Trace event.
	//
	// Even more fine-grained information than Debug events.
	// Loggers not supporting this level should fall back to Debug.
	Trace(msg string, keyvals ...interface{})

	// Debug logs a Debug event.
	//
	// A verbose series of information events.
	// They are useful when debugging the system.
	Debug(msg string, keyvals ...interface{})

	// Info logs an Info event.
	//
	// General information about what's happening inside the system.
	Info(msg string, keyvals ...interface{})

	// Warn logs a Warn(ing) event.
	//
	// Non-critical events that should be looked at.
	Warn(msg string, keyvals ...interface{})

	// Error logs an Error event.
	//
	// Critical events that require immediate attention.
	// Loggers commonly provide Fatal and Panic levels above Error level,
	// but exiting and panicing is out of scope for a logging library.
	Error(msg string, keyvals ...interface{})
}

type jwwLogger struct{}

func New() Logger {
	return jwwLogger{}
}

func init() {
	jww.SetPrefix("sail-client")
}

func (jwwLogger) Trace(msg string, keyvals ...interface{}) {
	jww.TRACE.Printf(jwwLogMessage(msg, keyvals...))
}

func (jwwLogger) Debug(msg string, keyvals ...interface{}) {
	jww.DEBUG.Printf(jwwLogMessage(msg, keyvals...))
}

func (jwwLogger) Info(msg string, keyvals ...interface{}) {
	jww.INFO.Printf(jwwLogMessage(msg, keyvals...))
}

func (jwwLogger) Warn(msg string, keyvals ...interface{}) {
	jww.WARN.Printf(jwwLogMessage(msg, keyvals...))
}

func (jwwLogger) Error(msg string, keyvals ...interface{}) {
	jww.ERROR.Printf(jwwLogMessage(msg, keyvals...))
}

func jwwLogMessage(msg string, keyvals ...interface{}) string {
	out := msg

	if len(keyvals) > 0 && len(keyvals)%2 == 1 {
		keyvals = append(keyvals, nil)
	}

	for i := 0; i <= len(keyvals)-2; i += 2 {
		out = fmt.Sprintf("%s %v=%v", out, keyvals[i], keyvals[i+1])
	}

	return out
}
