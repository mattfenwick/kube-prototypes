## Network policy API design

[Find out more about our goals!](https://jayunit100.blogspot.com/2020/08/daydreaming-about-perfect-networkpolicy.html)

## Setup

```
brew install kind

git clone git@github.com:mattfenwick/kube-prototypes.git

cd kube-prototypes/cmd/netpol-crd

go run main.go
```

## Interesting corner cases

### Asymmetry of target selectors

In namespace n1, deny all egress from namespace d1:

```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: allow-nothing-from-d1
  namespace: d1
spec:
  podSelector: {}
  policyTypes:
  - Egress
```

In namespace d2, allow all ingress:

```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: allow-all-to-d2
  namespace: d2
spec:
  ingress:
  - {}
  podSelector: {}
  policyTypes:
  - Ingress
```

Expected result: requests from `d1 -> d2` are allowed, since allows trump denies.

Actual result: requests from `d1 -> d2` are denied.  
 - Possible explanation: since the two policies have different targets, there must be something else going on? 
