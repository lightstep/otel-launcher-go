// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package global // import "github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/internal/global"

import (
	"log"
	"os"
	"sync"

	"github.com/go-logr/logr"
	"github.com/go-logr/stdr"
)

// globalLogger is the logging interface used within the otel api and sdk provide deatails of the internals.
//
// The default logger uses stdr which is backed by the standard `log.Logger`
// interface. This logger will only show messages at the Error Level.
var globalLogger logr.Logger = stdr.New(log.New(os.Stderr, "", log.LstdFlags|log.Lshortfile))
var globalLoggerLock = &sync.RWMutex{}

// SetLogger overrides the globalLogger with l.
//
// To see Info messages use a logger with `l.V(1).Enabled() == true`
// To see Debug messages use a logger with `l.V(5).Enabled() == true`.
func SetLogger(l logr.Logger) {
	globalLoggerLock.Lock()
	defer globalLoggerLock.Unlock()
	globalLogger = l
}

// Info prints messages about the general state of the API or SDK.
// This should usually be less then 5 messages a minute.
func Info(msg string, keysAndValues ...interface{}) {
	globalLoggerLock.RLock()
	defer globalLoggerLock.RUnlock()
	globalLogger.V(1).Info(msg, keysAndValues...)
}

// Error prints messages about exceptional states of the API or SDK.
func Error(err error, msg string, keysAndValues ...interface{}) {
	globalLoggerLock.RLock()
	defer globalLoggerLock.RUnlock()
	globalLogger.Error(err, msg, keysAndValues...)
}

// Debug prints messages about all internal changes in the API or SDK.
func Debug(msg string, keysAndValues ...interface{}) {
	globalLoggerLock.RLock()
	defer globalLoggerLock.RUnlock()
	globalLogger.V(5).Info(msg, keysAndValues...)
}
