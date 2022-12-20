package runtime

import (
	"fmt"
	"strings"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric/unit"
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
	matches int
	kind    builtinKind
	attrs   []attribute.Set
}

type builtinDescriptor struct {
	families []*builtinMetricFamily
}

func newBuiltinDescriptor() *builtinDescriptor {
	return &builtinDescriptor{}
}

func (bd *builtinDescriptor) add(pattern string, kind builtinKind, attrs ...attribute.Set) {
	bd.families = append(bd.families, &builtinMetricFamily{
		pattern: pattern,
		kind:    kind,
		attrs:   attrs,
	})
}

func toOTelNameAndStatedUnit(nameAndUnit string) (on, un string) {
	on, un, _ = strings.Cut(nameAndUnit, ":")
	return toOTelName(on), un
}

func toOTelName(name string) string {
	return namePrefix + strings.ReplaceAll(name, "/", ".")
}

func attributeName(order int) string {
	base := "class"
	for ; order > 0; order-- {

		base = "sub" + base
	}
	return base
}

func (bd *builtinDescriptor) singleCounter(pattern string) {
	bd.add(pattern, builtinCounter)
}

func (bd *builtinDescriptor) classesCounter(pattern string, attrs ...attribute.Set) {
	if len(attrs) < 2 {
		panic(fmt.Sprintf("must have >1 attrs: %s", attrs))
	}
	bd.add(pattern, builtinCounter, attrs...)
}

func (bd *builtinDescriptor) classesUpDownCounter(pattern string, attrs ...attribute.Set) {
	if len(attrs) < 2 {
		panic(fmt.Sprintf("must have >1 attrs: %s", attrs))
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

func (bd *builtinDescriptor) findMatch(goname string) (mname, munit, descPattern string, attrs []attribute.KeyValue, _ error) {
	fam, err := bd.findFamily(goname)
	if err != nil {
		return "", "", "", nil, err
	}
	fam.matches++

	// Set the name, unit and pattern.
	if wildCnt := strings.Count(fam.pattern, "*"); wildCnt == 0 {
		mname, munit = toOTelNameAndStatedUnit(goname)
		descPattern = goname
	} else if strings.HasSuffix(fam.pattern, ":*") {
		// Special case for bytes/objects w/ same prefix.
		mname, munit = toOTelNameAndStatedUnit(goname)
		descPattern = goname

		if munit == "objects" {
			mname += "." + munit
			munit = ""
		}
	} else {
		pfx, sfx, _ := strings.Cut(fam.pattern, "/*:")
		mname = toOTelName(pfx)
		munit = sfx
		asubstr := goname[len(pfx):]
		asubstr = asubstr[1 : len(asubstr)-len(sfx)-1]
		splitVals := strings.Split(asubstr, "/")
		for order, val := range splitVals {
			attrs = append(attrs, attribute.Key(attributeName(order)).String(val))
		}
		// Ignore subtotals
		if splitVals[len(splitVals)-1] == "total" {
			return "", "", "", nil, nil
		}
		descPattern = fam.pattern
	}

	// Fix the units for UCUM.
	switch munit {
	case "bytes":
		munit = string(unit.Bytes)
	case "seconds":
		munit = "s"
	case "":
	default:
		// Pseudo-units
		munit = "{" + munit + "}"
	}

	// Fix the name if it ends in ".classes"
	if strings.HasSuffix(mname, ".classes") {

		s := mname[:len(mname)-len("classes")]

		// Note that ".classes" is (apparently) intended as a generic
		// suffix, while ".cycles" is an exception.
		// The ideal name depends on what we know.
		switch munit {
		case "By":
			// OTel has similar conventions for memory usage, disk usage, etc, so
			// for metrics with /classes/*:bytes we create a .usage metric.
			mname = s + "usage"
		case "{cpu-seconds}":
			// Same argument above, except OTel uses .time for
			// cpu-timing metrics instead of .usage.
			mname = s + "time"
		}
	}

	return mname, munit, descPattern, attrs, nil
}

func (bd *builtinDescriptor) findFamily(name string) (family *builtinMetricFamily, _ error) {
	matches := 0

	for _, f := range bd.families {
		pat := f.pattern
		wilds := strings.Count(pat, "*")
		if wilds > 1 {
			return nil, fmt.Errorf("too many wildcards: %s", pat)
		}
		if wilds == 0 && name == pat {
			matches++
			family = f
			continue
		}
		pfx, sfx, _ := strings.Cut(pat, "*")

		if len(name) > len(pat) && strings.HasPrefix(name, pfx) && strings.HasSuffix(name, sfx) {
			matches++
			family = f
			continue
		}
	}
	if matches == 0 {
		return nil, fmt.Errorf("%s: %w", name, ErrUnmatchedBuiltin)
	}
	if matches > 1 {
		return nil, fmt.Errorf("%s: %w", name, ErrOvermatchedBuiltin)
	}
	family.matches++
	return family, nil
}
