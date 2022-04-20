package client

import (
	"bytes"
	"context"
	"io"
	"log"
	"strings"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/rest"
)

// k8s interface
type k8s interface {
	lister
	getter
}

// lister interface
type lister interface {
	ListPods(namespace, selector string) (*corev1.PodList, error)
}

// getter interface
type getter interface {
	GetPodLogs(pod corev1.Pod, podLogOptions corev1.PodLogOptions) ([]string, error)
}

var KubeClient k8s = getKubeClientset()

// k8s client struct
type client struct {
	*kubernetes.Clientset
}

// list pods method
func (c *client) ListPods(namespace, selector string) (*corev1.PodList, error) {
	pods, err := c.CoreV1().Pods(namespace).List(context.TODO(), v1.ListOptions{
		LabelSelector: selector,
	})
	if err != nil {
		return nil, err
	}

	return pods, nil
}

// get pods method
func (c *client) GetPodLogs(pod corev1.Pod, podLogOptions corev1.PodLogOptions) ([]string, error) {
	req := c.CoreV1().Pods(pod.Namespace).GetLogs(pod.Name, &podLogOptions)
	podLogs, err := req.Stream(context.TODO())
	if err != nil {
		return nil, err
	}

	defer podLogs.Close()

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, podLogs)
	if err != nil {
		return nil, err
	}

	return strings.Split(buf.String(), "\n"), nil
}

func getKubeClientset() *client {
	// creates the in-cluster config
	conf, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}

	cs, err := kubernetes.NewForConfig(conf)
	if err != nil {
		log.Printf("error in getting clientset from Kubeconfig: %v", err)
	}

	return &client{cs}
}