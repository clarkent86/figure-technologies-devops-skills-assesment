package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
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

	namespaces, err := clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		panic(err)
	}

	for _, ns := range namespaces.Items {
		namespace := ns.Name

		deployments, err := clientset.AppsV1().Deployments(namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to list deployments in %s: %v\n", namespace, err)
			continue
		}

		for _, deploy := range deployments.Items {
			if strings.Contains(deploy.Name, "database") {
				fmt.Printf("Matched deployment: %s/%s\n", namespace, deploy.Name)

				if deploy.Spec.Template.Annotations == nil {
					deploy.Spec.Template.Annotations = map[string]string{}
				}
				deploy.Spec.Template.Annotations["kubectl.kubernetes.io/restartedAt"] = time.Now().Format(time.RFC3339)

				_, err = clientset.AppsV1().Deployments(namespace).Update(ctx, &deploy, metav1.UpdateOptions{})
				if err != nil {
					fmt.Fprintf(os.Stderr, "Failed to update %s/%s: %v\n", namespace, deploy.Name, err)
				} else {
					fmt.Printf("Restarted deployment: %s/%s\n", namespace, deploy.Name)
				}
			}
		}
	}
}
