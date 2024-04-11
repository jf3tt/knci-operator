package controller

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func CreatePod(repoUrl string, repoAccessToken string) {

	var kubeconfig string
	if home, err := os.UserHomeDir(); err == nil {
		kubeconfig = filepath.Join(home, ".kube", "config")
	}

	gitCloneCommand := "git clone " + repoUrl + " /shared && tree"

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	// Создайте клиент Kubernetes используя конфигурацию
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	jobs := clientset.BatchV1().Jobs("sandbox")
	// Создание спецификации пода
	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "job",
			Namespace: "sandbox",
		},
		Spec: batchv1.JobSpec{
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					Volumes: []v1.Volume{
						{
							Name: "shared-data",
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
									Name:      "shared-data",
									MountPath: "/shared",
								},
							},
						},
					},
					Containers: []v1.Container{
						{
							Name:    "job",
							Image:   "alpine:latest",
							Command: []string{"sh", "-c", "tree /shared"},
							VolumeMounts: []v1.VolumeMount{
								{
									Name:      "shared-data",
									MountPath: "/shared",
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
	job1, err := jobs.Create(context.TODO(), job, metav1.CreateOptions{})
	if err != nil {
		log.Fatalln("Failed to create K8s job.")
	}
	// jobsClient := clientset.CoreV1().Jods("sandbox")
	// job, err = jobsClient.Create(context.TODO(), job, metav1.CreateOptions{})
	// if err != nil {
	// 	panic(err.Error())
	// }
	fmt.Printf("Pod created: %s\n", job.ObjectMeta.Name)
	fmt.Println(job1)
	// return "dsds"
}
