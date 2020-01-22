package k8s

import (
	"os"
	"path/filepath"
)

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}

func GetKubeConfig(path string) string {
	var kubeconfig string
	if path == "" {
		home := homeDir()
		if home != "" {
			kubeconfig = filepath.Join(home, ".kube", "config")
		}
	} else {
		kubeconfig = path
	}

	return kubeconfig
}
