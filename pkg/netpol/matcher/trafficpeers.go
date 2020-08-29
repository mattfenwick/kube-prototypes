package matcher

type TrafficPeers struct {
	// TODO change this to:
	//   SourceDests map[string]*SourceDestAndPort
	//   where the key is the PK of SourceDestAndPort
	//   and add a (tp *TrafficPeers)Add(sdap *SourceDest, port Port) method or something
	//   goal: nest ports under SourceDests
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

func (tp *TrafficPeers) Allows(td *ResolvedTraffic) bool {
	for _, sd := range tp.SourcesOrDests {
		if sd.Allows(td) {
			return true
		}
	}
	return false
}
