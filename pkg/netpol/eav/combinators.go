package eav

import (
	"github.com/mattfenwick/kube-prototypes/pkg/netpol/kube"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"reflect"
)

// All matches if all its subterms match.  If no subterms, it will match.
type All struct {
	Terms []TrafficMatcher
}

func NewAll(terms ...TrafficMatcher) *All {
	return &All{Terms: terms}
}

func (a *All) Matches(t *Traffic) bool {
	for _, term := range a.Terms {
		if !term.Matches(t) {
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

func (a *Any) Matches(t *Traffic) bool {
	for _, term := range a.Terms {
		if term.Matches(t) {
			return true
		}
	}
	return false
}

// Not matches if its subterm doesn't, and doesn't match if its subterm does.
type Not struct {
	Term TrafficMatcher
}

func (n *Not) Matches(t *Traffic) bool {
	return !n.Matches(t)
}

// InArray
type InArray struct {
	Selector Selector
	Values   []interface{}
}

func (i *InArray) Matches(t *Traffic) bool {
	v := i.Selector(t)
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

func (e *Equal) Matches(t *Traffic) bool {
	if len(e.Selectors) < 2 {
		// TODO should there be an error?
		return true
	}
	prev := e.Selectors[0]
	for _, next := range e.Selectors[1:] {
		if !reflect.DeepEqual(prev(t), next(t)) {
			return false
		}
		prev = next
	}
	return true
}

// Matchers

// Equal verifies that all selectors return the same thing.
// It's not useful to have fewer than 2 selectors, maybe that will be illegal in the future.
type LabelMatcher struct {
	Selector Selector
	Key      string
	Value    string
}

func (lm *LabelMatcher) Matches(t *Traffic) bool {
	labels, ok := lm.Selector(t).(map[string]string)
	if !ok {
		panic(errors.Errorf("expected map[string]string, found %T", lm.Selector(t)))
	}
	value, ok := labels[lm.Key]
	return ok && value == lm.Value
}

// IPMatcher matches an IP address using a cidr
type IPMatcher struct {
	Selector Selector
	CIDR     string
}

func (ipm *IPMatcher) Matches(t *Traffic) bool {
	ip, ok := ipm.Selector(t).(string)
	if !ok {
		panic(errors.Errorf("expected string, found %T", ipm.Selector(t)))
	}
	return kube.IsIPInCIDR(ip, ipm.CIDR)
}

func IPBlockMatcher(selector Selector, cidr string, except []string) TrafficMatcher {
	// TODO wow this is unpleasant, is there a way to not have to do this?
	var values []interface{}
	for _, e := range except {
		values = append(values, e)
	}
	return NewAll(
		&IPMatcher{Selector: selector, CIDR: cidr},
		&Not{Term: &InArray{
			Selector: selector,
			Values:   values}})
}

type Bool struct {
	Selector Selector
}

func (b *Bool) Matches(t *Traffic) bool {
	return b.Selector(t).(bool)
}

func KubeMatchLabels(selector Selector, labels map[string]string) TrafficMatcher {
	var terms []TrafficMatcher
	for key, val := range labels {
		terms = append(terms, &LabelMatcher{
			Selector: selector,
			Key:      key,
			Value:    val,
		})
	}
	return NewAll(terms...)
}

type KubeMatchExpressionMatcher struct {
	Selector   Selector
	Expression metav1.LabelSelectorRequirement
}

func (kmem *KubeMatchExpressionMatcher) Matches(t *Traffic) bool {
	labels := kmem.Selector(t).(map[string]string)
	return kube.IsMatchExpressionMatchForLabels(labels, kmem.Expression)
}

func KubeMatchExpressions(selector Selector, mes []metav1.LabelSelectorRequirement) TrafficMatcher {
	var terms []TrafficMatcher
	for _, exp := range mes {
		terms = append(terms, &KubeMatchExpressionMatcher{
			Selector:   selector,
			Expression: exp,
		})
	}
	return NewAll(terms...)
}

func KubeMatchLabelSelector(selector Selector, ls metav1.LabelSelector) TrafficMatcher {
	return NewAll(
		KubeMatchLabels(selector, ls.MatchLabels),
		KubeMatchExpressions(selector, ls.MatchExpressions))
}
