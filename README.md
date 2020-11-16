# kube-prototypes

This repository is a demonstration of an RDF based data model for network policies, and is a work in progress prototype.

Note: it is not intended to be used in a production environment at this stage.

# Background

There are various shortcomings in the existing V1 networkpolicy API of Kubernetes, which are being addressed in a variety of manners.

One of these is this project, which attempts to skirt around *all* shortcomings by building an easy to use, high level operator that translates easy to write policies, based on a 'graphical' model, into lower level v1 network policy examples.  

To see many of the various use cases that have inspired this, see https://github.com/jayunit100/network-policy-subproject.  


# Specifics

The specific problems wed like to solve are:

- Implement a `cluster network policy`, on top of the existing v1 networkpolicy API, which are CNI independent.
- Implement `policy priorities`, on top of the v1 networkpolicy API, which are CNI independent.
- Implement `service selector` network policies, on top of the v1 networkpolicy api, which are CNI independent.

# Why 

- If we can demonstrate that *some subset* of the asks made in the network-policy-subproject are implementable WITHOUT CNI support, it might be very easy to, hand-in-hand, propose a new policy API which can be supported WITHOUT CNI vendor buy in, because a default wrapper implementation, such as this, proves their feasibility.
- Maybe this approach will make it easy for vendors to share a common CNI-policy operator which willl allow vendors to innovate on a shared security model in a way that can evolve rapidly alongside the K8s api, prototyping new features for the community before they go into k8s.
- One great way to get more experience with the corner cases of the NetworkPolicy V1 api is to try to build things on top of it.  In some ways, this project is an experiment in probing these corner cases for deeper insight that we can feed back into the network policy working group  
