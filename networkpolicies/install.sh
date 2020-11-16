NS=${NS:-default}

kubectl apply -n $NS -f api-allow.yaml
kubectl apply -n $NS -f blackduck-deny-all.yaml
kubectl apply -n $NS -f deny-all.yaml
kubectl apply -n $NS -f web-allow-all.yaml
kubectl apply -n $NS -f web-allow-all-2.yaml
