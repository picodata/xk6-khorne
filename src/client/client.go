package client

import (
	"os"
	"path/filepath"

	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

// New creates a new k8s client
func New() (*kubernetes.Clientset, error) {
	config, err := GetConfig()
	if err != nil {
		return nil, err
	}

	if clientset, err := kubernetes.NewForConfig(config); err != nil {
		return nil, err
	} else {
		return clientset, nil
	}
}

// GetConfigPath fetches the path to the users kubeconfig
func GetConfigPath() string {
	if configPath := os.Getenv("KHORNE_KUBECONFIG"); configPath != "" {
		return configPath
	}

	return filepath.Join(homedir.HomeDir(), ".kube", "config")
}

// GetConfig creates a new k8s config
func GetConfig() (*rest.Config, error) {
	configPath := GetConfigPath()
	if _, err := fileExists(configPath); err != nil {
		return nil, err
	}

	config, _ := clientcmd.BuildConfigFromFlags("", configPath)
	return config, nil
}

func fileExists(filename string) (bool, error) {
	_, err := os.Stat(filename)
	if err == nil {
		return true, nil
	}
	return false, err
}
