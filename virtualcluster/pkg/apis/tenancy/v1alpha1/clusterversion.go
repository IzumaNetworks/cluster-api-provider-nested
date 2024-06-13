/*
Copyright 2019 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	"fmt"

	vcdebug "sigs.k8s.io/cluster-api-provider-nested/virtualcluster/pkg/debug"
)

// GetEtcdDomain returns the dns of etcd service, note that, though the
// complete etcd svc dns is {etcdSvcName}.{namespace}.svc.{clusterdomain},
// this EtcdDomain is only used by apiserver that in the same namespace,
// so the etcdSvcName is adequate
func (cv *ClusterVersion) GetEtcdDomain() string {
	vcdebug.Info("etcdSvcName", cv.Spec.ETCD.Service.Name)
	return cv.Spec.ETCD.Service.Name
}

// GetEtcdServers returns the list of hostnames of etcd pods
func (cv *ClusterVersion) GetEtcdServers() (etcdServers []string) {
	etcdStsName := cv.Spec.ETCD.StatefulSet.Name
	replicas := cv.Spec.ETCD.StatefulSet.Spec.Replicas
	var i int32
	for ; i < *replicas; i++ {
		etcdServers = append(etcdServers, fmt.Sprintf("%s-%d.%s", etcdStsName, i, cv.GetEtcdDomain()))
	}
	vcdebug.Info("etcdServers", etcdServers)
	return etcdServers
}

// GetAPIServerDomain returns the dns of the apiserver service
//
// TODO support NodePort and ClusterIP for accessing apiserver from
// outside the cluster
func (cv *ClusterVersion) GetAPIServerDomain(namespace string) string {
	vcdebug.Info("GetAPIServerDomain", cv.Spec.APIServer.Service.Name+"."+namespace)
	return cv.Spec.APIServer.Service.Name + "." + namespace
}
