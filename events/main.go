package main

import (
	"context"
	guuid "github.com/google/uuid"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"os"
)

const (
	crdNamespace = "default" // Namespace where the CRD instances are created
)

func main() {
	ctx := context.Background()
	kubeconfigPath := os.Getenv("KUBECONFIG")
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		logrus.Fatal(err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		logrus.Fatal(err)
	}
	newEvent := &corev1.Event{
		ObjectMeta: v1.ObjectMeta{
			Name:      "map-data-creation-event-" + (guuid.New()).String(),
			Namespace: crdNamespace,
		},
		InvolvedObject: corev1.ObjectReference{
			Kind:      "MapData",
			Namespace: crdNamespace,
			Name:      "mapdata-sample1",
		},
		Reason:  "CreationNotification",
		Message: "The MapData instance was created",
		Type:    corev1.EventTypeWarning,
	}
	newEvent, err = clientset.CoreV1().Events(crdNamespace).Create(ctx, newEvent, v1.CreateOptions{})
	if err != nil {
		logrus.Fatal(err)
	}
	opts := v1.ListOptions{
		FieldSelector: "involvedObject.kind=MapData",
		Watch:         true,
	}
	watcher, err := clientset.CoreV1().Events(crdNamespace).Watch(ctx, opts)
	if err != nil {
		logrus.Fatal(err)
	}
	ch := watcher.ResultChan()
	for event := range ch {
		eventObj, ok := event.Object.(*corev1.Event)
		if ok {
			logrus.Info("Received event named: " + eventObj.Name + " of type: " + eventObj.Type)
		}
	}
	watcher.Stop()
}
