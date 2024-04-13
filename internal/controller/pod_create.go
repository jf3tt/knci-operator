package controller

import (
	"context"
	"fmt"
	civ1 "knci/api/v1"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func boolPtr(b bool) *bool {
	return &b
}

func hostPathTypePtr(t v1.HostPathType) *v1.HostPathType {
	return &t
}

func GenerateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"

	var seededRand *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))

	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

// +kubebuilder:rbac:groups=core,resources=pods,verbs=create;list;watch

func CreatePod(ci civ1.CI) {

	var config *rest.Config
	var err error

	if home := homedir.HomeDir(); home != "" {
		kubeconfig := filepath.Join(home, ".kube", "config")
		if _, err := os.Stat(kubeconfig); err == nil {
			config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
			if err != nil {
				fmt.Println("error: fetched local config")
				panic(err.Error())
			}
		} else {
			config, err = rest.InClusterConfig()
			if err != nil {
				fmt.Println("error: fetched k8s config")
				panic(err.Error())
			}
		}
	} else {
		config, err = rest.InClusterConfig()
		fmt.Println("fetched k8s config")
		if err != nil {
			fmt.Println("error: fetched k8s config")
			panic(err.Error())
		}
	}

	gitCloneCommand := "git clone " + ci.Spec.Repo.URL + " /repo && tree /repo"

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	podId := GenerateRandomString(5)

	var containers []v1.Container //nolint:prealloc
	for _, pod := range ci.Spec.Repo.Jobs {
		container := v1.Container{
			Name:    pod.Name,
			Image:   pod.Image,
			Command: pod.Commands,
			VolumeMounts: []v1.VolumeMount{
				{
					Name:      "git-repo",
					MountPath: "/repo",
				},
				{
					Name:      "docker-sock",
					MountPath: "/var/run/docker.sock",
				},
			},
			SecurityContext: &v1.SecurityContext{
				Privileged: boolPtr(true),
			},
		}
		containers = append(containers, container)
	}

	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ci.ObjectMeta.Name + "-job-" + podId,
			Namespace: "knci-system",
			Labels: map[string]string{
				"ci.knci.io/name": ci.ObjectMeta.Name,
			},
		},
		Spec: v1.PodSpec{
			Volumes: []v1.Volume{
				{
					Name: "git-repo",
					VolumeSource: v1.VolumeSource{
						EmptyDir: &v1.EmptyDirVolumeSource{},
					},
				},
				{
					Name: "docker-sock",
					VolumeSource: v1.VolumeSource{
						HostPath: &v1.HostPathVolumeSource{
							Path: "/run/k3s/containerd/containerd.sock",
							Type: hostPathTypePtr(v1.HostPathSocket),
						},
					},
				},
			},
			InitContainers: []v1.Container{
				{
					Name:    "init",
					Image:   "alpine/git:latest",
					Command: []string{"sh", "-c", gitCloneCommand},
					VolumeMounts: []v1.VolumeMount{
						{
							Name:      "git-repo",
							MountPath: "/repo",
						},
					},
				},
			},
			Containers:    containers,
			RestartPolicy: v1.RestartPolicyNever,
		},
	}

	_, err = clientset.CoreV1().Pods("knci-system").Create(context.Background(), pod, metav1.CreateOptions{})
	if err != nil {
		panic(err)
	}

}
