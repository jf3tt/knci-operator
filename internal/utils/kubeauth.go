package kubeauth

import (
	"fmt"
	"os"
	"path/filepath"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

var config *rest.Config
var err error

func GetKubeConfig() (*rest.Config, error) {
	fmt.Println("Getting Kubeconfig")
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
	return config, err
}
