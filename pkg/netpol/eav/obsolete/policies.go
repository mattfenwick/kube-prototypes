package obsolete

type Policies struct {
	Policies []*Policy
}

// Allows decides whether traffic is allowed based on:
// - did any policy match the Traffic target
//   match = same PolicyType and Traffic target matches rule target
// - if no policy matches the traffic target => automatic allow
// - if >= 1 policy matches the traffic target:
//     if >=1 policy matches the Peer and Port/Protocol => allow
//     if 0 policy matches Peer+Port+Protocol => deny
func (ps *Policies) Allows(t *Traffic) bool {
	didFindMatch := false
	for _, policy := range ps.Policies {
		isMatch, allows := policy.Allows(t)
		if allows {
			return true
		}
		if isMatch {
			didFindMatch = true
		}
	}
	return !didFindMatch
}
