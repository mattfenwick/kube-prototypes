NS=$1

kubectl create ns $NS

kubectl create -f daemonset.yaml -n $NS
