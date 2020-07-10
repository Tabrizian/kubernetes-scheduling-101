package scheduler

import (
	"context"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	coreV1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"math/rand"
	"time"
)

// Scheduler struct containing main attributes of scheduling
type Scheduler struct {
	Name    string                // Scheduler Name
	client  *kubernetes.Clientset // k8s Client
	watcher watch.Interface       // Watcher
	queue	[]*coreV1.Pod
	nodes	*coreV1.NodeList
}

// NewScheduler creates a new scheduler 
func NewScheduler(name string, client *kubernetes.Clientset) *Scheduler {
	scheduler := &Scheduler{
		Name:   name,
		client: client,
	}

	go scheduler.Schedule()
	return scheduler
}

// Schedule is the background coroutine
func (s *Scheduler) Schedule() {
	for {
		log.Info("Queue Length is ", len(s.queue))
		s.Random()
		time.Sleep(2 * time.Second)
	}
}

// Register is the function
func (s *Scheduler) Register() error {
	watcher, err := s.client.CoreV1().Pods("").Watch(context.TODO(), metav1.ListOptions{
		FieldSelector: "spec.schedulerName=" + s.Name + ",status.phase=Pending",
	})
	log.Info("Watching for uncscheduled pods")
	if err != nil {
		return errors.Wrap(err, "Failed to create a watch on Pods")
	}
	s.watcher = watcher
	log.Info("Watching for new Pods to be scheduled for ", s.Name)
	nodes, err := s.client.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return errors.Wrap(err, "Failed to list the nodes")
	}
	s.nodes = nodes
	ch := s.watcher.ResultChan()
	for event := range ch {
		if event.Type != "ADDED" {
			continue
		}
		pod, ok := event.Object.(*coreV1.Pod)
		if !ok {
			panic(err.Error())
		}
		log.Info("Pod added to the queue ", pod.Name)
		s.queue = append(s.queue, pod)
	}

	return err
}

// Random randomly schedules the first item on the queue
// on the nodes
func (s *Scheduler) Random() error {
	if s.nodes == nil || len(s.nodes.Items) == 0 {
		log.Info("There is no node to schedule on!")
		return nil
	}

	if len(s.queue) == 0 {
		log.Info("There is no Pod to be scheduled.")
		return nil
	}

	nodeItems := s.nodes.Items

	pod := s.queue[0]
	s.queue = s.queue[1:]
	randomNumber := rand.Intn(len(nodeItems))
	randomNode := nodeItems[randomNumber]
	_, err := s.client.CoreV1().Pods(pod.Namespace).Update(context.TODO(), pod, metav1.UpdateOptions{})
	s.client.CoreV1().Pods(pod.Namespace).Bind(context.TODO(), &coreV1.Binding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      pod.Name,
			Namespace: pod.Namespace,
		},
		Target: coreV1.ObjectReference{
			APIVersion: "v1",
			Kind:       "Node",
			Name:       randomNode.GetName(),
		},
	}, metav1.CreateOptions{})

	timestamp := time.Now().UTC()
	s.client.CoreV1().Events(pod.Namespace).Create(context.TODO(), &coreV1.Event{
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
	}, metav1.CreateOptions{})

	log.Info("Scheduled Pod ", pod.GetName(), " on node ", randomNode.GetName())
	if err != nil {
		log.Error("Failed to schedule pod", err.Error())
		return errors.Wrap(err, "Failed to list the nodes")
	}

	return nil
}
