package restart

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// RestartKeyword will gracefully rollout-restart every Deployment
// whose name contains the given keyword, across all namespaces.
func RestartKeyword(ctx context.Context, clientset kubernetes.Interface, keyword string) error {
	nsList, err := clientset.CoreV1().
		Namespaces().
		List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}

	sem := make(chan struct{}, 5)
	g, ctx := errgroup.WithContext(ctx)
	var mu sync.Mutex
	var total, restarted int

	for _, ns := range nsList.Items {
		namespace := ns.Name
		sem <- struct{}{}

		g.Go(func() error {
			defer func() { <-sem }()
			deps, err := clientset.AppsV1().
				Deployments(namespace).
				List(ctx, metav1.ListOptions{})
			if err != nil {
				return fmt.Errorf("list %s: %w", namespace, err)
			}

			mu.Lock()
			total += len(deps.Items)
			mu.Unlock()

			for _, d := range deps.Items {
				if strings.Contains(d.Name, keyword) {
					if d.Spec.Template.Annotations == nil {
						d.Spec.Template.Annotations = map[string]string{}
					}
					d.Spec.Template.Annotations["kubectl.kubernetes.io/restartedAt"] =
						time.Now().Format(time.RFC3339)

					if _, err := clientset.AppsV1().
						Deployments(namespace).
						Update(ctx, &d, metav1.UpdateOptions{}); err != nil {
						return err
					}

					mu.Lock()
					restarted++
					mu.Unlock()
				}
			}
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return err
	}

	fmt.Printf("Checked %d, restarted %d\n", total, restarted)
	return nil
}
