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
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/aggregator"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/aggregator/aggregation"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/aggregator/gauge"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/aggregator/histogram"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/aggregator/minmaxsumcount"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/aggregator/sum"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/data"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/number"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/sdkinstrument"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/view"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/instrumentation"
)

// Compiler implements Views for a single Meter.  A single Compiler
// controls the namespace of metric instruments output and reports
// conflicting definitions for the same name.
//
// Information flows through the Compiler as follows:
//
// When new instruments are created:
// - The Compiler.Compile() method returns an Instrument value and possible
//   duplicate or semantic conflict error.
//
// When instruments are used:
// - The Instrument.NewAccumulator() method returns an Accumulator value for each attribute.Set used
// - The Accumulator.Update() aggregates one value for each measurement.
//
// During collection:
// - The Accumulator.SnapshotAndProcess() method captures the current value
//   and conveys it to the output storage
// - The Compiler.Collectors() interface returns one Collector per output
//   Metric in the Meter (duplicate definitions included).
// - The Collector.Collect() method outputs one Point for each attribute.Set
//   in the result.
type Compiler struct {
	// views is the configuration of this compiler.
	views *view.Views

	// library is the value used fr matching
	// instrumentation library information.
	library instrumentation.Library

	// lock protects collectors and names.
	compilerLock sync.Mutex

	// collectors is the de-duplicated list of metric outputs, which may
	// contain conflicting identities.
	collectors []data.Collector

	// names is the map of output names for metrics
	// produced by this compiler.
	names map[string][]leafInstrument
}

// Instrument is a compiled implementation of an instrument
// corresponding with one or more instrument-view behaviors.
type Instrument interface {
	// NewAccumulator returns an Accumulator and an Updater[N]
	// matching the number type of the API-level instrument.
	//
	// Callers are expected to type-assert Updater[int64] or
	// Updater[float64] before calling Update().
	//
	// The caller's primary responsibility is to maintain
	// the collection of Accumulators that had Update()
	// called since the last collection and to ensure that each
	// of them has SnapshotAndProcess() called.
	NewAccumulator(kvs attribute.Set) Accumulator
}

// Updater captures single measurements, for N an int64 or float64.
type Updater[N number.Any] interface {
	// Update captures a single measurement.  For synchronous
	// instruments, this passes directly through to the
	// aggregator.  For asynchronous instruments, the last value
	// is captured by the accumulator snapshot.
	Update(value N)
}

// Accumulator is an intermediate interface used for short-term
// aggregation.  Every Accumulator is also an Updater.  The owner of
// an Accumulator is responsible for maintaining the current set
// of Accumulators, defined as those which have been Updated and not
// yet had SnapshotAndProcess() called.
type Accumulator interface {
	// SnapshotAndProcess() takes a snapshot of data aggregated
	// through Update() and simultaneously resets the current
	// aggregator.  The attribute.Set is possibly filtered, after
	// which the snapshot is merged into the output.
	//
	// There is no return value from this method; the caller can
	// safely forget an Accumulator after this method is called,
	// provided Update is not used again.
	//
	// When `release` is true, this is the last time the Accumulator
	// will be snapshot/processed (according to the caller's
	// reference counting) and it can be forgotten.
	SnapshotAndProcess(release bool)
}

// leafInstrument is one of the (synchronous or asynchronous),
// (cumulative or delta) instrument implementations.  This is used in
// duplicate conflict detection and resolution.
type leafInstrument interface {
	// Instrument is the form returned by Compile().
	Instrument
	// Collector is the form returned in Collectors().
	data.Collector
	// Duplicate is how other instruments this in a conflict.
	Duplicate

	// mergeDescription handles the special case allowing
	// descriptions to be merged instead of conflict.
	mergeDescription(string)
}

// singleBehavior is one instrument-view behavior, including the
// original instrument details, the aggregation kind and temporality,
// aggregator configuration, and optional keys to filter.
type singleBehavior struct {
	// fromName is the original instrument name
	fromName string

	// desc is the output of the view, including naming,
	// description and unit.  This includes the original
	// instrument's instrument kind and number kind.
	desc sdkinstrument.Descriptor

	// kind is the aggregation indicated by this view behavior.
	kind aggregation.Kind

	// tempo is the configured aggregation temporality.
	tempo aggregation.Temporality

	// acfg is the aggregator configuration.
	acfg aggregator.Config

	// keysSet (if non-nil) is an attribute set containing each
	// key being filtered with a zero value.  This is used to
	// compare against potential duplicates for having the
	// same/different filter.
	keysSet *attribute.Set // With Int(0)

	// keysFilter (if non-nil) is the constructed keys filter.
	keysFilter *attribute.Filter

	// hinted is true when the aggregation was set
	// programmatically via a hint. this bypasses semantic
	// compatibility checking and allows hints to create a
	// synchronous gauge instrument, for example.
	hinted bool
}

