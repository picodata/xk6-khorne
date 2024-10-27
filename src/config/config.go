package config

import (
	"os"
	"path/filepath"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

// GetConfig creates a new k8s config
func GetConfig() (*rest.Config, error) {
	configPath := getConfigPath()
	if _, err := fileExists(configPath); err != nil {
		return nil, err
	}

	config, _ := clientcmd.BuildConfigFromFlags("", configPath)
	return config, nil
}

// GetConfigPath fetches the path to the users kubeconfig
func getConfigPath() string {
	if configPath := os.Getenv("KHORNE_KUBECONFIG"); configPath != "" {
		return configPath
	}

	return filepath.Join(homedir.HomeDir(), ".kube", "config")
}

func fileExists(filename string) (bool, error) {
	_, err := os.Stat(filename)
	if err == nil {
		return true, nil
	}
	return false, err
}
