package controller

import (
	"fmt"
	civ1 "knci/api/v1"

	v1 "k8s.io/api/core/v1"
)

func CreateNewPipeline(ci *civ1.CI, pod v1.Pod) civ1.Pipeline {
	fmt.Println(ci.ObjectMeta.Name)
	GetPodTemplate(*ci)
	var pipeline civ1.Pipeline
	return pipeline
}
