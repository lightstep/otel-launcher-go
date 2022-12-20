package runtime

import (
	"fmt"
	"path"

	"go.opentelemetry.io/otel/attribute"
)

type builtinKind int

var (
	singleton = []attribute.Set{emptySet}

	emptySet = attribute.NewSet()

	ErrUnmatchedBuiltin   = fmt.Errorf("builtin unmatched")
	ErrOvermatchedBuiltin = fmt.Errorf("builtin overmatched")
)

const (
	builtinCounter builtinKind = iota
	builtinObjectBytesCounter
	builtinUpDownCounter
	builtinGauge
)

type builtinMetricFamily struct {
	pattern string
	kind    builtinKind
	attrs   []attribute.Set
}

type builtinDescriptor struct {
	families []builtinMetricFamily
}

func newBuiltinsDescriptor() *builtinDescriptor {
	return &builtinDescriptor{}
}

func (bd *builtinDescriptor) add(pattern string, kind builtinKind, attrs ...attribute.Set) {

	bd.families = append(bd.families, builtinMetricFamily{
		pattern: pattern,
		kind:    kind,
		attrs:   attrs,
	})
}

// func toOTelName(n string) string {
// 	n, _, _ = strings.Cut(n, ":")
// 	n = namePrefix + strings.ReplaceAll(n, "/", ".")
// 	return n
// }

// func toOTelUnit(n string) string {
// 	_, u, _ := strings.Cut(n, ":")
// 	return u
// }

func (bd *builtinDescriptor) singleCounter(pattern string) {
	bd.add(pattern, builtinCounter)
}

func (bd *builtinDescriptor) classesCounter(pattern string, attrs ...attribute.Set) {
	if len(attrs) < 2 {
		panic("must have >1 attrs")
	}
	bd.add(pattern, builtinCounter, attrs...)
}

func (bd *builtinDescriptor) classesUpDownCounter(pattern string, attrs ...attribute.Set) {
	if len(attrs) < 2 {
		panic("must have >1 attrs")
	}
	bd.add(pattern, builtinUpDownCounter, attrs...)
}

func (bd *builtinDescriptor) objectBytesCounter(pattern string) {
	bd.add(pattern, builtinObjectBytesCounter)
}

func (bd *builtinDescriptor) singleUpDownCounter(pattern string) {
	bd.add(pattern, builtinUpDownCounter)
}

func (bd *builtinDescriptor) singleGauge(pattern string) {
	bd.add(pattern, builtinGauge)
}

func (bd *builtinDescriptor) find(name string) error {
	fam, err := bd.findFamily(name)
	if err != nil {

	}

	return
}

func (bd *builtinDescriptor) findFamily(name string) (builtinMetricFamily, error) {
	var first builtinMetricFamily
	matches := 0

	for _, f := range bd.families {
		matched, err := path.Match(f.pattern, name)
		if err != nil {
			return first, err
		}
		if matched {
			matches++
		}
	}
	if matches == 0 {
		return first, fmt.Errorf("%s: %w", name, ErrUnmatchedBuiltin)
	}
	if matches > 1 {
		return first, fmt.Errorf("%s: %w", name, ErrOvermatchedBuiltin)
	}
	return first, nil
}
