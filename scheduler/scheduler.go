package scheduler

import (
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	coreV1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"math/rand"
	"time"
)

type Scheduler struct {
	Name    string                // Scheduler Name
	client  *kubernetes.Clientset // k8s Client
	watcher watch.Interface       // Watcher

}

func NewScheduler(name string, client *kubernetes.Clientset) *Scheduler {
	scheduler := &Scheduler{
		Name:   name,
		client: client,
	}

	return scheduler
}

func (s *Scheduler) Register() error {
	watcher, err := s.client.CoreV1().Pods("").Watch(metav1.ListOptions{
		FieldSelector: "spec.schedulerName=" + s.Name + ",status.phase=Pending",
	})
	log.Info("Watching for uncscheduled pods")
	if err != nil {
		return errors.Wrap(err, "Failed to create a watch on Pods")
	}
	s.watcher = watcher
	log.Info("Watching for new Pods to be scheduled for", s.Name)
	nodes, err := s.client.CoreV1().Nodes().List(metav1.ListOptions{})
	if err != nil {
		return errors.Wrap(err, "Failed to list the nodes")
	}
	ch := s.watcher.ResultChan()
	for event := range ch {
		if event.Type != "ADDED" {
			continue
		}
		pod, ok := event.Object.(*coreV1.Pod)
		if !ok {
			panic(err.Error())
		}
		log.Info("Pod is ", pod.Name)
		s.Random(pod, nodes)
	}

	return err
}

func (s *Scheduler) Random(pod *coreV1.Pod, nodes *coreV1.NodeList) (string, error) {
	nodeItems := nodes.Items
	randomNumber := rand.Intn(len(nodeItems))
	randomNode := nodeItems[randomNumber]
	_, err := s.client.CoreV1().Pods(pod.Namespace).Update(pod)
	s.client.CoreV1().Pods(pod.Namespace).Bind(&coreV1.Binding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      pod.Name,
			Namespace: pod.Namespace,
		},
		Target: coreV1.ObjectReference{
			APIVersion: "v1",
			Kind:       "Node",
			Name:       randomNode.GetName(),
		},
	})
	timestamp := time.Now().UTC()
	s.client.CoreV1().Events(pod.Namespace).Create(&coreV1.Event{
		Count:          1,
		Message:        "Pod Scheduled",
		Reason:         "Scheduled",
		LastTimestamp:  metav1.NewTime(timestamp),
		FirstTimestamp: metav1.NewTime(timestamp),
		Type:           "Normal",
		Source: coreV1.EventSource{
			Component: s.Name,
		},
		InvolvedObject: coreV1.ObjectReference{
			Kind:      "Pod",
			Name:      pod.Name,
			Namespace: pod.Namespace,
			UID:       pod.UID,
		},
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: pod.Name + "-",
		},
	})
	log.Info("Scheduled Pod ", pod.GetName(), " on node ", randomNode.GetName())
	if err != nil {
		log.Error("Failed to schedule pod", err.Error())
		return "", errors.Wrap(err, "Failed to list the nodes")
	}
	return nodes.Items[randomNumber].GetName(), nil
}
