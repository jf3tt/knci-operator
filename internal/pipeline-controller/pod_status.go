package pipelineController

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func checkContainerStatus(pod v1.Pod) {
	pipelineLogger := log.WithFields(log.Fields{
		"pod": pod.Name,
	})
	podStatus := pod.Status.Conditions
	for _, condition := range podStatus {
		if condition.Reason == "PodCompleted" {
			fmt.Println(condition.Reason)
			pipelineLogger.Debug("Pod status changed to ", condition.Reason)
			break
		}
	}

	if pod.Status.Phase == "Failed" {
		for _, condition := range podStatus {
			pipelineLogger.Debug("Pod failed with error: ", condition.Reason)
			break
		}
	}
}

func checkPodStatus(pod v1.Pod) {

	clientset := kubernetesAuth()
	watcher, err := clientset.CoreV1().Pods("knci-system").Watch(context.TODO(), metav1.ListOptions{
		FieldSelector: fmt.Sprintf("metadata.name=%s", pod.Name),
	})
	if err != nil {
		log.Fatalf("Error setting up watch for pod: %s", err.Error())
	}

	defer watcher.Stop()

	for event := range watcher.ResultChan() {
		p := event.Object.(*v1.Pod)
		checkContainerStatus(*p)
	}
}
