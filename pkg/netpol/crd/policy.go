package crd

import (
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Directive string

const (
	DirectiveAllow Directive = "Allow"
	DirectiveDeny  Directive = "Deny"
)

type Policy struct {
	metav1.ObjectMeta
	Spec PolicySpec
}

type PolicySpec struct {
	Compatibility  []networkingv1.PolicyType
	Priority       int
	TrafficMatcher *TrafficEdge
	Directive      Directive
}

// Allows returns:
// - false, "" if no match
// - true, Deny if matched and denies
// - true, Allow if matched and allowed
func (ps *PolicySpec) Allows(tm *Traffic) (bool, Directive) {
	if ps.TrafficMatcher.Matches(tm) {
		return true, ps.Directive
	}
	return false, ""
}
