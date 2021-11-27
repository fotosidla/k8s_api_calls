package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"log"
	"os"
	"path/filepath"
)

type app struct {
	client *kubernetes.Clientset
	redis *redis.Client
}


func main() {
	ctx := context.Background()
	events, err := connK8s().CoreV1().Events(v1.NamespaceAll).Watch(ctx, metav1.ListOptions{})


	app := &app{
		client: connK8s(),
		redis: redisCon(),
	}
	

	if err != nil {
		panic(err)
	}
	resultChan := events.ResultChan()
	app.processEvents(resultChan)


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
	pods, err := connK8s().CoreV1().Pods(v1.NamespaceAll).List(context.TODO(), metav1.ListOptions{})
	test := asdf.getOwner(pods)
	//fmt.Println("Stored mall event.")
	bytes, err := json.Marshal(MallEvent{
		Name: event.Name,
		Time: event.Time,
		Reason: event.Reason,
		Message: event.Message,
		Owner: test.Owner,
		Kind: test.Kind,
	})
	println(string(bytes),err)
	// TODO store logic
}

type MallEvent struct {
	Owner string `json:"owner"`
	Kind  string `json:"kind"`
	Name  string `json:"name"`
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
				Owner: replica.OwnerReferences[0].Name,
				Kind:  "Deployment",
			}
		case "DaemonSet", "StatefulSet":
			return &MallEvent{
				Owner: pod.OwnerReferences[0].Name,
				Kind:  pod.OwnerReferences[0].Kind,
			}
		default:
			continue
		}
	}
	return nil
	}


func (c *app) mapEvent(event *v1.Event) *MallEvent {

	//fmt.Println("Mapping mall event.")

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

func redisCon() *redis.Client{
	redisset := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
	return redisset
}
