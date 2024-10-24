package chaos

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/picodata/xk6-khorne/src/client"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/dynamic"
)

type ChaosController struct {
	DynamicClient dynamic.Interface
	Namespace     string
}

// NewController creates a new instance of ChaosController
func NewController(namespace string) *ChaosController {
	config, err := client.GetConfig()
	if err != nil {
		log.Fatal(err.Error())
		return nil
	}

	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		log.Fatalf("Failed to create Kubernetes client: %v", err)
	}

	return &ChaosController{dynamicClient, namespace}
}

func (cc *ChaosController) listNodes(namespace string) ([]string, error) {
	podGVR := schema.GroupVersionResource{
		Group:    "", // Core API group
		Version:  "v1",
		Resource: "pods",
	}

	podList, err := cc.DynamicClient.Resource(podGVR).Namespace(namespace).List(context.TODO(), v1.ListOptions{})
	if err != nil {
		return nil, err
	}

	alivePods := make([]string, 0)
	for _, pod := range podList.Items {
		alivePods = append(alivePods, pod.GetName())
	}

	return alivePods, nil
}

func (cc *ChaosController) deleteNode(namespace string, podName string) error {
	podGVR := schema.GroupVersionResource{
		Group:    "", // Core API group
		Version:  "v1",
		Resource: "pods",
	}

	return cc.DynamicClient.Resource(podGVR).Namespace(namespace).Delete(context.TODO(), podName, v1.DeleteOptions{})
}

// configConfigPath    - path to k8s cluster config
// killNodes 	- to kill or not to kill affected nodes after each run
func RunChaosExperiment(namespace string, chaosConfigPath string) error {
	cc := NewController(namespace)

	// Work with experiment yaml file
	experimentFile, err := os.ReadFile(chaosConfigPath)
	if err != nil {
		log.Fatalf("Failed to read manifest file: %v", err)
		return err
	}
	reader := bytes.NewReader(experimentFile)
	decoder := yaml.NewYAMLOrJSONDecoder(reader, 1000)
	parsedYaml := &unstructured.Unstructured{}

	if err := decoder.Decode(parsedYaml); err != nil {
		log.Fatalf("Failed to decode YAML: %v", err)
		return err
	}

	expType, _, err := unstructured.NestedString(parsedYaml.Object, "kind")
	if err != nil {
		return err
	}
	gvr := schema.GroupVersionResource{
		Group:    "chaos-mesh.org",         // ChaosMesh group
		Version:  "v1alpha1",               // Version of the ChaosMesh resource
		Resource: strings.ToLower(expType), // Resource is "podchaos"
	}

	// Apply the resource (equivalent to `kubectl apply -f`)
	_, err = cc.DynamicClient.Resource(gvr).Namespace(cc.Namespace).Create(context.TODO(), parsedYaml, v1.CreateOptions{})
	if err != nil {
		log.Fatalf("Failed to apply ChaosMesh resource: %v", err)
	}

	return err
}

// Deleting the experiment from chaosmesh list
// to avoid collisions, when the same experiment is started
func ClearChaosCache(namespace string) error {
	cc := NewController(namespace)
	var chaosTypes = []string{"podchaos", "networkchaos"}

	for _, chaosType := range chaosTypes {
		gvr := schema.GroupVersionResource{
			Group:    "chaos-mesh.org", // ChaosMesh group
			Version:  "v1alpha1",       // Version of the ChaosMesh resource
			Resource: chaosType,
		}

		err := cc.DynamicClient.Resource(gvr).Namespace(cc.Namespace).DeleteCollection(
			context.TODO(),
			v1.DeleteOptions{},
			v1.ListOptions{},
		)
		if err != nil {
			log.Fatalf("Failed to delete Network resources: %v", err)
			return err
		}
	}

	return nil
}

// Check the health of each node in cluster
// Returning string so that JS can handle it correctly
func CheckClusterHealth(namespace string) error {
	cc := NewController(namespace)

	allPods, err := cc.listNodes(namespace)
	if err != nil {
		return err
	}

	for _, podName := range allPods {
		if err := cc.checkSingleNodeHealth(podName); err != nil {
			return err
		}
	}

	return nil
}

// Check node health, params are suited for go tests
// Use k8s interface to pass both kubernetes.client and fake.client
func (cc *ChaosController) checkSingleNodeHealth(podName string) error {
	podGVR := schema.GroupVersionResource{
		Group:    "", // Core API group
		Version:  "v1",
		Resource: "pods",
	}

	// Get the pod
	pod, err := cc.DynamicClient.Resource(podGVR).Namespace(cc.Namespace).Get(context.TODO(), podName, v1.GetOptions{})
	if err != nil {
		return err
	}

	status, found, err := unstructured.NestedMap(pod.Object, "status")
	if err != nil || !found {
		return fmt.Errorf("failed to get status for pod %s: %v", podName, err)
	}

	conditions, found, err := unstructured.NestedSlice(status, "conditions")
	if err != nil || !found {
		return fmt.Errorf("failed to get conditions for pod %s: %v", podName, err)
	}

	for _, conditionObj := range conditions {
		condition := conditionObj.(map[string]interface{})

		if conditionType, found := condition["type"]; found && conditionType == "Ready" {
			if conditionStatus, found := condition["status"]; found && conditionStatus != "True" {
				reason, _ := condition["reason"].(string)
				message, _ := condition["message"].(string)
				return fmt.Errorf("pod %s is not ready. status: %s, reason: %s, message: %s",
					podName, conditionStatus, reason, message,
				)
			}
		}
	}

	return nil
}

// Get all podes, affected by chaos experiment and reload them
// This is done in order to avoid CrashLoopBackOff pod status
func RestartPods(namespace string, pods []string) error {
	cc := NewController(namespace)

	for _, podName := range pods {
		if err := cc.deleteNode(cc.Namespace, podName); err != nil {
			return err
		}
	}

	return nil
}
