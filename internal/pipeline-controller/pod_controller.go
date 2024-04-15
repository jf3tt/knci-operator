package pipelineController

import (
	"context"
	"fmt"
	civ1 "knci/api/v1"
	kauth "knci/internal/utils"

	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type Pipeline struct {
	Name   string
	CI     civ1.CI
	Status string
}

type Log struct {
	Owner   string
	Message string
}

func CreatePipeline(ci *civ1.CI) Pipeline {
	var pipeline Pipeline

	pipeline.CreateJob(ci)

	return pipeline
}

// func (ci civ1.CI) Finalizer() {
// 	CheckForDeleting(ci, ctx, r)
// }

func kubernetesAuth() *kubernetes.Clientset {
	var config *rest.Config
	var err error

	config, err = kauth.GetKubeConfig()
	if err != nil {
		fmt.Println("Error getting kubernetes configuration)")
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		fmt.Println("Error getting kubernetes configuration)")
	}
	return clientset
}

func (p Pipeline) CreateJob(ci *civ1.CI) v1.Pod {
	var podSpec v1.Pod

	var err error
	clientset := kubernetesAuth()

	podSpec = getPodTemplate(*ci)

	pod, err := clientset.CoreV1().Pods("knci-system").Create(context.TODO(), &podSpec, metav1.CreateOptions{})
	if err != nil {
		log.Fatalf("Error creating pod: %s", err.Error())
	}
	log.Debug("Pod created: ", pod.ObjectMeta.Name)
	return *pod
}