// New returns a compiler for library given configured views.
func New(library instrumentation.Library, views *view.Views) *Compiler {
	return &Compiler{
		library: library,
		views:   views,
		names:   map[string][]leafInstrument{},
	}
}

func (v *Compiler) Collectors() []data.Collector {
	v.compilerLock.Lock()
	defer v.compilerLock.Unlock()
	return v.collectors
}

// tryToApplyHint looks for a Lightstep-specified hint structure
// encoded as JSON in the description.  If valid, returns the modified
// configuration, otherwise returns the default for the instrument.
func (v *Compiler) tryToApplyHint(instrument sdkinstrument.Descriptor) (_ sdkinstrument.Descriptor, akind aggregation.Kind, acfg aggregator.Config, hinted bool) {
	// These are the default behaviors, we'll use them unless there's a valid hint.
	akind = v.views.Defaults.Aggregation(instrument.Kind)
	acfg = v.views.Defaults.AggregationConfig(
		instrument.Kind,
		instrument.NumberKind,
	)

	// Check for required JSON symbols, empty strings, ...
	if !strings.Contains(instrument.Description, "{") {
		return instrument, akind, acfg, hinted
	}

	var hint view.Hint
	if err := json.Unmarshal([]byte(instrument.Description), &hint); err != nil {
		// This could be noisy if valid descriptions contain spurious '{' chars.
		otel.Handle(fmt.Errorf("hint parse error: %w", err))
		return instrument, akind, acfg, hinted
	}

	// Replace the hint input with its embedded description.
	instrument.Description = hint.Description

	// Bypass semantic compatibility check.
	hinted = true

	// Potentially set the aggregation kind and aggregator config.
	if hint.Aggregation != "" {
		parseKind, ok := aggregation.ParseKind(hint.Aggregation)
		if !ok {
			otel.Handle(fmt.Errorf("hint invalid aggregation: %v", hint.Aggregation))
		} else if parseKind != aggregation.UndefinedKind {
			akind = parseKind
		}
	}

	if hint.Config != (aggregator.JSONConfig{}) {
		cfg := hint.Config.ToConfig()
		cfg, err := cfg.Validate()
		if err != nil {
			otel.Handle(fmt.Errorf("hint invalid aggregator config: %w", err))
		}
		acfg = cfg
	}

	return instrument, akind, acfg, hinted
}

