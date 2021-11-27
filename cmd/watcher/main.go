package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type app struct {
	client *kubernetes.Clientset
	redis  *kubernetes.Clientset
	demo   demo
}

type demo struct {
	client *kubernetes.Clientset
}

func main() {
	ctx := context.Background()
	events, err := connK8s().CoreV1().Events(v1.NamespaceAll).Watch(ctx, metav1.ListOptions{})
	app := &app{
		client: connK8s(),
	}
	if err != nil {
		panic(err)
	}
	resultChan := events.ResultChan()
	app.processEvents(resultChan)
	pods, err := connK8s().CoreV1().Pods("").List(ctx, metav1.ListOptions{})
	app.getOwner(pods)
	println(pods.Items)

}

func (c *app) processEvents(events <-chan watch.Event) {
	for event := range events {
		mallEvent := c.mapEvent(event.Object.(*v1.Event))

		if mallEvent == nil {
			fmt.Println("Skipped processing of event.")
			continue
		}

		c.storeEvent(mallEvent)

	}
}

func (asdf *app) storeEvent(event *MallEvent) {
	fmt.Println("Stored mall event.")

	// TODO store logic
}

type MallEvent struct {
	owner   string `json:"owner"`
	kind    string `json:"kind"`
	Name    string `json:"name"`
	Time    string `json:"time"`
	Reason  string `json:"reason"`
	Message string `json:"message"`
	// TODO some other fields
}

func (c *app) getOwner(pods *v1.PodList) *MallEvent {

	cAPP := c.client.AppsV1()
	for _, pod := range pods.Items {
		if len(pod.OwnerReferences) == 0 {
			return nil

		}

		switch pod.OwnerReferences[0].Kind {
		case "ReplicaSet":
			replica, replicaERR := cAPP.ReplicaSets(pod.Namespace).Get(context.TODO(), pod.OwnerReferences[0].Name, metav1.GetOptions{})
			if replicaERR != nil {
				panic(replicaERR.Error())
			}
			return &MallEvent{
				owner: replica.OwnerReferences[0].Name,
				kind:  "Deployment",
			}
		case "DaemonSet", "StatefulSet":
			return &MallEvent{
				owner: pod.OwnerReferences[0].Name,
				kind:  pod.OwnerReferences[0].Kind,
			}
		default:
			continue
		}
	}
	return nil
	}


func (c *app) mapEvent(event *v1.Event) *MallEvent {

	fmt.Println("Mapping mall event.")

	// TODO example just for illustration of limited mapping support
	if event.Reason != "Scheduled" {
		return nil
	}

	return &MallEvent{
		Name:    event.Name,
		Time:    event.EventTime.Format("2 Jan 2006 15:04:05"),
		Reason:  event.Reason,
		Message: event.Message,
	}

	// TODO some mapping logic
}

// TODO extract into shared unit FIXME copypaste
func connK8s() *kubernetes.Clientset {
	kubeconfig := filepath.Join(
		os.Getenv("HOME"), ".kube", "config",
	)
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		log.Fatal(err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}
	return clientset
}
