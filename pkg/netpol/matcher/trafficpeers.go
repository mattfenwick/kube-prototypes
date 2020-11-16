package matcher

type TrafficPeers struct {
	// TODO change this to:
	//   SourceDests map[string]*PeerPortMatcher
	//   where the key is the PK of PeerPortMatcher
	//   and add a (tp *TrafficPeers)Add(sdap *PeerMatcher, port Port) method or something
	//   goal: nest ports under SourceDests
	// TODO is there a better way to represent 'nothing allowed' than an empty slice?
	SourcesOrDests []*PeerPortMatcher
}

func (tp *TrafficPeers) Combine(other *TrafficPeers) *TrafficPeers {
	if tp == nil && other == nil {
		// if both rules are (e|in)gress-only, combined rule should be (e|in)gress-only too
		return nil
	}
	var mine, theirs []*PeerPortMatcher
	if tp != nil {
		mine = tp.SourcesOrDests
	}
	if other != nil {
		theirs = other.SourcesOrDests
	}
	return &TrafficPeers{SourcesOrDests: append(mine, theirs...)}
}

func (tp *TrafficPeers) Allows(peer *TrafficPeer, portProtocol *PortProtocol) bool {
	if tp == nil {
		return true
	}
	for _, sd := range tp.SourcesOrDests {
		if sd.Allows(peer, portProtocol) {
			return true
		}
	}
	return false
}