// Compile is called during NewInstrument by the Meter
// implementation, the result saved in the instrument and used to
// construct new Accumulators throughout its lifetime.
func (v *Compiler) Compile(instrument sdkinstrument.Descriptor) (Instrument, ViewConflictsBuilder) {
	var behaviors []singleBehavior
	var matches []view.ClauseConfig

	for _, view := range v.views.Clauses {
		if !view.Matches(v.library, instrument) {
			continue
		}
		matches = append(matches, view)
	}

	for _, view := range matches {
		akind := view.Aggregation()
		if akind == aggregation.DropKind {
			continue
		}

		modified, hintAkind, hintAcfg, hinted := v.tryToApplyHint(instrument)
		instrument = modified // the hint erases itself from the description

		if akind == aggregation.UndefinedKind {
			akind = hintAkind
		}

		cf := singleBehavior{
			fromName: instrument.Name,
			desc:     viewDescriptor(instrument, view),
			kind:     akind,
			acfg:     pickAggConfig(hintAcfg, view.AggregatorConfig()),
			tempo:    v.views.Defaults.Temporality(instrument.Kind),
			hinted:   hinted,
		}

		keys := view.Keys()
		if keys != nil {
			cf.keysSet = keysToSet(view.Keys())
			cf.keysFilter = keysToFilter(view.Keys())
		}
		behaviors = append(behaviors, cf)
	}

	// If there were no matching views, set the default aggregation.
	if len(matches) == 0 {
		modified, akind, acfg, hinted := v.tryToApplyHint(instrument)
		instrument = modified // the hint erases itself from the description

		if akind != aggregation.DropKind {
			behaviors = append(behaviors, singleBehavior{
				fromName: instrument.Name,
				desc:     instrument,
				kind:     akind,
				acfg:     acfg,
				tempo:    v.views.Defaults.Temporality(instrument.Kind),
				hinted:   hinted,
			})
		}
	}

	v.compilerLock.Lock()
	defer v.compilerLock.Unlock()

	var conflicts ViewConflictsBuilder
	var compiled []Instrument

	for _, behavior := range behaviors {
		// the following checks semantic compatibility
		// and if necessary fixes the aggregation kind
		// to the default, via in place update.
		semanticErr := checkSemanticCompatibility(instrument.Kind, &behavior)

		existingInsts := v.names[behavior.desc.Name]
		var leaf leafInstrument

		// Scan the existing instruments for a match.
		for _, inst := range existingInsts {
			// Test for equivalence among the fields that we
			// cannot merge or will not convert, means the
			// testing everything except the description for
			// equality.

			if inst.Aggregation() != behavior.kind {
				continue
			}
			if inst.Descriptor().Kind.Synchronous() != behavior.desc.Kind.Synchronous() {
				continue
			}

			if inst.Descriptor().Unit != behavior.desc.Unit {
				continue
			}
			if inst.Descriptor().NumberKind != behavior.desc.NumberKind {
				continue
			}
			if !equalConfigs(inst.Config(), behavior.acfg) {
				continue
			}

			// For attribute keys, test for equal nil-ness or equal value.
			instKeys := inst.Keys()
			confKeys := behavior.keysSet
			if (instKeys == nil) != (confKeys == nil) {
				continue
			}
			if instKeys != nil && *instKeys != *confKeys {
				continue
			}
			// We can return the previously-compiled instrument,
			// we may have different descriptions and that is
			// specified to choose the longer one.
			inst.mergeDescription(behavior.desc.Description)
			leaf = inst
			break
		}
		if leaf == nil {
			switch behavior.desc.NumberKind {
			case number.Int64Kind:
				leaf = buildView[int64, number.Int64Traits](behavior)
			case number.Float64Kind:
				leaf = buildView[float64, number.Float64Traits](behavior)
			}

			v.collectors = append(v.collectors, leaf)
			existingInsts = append(existingInsts, leaf)
			v.names[behavior.desc.Name] = existingInsts
		}
		if len(existingInsts) > 1 || semanticErr != nil {
			c := Conflict{
				Semantic:   semanticErr,
				Duplicates: make([]Duplicate, len(existingInsts)),
			}
			for i := range existingInsts {
				c.Duplicates[i] = existingInsts[i]
			}
			conflicts.Add(v.views.Name, c)
		}
		compiled = append(compiled, leaf)
	}
	return Combine(instrument, compiled...), conflicts
}

// buildView compiles either a synchronous or asynchronous instrument
// given its behavior and generic number type/traits.
func buildView[N number.Any, Traits number.Traits[N]](behavior singleBehavior) leafInstrument {
	if behavior.desc.Kind.Synchronous() {
		return compileSync[N, Traits](behavior)
	}
	return compileAsync[N, Traits](behavior)
}

// newSyncView returns a compiled synchronous instrument.  If the view
// calls for delta temporality, a stateless instrument is returned,
// otherwise for cumulative temporality a stateful instrument will be
// used.  I.e., Delta->Stateless, Cumulative->Stateful.
func newSyncView[
	N number.Any,
	Storage any,
	Methods aggregator.Methods[N, Storage],
](behavior singleBehavior) leafInstrument {
	// Note: nolint:govet below is to avoid copylocks.  The lock
	// is being copied before the new object is returned to the
	// user, and the extra allocation cost here would be
	// noticeable.
	metric := instrumentBase[N, Storage, int64, Methods]{
		fromName:   behavior.fromName,
		desc:       behavior.desc,
		acfg:       behavior.acfg,
		data:       map[attribute.Set]*storageHolder[Storage, int64]{},
		keysSet:    behavior.keysSet,
		keysFilter: behavior.keysFilter,
	}
	instrument := compiledSyncBase[N, Storage, Methods]{
		instrumentBase: metric, //nolint:govet
	}
	if behavior.tempo == aggregation.DeltaTemporality {
		return &statelessSyncInstrument[N, Storage, Methods]{
			compiledSyncBase: instrument, //nolint:govet
		}
	}

	return &statefulSyncInstrument[N, Storage, Methods]{
		compiledSyncBase: instrument, //nolint:govet
	}
}

