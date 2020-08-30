package matcher

// This is the root type
type NetworkPolicies struct {
	Targets map[string]*Target
}

func NewNetworkPolicies() *NetworkPolicies {
	return &NetworkPolicies{Targets: map[string]*Target{}}
}

func (np *NetworkPolicies) AddTarget(target *Target) *Target {
	pk := target.GetPrimaryKey()
	if prev, ok := np.Targets[pk]; ok {
		combined := prev.Combine(target)
		np.Targets[pk] = combined
	} else {
		np.Targets[pk] = target
	}
	return np.Targets[pk]
}

func (np *NetworkPolicies) TargetsApplyingToPod(namespace string, podLabels map[string]string) []*Target {
	var targets []*Target
	for _, target := range np.Targets {
		if target.IsMatch(namespace, podLabels) {
			targets = append(targets, target)
		}
	}
	return targets
}

func (np *NetworkPolicies) TargetsApplyingToNamespace(namespace string) []*Target {
	var targets []*Target
	for _, t := range np.Targets {
		if t.Namespace == namespace {
			targets = append(targets, t)
		}
	}
	return targets
}

// IsTrafficAllowed returns:
// - whether the traffic is allowed
// - which rules allowed the traffic
// - which rules matched the traffic target
func (np *NetworkPolicies) IsTrafficAllowed(trafficDirection *ResolvedTraffic) (bool, []*Target, []*Target) {
	matchingTargets := np.TargetsApplyingToPod(trafficDirection.Target.Namespace, trafficDirection.Target.PodLabels)

	// No targets match => automatic allow
	if len(matchingTargets) == 0 {
		return true, nil, nil
	}

	// Check if any matching targets allow this traffic
	var allowers []*Target
	for _, match := range matchingTargets {
		if match.Allows(trafficDirection) {
			allowers = append(allowers, match)
		}
	}
	if len(allowers) > 0 {
		return true, allowers, matchingTargets
	}

	// Otherwise, deny
	return false, nil, matchingTargets
}
