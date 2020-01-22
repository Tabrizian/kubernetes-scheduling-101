package scheduler

import (
	"k8s.io/client-go/kubernetes"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"github.com/pkg/errors"
)

type Scheduler struct {
	Name string // Scheduler Name
	client *kubernetes.Clientset // k8s Client
	watcher watch.Interface // Watcher

}

func NewScheduler(name string, client *kubernetes.Clientset) (*Scheduler) {
	scheduler := &Scheduler{
		Name: name,
		client: client,
	}

	return scheduler
}

func (s *Scheduler) Register() error  {
	watcher, err := s.client.CoreV1().Pods("").Watch(metav1.ListOptions{})
	if err != nil {
		return errors.Wrap(err, "Failed to create a watch on Pods")
	}
	s.watcher = watcher

	return err
}