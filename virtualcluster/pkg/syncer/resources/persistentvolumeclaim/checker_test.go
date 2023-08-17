/*
Copyright 2020 The Kubernetes Authors.

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

package persistentvolumeclaim

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	core "k8s.io/client-go/testing"
	"k8s.io/utils/pointer"

	"sigs.k8s.io/cluster-api-provider-nested/virtualcluster/pkg/apis/tenancy/v1alpha1"
	"sigs.k8s.io/cluster-api-provider-nested/virtualcluster/pkg/syncer/conversion"
	"sigs.k8s.io/cluster-api-provider-nested/virtualcluster/pkg/syncer/util/featuregate"
	util "sigs.k8s.io/cluster-api-provider-nested/virtualcluster/pkg/syncer/util/test"
)

var (
	statusPending = &corev1.PersistentVolumeClaimStatus{
		Phase: corev1.ClaimPending,
	}
	statusBound = &corev1.PersistentVolumeClaimStatus{
		Phase: corev1.ClaimBound,
	}
	statusLost = &corev1.PersistentVolumeClaimStatus{
		Phase: corev1.ClaimLost,
	}
)

func applyStatusToPVC(pvc *corev1.PersistentVolumeClaim, pvs *corev1.PersistentVolumeClaimStatus) *corev1.PersistentVolumeClaim {
	pvc.Status.Phase = pvs.Phase
	return pvc
}

func TestPVCPatrol(t *testing.T) {
	defer util.SetFeatureGateDuringTest(t, featuregate.DefaultFeatureGate, featuregate.SyncTenantPVCStatusPhase, true)()
	testTenant := &v1alpha1.VirtualCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "tenant-1",
			UID:       "7374a172-c35d-45b1-9c8e-bf5c5b614937",
		},
		Spec: v1alpha1.VirtualClusterSpec{},
		Status: v1alpha1.VirtualClusterStatus{
			Phase: v1alpha1.ClusterRunning,
		},
	}

	spec1 := &corev1.PersistentVolumeClaimSpec{
		AccessModes: []corev1.PersistentVolumeAccessMode{
			corev1.ReadWriteOnce,
		},
		Resources: corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceStorage: resource.MustParse("20Gi"),
			},
		},
		StorageClassName: pointer.StringPtr("storage-class-1"),
		VolumeName:       "volume-1",
	}

	spec2 := &corev1.PersistentVolumeClaimSpec{
		AccessModes: []corev1.PersistentVolumeAccessMode{
			corev1.ReadWriteOnce,
		},
		Resources: corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceStorage: resource.MustParse("30Gi"),
			},
		},
		StorageClassName: pointer.StringPtr("storage-class-1"),
		VolumeName:       "volume-1",
	}

	defaultClusterKey := conversion.ToClusterKey(testTenant)
	superDefaultNSName := conversion.ToSuperClusterNamespace(defaultClusterKey, "default")

	testcases := map[string]struct {
		ExistingObjectInSuper  []runtime.Object
		ExistingObjectInTenant []runtime.Object
		ExpectedDeletedPObject []string
		ExpectedDeletedVObject []string
		ExpectedCreatedPObject []string
		ExpectedUpdatedPObject []runtime.Object
		ExpectedUpdatedVObject []runtime.Object
		ExpectedNoOperation    bool
		WaitDWS                bool // Make sure to set this flag if the test involves DWS.
		WaitUWS                bool // Make sure to set this flag if the test involves UWS.
	}{
		"pPVC not created by vc": {
			ExistingObjectInSuper: []runtime.Object{
				unknownPVC("pvc-1", superDefaultNSName),
			},
			ExpectedNoOperation: true,
		},
		"pPVC exists, vPVC does not exists": {
			ExistingObjectInSuper: []runtime.Object{
				superPVC("pvc-1", superDefaultNSName, "12345", defaultClusterKey),
			},
			ExpectedDeletedPObject: []string{
				superDefaultNSName + "/pvc-1",
			},
		},
		"pPVC exists, vPVC exists with different uid": {
			ExistingObjectInSuper: []runtime.Object{
				superPVC("pvc-2", superDefaultNSName, "12345", defaultClusterKey),
			},
			ExistingObjectInTenant: []runtime.Object{
				tenantPVC("pvc-2", "default", "123456"),
			},
			ExpectedDeletedPObject: []string{
				superDefaultNSName + "/pvc-2",
			},
		},
		"pPVC exists, vPVC exists with different spec": {
			ExistingObjectInSuper: []runtime.Object{
				applySpecToPVC(superPVC("pvc-3", superDefaultNSName, "12345", defaultClusterKey), spec2),
			},
			ExistingObjectInTenant: []runtime.Object{
				applySpecToPVC(tenantPVC("pvc-3", "default", "12345"), spec1),
			},
			ExpectedNoOperation: true,
			// notes: have not updated the different pPVC in patrol now.
		},
		"vPVC exists, pPVC does not exists": {
			ExistingObjectInTenant: []runtime.Object{
				tenantPVC("pvc-4", "default", "12345"),
			},
			ExpectedCreatedPObject: []string{
				superDefaultNSName + "/pvc-4",
			},
			WaitDWS: true,
		},
		"pPVC is lost, vPVC is bound": {
			ExistingObjectInSuper: []runtime.Object{
				applyStatusToPVC(superPVC("pvc-3", superDefaultNSName, "12345", defaultClusterKey), statusLost),
			},
			ExistingObjectInTenant: []runtime.Object{
				applyStatusToPVC(tenantPVC("pvc-3", "default", "12345"), statusBound),
			},
			// TODO: Set ExpectedUpdatedVObject with Status.Phase="Lost"
			ExpectedNoOperation: false,
			WaitUWS:             true,
		},
		"pPVC is bound, vPVC is pending": {
			ExistingObjectInSuper: []runtime.Object{
				applyStatusToPVC(superPVC("pvc-3", superDefaultNSName, "12345", defaultClusterKey), statusBound),
			},
			ExistingObjectInTenant: []runtime.Object{
				applyStatusToPVC(tenantPVC("pvc-3", "default", "12345"), statusPending),
			},
			ExpectedUpdatedVObject: []runtime.Object{},
			ExpectedNoOperation:    true,
		},
		"pPVC is pending, vPVC is pending": {
			ExistingObjectInSuper: []runtime.Object{
				applyStatusToPVC(superPVC("pvc-3", superDefaultNSName, "12345", defaultClusterKey), statusPending),
			},
			ExistingObjectInTenant: []runtime.Object{
				applyStatusToPVC(tenantPVC("pvc-3", "default", "12345"), statusPending),
			},
			ExpectedUpdatedVObject: []runtime.Object{},
			ExpectedNoOperation:    true,
		},
	}

	for k, tc := range testcases {
		t.Run(k, func(t *testing.T) {
			tenantActions, superActions, err := util.RunPatrol(NewPVCController, testTenant, tc.ExistingObjectInSuper, tc.ExistingObjectInTenant, nil, tc.WaitDWS, tc.WaitUWS, nil)
			if err != nil {
				t.Errorf("%s: error running patrol: %v", k, err)
				return
			}

			if tc.ExpectedNoOperation {
				if len(superActions) != 0 {
					t.Errorf("%s: Expect no operation, got %v in super cluster", k, superActions)
					return
				}
				if len(tenantActions) != 0 {
					t.Errorf("%s: Expect no operation, got %v tenant cluster", k, tenantActions)
					return
				}
				return
			}

			if tc.ExpectedDeletedPObject != nil {
				if len(tc.ExpectedDeletedPObject) != len(superActions) {
					t.Errorf("%s: Expected to delete pPVC %#v. Actual actions were: %#v", k, tc.ExpectedDeletedPObject, superActions)
					return
				}
				for i, expectedName := range tc.ExpectedDeletedPObject {
					action := superActions[i]
					if !action.Matches("delete", "persistentvolumeclaims") {
						t.Errorf("%s: Unexpected action %s", k, action)
						continue
					}
					fullName := action.(core.DeleteAction).GetNamespace() + "/" + action.(core.DeleteAction).GetName()
					if fullName != expectedName {
						t.Errorf("%s: Expect to delete pPVC %s, got %s", k, expectedName, fullName)
					}
				}
			}
			if tc.ExpectedDeletedVObject != nil {
				if len(tc.ExpectedDeletedVObject) != len(tenantActions) {
					t.Errorf("%s: Expected to delete VPVC %#v. Actual actions were: %#v", k, tc.ExpectedDeletedVObject, tenantActions)
					return
				}
				for i, expectedName := range tc.ExpectedDeletedVObject {
					action := tenantActions[i]
					if !action.Matches("delete", "persistentvolumeclaims") {
						t.Errorf("%s: Unexpected action %s", k, action)
						continue
					}
					fullName := action.(core.DeleteAction).GetNamespace() + "/" + action.(core.DeleteAction).GetName()
					if fullName != expectedName {
						t.Errorf("%s: Expect to delete pPVC %s, got %s", k, expectedName, fullName)
					}
				}
			}
			if tc.ExpectedCreatedPObject != nil {
				if len(tc.ExpectedCreatedPObject) != len(superActions) {
					t.Errorf("%s: Expected to create PPVC %#v. Actual actions were: %#v", k, tc.ExpectedCreatedPObject, superActions)
					return
				}
				for i, expectedName := range tc.ExpectedCreatedPObject {
					action := superActions[i]
					if !action.Matches("create", "persistentvolumeclaims") {
						t.Errorf("%s: Unexpected action %s", k, action)
						continue
					}
					created := action.(core.CreateAction).GetObject().(*corev1.PersistentVolumeClaim)
					fullName := created.Namespace + "/" + created.Name
					if fullName != expectedName {
						t.Errorf("%s: Expect to create pPVC %s, got %s", k, expectedName, fullName)
					}
				}
			}
			if tc.ExpectedUpdatedPObject != nil {
				if len(tc.ExpectedUpdatedPObject) != len(superActions) {
					t.Errorf("%s: Expected to update PPVC %#v. Actual actions were: %#v", k, tc.ExpectedUpdatedPObject, superActions)
					return
				}
				for i, obj := range tc.ExpectedUpdatedPObject {
					action := superActions[i]
					if !action.Matches("update", "persistentvolumeclaims") {
						t.Errorf("%s: Unexpected action %s", k, action)
					}
					actionObj := action.(core.UpdateAction).GetObject()
					if !equality.Semantic.DeepEqual(obj, actionObj) {
						t.Errorf("%s: Expected updated pPVC is %v, got %v", k, obj, actionObj)
					}
				}
			}
			if tc.ExpectedUpdatedVObject != nil {
				if len(tc.ExpectedUpdatedVObject) != len(tenantActions) {
					t.Errorf("%s: Expected to update VPVC %#v. Actual actions were: %#v", k, tc.ExpectedUpdatedVObject, tenantActions)
					return
				}
				for i, obj := range tc.ExpectedUpdatedVObject {
					action := tenantActions[i]
					if !action.Matches("update", "persistentvolumeclaims") {
						t.Errorf("%s: Unexpected action %s", k, action)
					}
					actionObj := action.(core.UpdateAction).GetObject()
					if !equality.Semantic.DeepEqual(obj, actionObj) {
						t.Errorf("%s: Expected updated vPVC is %v, got %v", k, obj, actionObj)
					}
				}
			}
		})
	}
}
