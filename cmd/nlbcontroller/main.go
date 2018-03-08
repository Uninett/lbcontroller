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

	"log"
	"os"
	"time"

	"github.com/folago/nlb"
	"github.com/pkg/errors"

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

var (
	// max retries for a queued object
	maxRetries = 3
	lbendpoint string
)

func main() {
	flag.Parse()

	//check is we have a load balancer endpoint
	lbendpoint = os.Getenv("NLB_ENDPOINT")
	if len(lbendpoint) == 0 {

		log.Fatalln("No load balancer endpoint defined, NLB_ENDPOINT not set in environment")
	}
	log.Printf("Load balancer endpoint: %s\n", lbendpoint)

	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	// creates the clientset
	// creates the clientset
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
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

	log.Println("Starting l4lb controller")

	go c.informer.Run(stopCh)

	// wait for the caches to synchronize before starting the worker
	if !cache.WaitForCacheSync(stopCh, c.HasSynced) {
		utilruntime.HandleError(fmt.Errorf("Timed out waiting for caches to sync"))
		return
	}

	log.Println("l4lb controller synced and ready")

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
		log.Printf("Error processing %s (will retry): %v", key, err)
		// requeue the item to work on later
		c.queue.AddRateLimited(key)
	} else {
		// err != nil and too many retries
		log.Printf("Error processing %s (giving up): %v", key, err)
		c.queue.Forget(key)
		utilruntime.HandleError(err)
	}

	return true
}

func (c *Controller) processItem(key string) error {
	log.Printf("Processing change to Service %s", key)

	obj, exists, err := c.informer.GetIndexer().GetByKey(key)
	if err != nil {
		return errors.Errorf("Error fetching object with key %s from store: %v", key, err)
	}

	if !exists {
		//fmt.Printf("Deleted Service %s\n", obj.(*v1.Service).GetName())
		//obj is nil
		log.Printf("Deleted Service with key %s\n", key)
		return nil
	}
	svc := obj.(*v1.Service)

	sname := svc.GetName()
	log.Printf("Sync/Add/Update for Service %s\n", sname)

	if svc.Spec.Type == v1.ServiceTypeLoadBalancer {
		log.Printf("service %s of type %s\n", sname, svc.Spec.Type)
		lbSvc, found, err := nlb.GetService(sname, lbendpoint+"/services")
		if err != nil {
			log.Println("ERROR: ", err)
			return errors.Wrap(err, "Error getting list of load balancer services")
		}
		fmt.Printf("Service received: %v\n", lbSvc)

		//create the service, get the frontend, populate the service.status.loadBalancer.ingress
		/*
			get the lbservice, the name should be namespace.service
			if not present create a new lbservice.
				first chech if the k8service has a loadBalancerIP set, if yes
					first create a frontend with that IP and referencr it
					in the lbservice json object
				create a new lbservice
		*/
		if !found {
			lbSvc.Type = nlb.TCP
			lbSvc.Metadata.Name = sname
			lbSvc.Config = nlb.TCPConfig{
				Method: "least_conn",
				//Ports: svc.Spec.Ports
			}
			meta, err := nlb.NewService(*lbSvc, lbendpoint+"/services")
			if err != nil {
				//log.Println(err)
				return errors.Wrap(err, "Error creating a new load balancer service")
			}
			fmt.Println("created load balancer", meta)
		} else { //update here?
			/*
			 */
		}

		fmt.Println("service status: ", svc.Status)
	}
	return nil
}
