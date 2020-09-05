#!/bin/bash

CALICO_CONF=kind-calico.yaml

# See https://alexbrand.dev/post/creating-a-kind-cluster-with-calico-networking/
cat << EOF > $CALICO_CONF
kind: Cluster
apiVersion: kind.sigs.k8s.io/v1alpha3
networking:
  disableDefaultCNI: true # disable kindnet
  podSubnet: 192.168.0.0/16 # set to Calico's default subnet
nodes:
- role: control-plane
- role: worker
- role: worker
- role: worker
EOF


kind create cluster --name=calico --config=$CALICO_CONF

sleep 5

until kubectl cluster-info;  do
    echo "`date`waiting for cluster..."
    sleep 2
done


kubectl get pods
kubectl apply -f ./calico_3_8.yaml
kubectl get pods -n kube-system
kubectl -n kube-system set env daemonset/calico-node FELIX_IGNORELOOSERPF=true

# No idea what this does: https://docs.projectcalico.org/reference/felix/configuration
kubectl -n kube-system set env daemonset/calico-node FELIX_XDPENABLED=false
