package v1

import (
	"fmt"

	v1 "k8s.io/api/core/v1"
)

type Pipeline struct {
	Name string
	Pods v1.Pod
}

func (p Pipeline) CreatePipeline(ci CI) Pipeline {
	// p.getPodTemplate(ci)
	// p.CreatePod()
	fmt.Println("Creating pipeline")
	return p
}
