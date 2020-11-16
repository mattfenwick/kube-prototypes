package matcher

// This is the root type
type Policy struct {
	Targets map[string]*Target
}

func NewPolicy() *Policy {
	return &Policy{Targets: map[string]*Target{}}
}

func (np *Policy) AddTarget(target *Target) *Target {
	pk := target.GetPrimaryKey()
	if prev, ok := np.Targets[pk]; ok {
		combined := prev.Combine(target)
		np.Targets[pk] = combined
	} else {
		np.Targets[pk] = target
	}
	return np.Targets[pk]
}

func (np *Policy) TargetsApplyingToPod(namespace string, podLabels map[string]string) []*Target {
	var targets []*Target
	for _, target := range np.Targets {
		if target.IsMatch(namespace, podLabels) {
			targets = append(targets, target)
		}
	}
	return targets
}

func (np *Policy) TargetsApplyingToNamespace(namespace string) []*Target {
	var targets []*Target
	for _, t := range np.Targets {
		if t.Namespace == namespace {
			targets = append(targets, t)
		}
	}
	return targets
}

type DirectionResult struct {
	IsAllowed       bool
	AllowingTargets []*Target
	MatchingTargets []*Target
}

type AllowedResult struct {
	Ingress *DirectionResult
	Egress  *DirectionResult
}

func (ar *AllowedResult) IsAllowed() bool {
	return ar.Ingress.IsAllowed && ar.Egress.IsAllowed
}

// IsTrafficAllowed returns:
// - whether the traffic is allowed
// - which rules allowed the traffic
// - which rules matched the traffic target
func (np *Policy) IsTrafficAllowed(traffic *Traffic) *AllowedResult {
	return &AllowedResult{
		Ingress: np.IsIngressOrEgressAllowed(traffic, true),
		Egress:  np.IsIngressOrEgressAllowed(traffic, false),
	}
}

func (np *Policy) IsIngressOrEgressAllowed(traffic *Traffic, isIngress bool) *DirectionResult {
	var target *TrafficPeer
	var peer *TrafficPeer
	if isIngress {
		target = traffic.Destination
		peer = traffic.Source
	} else {
		target = traffic.Source
		peer = traffic.Destination
	}

	// 1. if target is external to cluster -> allow
	if target.Internal == nil {
		return &DirectionResult{IsAllowed: true, AllowingTargets: nil, MatchingTargets: nil}
	}

	matchingTargets := np.TargetsApplyingToPod(target.Internal.Namespace, target.Internal.PodLabels)

	// No targets match => automatic allow
	if len(matchingTargets) == 0 {
		return &DirectionResult{IsAllowed: true, AllowingTargets: nil, MatchingTargets: nil}
	}

	// Check if any matching targets allow this traffic
	var allowers []*Target
	for _, target := range matchingTargets {
		if isIngress {
			if target.Ingress.Allows(peer, traffic.PortProtocol) {
				allowers = append(allowers, target)
			}
		} else {
			if target.Egress.Allows(peer, traffic.PortProtocol) {
				allowers = append(allowers, target)
			}
		}
	}
	if len(allowers) > 0 {
		return &DirectionResult{IsAllowed: true, AllowingTargets: allowers, MatchingTargets: matchingTargets}
	}

	// Otherwise, deny
	return &DirectionResult{IsAllowed: false, AllowingTargets: nil, MatchingTargets: matchingTargets}
}
