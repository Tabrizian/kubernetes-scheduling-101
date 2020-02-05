/*
Copyright 2016 The Kubernetes Authors.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Note: the example only works with the code within the same release/branch.
package main

import (
	"flag"

	k8sConfig "github.com/Tabrizian/kubernetes-scheduling-101/k8s"
	"github.com/Tabrizian/kubernetes-scheduling-101/scheduler"
	log "github.com/sirupsen/logrus"
	//
	// Uncomment to load all auth plugins
	// _ "k8s.io/client-go/plugin/pkg/client/auth"
	//
	// Or uncomment to load specific auth plugins
	// _ "k8s.io/client-go/plugin/pkg/client/auth/azure"
	// _ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	// _ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
	// _ "k8s.io/client-go/plugin/pkg/client/auth/openstack"
)

func main() {
	var kubeconfig *string

	kubeconfig = flag.String("kubeconfig", "", "(optional) absolute path to the kubeconfig file")
	flag.Parse()
	log.Info("Kubeconfig flag is loaded")
	clientset, err := k8sConfig.GetKubeConfig(kubeconfig)
	if err != nil {
		log.Error("Failed to create clientset")
		panic(err.Error())
	}
	imanScheduler := scheduler.NewScheduler("iman-scheduler", clientset)
	log.Info("Registering new scheduler")
	imanScheduler.Register()
}
