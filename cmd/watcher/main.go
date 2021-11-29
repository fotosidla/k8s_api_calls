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
	app.processEvents(resultChan,ctx)


}

func (c *app) processEvents(events <-chan watch.Event,ctx context.Context) {

	for event := range events {
		mallEvent := c.mapEvent(event.Object.(*v1.Event),ctx)

		if mallEvent == nil {
			fmt.Println("Skipped processing of event.")
			continue
		}

		c.storeEvent(mallEvent)



	}
}

func (asdf *app) storeEvent(event *MallEvent) {





	//fmt.Println("Stored mall event.")
	bytes, err := json.Marshal(MallEvent{

		Name: event.Name,
		Time: event.Time,
		Reason: event.Reason,
		Message: event.Message,

	})
	println(string(bytes),err)
	// TODO store logic
	//TODO redis store -> repSetOwn.Owner need to be gathered exactly for one event -> logic implementation

	if err != nil {
		fmt.Println(err)
	}
}



type MallEvent struct {
	Owner string `json:"Owner"`
	Kind string `json:"Kind"`
	Name  string `json:"name"`
	Time    string `json:"time"`
	Reason  string `json:"reason"`
	Message string `json:"message"`
	// TODO some other fields
}



func (c *app) mapEvent(event *v1.Event, ctx context.Context) *MallEvent {

	//fmt.Println("Mapping mall event.")

	// TODO example just for illustration of limited mapping support
	if event.Reason != "Scheduled" {
		return nil
	}
	pod, err := c.client.CoreV1().Pods(event.InvolvedObject.Namespace).Get(ctx, event.InvolvedObject.Name, metav1.GetOptions{})
	test := err.Error()
	if test == "NotFound" {
		println("POD NOT FOUND")
	}
	if  len(pod.ObjectMeta.OwnerReferences) > 1 {
		panic("MORE REFERENCES THAN ONE")
	}
	if  len(pod.ObjectMeta.OwnerReferences) == 0 {
		return nil
	}

	podRef := pod.ObjectMeta.OwnerReferences[0]
	if podRef.Kind != "ReplicaSet"{
		return nil
	}
	repSet, err := c.client.AppsV1().ReplicaSets(event.InvolvedObject.Namespace).Get(ctx, podRef.Name, metav1.GetOptions{})
	if err != nil {
		panic(err)
	}
	if repSet.ObjectMeta.OwnerReferences[0].Kind != "Deployment" {
		panic("Owner reference is not DEPLOYMENT")
	}
	fmt.Printf("POD NAME %s POD OWNER %s\n" ,pod.Name,repSet.ObjectMeta.OwnerReferences[0].Name)



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
