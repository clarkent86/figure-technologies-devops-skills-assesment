package main

import (
	"context"
	"flag"
	"os"
	"path/filepath"

	"github.com/figure-technologies-devops-skills-assesment/restart"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	keyword = "database" // keyword to filter deployments
)

func main() {
	ctx := context.TODO()

	// Load kubeconfig
	home, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	kubeconfig := flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "path to kubeconfig")
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	restart.RestartDatabases(ctx, clientset, keyword)
}
