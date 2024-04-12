package controller

import (
	"context"
	"fmt"
	civ1 "knci/api/v1"
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

func CreatePod(ci civ1.CI) {

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

	gitCloneCommand := "git clone " + ci.Spec.Repo.URL + " /repo && tree /repo"
	fmt.Println("Commands: ", ci.Spec.Repo.Jobs[1].Commands[0])

	// Создайте клиент Kubernetes используя конфигурацию
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	jobs := clientset.BatchV1().Jobs("knci-system")
	// Создание спецификации пода
	jobId := GenerateRandomString(5)
	for i := 0; i < len(ci.Spec.Repo.Jobs); i++ {
		fmt.Println(ci.Spec.Repo.Jobs[i].Image)
	}

	var containers []v1.Container
	for _, job := range ci.Spec.Repo.Jobs {
		container := v1.Container{
			Name:    job.Name,
			Image:   job.Image,
			Command: job.Commands,
			// SecurityContext: &v1.SecurityContext{
			// 	Privileged: boolPtr(true),
			// },
			VolumeMounts: []v1.VolumeMount{
				{
					Name:      "git-repo",
					MountPath: "/repo",
				},
			},
		}
		containers = append(containers, container)
	}

	fmt.Println("CONTAINERS: ", containers)

	// func boolPtr(b bool) *bool {
	// 	return &b
	// }

	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ci.ObjectMeta.Name + "-job-" + jobId,
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
					Containers:    containers,
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

	// return buildJob.Name
}
