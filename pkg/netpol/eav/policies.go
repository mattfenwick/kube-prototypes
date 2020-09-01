package eav

/*
Overview

make an algebra of matchers
 - AND
 - OR
 - EXCEPT

give matchers access to the target as well
 - so they can compare Peer info to Target info
   examples:
     allow all traffic where the Peer is in the same namespace as the Target
       EQUALS(TargetNamespace(), PeerNamespace())
     allow all traffic where the Peer is on the same node as the Target

What do we have:
 - selectors (perhaps described by a path?)
   - traffic target namespace
   - peer ip
   - port
   - type (ingress / egress)
 - comparators
   - key/val in labels
   - labels in label selector
   - equal
 - combiners
   - AND
   - OR
   - NOT
   - XOR

Key problems:
 - do we want to specially privilege Target matching?
   - ideally no, but how do we deal with the concept of 'no targets match => automatic allow'
     probably by having explicit deny/allow

DenyAll (ingress and egress):
  Rule(AnyTraffic, Deny)

DenyAllIngressToNamespace:
  Rule(Equals(DestinationNamespace, x), Deny)

Allow DNS from any pod
  Rule(Equals(DestinationPort, 53), Allow)

Block UDP and SCMP on any port from any pod for egress
  Rule
    IsIn
      SourceProtocol
      List
        UDP
        SCMP
    Deny

Allow traffic from any pod to any other pod *on the same node*
  Rule(Equal(Destination.Pod.Namespace, Source.Pod.Namespace), Allow)
  - avoid corner case where both namespaces are empty

Split IPBlock into positive CIDR and Negative Except
  AND(CIDR("1.2.3.4/24"), NOT(IPS("1.2.3.4", "1.2.3.5")))
  Rule(
    And(

*/

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
		isMatch, directive := policy.Allows(t)
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
