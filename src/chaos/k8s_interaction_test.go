package chaos

import (
	"context"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/kubernetes/scheme"
)

func TestCheckUnhealthyNodeValidation(t *testing.T) {
	clientset := fake.NewSimpleDynamicClient(scheme.Scheme)
	// Create a fake dynamic client
	// Create an unstructured Pod
	p := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "Pod",
			"metadata": map[string]interface{}{
				"name":      "ill-pod",
				"namespace": "test-namespace",
			},
			"status": map[string]interface{}{
				"conditions": []interface{}{
					map[string]interface{}{
						"type":    "Ready",
						"status":  "False",
						"reason":  "Feeling lazy today",
						"message": "Don't worry - be happy",
					},
				},
			},
		},
	}

	// Define the GroupVersionResource for Pods
	gvr := schema.GroupVersionResource{Group: "", Version: "v1", Resource: "pods"}

	// Create the pod using the dynamic client
	_, err := clientset.Resource(gvr).Namespace("test-namespace").Create(context.TODO(), p, metav1.CreateOptions{})
	if err != nil {
		panic(err) // Handle error appropriately
	}
	chaosController := NewController("test-namespace")
	chaosController.DynamicClient = clientset

	err = chaosController.checkSingleNodeHealth("ill-pod")
	if err == nil {
		t.Fatalf("Unhealthy pod is determined as healthy")
	}
}

func TestCheckHealthyNodeValidation(t *testing.T) {
	clientset := fake.NewSimpleDynamicClient(scheme.Scheme)
	// Create a fake dynamic client
	// Create an unstructured Pod
	p := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "Pod",
			"metadata": map[string]interface{}{
				"name":      "ill-pod",
				"namespace": "test-namespace",
			},
			"status": map[string]interface{}{
				"conditions": []interface{}{
					map[string]interface{}{
						"type":    "Ready",
						"status":  "True",
						"reason":  "Grind mood",
						"message": "Work work work",
					},
				},
			},
		},
	}

	// Define the GroupVersionResource for Pods
	gvr := schema.GroupVersionResource{Group: "", Version: "v1", Resource: "pods"}

	// Create the pod using the dynamic client
	_, err := clientset.Resource(gvr).Namespace("test-namespace").Create(context.TODO(), p, metav1.CreateOptions{})
	if err != nil {
		panic(err) // Handle error appropriately
	}
	chaosController := NewController("test-namespace")
	chaosController.DynamicClient = clientset

	err = chaosController.checkSingleNodeHealth("ill-pod")
	if err != nil {
		t.Fatalf("Node is excepcted to be healthy, error: " + err.Error())
	}
}
