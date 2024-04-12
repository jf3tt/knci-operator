package controller

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"path/filepath"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func GenerateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"

	var seededRand *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))

	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func CreatePod(repoName string, repoUrl string, repoAccessToken string) {

	// var kubeconfig string
	var config *rest.Config
	var err error

	if home := homedir.HomeDir(); home != "" && filepath.Join(home, ".kube", "config") != "" {
		kubeconfig := filepath.Join(home, ".kube", "config")
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		fmt.Println("fetched local config")
		if err != nil {
			fmt.Println("error: fetched local config")
			panic(err.Error())
		}
	} else {
		config, err = rest.InClusterConfig()
		fmt.Println("fetched k8s config")
		if err != nil {
			fmt.Println("error: fetched k8s config")
			panic(err.Error())
		}
	}

	gitCloneCommand := "git clone " + repoUrl + " /repo && tree"

	// config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	// if err != nil {
	// 	panic(err.Error())
	// }

	// Создайте клиент Kubernetes используя конфигурацию
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	jobs := clientset.BatchV1().Jobs("knci-system")
	// Создание спецификации пода
	jobId := GenerateRandomString(5)
	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      repoName + "-job-" + jobId,
			Namespace: "knci-system",
		},
		Spec: batchv1.JobSpec{
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					Volumes: []v1.Volume{
						{
							Name: "git-repo",
							VolumeSource: v1.VolumeSource{
								EmptyDir: &v1.EmptyDirVolumeSource{},
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
					Containers: []v1.Container{
						{
							Name:    "job",
							Image:   "alpine:latest",
							Command: []string{"sh", "-c", "tree /repo"},
							VolumeMounts: []v1.VolumeMount{
								{
									Name:      "git-repo",
									MountPath: "/repo",
								},
							},
						},
					},
					RestartPolicy: v1.RestartPolicyNever,
				},
			},
		},
	}

	// Создание пода
	buildJob, err := jobs.Create(context.TODO(), job, metav1.CreateOptions{})
	if err != nil {
		log.Fatalln("Failed to create K8s job.")
	}
	fmt.Printf("Pod created: %s\n", buildJob.Name)

	// return "dsds"
}
