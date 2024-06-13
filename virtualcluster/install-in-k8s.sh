#!/bin/bash

# boiler plate to get the directory this script is in
SOURCE=${BASH_SOURCE[0]}
while [ -L "$SOURCE" ]; do 
  DIR=$( cd -P "$( dirname "$SOURCE" )" >/dev/null 2>&1 && pwd )
  SOURCE=$(readlink "$SOURCE")
  [[ $SOURCE != /* ]] && SOURCE=$DIR/$SOURCE 
done
DIR=$( cd -P "$( dirname "$SOURCE" )" >/dev/null 2>&1 && pwd )

echo "Install Cluster API Provider CRDs"
kubectl apply -f https://raw.githubusercontent.com/kubernetes-sigs/cluster-api-provider-nested/main/virtualcluster/config/crd/tenancy.x-k8s.io_clusterversions.yaml
echo "Install Cluster API Provider CRDs 2"
kubectl apply -f https://raw.githubusercontent.com/kubernetes-sigs/cluster-api-provider-nested/main/virtualcluster/config/crd/tenancy.x-k8s.io_virtualclusters.yaml

echo "Apply all_in_one-izuma-debug.yaml w/ debug"
kubectl apply -f $DIR/all_in_one-izuma-debug.yaml

echo "Apply clusterversion_v1_nodeport.yaml w/ fix"
kubectl apply -f $DIR/clusterversion_v1_nodeport.yaml

echo "Ready to create virtual cluster... done."
echo "Try  kubectl vc create -f virtualcluster_1_nodeport.yaml  -o vc-1.kubeconfig   or similar"

