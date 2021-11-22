package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/go-redis/redis"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type Output struct {
	Name    string `json:"name"`
	Time    string `json:"time"`
	Reason  string `json:"reason"`
	Message string `json:"message"`
}

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

func GetEventsNSA(clientset *kubernetes.Clientset, client *redis.Client, ctx context.Context) {

	api := clientset.CoreV1()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	watcher, err := api.Events(v1.NamespaceAll).Watch(ctx, metav1.ListOptions{})
	if err != nil {
		panic(err)
	}
	for event := range watcher.ResultChan() {
		svc := event.Object.(*v1.Event)
		switch event.Type {
		case watch.Added:
			if svc.Reason == "ScalingReplicaSet" {
				bytes, err := json.Marshal(Output{
					Name:    svc.Name,
					Time:    svc.EventTime.Format("2 Jan 2006 15:04:05"),
					Reason:  svc.Reason,
					Message: svc.Message,
				})
				if err != nil {
					panic(err)
				}
				err = client.Set(string(svc.UID), bytes, 0).Err()
				if err != nil {
					fmt.Println(err)
				}
			}
		}
	}
}

func main() {
	//TODO: Jak na flagy? -> nefungují ani po kompilaci ani při go run main.go

	//REDIS CONNECTION
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	//Redis test connection
	pong, err := client.Ping().Result()
	fmt.Println(pong, err)

	var ns string
	flag.StringVar(&ns, "namespace", "", "namespace")
	flag.Parse()

	// Popis cesty k kube configu - HOME/.kube/config
	// Filepath join -> vytvoří cestu ze zadaných stringů
	ctx := context.Background()
	clientset := connK8s()
	GetEventsNSA(clientset, client, ctx)
}
