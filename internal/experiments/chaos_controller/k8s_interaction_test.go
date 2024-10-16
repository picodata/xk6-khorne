package chaos_controller

import (
	"context"
	"strings"
	"testing"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestCheckUnhealthyNodeValidation(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	p := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: "ill-pod",
		},
		Status: v1.PodStatus{
			Conditions: []v1.PodCondition{
				{
					Type:   v1.PodReady,
					Status: v1.ConditionFalse,
				},
			},
		}}
	clientset.CoreV1().Pods("test-ns").Create(context.TODO(), p, metav1.CreateOptions{})
	err := CheckSingleNodeHealth(clientset, "test-ns", "ill-pod")
	if err != nil {
		if !strings.Contains(err.Error(), "node ill-pod is not ready. Status: False") {
			t.Fatalf("Invalid error message. Got: " + err.Error())
		}
	} else {
		t.Fatalf("Unhealthy pod is determined as healthy")
	}
}

func TestCheckHealthyNodeValidation(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	p := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: "ill-pod",
		},
		Status: v1.PodStatus{
			Conditions: []v1.PodCondition{
				{
					Type:   v1.PodReady,
					Status: v1.ConditionTrue,
				},
			},
		}}
	clientset.CoreV1().Pods("test-ns").Create(context.TODO(), p, metav1.CreateOptions{})
	err := CheckSingleNodeHealth(clientset, "test-ns", "ill-pod")

	if err != nil {
		t.Fatalf("Node is excepcted to be healthy, error: " + err.Error())
	}
}
