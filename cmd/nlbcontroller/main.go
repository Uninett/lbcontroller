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

// https://engineering.bitnami.com/articles/kubewatch-an-example-of-kubernetes-custom-controller.html
package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/golang/glog"
	"k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

// max retries for a queued object
var maxRetries = 3

func main() {
	flag.Parse()
	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	// creates the clientset
	// creates the clientset
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		glog.Fatal(err)
	}

	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())

	informer := cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options meta_v1.ListOptions) (runtime.Object, error) {
				return client.CoreV1().Services(meta_v1.NamespaceAll).List(options)
			},
			WatchFunc: func(options meta_v1.ListOptions) (watch.Interface, error) {
				return client.CoreV1().Services(meta_v1.NamespaceAll).Watch(options)
			},
		},
		&v1.Service{},
		0, //Skip resync
		cache.Indexers{},
	)

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(obj)
			if err == nil {
				queue.Add(key)
			}
		},
		UpdateFunc: func(old, new interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(new)
			if err == nil {
				queue.Add(key)
			}
		},
		DeleteFunc: func(obj interface{}) {
			key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
			if err == nil {
				queue.Add(key)
			}
		},
	})

	controller := &Controller{
		clientset: client,
		informer:  informer,
		queue:     queue,
	}

	stop := make(chan struct{})
	defer close(stop)
	go controller.Run(stop)

	// Wait forever
	select {}

}

// Controller object
type Controller struct {
	clientset kubernetes.Interface
	queue     workqueue.RateLimitingInterface
	informer  cache.SharedIndexInformer
	//eventHandler handlers.Handler
}

// Run will start the controller.
// StopCh channel is used to send interrupt signal to stop it.
func (c *Controller) Run(stopCh <-chan struct{}) {
	// don't let panics crash the process
	defer utilruntime.HandleCrash()
	// make sure the work queue is shutdown which will trigger workers to end
	defer c.queue.ShutDown()

	glog.Info("Starting l4lb controller")

	go c.informer.Run(stopCh)

	// wait for the caches to synchronize before starting the worker
	if !cache.WaitForCacheSync(stopCh, c.HasSynced) {
		utilruntime.HandleError(fmt.Errorf("Timed out waiting for caches to sync"))
		return
	}

	glog.Info("l4lb controller synced and ready")

	// runWorker will loop until "something bad" happens.  The .Until will
	// then rekick the worker after one second
	wait.Until(c.runWorker, time.Second, stopCh)
}

// HasSynced is required for the cache.Controller interface.
func (c *Controller) HasSynced() bool {
	return c.informer.HasSynced()
}
func (c *Controller) runWorker() {
	// processNextWorkItem will automatically wait until there's work available
	for c.processNextItem() {
		// continue looping
	}
}

// processNextWorkItem deals with one key off the queue.  It returns false
// when it's time to quit.
func (c *Controller) processNextItem() bool {
	// pull the next work item from queue.  It should be a key we use to lookup
	// something in a cache
	key, quit := c.queue.Get()
	if quit {
		return false
	}

	// you always have to indicate to the queue that you've completed a piece of
	// work
	defer c.queue.Done(key)

	// do your work on the key.
	err := c.processItem(key.(string))

	if err == nil {
		// No error, tell the queue to stop tracking history
		c.queue.Forget(key)
	} else if c.queue.NumRequeues(key) < maxRetries {
		glog.Errorf("Error processing %s (will retry): %v", key, err)
		// requeue the item to work on later
		c.queue.AddRateLimited(key)
	} else {
		// err != nil and too many retries
		glog.Errorf("Error processing %s (giving up): %v", key, err)
		c.queue.Forget(key)
		utilruntime.HandleError(err)
	}

	return true
}

func (c *Controller) processItem(key string) error {
	glog.Infof("Processing change to Service %s", key)

	obj, exists, err := c.informer.GetIndexer().GetByKey(key)
	if err != nil {
		return fmt.Errorf("Error fetching object with key %s from store: %v", key, err)
	}

	if !exists {
		//fmt.Printf("Deleted Service %s\n", obj.(*v1.Service).GetName())
		//obj is nil
		glog.Infof("Deleted Service with key %s\n", key)
		return nil
	}

	svc := obj.(*v1.Service)
	glog.Infof("Sync/Add/Update for Service %s\n", svc.GetName())
	fmt.Println("service type: ", svc.Spec.Type)
	fmt.Println("service status: ", svc.Status)

	return nil
}
