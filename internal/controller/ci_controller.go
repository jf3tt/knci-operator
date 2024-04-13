/*
Copyright 2024.

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

package controller

import (
	"context"
	civ1 "knci/api/v1"

	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// CIReconciler reconciles a CI object
type CIReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=ci.knci,resources=cis,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=ci.knci,resources=cis/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=ci.knci,resources=cis/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the CI object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.17.2/pkg/reconcile

func removeString(slice []string, s string) []string {
	newSlice := []string{}
	for _, item := range slice {
		if item != s {
			newSlice = append(newSlice, item)
		}
	}
	return newSlice
}

func (r *CIReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	var ci civ1.CI
	err := r.Get(ctx, req.NamespacedName, &ci)
	if err != nil {
		if apierrors.IsNotFound(err) {
			log.Info("CI resource not found. Ignoring since object must be deleted", "name", req.NamespacedName.Name, "namespace", req.NamespacedName.Namespace)
			return ctrl.Result{}, nil
		}
		log.Error(err, "unable to fetch CI")
		return ctrl.Result{}, err
	}

	if !ci.ObjectMeta.DeletionTimestamp.IsZero() {
		log.Info("Deleting CI", "CI Name", ci.ObjectMeta.Name)

		podList := &v1.PodList{}

		listOpts := []client.ListOption{
			client.InNamespace("knci-system"),
			client.MatchingLabels(map[string]string{
				"ci.knci.io/name": ci.ObjectMeta.Name,
			}),
		}

		if err := r.List(ctx, podList, listOpts...); err != nil {
			return ctrl.Result{}, nil
		}

		for _, pod := range podList.Items {
			if err := r.Delete(ctx, &pod); err != nil {
				log.Info("Deleting error")
			}
		}

		ci.ObjectMeta.Finalizers = removeString(ci.ObjectMeta.Finalizers, "ci.knci.io/finalizer")
		if err := r.Update(ctx, &ci); err != nil {
			return ctrl.Result{}, err
		}

	} else {
		log.Info("Processing CI", "CI Name", ci.ObjectMeta.Name, "Repo URL", ci.Spec.Repo.URL, "Scrape Interval", ci.Spec.Repo.ScrapeInterval)

		CreatePod(ci)
	}

	return ctrl.Result{}, nil
}

// deleteAssociatedPods удаляет все поды, связанные с данной CI
func (r *CIReconciler) deleteAssociatedPods(ctx context.Context, ci *civ1.CI) error {
	podList := &v1.PodList{}
	listOpts := []client.ListOption{
		client.InNamespace("knci-system"),
		client.MatchingLabels(map[string]string{"app": "knci"}),
	}
	if err := r.List(ctx, podList, listOpts...); err != nil {
		return err
	}

	for _, pod := range podList.Items {
		if err := r.Delete(ctx, &pod); err != nil {
			return err
		}
	}

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *CIReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&civ1.CI{}).
		Complete(r)
}
