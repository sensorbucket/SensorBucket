package main

import (
	"context"
	"fmt"

	fission "github.com/fission/fission/pkg/crd"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	if err := Run(); err != nil {
		panic(err)
	}
}

func Run() error {
	config, err := clientcmd.BuildConfigFromFlags("", "/home/timvosch/.kube/config")
	if err != nil {
		return err
	}

	cg := fission.NewClientGeneratorWithRestConfig(config)
	fClient, err := cg.GetFissionClient()
	if err != nil {
		return err
	}

	fls, err := fClient.CoreV1().Functions("").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return err
	}

	fmt.Printf("fls: %v\n", fls)

	return nil
}
