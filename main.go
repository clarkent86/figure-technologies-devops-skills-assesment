package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/clarkent86/figure-technologies-devops-skills-assesment/restart"
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
		fmt.Println("Error building kubeconfig:", err)
		os.Exit(0)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	err = restart.RestartKeyword(ctx, clientset, keyword)
	if err != nil {
		fmt.Println("Error restarting databases:", err)
	}
}
