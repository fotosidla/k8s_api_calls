package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func printPvCs(pvcs *v1.PersistentVolumeClaimList) {
	template := "%-32s%-8s%-8s\n"
	fmt.Printf(template, "NAME", "STATUS", "CAPACITY")
	for _, pvc := range pvcs.Items {
		quant := pvc.Spec.Resources.Requests[v1.ResourceStorage]
		fmt.Printf(template, pvc.Name, string(pvc.Status.Phase), quant.String())
	}
}

func main() {
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
	api := clientset.CoreV1()

	TODO: JAK FUNGUJE CONTEXT? Pokud není CTX jako argument funkce funkce neproběhne
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	listOptions := metav1.ListOptions{
		LabelSelector: label,
	}
	//pvcs, err := api.PersistentVolumeClaims(ns).List(ctx, listOptions)
	TODO: : Proč je tu ctx a k čemu slouží
	pods, err := api.Pods("default").List(ctx, listOptions)
	for _, PodList := range (*pods).Items {
		fmt.Printf("pods-name=%v\n", PodList.Name)
	}
	//printPvCs(pvcs)

}
