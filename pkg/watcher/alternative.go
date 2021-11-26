package watcher

import (
	"context"
	"fmt"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"log"
	"os"
	"path/filepath"
)

func Main() {
	ctx := context.Background()
	events, err := connK8s().CoreV1().Events(v1.NamespaceAll).Watch(ctx, metav1.ListOptions{})
	if err != nil {
		panic(err)
	}
	resultChan := events.ResultChan()
	processEvents(resultChan)
}

func processEvents(events <-chan watch.Event) {
	for event := range events {
		mallEvent := mapEvent(event.Object.(*v1.Event))
		if mallEvent == nil {
			fmt.Println("Skipped processing of event.")
			continue
		}
		storeEvent(mallEvent)
	}
}

func storeEvent(event *MallEvent) {
	fmt.Println("Stored mall event.")
	// TODO store logic
}

type MallEvent struct {
	deployment  string
	otherFields string
	// TODO some other fields
}

func mapEvent(event *v1.Event) *MallEvent {
	fmt.Println("Mapping mall event.")
	// TODO example just for illustration of limited mapping support
	if event.Reason != "Scheduled" {
		return nil
	}
	// TODO some mapping logic
	return &MallEvent{
		deployment:  "TODO somehow",
		otherFields: "TODO something else",
	}
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
