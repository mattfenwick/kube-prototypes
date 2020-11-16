package matcher

type IngressEgressMatcher struct {
	// TODO change this to:
	//   SourceDests map[string]*PeerPortMatcher
	//   where the key is the PK of PeerPortMatcher
	//   and add a (tp *IngressEgressMatcher)Add(sdap *PeerMatcher, port Port) method or something
	//   goal: nest ports under SourceDests
	// TODO is there a better way to represent 'nothing allowed' than an empty slice?
	Matchers []*PeerPortMatcher
}

func (tp *IngressEgressMatcher) Combine(other *IngressEgressMatcher) *IngressEgressMatcher {
	if tp == nil && other == nil {
		// if both rules are (e|in)gress-only, combined rule should be (e|in)gress-only too
		return nil
	}
	var mine, theirs []*PeerPortMatcher
	if tp != nil {
		mine = tp.Matchers
	}
	if other != nil {
		theirs = other.Matchers
	}
	return &IngressEgressMatcher{Matchers: append(mine, theirs...)}
}

func (tp *IngressEgressMatcher) Allows(peer *TrafficPeer, portProtocol *PortProtocol) bool {
	if tp == nil {
		return true
	}
	for _, sd := range tp.Matchers {
		if sd.Allows(peer, portProtocol) {
			return true
		}
	}
	return false
}