// compileSync calls newSyncView to compile a synchronous
// instrument with specific aggregator storage and methods.
func compileSync[N number.Any, Traits number.Traits[N]](behavior singleBehavior) leafInstrument {
	switch behavior.kind {
	case aggregation.HistogramKind:
		return newSyncView[
			N,
			histogram.Histogram[N, Traits],
			histogram.Methods[N, Traits],
		](behavior)
	case aggregation.MinMaxSumCountKind:
		return newSyncView[
			N,
			minmaxsumcount.State[N, Traits],
			minmaxsumcount.Methods[N, Traits],
		](behavior)
	case aggregation.NonMonotonicSumKind:
		return newSyncView[
			N,
			sum.State[N, Traits, sum.NonMonotonic],
			sum.Methods[N, Traits, sum.NonMonotonic],
		](behavior)
	case aggregation.GaugeKind:
		// Note: off-spec synchronous gauge support
		return newSyncView[
			N,
			gauge.State[N, Traits],
			gauge.Methods[N, Traits],
		](behavior)
	default:
		fallthrough
	case aggregation.MonotonicSumKind:
		return newSyncView[
			N,
			sum.State[N, Traits, sum.Monotonic],
			sum.Methods[N, Traits, sum.Monotonic],
		](behavior)
	}
}

// newAsyncView returns a compiled asynchronous instrument.  If the
// view calls for delta temporality, a stateful instrument is
// returned, otherwise for cumulative temporality a stateless
// instrument will be used.  I.e., Cumulative->Stateless,
// Delta->Stateful.
func newAsyncView[
	N number.Any,
	Storage any,
	Methods aggregator.Methods[N, Storage],
](behavior singleBehavior) leafInstrument {
	// Note: nolint:govet below is to avoid copylocks.  The lock
	// is being copied before the new object is returned to the
	// user, and the extra allocation cost here would be
	// noticeable.
	metric := instrumentBase[N, Storage, notUsed, Methods]{
		fromName:   behavior.fromName,
		desc:       behavior.desc,
		acfg:       behavior.acfg,
		data:       map[attribute.Set]*storageHolder[Storage, notUsed]{},
		keysSet:    behavior.keysSet,
		keysFilter: behavior.keysFilter,
	}
	instrument := compiledAsyncBase[N, Storage, Methods]{
		instrumentBase: metric, //nolint:govet
	}

	if behavior.tempo == aggregation.DeltaTemporality {
		var methods Methods
		if methods.Kind() != aggregation.GaugeKind {
			return &statefulAsyncInstrument[N, Storage, Methods]{
				compiledAsyncBase: instrument, //nolint:govet
			}
		}
		// Gauges fall through to the stateless behavior
		// regardless of delta temporality.
	}

	return &statelessAsyncInstrument[N, Storage, Methods]{
		compiledAsyncBase: instrument, //nolint:govet
	}
}

// compileAsync calls newAsyncView to compile an asynchronous
// instrument with specific aggregator storage and methods.
func compileAsync[N number.Any, Traits number.Traits[N]](behavior singleBehavior) leafInstrument {
	switch behavior.kind {
	case aggregation.MonotonicSumKind:
		return newAsyncView[
			N,
			sum.State[N, Traits, sum.Monotonic],
			sum.Methods[N, Traits, sum.Monotonic],
		](behavior)
	case aggregation.NonMonotonicSumKind:
		return newAsyncView[
			N,
			sum.State[N, Traits, sum.NonMonotonic],
			sum.Methods[N, Traits, sum.NonMonotonic],
		](behavior)
	default:
		fallthrough
	case aggregation.GaugeKind:
		return newAsyncView[
			N,
			gauge.State[N, Traits],
			gauge.Methods[N, Traits],
		](behavior)
	}
}

// Combine accepts a variable number of Instruments to combine.  If 0
// items, nil is returned. If 1 item, the item itself is return.
// otherwise, a multiInstrument of the appropriate number kind is returned.
func Combine(desc sdkinstrument.Descriptor, insts ...Instrument) Instrument {
	if len(insts) == 0 {
		return nil
	}
	if len(insts) == 1 {
		return insts[0]
	}
	if desc.NumberKind == number.Float64Kind {
		return multiInstrument[float64](insts)
	}
	return multiInstrument[int64](insts)
}

