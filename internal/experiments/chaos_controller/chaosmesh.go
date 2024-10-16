package chaos_controller

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
	"xk6-khorne/internal/k8s/pods"
	"xk6-khorne/pkg/k8s/client"

	coreV1 "k8s.io/api/core/v1"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

type ChaosController struct {
	cluster_config_path string
	pod                 *pods.Pods
	dynamicClient       *dynamic.DynamicClient
	namespace           string
	curExp              *unstructured.Unstructured
	gvr                 *schema.GroupVersionResource
}

// New creates a new podkiller
func New(namespace string) *ChaosController {
	c, err := client.New()
	p := pods.New(c)
	if err != nil {
		return nil
	}

	config_path := getConfigPath()

	config, err := clientcmd.BuildConfigFromFlags("", config_path)
	if err != nil {
		log.Fatalf("Failed to build kubeconfig: %v", err)
	}
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		log.Fatalf("Failed to create Kubernetes client: %v", err)
	}

	return &ChaosController{config_path, p, dynamicClient, namespace, nil, nil}
}

// configPath    - path to k8s cluster config
// killNodes 	- to kill or not to kill affected nodes after each run
// recoveryTime - how much time to wait for cluster to recover after experiment
func (cc *ChaosController) RunChaosExperiment(configPath string, restartNodes bool, recoveryTime string) error {
	// Work with experiment yaml file
	experiment_file, err := os.ReadFile(configPath)
	if err != nil {
		log.Fatalf("Failed to read manifest file: %v", err)
	}
	reader := bytes.NewReader(experiment_file)
	decoder := yaml.NewYAMLOrJSONDecoder(reader, 1000)
	parsedYaml := &unstructured.Unstructured{}

	if err := decoder.Decode(parsedYaml); err != nil {
		log.Fatalf("Failed to decode YAML: %v", err)
	}
	cc.curExp = parsedYaml

	exp_type, _, err := unstructured.NestedString(cc.curExp.Object, "kind")
	if err != nil {
		return err
	}
	gvr := schema.GroupVersionResource{
		Group:    "chaos-mesh.org",          // ChaosMesh group
		Version:  "v1alpha1",                // Version of the ChaosMesh resource
		Resource: strings.ToLower(exp_type), // Resource is "podchaos"
	}
	cc.gvr = &gvr

	// Apply the resource (equivalent to `kubectl apply -f`)
	_, err = cc.dynamicClient.Resource(gvr).Namespace(cc.namespace).Create(context.TODO(), parsedYaml, v1.CreateOptions{})
	if err != nil {
		log.Fatalf("Failed to apply ChaosMesh resource: %v", err)
	}

	cc.ExperimentSleep(recoveryTime)

	if restartNodes {
		RestartAffectedPods(cc)
	}
	return err
}

// Deleting the experiment from chaosmesh list
// to avoid collisions, when the same experiment is started
func (cc *ChaosController) ClearExperimentData(chaosType string) {
	gvr := schema.GroupVersionResource{
		Group:    "chaos-mesh.org",           // ChaosMesh group
		Version:  "v1alpha1",                 // Version of the ChaosMesh resource
		Resource: strings.ToLower(chaosType), // Resource is "podchaos"
	}

	err := cc.dynamicClient.Resource(gvr).Namespace(cc.namespace).DeleteCollection(
		context.TODO(),
		v1.DeleteOptions{},
		v1.ListOptions{},
	)
	if err != nil {
		log.Fatalf("Failed to delete PodChaos resources: %v", err)
	}
}

// Check the health of each node in cluster
// Returning string so that JS can handle it correctly
func (cc *ChaosController) CheckClusterHealth() string {
	allPods, err := cc.pod.List(context.Background(), cc.namespace)
	if err != nil {
		return err.Error()
	}

	err = cc.CheckPodsHealth(allPods)
	if err != nil {
		return err.Error()
	}
	return ""

}

// Check the status of pods, affected by the chaos experiment
func (cc *ChaosController) CheckAffectedPodsLiveness() string {
	affectedNodes, ok, err := unstructured.NestedStringSlice(cc.curExp.Object, "spec", "selector", "pods", extractNamespace(cc.curExp))
	if !ok {
		return err.Error()
	}

	err = cc.CheckPodsHealth(affectedNodes)
	if err != nil {
		return err.Error()
	}
	return ""
}

func (cc *ChaosController) CheckPodsHealth(pods []string) error {
	cur_namespace := extractNamespace(cc.curExp)
	for _, podName := range pods {
		err := CheckSingleNodeHealth(cc.pod.Client, cur_namespace, podName)
		if err != nil {
			return err
		}
	}

	return nil
}

// Check node health, params are suited for go tests
// Use k8s interface to pass both kubernetes.client and fake.client
func CheckSingleNodeHealth(clientset kubernetes.Interface, namespace, podName string) error {
	if clientset == nil {
		return fmt.Errorf("clientset is nil")
	}

	pod, err := clientset.CoreV1().Pods(namespace).Get(context.TODO(), podName, v1.GetOptions{})
	if err != nil {
		return err
	}

	for _, condition := range pod.Status.Conditions {
		if condition.Type == coreV1.PodConditionType(coreV1.NodeReady) {
			if condition.Status != coreV1.ConditionTrue {
				return fmt.Errorf("node %s is not ready. Status: %s, Reason: %s, Message: %s",
					podName, condition.Status, condition.Reason, condition.Message)
			}
		}
	}

	return nil
}

// Get all podes, affected by chaos experiment and reload them
// This is done in order to avoid CrashLoopBackOff pod status
func RestartAffectedPods(cc *ChaosController) error {
	affectedNodes, ok, err := unstructured.NestedStringSlice(cc.curExp.Object, "spec", "selector", "pods", extractNamespace(cc.curExp))
	if !ok {
		return err
	}

	// TODO implement case with pods from 2+ namespaces
	for _, podName := range affectedNodes {
		cc.pod.KillByName(context.Background(), cc.namespace, podName)
	}

	return nil
}

// Wait cluster to reconfigure after chaos experiment
func (cc *ChaosController) ExperimentSleep(durationStr string) error {
	experimentDuration, err := time.ParseDuration(durationStr)
	if err != nil {
		return err
	}

	time.Sleep(experimentDuration)

	return nil
}

// Despite namespace variable is stored in controller instance,
// chaos experiment might have different namespace in its deployment
func extractNamespace(experimentYaml *unstructured.Unstructured) string {
	namespace := experimentYaml.GetNamespace()
	if namespace == "" {
		namespace = "default"
	}

	return namespace
}

// getConfigPath fetches the path to the users kubeconfig
func getConfigPath() string {
	if configPath := os.Getenv("K6_CHAOS_KUBECONFIG"); configPath != "" {
		return configPath
	}

	return filepath.Join(homedir.HomeDir(), ".kube", "config")
}
