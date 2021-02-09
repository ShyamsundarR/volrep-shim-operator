/*
Copyright 2021 The RamenDR authors.

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

package controllers

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	replicationv1alpha1 "github.com/volrep-shim-operator/api/v1alpha1"
)

// VolumeReplicationReconciler reconciles a VolumeReplication object
type VolumeReplicationReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

const (
	pvcKind = "PersistentVolumeClaim"
)

// +kubebuilder:rbac:groups=replication.storage.ramen.io,resources=volumereplications,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=replication.storage.ramen.io,resources=volumereplications/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=replication.storage.ramen.io,resources=volumereplications/finalizers,verbs=update
// +kubebuilder:rbac:groups=core,resources=persistentvolumeclaims,verbs=get;list;watch
// +kubebuilder:rbac:groups=core,resources=persistentvolumes,verbs=get;list;watch

// High level TODOs
// - Add Ceph secrets to operator, for ceph commands
// - Build operator from a ceph client container base image for required Ceph commands
// - Generate RBD image name from PV CSI volumeHandle
// - Run a sample ceph command to test image availability and reflect the same in CRD status
// - More from here on...

// High level questions
// - Why does any resource that we get add a watch to the same? Hence requiring a RBAC for watch as well?
// - How to update sample.yaml
// - How to make certain CRD fields mandatory and with some validation

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the VolumeReplication object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.7.0/pkg/reconcile
func (r *VolumeReplicationReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var statusErr error
	logger := r.Log.WithValues("volumereplication", req.NamespacedName)
	logger.Info("reconcile started")

	// Get the CR for this reconcile instance
	volReplication := &replicationv1alpha1.VolumeReplication{}
	err := r.Get(ctx, req.NamespacedName, volReplication)
	if err != nil {
		// NOTE: The reconciler manager puts this back in the queue with an exponential
		// backoff, so no requeue from our end. Further, there is a stack trace that is printed on
		// error returns in the logs, see: https://github.com/operator-framework/operator-sdk/issues/1615
		// so avoiding error returns unless it is critcal (and that is possibly correct as well)
		if !kerrors.IsNotFound(err) {
			logger.Error(err, "failed to get VolumeReplication CR")
		}
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	logger.Info("will reconcile", "spec", volReplication.Spec)

	conditionSuccess := metav1.Condition{
		Type:    "Reconciled",
		Status:  metav1.ConditionTrue,
		Reason:  "Complete",
		Message: "Reconcile complete",
	}

	conditionFailure := metav1.Condition{
		Type:    "Reconciled",
		Status:  metav1.ConditionFalse,
		Reason:  "Error",
		Message: "Reconcile error",
	}

	defer func() {
		if err != nil && statusErr == nil {
			// Update failure status
			conditionFailure.Message = err.Error()
			meta.SetStatusCondition(&volReplication.Status.Conditions, conditionFailure)
			// May not succeed, but err is non-nil and will be retried
			_ = r.Status().Update(ctx, volReplication)
		}
	}()

	if volReplication.Spec.DataSource == nil {
		err = fmt.Errorf("missing dataSource %v", volReplication.Spec)
		return ctrl.Result{}, err
	}

	if volReplication.Spec.DataSource.Kind != pvcKind {
		err = fmt.Errorf("unsupported data source in resource %v", volReplication.Spec.DataSource.Kind)
		return ctrl.Result{}, err
	}

	if volReplication.Spec.State != replicationv1alpha1.ReplicationPrimary &&
		volReplication.Spec.State != replicationv1alpha1.ReplicationSecondary {
		err = fmt.Errorf("unsupported state in resource %v", volReplication.Spec.State)
		return ctrl.Result{}, err
	}

	// Get and validate PVC for VolumeReplication reconcile instance
	replicationPVC := &corev1.PersistentVolumeClaim{}
	pvcObjectKey := client.ObjectKey{
		Namespace: req.NamespacedName.Namespace,
		Name:      volReplication.Spec.DataSource.Name,
	}
	err = r.Get(ctx, pvcObjectKey, replicationPVC)
	if err != nil {
		err = fmt.Errorf("failed to get PersistentVolumeClaim for PVC %+v", pvcObjectKey)
		return ctrl.Result{}, err
	}
	logger.Info("found", "PVC", replicationPVC)

	// Check if PVC is bound
	if replicationPVC.Status.Phase != corev1.ClaimBound {
		err = fmt.Errorf("PVC is not yet bound, status %v", replicationPVC.Status.Phase)
		return ctrl.Result{}, err
	}

	// Get the PV
	replicationPV := &corev1.PersistentVolume{}
	pvObjectKey := client.ObjectKey{
		Name: replicationPVC.Spec.VolumeName,
	}
	if pvObjectKey.Name == "" {
		err = fmt.Errorf("invalid PVC state, bound with no volume name")
		return ctrl.Result{}, err
	}

	err = r.Get(ctx, pvObjectKey, replicationPV)
	if err != nil {
		err = fmt.Errorf("failed to get PersistentVolume %+v for PVC %+v", pvObjectKey, pvcObjectKey)
		return ctrl.Result{}, err
	}
	logger.Info("found", "PV", replicationPV)

	// Update status
	meta.SetStatusCondition(&volReplication.Status.Conditions, conditionSuccess)
	statusErr = r.Status().Update(ctx, volReplication)
	if err == nil {
		err = statusErr
	}

	logger.Info("Reconcile finished")
	return ctrl.Result{}, err
}

// SetupWithManager sets up the controller with the Manager.
func (r *VolumeReplicationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&replicationv1alpha1.VolumeReplication{}).
		Complete(r)
}
