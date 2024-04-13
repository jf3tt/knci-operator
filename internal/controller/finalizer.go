package controller

import (
	"context"
	civ1 "knci/api/v1"

	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func removeString(slice []string, s string) []string {
	newSlice := []string{}
	for _, item := range slice {
		if item != s {
			newSlice = append(newSlice, item)
		}
	}
	return newSlice
}

func CheckForDeleting(ci civ1.CI, ctx context.Context, r *CIReconciler) {
	log := log.FromContext(ctx)

	podList := &v1.PodList{}

	// get a list of pods by labels
	listOpts := []client.ListOption{
		client.InNamespace("knci-system"),
		client.MatchingLabels(map[string]string{
			"ci.knci.io/name": ci.ObjectMeta.Name,
		}),
	}

	if err := r.List(ctx, podList, listOpts...); err != nil {

	}

	// deleting pods produced by ci crd
	for _, pod := range podList.Items {
		if err := r.Delete(ctx, &pod); err != nil {
			log.Info("Deleting error")
		}
	}

	// deleting finalizers from ci crd
	ci.ObjectMeta.Finalizers = removeString(ci.ObjectMeta.Finalizers, "ci.knci.io/finalizer")
	if err := r.Update(ctx, &ci); err != nil {
		// return ctrl.Result{}, err
		log.Info("ERROR: UPDATE")
	}
	log.Info("Deleting completed", "CI Name", ci.ObjectMeta.Name)
}