// multiInstrument is used by Combine() to combine the effects of
// multiple instrument-view behaviors.  These instruments produce
// multiAccumulators in NewAccumulator.
type multiInstrument[N number.Any] []Instrument

// NewAccumulator returns a Accumulator for multiple views of the same instrument.
func (mi multiInstrument[N]) NewAccumulator(kvs attribute.Set) Accumulator {
	accs := make([]Accumulator, 0, len(mi))

	for _, inst := range mi {
		accs = append(accs, inst.NewAccumulator(kvs))
	}
	return multiAccumulator[N](accs)
}

// Uses a int(0)-value attribute to identify distinct key sets.
func keysToSet(keys []attribute.Key) *attribute.Set {
	attrs := make([]attribute.KeyValue, len(keys))
	for i, key := range keys {
		attrs[i] = key.Int(0)
	}
	ns := attribute.NewSet(attrs...)
	return &ns
}

// keyFilter provides an attribute.Filter implementation based on a
// map[attribute.Key].
type keyFilter map[attribute.Key]struct{}

// filter is an attribute.Filter.
func (ks keyFilter) filter(kv attribute.KeyValue) bool {
	_, has := ks[kv.Key]
	return has
}

// keysToFilter constructs a keyFilter.
func keysToFilter(keys []attribute.Key) *attribute.Filter {
	kf := keyFilter{}
	for _, k := range keys {
		kf[k] = struct{}{}
	}
	var af attribute.Filter = kf.filter
	return &af
}

// equalConfigs compares two aggregator configurations.
func equalConfigs(a, b aggregator.Config) bool {
	return a == b
}

// pickAggConfig returns the aggregator configuration prescribed by a view clause
// if it is not empty, otherwise the default value.
func pickAggConfig(def, vcfg aggregator.Config) aggregator.Config {
	if vcfg != (aggregator.Config{}) {
		return vcfg
	}
	return def
}

// checkSemanticCompatibility checks whether an instrument /
// aggregator pairing is well defined.
func checkSemanticCompatibility(ik sdkinstrument.Kind, behavior *singleBehavior) error {
	if behavior.hinted {
		// Anything goes!
		return nil
	}

	agg := behavior.kind
	cat := agg.Category(ik)

	if agg == aggregation.AnySumKind {
		switch cat {
		case aggregation.MonotonicSumCategory, aggregation.HistogramCategory:
			agg = aggregation.MonotonicSumKind
		case aggregation.NonMonotonicSumCategory:
			agg = aggregation.NonMonotonicSumKind
		default:
			agg = aggregation.UndefinedKind
		}
		behavior.kind = agg
	}

	switch ik {
	case sdkinstrument.SyncCounter, sdkinstrument.SyncHistogram:
		switch cat {
		case aggregation.MonotonicSumCategory, aggregation.NonMonotonicSumCategory, aggregation.HistogramCategory:
			return nil
		}

	case sdkinstrument.SyncUpDownCounter, sdkinstrument.AsyncUpDownCounter:
		switch cat {
		case aggregation.NonMonotonicSumCategory:
			return nil
		}

	case sdkinstrument.AsyncCounter:
		switch cat {
		case aggregation.NonMonotonicSumCategory, aggregation.MonotonicSumCategory:
			return nil
		}

	case sdkinstrument.AsyncGauge:
		switch cat {
		case aggregation.GaugeCategory:
			return nil
		}
	}

	behavior.kind = view.StandardAggregationKind(ik)
	return SemanticError{
		Instrument:  ik,
		Aggregation: agg,
	}
}

// viewDescriptor returns the modified sdkinstrument.Descriptor of a
// view.  It retains the original instrument kind, numebr kind, and
// unit, while allowing the name and description to change.
func viewDescriptor(instrument sdkinstrument.Descriptor, v view.ClauseConfig) sdkinstrument.Descriptor {
	ikind := instrument.Kind
	nkind := instrument.NumberKind
	name := instrument.Name
	description := instrument.Description
	unit := instrument.Unit
	if v.HasName() {
		name = v.Name()
	}
	if v.Description() != "" {
		description = v.Description()
	}
	return sdkinstrument.NewDescriptor(name, ikind, nkind, description, unit)
}
