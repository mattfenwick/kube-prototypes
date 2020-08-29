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

func (np *NetworkPolicies) TargetsApplyingToPod(podLabels map[string]string, namespaceLabels map[string]string) []*Target {
	panic("TODO")
}

func (np *NetworkPolicies) TargetsApplyingToNamespace(namespaceLabels map[string]string) []*Target {
	panic("TODO")
}

func (np *NetworkPolicies) IsTrafficAllowed(trafficDirection *ResolvedTraffic) (bool, []*Target) {
	var matchingTargets []*Target
	for _, target := range np.Targets {
		if target.Namespace == trafficDirection.Target.Namespace &&
			isLabelsMatchLabelSelector(trafficDirection.Target.PodLabels, target.PodSelector) {
			matchingTargets = append(matchingTargets, target)
		}
	}
	// No targets match => automatic allow
	if len(matchingTargets) == 0 {
		return true, nil
	}

	// Check if any matching targets allow this traffic
	var allowers []*Target
	for _, match := range matchingTargets {
		if match.Allows(trafficDirection) {
			allowers = append(allowers, match)
		}
	}
	if len(allowers) > 0 {
		return true, allowers
	}

	// Otherwise, deny
	return false, nil
}
