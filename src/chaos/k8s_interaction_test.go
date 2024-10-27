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
	// Create a fake dynamic client
	clientset := fake.NewSimpleDynamicClient(scheme.Scheme)
	// Create unhealthy pod
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

	chaosController := ChaosController{clientset, "test-namespace"}

	err = chaosController.checkSingleNodeHealth("ill-pod")
	if err == nil {
		t.Fatalf("Unhealthy pod is determined as healthy")
	}
}

func TestCheckHealthyNodeValidation(t *testing.T) {
	// Create a fake dynamic client
	clientset := fake.NewSimpleDynamicClient(scheme.Scheme)
	// Create healthy pod
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

	chaosController := ChaosController{clientset, "test-namespace"}

	err = chaosController.checkSingleNodeHealth("ill-pod")
	if err != nil {
		t.Fatalf("Node is excepcted to be healthy, error: " + err.Error())
	}
}
