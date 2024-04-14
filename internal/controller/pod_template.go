package controller

import (
	civ1 "knci/api/v1"
	"math/rand"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func getBoolPtr(b bool) *bool {
	return &b
}

func GetHostPathTypePtr(t v1.HostPathType) *v1.HostPathType {
	return &t
}

func GenerateId(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"

	var seededRand *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))

	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func GetPodTemplate(ci civ1.CI) v1.Pod {
	gitCloneCommand := "git clone " + "---BLANK---" + " /repo && tree /repo"
	podId := GenerateId(5)
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
					MountPath: "/run/k3s/containerd/containerd.sock",
				},
			},
			SecurityContext: &v1.SecurityContext{
				Privileged: getBoolPtr(true),
			},
			Env: []v1.EnvVar{
				{
					Name:  "DOCKER_HOST",
					Value: "unix://var/run/docker.sock",
				},
			},
		}
		containers = append(containers, container)
	}
	// var pods v1.Pod

	podTemplate := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "ci.ObjectMeta.Name" + podId,
			Namespace: "knci-system",
			Labels: map[string]string{
				"ci.knci.io/name": "ci.ObjectMeta.Name",
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
							Type: GetHostPathTypePtr(v1.HostPathSocket),
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
	return *podTemplate
}
