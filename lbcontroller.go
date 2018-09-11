package main

import (
	"bytes"
	"net/http"
	"time"

	"github.com/koki/json"
	"github.com/pkg/errors"
)

const (
	jsonContent = "application/json"
)

var (
	servicePath  = "services"
	frontendPath = "frontends"
)

//Metadata of messages sent to the API
//"metadata": {
//	"name": "testservice",
//	"created_at": "......",
//	"updated_at": "......",
//	...
//}
type Metadata struct {
	Name      string    `json:"name,omitempty"`
	CreatedAt time.Time `json:"created_at,omitempty"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
}

// Config represent the configuration of a TCP load balanced service, e.g.:
// "config": {
// 	"method": "least_conn",
// 	"ports": [80, 443],
// 	"backends": [
// 		 "hostname1.example.com": {
// 			 "addrs": ["10.3.2.43", "2001:700:f00d::8"]
// 		 },
// 		 "hostname2.example.com": {
// 			 "addrs": ["10.3.2.53", "2001:700:f00d::18"]
// 		 }
// 	],
// 	"upstream_max_conns": 100,
// 	"acl": ["10.10.20.0/24", "2001:700:1337::/48"],
// 	"health_check": {
// 		 "port": 1337,
// 		 "send": "healthz\n",
// 		 "expect": "^OK$"
// },
// 	"frontend": "foobar"
// }
type Config struct {
	Method           string           `json:"method,omitempty"`
	Ports            map[string]int32 `json:"ports,omitempty"`
	Backends         []Backend        `json:"backends,omitempty"`
	UpstreamMaxConns int              `json:"upstream_max_conns,omitempty"`
	ACL              []string         `json:"acl,omitempty"`
	HealthCheck      HealthCheck      `json:"health_check,omitempty"`
	Frontend         string           `json:"frontend,omitempty"`
}

// Backend represents a backend in the loadbalancer configuration
type Backend struct {
	Host  string   `json:"host,omitempty"`
	Addrs []string `json:"addrs,omitempty"`
}

// HealthCheck is a loadbalancer heath check for TCP services
type HealthCheck struct {
	Port   int32  `json:"port,omitempty"`
	Send   string `json:"send,omitempty"`
	Expect string `json:"expect,omitempty"`
}

//prepare the http request and marchal the object to send
func prepareRequest(obj Service, url, method string) (*http.Request, error) {
	data, err := json.Marshal(obj)
	if err != nil {
		return nil, errors.Wrap(err, "error marshalling Frontend")
	}
	buf := bytes.NewBuffer(data)

	req, err := http.NewRequest(method, url, buf)
	if err != nil {
		return nil, errors.Wrap(err, "error creatign http.Request")
	}
	req.Header.Set("Content-Type", jsonContent)
	return req, nil
}

// ServiceType represent the type of service offered by the load balancers.
type ServiceType string

// Known types of service
const (
	TCP              ServiceType = "tcp"
	UDP              ServiceType = "udp"
	TCPProxyProtocol ServiceType = "tcp_proxy_protocol"
)
