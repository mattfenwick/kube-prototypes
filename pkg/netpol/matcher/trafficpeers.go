package matcher

type TrafficPeers struct {
	SourcesOrDests []*SourceDestAndPort
}

func (tp *TrafficPeers) Combine(other *TrafficPeers) *TrafficPeers {
	if tp == nil && other == nil {
		// if both rules are (e|in)gress-only, combined rule should be (e|in)gress-only too
		return nil
	}
	var mine, theirs []*SourceDestAndPort
	if tp != nil {
		mine = tp.SourcesOrDests
	}
	if other != nil {
		theirs = other.SourcesOrDests
	}
	return &TrafficPeers{SourcesOrDests: append(mine, theirs...)}
}

func (tp *TrafficPeers) Allows(td *TrafficDirection) bool {
	for _, sd := range tp.SourcesOrDests {
		if sd.Allows(td) {
			return true
		}
	}
	return false
}
