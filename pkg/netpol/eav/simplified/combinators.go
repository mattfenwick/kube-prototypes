package eav

import (
	"reflect"
)

// All matches if all its subterms match.  If no subterms, it will match.
type All struct {
	Terms []TrafficMatcher
}

func NewAll(terms ...TrafficMatcher) *All {
	return &All{Terms: terms}
}

func (a *All) Matches(tm *TrafficMap) bool {
	for _, term := range a.Terms {
		if !term.Matches(tm) {
			return false
		}
	}
	return true
}

// Any matches if all its subterms match.  If no subterms, it will *not* match.
type Any struct {
	Terms []TrafficMatcher
}

func NewAny(terms ...TrafficMatcher) *Any {
	return &Any{Terms: terms}
}

func (a *Any) Matches(tm *TrafficMap) bool {
	for _, term := range a.Terms {
		if term.Matches(tm) {
			return true
		}
	}
	return false
}

// Not matches if its subterm doesn't, and doesn't match if its subterm does.
type Not struct {
	Term TrafficMatcher
}

func (n *Not) Matches(tm *TrafficMap) bool {
	return !n.Matches(tm)
}

// InArray
type InArray struct {
	Selector Selector
	Values   []interface{}
}

func (i *InArray) Matches(tm *TrafficMap) bool {
	v := i.Selector.Select(tm)
	for _, val := range i.Values {
		if reflect.DeepEqual(v, val) {
			return true
		}
	}
	return false
}

// Equal verifies that all selectors return the same thing.
// It's not useful to have fewer than 2 selectors, maybe that will be illegal in the future.
type Equal struct {
	Selectors []Selector
}

func NewEqual(selectors ...Selector) *Equal {
	return &Equal{Selectors: selectors}
}

func (e *Equal) Matches(tm *TrafficMap) bool {
	if len(e.Selectors) < 2 {
		// TODO should there be an error?
		return true
	}
	prev := e.Selectors[0]
	for _, next := range e.Selectors[1:] {
		if !reflect.DeepEqual(prev.Select(tm), next.Select(tm)) {
			return false
		}
		prev = next
	}
	return true
}

type Bool struct {
	Selector Selector
}

func (b *Bool) Matches(tm *TrafficMap) bool {
	return b.Selector.Select(tm).(bool)
}
