package restart_test

import (
	"context"
	"errors"
	"testing"

	"github.com/clarkent86/figure-technologies-devops-skills-assesment/restart"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/fake"
	clienttesting "k8s.io/client-go/testing"
)

func TestRestartDatabases(t *testing.T) {
	tests := []struct {
		name                string
		namespaces          []string
		deployments         map[string][]string
		namespaceListErr    bool
		deploymentListErrNs map[string]bool
		wantErr             bool
	}{
		{
			name:             "namespace list error",
			namespaceListErr: true,
			wantErr:          true,
		},
		{
			name:       "no namespaces",
			namespaces: []string{},
			wantErr:    false,
		},
		{
			name:       "no matching deployments",
			namespaces: []string{"ns1"},
			deployments: map[string][]string{
				"ns1": {"app1", "dbapp"},
			},
			wantErr: false,
		},
		{
			name:       "matching deployments",
			namespaces: []string{"ns1"},
			deployments: map[string][]string{
				"ns1": {"database-app", "other"},
			},
			wantErr: false,
		},
		{
			name:                "deployment list error",
			namespaces:          []string{"ns1"},
			deploymentListErrNs: map[string]bool{"ns1": true},
			wantErr:             true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create fake clientset
			cs := fake.NewSimpleClientset()

			// Reactor for namespace list error
			if tt.namespaceListErr {
				cs.PrependReactor("list", "namespaces", func(action clienttesting.Action) (handled bool, ret runtime.Object, err error) {
					return true, nil, errors.New("list namespaces error")
				})
			} else {
				// Add namespaces
				for _, ns := range tt.namespaces {
					cs.Tracker().Create(
						schema.GroupVersionResource{Group: "", Version: "v1", Resource: "namespaces"},
						&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: ns}},
						"", metav1.CreateOptions{},
					)

				}
			}

			// Setup deployments or reactors for each namespace
			for _, ns := range tt.namespaces {
				if tt.deploymentListErrNs[ns] {
					cs.PrependReactor("list", "deployments", func(action clienttesting.Action) (bool, runtime.Object, error) {
						la := action.(clienttesting.ListAction)
						if la.GetNamespace() == ns {
							return true, nil, errors.New("list deployments error")
						}
						return false, nil, nil
					})
				} else {
					// Add deployment objects
					for _, d := range tt.deployments[ns] {
						cs.Tracker().Create(
							schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"},
							&appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Name: d, Namespace: ns}},
							ns, metav1.CreateOptions{},
						)
					}
				}
			}

			err := restart.RestartDatabases(context.Background(), cs, "database")
			if (err != nil) != tt.wantErr {
				t.Fatalf("restartDatabases() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
