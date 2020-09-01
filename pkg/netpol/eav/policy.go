package eav

type Directive string

const (
	DirectiveAllow Directive = "Allow"
	DirectiveDeny  Directive = "Deny"
)

type Policy struct {
	TrafficMatcher TrafficMatcher
	Directive      Directive
}

// Allows returns:
// - false, "" if no match
// - true, Deny if matched and denies
// - true, Allow if matched and allowed
func (p *Policy) Allows(t *Traffic) (bool, Directive) {
	if p.TrafficMatcher.Matches(t) {
		return true, p.Directive
	}
	return false, ""
}
