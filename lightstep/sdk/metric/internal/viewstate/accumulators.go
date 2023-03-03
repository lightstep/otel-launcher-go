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

package viewstate // import "github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/internal/viewstate"

import (
	"sync"
	"sync/atomic"

	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/aggregator"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/number"
	"go.opentelemetry.io/otel/attribute"
)

// compiledSyncBase is any synchronous instrument view.
type compiledSyncBase[N number.Any, Storage any, Methods aggregator.Methods[N, Storage]] struct {
	instrumentBase[N, Storage, int64, Methods]
}

// NewAccumulator returns a Accumulator for a synchronous instrument view.
func (c *compiledSyncBase[N, Storage, Methods]) NewAccumulator(kvs attribute.Set) Accumulator {
	sc := &syncAccumulator[N, Storage, Methods]{}
	c.initStorage(&sc.current)
	c.initStorage(&sc.snapshot)

	sc.holder = c.findStorage(kvs)
	return sc
}

// findStorage locates the output Storage and adds to the auxiliary
// reference count for synchronous instruments.
func (c *compiledSyncBase[N, Storage, Methods]) findStorage(
	kvs attribute.Set,
) *storageHolder[Storage, int64] {
	kvs = c.applyKeysFilter(kvs)

	c.instLock.Lock()
	defer c.instLock.Unlock()

	entry := c.getOrCreateEntry(kvs)
	atomic.AddInt64(&entry.auxiliary, 1)
	return entry
}

// compiledAsyncBase is any asynchronous instrument view.
type compiledAsyncBase[N number.Any, Storage any, Methods aggregator.Methods[N, Storage]] struct {
	instrumentBase[N, Storage, notUsed, Methods]
}

// NewAccumulator returns a Accumulator for an asynchronous instrument view.
func (c *compiledAsyncBase[N, Storage, Methods]) NewAccumulator(kvs attribute.Set) Accumulator {
	ac := &asyncAccumulator[N, Storage, Methods]{}

	ac.holder = c.findStorage(kvs)
	return ac
}

// findStorage locates the output Storage for asynchronous instruments.
func (c *compiledAsyncBase[N, Storage, Methods]) findStorage(
	kvs attribute.Set,
) *storageHolder[Storage, notUsed] {
	kvs = c.applyKeysFilter(kvs)

	c.instLock.Lock()
	defer c.instLock.Unlock()

	return c.getOrCreateEntry(kvs)
}

// multiAccumulator
type multiAccumulator[N number.Any] []Accumulator

func (a multiAccumulator[N]) SnapshotAndProcess(release bool) {
	for _, coll := range a {
		coll.SnapshotAndProcess(release)
	}
}

func (a multiAccumulator[N]) Update(value N) {
	for _, coll := range a {
		coll.(Updater[N]).Update(value)
	}
}

// syncAccumulator
type syncAccumulator[N number.Any, Storage any, Methods aggregator.Methods[N, Storage]] struct {
	// syncLock prevents two readers from calling
	// SnapshotAndProcess at the same moment.
	syncLock sync.Mutex
	current  Storage
	snapshot Storage
	holder   *storageHolder[Storage, int64]
}

func (a *syncAccumulator[N, Storage, Methods]) Update(number N) {
	var methods Methods
	methods.Update(&a.current, number)
}

func (a *syncAccumulator[N, Storage, Methods]) SnapshotAndProcess(release bool) {
	var methods Methods
	a.syncLock.Lock()
	defer a.syncLock.Unlock()
	methods.Move(&a.current, &a.snapshot)
	methods.Merge(&a.snapshot, &a.holder.storage)
	if release {
		// On the final snapshot-and-process, decrement the auxiliary reference count.
		atomic.AddInt64(&a.holder.auxiliary, -1)
	}
}

// asyncAccumulator
type asyncAccumulator[N number.Any, Storage any, Methods aggregator.Methods[N, Storage]] struct {
	asyncLock sync.Mutex
	current   N
	holder    *storageHolder[Storage, notUsed]
}

func (a *asyncAccumulator[N, Storage, Methods]) Update(number N) {
	a.asyncLock.Lock()
	defer a.asyncLock.Unlock()
	a.current = number
}

func (a *asyncAccumulator[N, Storage, Methods]) SnapshotAndProcess(_ bool) {
	a.asyncLock.Lock()
	defer a.asyncLock.Unlock()

	var methods Methods
	methods.Update(&a.holder.storage, a.current)
}
