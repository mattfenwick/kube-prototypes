# Enabling Calico on GCP

[See here](https://cloud.google.com/blog/products/gcp/network-policy-support-for-kubernetes-with-calico):

```
CLUSTER=TODO
PROJECT_ID=TODO
ZONE=TODO

# Create a cluster with Network Policy Enabled
# Enable the addon
gcloud beta container clusters update $CLUSTER --project=$PROJECT_ID --zone=$ZONE --update-addons=NetworkPolicy=ENABLE

# Enable on nodes (This re-creates the node pools)
gcloud beta container clusters update $CLUSTER --project=$PROJECT_ID --zone=$ZONE --enable-network-policy
```