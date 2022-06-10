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
	"regexp"

	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/aggregator"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/aggregator/aggregation"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/number"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/sdkinstrument"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/instrumentation"
)

// ClauseConfig contains each of the configurable aspects of a
// single Views clause.
type ClauseConfig struct {
	// Matchers for the instrument
	instrumentName       string
	instrumentNameRegexp *regexp.Regexp
	instrumentKind       sdkinstrument.Kind
	numberKind           number.Kind
	library              instrumentation.Library

	// Properties of the view
	keys        []attribute.Key // nil implies all keys, []attribute.Key{} implies none
	name        string
	description string
	aggregation aggregation.Kind
	acfg        aggregator.Config
}

const (
	unsetInstrumentKind = sdkinstrument.Kind(-1)
	unsetNumberKind     = number.Kind(-1)
)

// ClauseOption applies a configuration option value to a view Config.
type ClauseOption interface {
	apply(ClauseConfig) ClauseConfig
}

// clauseOptionFunction makes a functional ClauseOption out of a function object.
type clauseOptionFunction func(cfg ClauseConfig) ClauseConfig

// apply implements ClauseOption.
func (of clauseOptionFunction) apply(in ClauseConfig) ClauseConfig {
	return of(in)
}

// Matchers

func MatchInstrumentName(name string) ClauseOption {
	return clauseOptionFunction(func(clause ClauseConfig) ClauseConfig {
		clause.instrumentName = name
		return clause
	})
}

func MatchInstrumentNameRegexp(re *regexp.Regexp) ClauseOption {
	return clauseOptionFunction(func(clause ClauseConfig) ClauseConfig {
		clause.instrumentNameRegexp = re
		return clause
	})
}

func MatchInstrumentKind(k sdkinstrument.Kind) ClauseOption {
	return clauseOptionFunction(func(clause ClauseConfig) ClauseConfig {
		clause.instrumentKind = k
		return clause
	})
}

func MatchNumberKind(k number.Kind) ClauseOption {
	return clauseOptionFunction(func(clause ClauseConfig) ClauseConfig {
		clause.numberKind = k
		return clause
	})
}

func MatchInstrumentationLibrary(lib instrumentation.Library) ClauseOption {
	return clauseOptionFunction(func(clause ClauseConfig) ClauseConfig {
		clause.library = lib
		return clause
	})
}

// Properties

// WithKeys overwrites; nil is distinct from empty non-nil.
func WithKeys(keys []attribute.Key) ClauseOption {
	return clauseOptionFunction(func(clause ClauseConfig) ClauseConfig {
		clause.keys = keys
		return clause
	})
}

func WithName(name string) ClauseOption {
	return clauseOptionFunction(func(clause ClauseConfig) ClauseConfig {
		clause.name = name
		return clause
	})
}

func WithDescription(desc string) ClauseOption {
	return clauseOptionFunction(func(clause ClauseConfig) ClauseConfig {
		clause.description = desc
		return clause
	})
}

func WithAggregation(kind aggregation.Kind) ClauseOption {
	return clauseOptionFunction(func(clause ClauseConfig) ClauseConfig {
		clause.aggregation = kind
		return clause
	})
}

func WithAggregatorConfig(acfg aggregator.Config) ClauseOption {
	return clauseOptionFunction(func(clause ClauseConfig) ClauseConfig {
		clause.acfg = acfg
		return clause
	})
}

// IsSingleInstrument is a requirement when HasName().
func (c *ClauseConfig) IsSingleInstrument() bool {
	return c.instrumentName != ""
}

// HasName implies IsSingleInstrument SHOULD be required.
func (c *ClauseConfig) HasName() bool {
	return c.name != ""
}

func (c *ClauseConfig) Name() string {
	return c.name
}

func (c *ClauseConfig) Keys() []attribute.Key {
	return c.keys
}

func (c *ClauseConfig) Description() string {
	return c.description
}

func (c *ClauseConfig) Aggregation() aggregation.Kind {
	return c.aggregation
}

func (c *ClauseConfig) AggregatorConfig() aggregator.Config {
	return c.acfg
}

func stringMismatch(test, value string) bool {
	return test != "" && test != value
}

func ikindMismatch(test, value sdkinstrument.Kind) bool {
	return test != unsetInstrumentKind && test != value
}

func nkindMismatch(test, value number.Kind) bool {
	return test != unsetNumberKind && test != value
}

func regexpMismatch(test *regexp.Regexp, value string) bool {
	return test != nil && !test.MatchString(value)
}

func (c *ClauseConfig) libraryMismatch(lib instrumentation.Library) bool {
	hasName := c.library.Name != ""
	hasVersion := c.library.Version != ""
	hasSchema := c.library.SchemaURL != ""

	if !hasName && !hasVersion && !hasSchema {
		return false
	}
	return stringMismatch(c.library.Name, lib.Name) ||
		stringMismatch(c.library.Version, lib.Version) ||
		stringMismatch(c.library.SchemaURL, lib.SchemaURL)
}

func (c *ClauseConfig) Matches(lib instrumentation.Library, desc sdkinstrument.Descriptor) bool {
	mismatch := c.libraryMismatch(lib) ||
		stringMismatch(c.instrumentName, desc.Name) ||
		ikindMismatch(c.instrumentKind, desc.Kind) ||
		nkindMismatch(c.numberKind, desc.NumberKind) ||
		regexpMismatch(c.instrumentNameRegexp, desc.Name)
	return !mismatch
}
