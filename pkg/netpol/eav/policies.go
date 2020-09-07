package eav

type Policies struct {
	Policies []*Policy
}

// Allows searches through policies for matches, from which it takes
// directives.  Some corner cases:
// - no matches => allowed (traffic must be explicitly denied)
// - allows take precedence over denies (TODO maybe rules need precedence?)
func (ps *Policies) Allows(t *Traffic) bool {
	isDenied := false
	for _, policy := range ps.Policies {
		isMatch, directive := policy.Spec.Allows(t)
		if isMatch {
			if directive == DirectiveAllow {
				return true
			} else if directive == DirectiveDeny {
				isDenied = true
			}
		}
	}
	return !isDenied
}
