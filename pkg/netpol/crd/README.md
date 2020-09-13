# Network policies: experimental CRD

## Model

[*Traffic*](https://github.com/mattfenwick/kube-prototypes/blob/master/pkg/netpol/crd/traffic.go)

"Traffic" is a request from a *Source* to a *Destination*, over a specified port and protocol.
Sources and Destinations are both *Peer*s.  If a Peer is internal to the Kubernetes cluster, its
data includes both names and labels from the Pod, Node, and Namespace.  If a peer is external,
this data will not be included.  Peers also have IP addresses. 

[*TrafficEdge*](https://github.com/mattfenwick/kube-prototypes/blob/master/pkg/netpol/crd/trafficedge.go)

"TrafficEdge"s match (or don't match) Traffic requests, based on a variety of attributes.  TrafficEdges
are the heart of network policies, as they give policies the ability to accept or reject Traffic based
on attributes of the Traffic.

One big difference between Kubernetes network policies and this experimental model is that Kubernetes
privileges a "target", which is treated differently than the other Peer.  On the other hand, this
model is symmetrical between source and destination, and ingress and egress.

[*Policy*](https://github.com/mattfenwick/kube-prototypes/blob/master/pkg/netpol/crd/policy.go)

A policy is a TrafficEdge -- used to match traffic -- and a directive, along with compatibility hints
to aid in translation into kubernetes network policies.  If a policy matches traffic, it can choose to
allow or deny the traffic.  If multiple policies match traffic, the policy priority is taken into account.

## Code

[*Examples*](https://github.com/mattfenwick/kube-prototypes/blob/master/pkg/netpol/crd/examples.go): practical uses of the experimental network policy API to:
 - flush out corner cases
 - compare to Kubernetes network policies
 - test accuracy of Builder/Reducer conversions

[*Builder*](https://github.com/mattfenwick/kube-prototypes/blob/master/pkg/netpol/crd/builder.go): convert Kubernetes network policies to experimental model (reverse of Reducer)

[*Reducer*](https://github.com/mattfenwick/kube-prototypes/blob/master/pkg/netpol/crd/reducer.go): convert experimental network policies to Kubernetes model (reverse of Builder)
