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

package view // import "github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/view"

import (
	"fmt"

	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/aggregator"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/aggregator/aggregation" // Views is a configured set of view clauses with an associated Name
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/aggregator/histogram"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/sdkinstrument"
	"go.uber.org/multierr"
)

// that is used for debugging.
type Views struct {
	// Name of these views, used in error reporting.
	Name string

	// Config is the configuration for these views.
	Config
}

// New configures the clauses and default settings of a Views.
func New(name string, opts ...Option) *Views {
	return &Views{
		Name:   name,
		Config: NewConfig(opts...),
	}
}

func checkAggregation(err error, agg *aggregation.Kind, def aggregation.Kind) error {
	if !agg.Valid() {
		err = multierr.Append(err, fmt.Errorf("invalid aggregation: %v", *agg))
		*agg = def
	}
	return err
}

func checkTemporality(err error, tempo *aggregation.Temporality, def aggregation.Temporality) error {
	if !tempo.Valid() {
		err = multierr.Append(err, fmt.Errorf("invalid temporality: %v", *tempo))
		*tempo = def
	}
	return err
}

func checkAggConfig(err error, acfg *aggregator.Config) error {
	if !acfg.Valid() {
		err = multierr.Append(err, fmt.Errorf("invalid histogram size: %v", acfg.Histogram.MaxSize))
		acfg.Histogram.MaxSize = histogram.DefaultMaxSize
	}
	return err
}

// Validate checks for inconsistent view settings and returns any
// errors with the nearest consistent configuration for use.
func Validate(v *Views) (*Views, error) {
	var err error

	// Make a deep copy
	valid := &Views{
		Name: v.Name,
	}

	valid.Clauses = make([]ClauseConfig, len(v.Clauses))
	valid.Defaults = v.Defaults

	for i := range valid.Clauses {
		valid.Clauses[i] = v.Clauses[i]
	}

	// Validate default settings
	for i := range valid.Defaults.ByInstrumentKind {
		kind := sdkinstrument.Kind(i)

		err = checkAggregation(err, &valid.Defaults.ByInstrumentKind[i].Aggregation, StandardAggregationKind(kind))
		err = checkTemporality(err, &valid.Defaults.ByInstrumentKind[i].Temporality, StandardTemporality(kind))
		err = checkAggConfig(err, &valid.Defaults.ByInstrumentKind[i].Int64)
		err = checkAggConfig(err, &valid.Defaults.ByInstrumentKind[i].Float64)
	}

	for i := range valid.Clauses {
		clause := &valid.Clauses[i]

		if !clause.IsSingleInstrument() && clause.HasName() {
			// Note: no correction, this condition creates conflicts.
			err = multierr.Append(err, fmt.Errorf("multi-instrument view specifies a single name"))
		}

		err = checkAggregation(err, &clause.aggregation, aggregation.UndefinedKind)
		err = checkAggConfig(err, &clause.acfg)

		if clause.instrumentName != "" && clause.instrumentNameRegexp != nil {
			err = multierr.Append(err, fmt.Errorf("view has instrument name and regexp matches"))
			// Note: prefer the name over the regexp.
			clause.instrumentNameRegexp = nil
		}

		for i := range clause.keys {
			if clause.keys[i] == "" {
				err = multierr.Append(err, fmt.Errorf("view has empty string in keys"))
			}
		}
	}

	return valid, err
}
