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
func (csv *compiledSyncBase[N, Storage, Methods]) NewAccumulator(kvs attribute.Set) Accumulator {
	sc := &syncAccumulator[N, Storage, Methods]{}
	csv.initStorage(&sc.current)
	csv.initStorage(&sc.snapshot)

	holder := csv.findStorage(kvs)

	sc.holder = holder

	return sc
}

func (csv *compiledSyncBase[N, Storage, Methods]) findStorage(
	kvs attribute.Set,
) *storageHolder[Storage, int64] {
	kvs = csv.applyKeysFilter(kvs)

	csv.instLock.Lock()
	defer csv.instLock.Unlock()

	entry := csv.getOrCreateEntry(kvs)
	atomic.AddInt64(&entry.auxiliary, 1)
	return entry
}

// compiledAsyncBase is any asynchronous instrument view.
type compiledAsyncBase[N number.Any, Storage any, Methods aggregator.Methods[N, Storage]] struct {
	instrumentBase[N, Storage, notUsed, Methods]
}

// NewAccumulator returns a Accumulator for an asynchronous instrument view.
func (cav *compiledAsyncBase[N, Storage, Methods]) NewAccumulator(kvs attribute.Set) Accumulator {
	ac := &asyncAccumulator[N, Storage, Methods]{}

	ac.holder = cav.findStorage(kvs)

	return ac
}

func (cav *compiledAsyncBase[N, Storage, Methods]) findStorage(
	kvs attribute.Set,
) *storageHolder[Storage, notUsed] {
	kvs = cav.applyKeysFilter(kvs)

	cav.instLock.Lock()
	defer cav.instLock.Unlock()

	entry := cav.getOrCreateEntry(kvs)
	return entry
}

// multiAccumulator
type multiAccumulator[N number.Any] []Accumulator

func (acc multiAccumulator[N]) SnapshotAndProcess(final bool) {
	for _, coll := range acc {
		coll.SnapshotAndProcess(final)
	}
}

func (acc multiAccumulator[N]) Update(value N) {
	for _, coll := range acc {
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

func (acc *syncAccumulator[N, Storage, Methods]) Update(number N) {
	var methods Methods
	methods.Update(&acc.current, number)
}

func (acc *syncAccumulator[N, Storage, Methods]) SnapshotAndProcess(final bool) {
	var methods Methods
	acc.syncLock.Lock()
	defer acc.syncLock.Unlock()
	methods.Move(&acc.current, &acc.snapshot)
	methods.Merge(&acc.snapshot, &acc.holder.storage)
	if final {
		atomic.AddInt64(&acc.holder.auxiliary, -1)
	}
}

type notUsed struct{}

// asyncAccumulator
type asyncAccumulator[N number.Any, Storage any, Methods aggregator.Methods[N, Storage]] struct {
	asyncLock sync.Mutex
	current   N
	holder    *storageHolder[Storage, notUsed]
}

func (acc *asyncAccumulator[N, Storage, Methods]) Update(number N) {
	acc.asyncLock.Lock()
	defer acc.asyncLock.Unlock()
	acc.current = number
}

func (acc *asyncAccumulator[N, Storage, Methods]) SnapshotAndProcess(_ bool) {
	acc.asyncLock.Lock()
	defer acc.asyncLock.Unlock()

	var methods Methods
	methods.Update(&acc.holder.storage, acc.current)
}
