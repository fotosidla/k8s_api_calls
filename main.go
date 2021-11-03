package main

import (
	"context"
	"flag"
	"log"
	"os"
	"path/filepath"
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	//TODO: Jak na flagy? -> nefungují ani po kompilaci ani při go run main.go
	var ns, label, field string
	flag.StringVar(&ns, "namespace", "", "namespace")
	flag.StringVar(&label, "l", "", "Label selector")
	flag.StringVar(&field, "f", "", "Field selector")

	// Popis cesty k kube configu - HOME/.kube/config
	// Filepath join -> vytvoří cestu ze zadaných stringů
	kubeconfig := filepath.Join(
		os.Getenv("HOME"), ".kube", "config",
	)
	// Vytvoření klienta z configu ->  .kube/config
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		log.Fatal(err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}
	//api := clientset.CoreV1()

	//TODO: JAK FUNGUJE CONTEXT? Pokud není CTX jako argument funkce funkce neproběhne
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//TODO: Prostudovat více listOptions a zjistit jaké má možnosti filtrování
	/* 	EvtOptions := metav1.ListOptions{
		TypeMeta: metav1.TypeMeta{Kind: "Pod"},
	} */
	/* events, _ := api.Events(ns).List(ctx, EvtOptions)
	for _, item := range events.Items {
		fmt.Println(item.Name, "LAST SEEN - ", item.LastTimestamp, "MESSAGE - ", item.Message, "REASON - ", item.Reason)
	} */

	//Funkční watcher na events
	restClient := clientset.CoreV1().RESTClient()
	lw := cache.NewListWatchFromClient(restClient, "events", v1.NamespaceAll, fields.Everything())
	_, controller := cache.NewInformer(lw,
		&v1.Event{},
		time.Millisecond*1,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				event, ok := obj.(*v1.Event)
				if !ok {
					log.Fatalf("list/watch returned non-event object: %T", event)
				}
				time.Sleep(time.Millisecond * 1)
				log.Printf("Name:", event.Name, "What happend %s", event.Message, "KIND:", event.Kind)
			},
		},
	)
	controller.Run(ctx.Done())

	//TODO: : Proč je tu ctx a k čemu slouží
	//pods, err := api.Pods("default").List(ctx, listOptions)
	//for _, PodList := range (*pods).Items {
	//fmt.Printf("pods-name=%v\n", PodList.Name)

	//}

}
