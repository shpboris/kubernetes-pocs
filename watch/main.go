package main

import (
	"context"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"
	"os"
)

const (
	crdPlural    = "mapdata"
	crdGroup     = "infra.shpboris"
	crdVersion   = "v1"
	crdNamespace = "default"
)

func main() {
	ctx := context.Background()
	kubeconfigPath := os.Getenv("KUBECONFIG")
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		logrus.Fatal(err)
	}
	client, err := dynamic.NewForConfig(config)
	if err != nil {
		logrus.Fatal(err)
	}
	crdResource := schema.GroupVersionResource{
		Group:    crdGroup,
		Version:  crdVersion,
		Resource: crdPlural,
	}
	watcher, err := client.Resource(crdResource).Namespace(crdNamespace).Watch(ctx, v1.ListOptions{
		Watch: true,
	})
	if err != nil {
		logrus.Fatal(err)
	}
	ch := watcher.ResultChan()
	for event := range ch {
		if event.Type == watch.Added || event.Type == watch.Modified || event.Type == watch.Deleted {
			logrus.Info("Received new CRD event of type: " + event.Type)
			obj := event.Object.(*unstructured.Unstructured)
			logrus.Info("Object name is: " + obj.GetName())
			objMap := obj.UnstructuredContent()
			specMap := objMap["spec"].(map[string]interface{})
			logrus.Info("Map name is: " + specMap["mapname"].(string))
			logrus.Info("Key1 is: " + specMap["key1"].(string))
			logrus.Info("Key2 is: " + specMap["key2"].(string))
		}
	}
	watcher.Stop()
}
