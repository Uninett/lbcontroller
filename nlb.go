package nlb

import (
	"bytes"
	"net/http"
	"time"

	"github.com/koki/json"
	"github.com/pkg/errors"
)

const jsonContent = "application/json"

//Message from/to the API endpoint, e.g.
//{
//	"type": "frontend"
//	"metadata": ...
//	"config": {
//		 "addresses": ["10.40.50.23","2001:700:fffd::23"]
//	}
//}
//different types comes with different configurations.
//The Config field should be unmarshalled after a Message is
//in this way the type is known
type Message struct {
	Type     ServiceType     `json:"type,omitempty"`
	Metadata Metadata        `json:"metadata,omitempty"`
	Config   json.RawMessage `json:"config,omitempty"`
}

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

//FrontendConfig is the configuration of a Frontend object
//{
//	"type": "frontend"
//	"metadata": ...
//	"config": {
//		 "addresses": ["10.40.50.23","2001:700:fffd::23"]
//	}
//}
type FrontendConfig struct {
	Addresses []string `json:"addresses,omitempty"`
}

//Frontend type for the load balancers
type Frontend struct {
	Type     string         `json:"type,omitempty"`
	Metadata Metadata       `json:"metadata,omitempty"`
	Config   FrontendConfig `json:"config,omitempty"`
}

// TCPConfig represent the configuration of a TCP load balanced service, e.g.:
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
type TCPConfig struct {
	Method           string         `json:"method,omitempty"`
	Ports            []uint16       `json:"ports,omitempty"`
	Backends         []Backend      `json:"backends,omitempty"`
	UpstreamMaxConns int            `json:"upstream_max_conns,omitempty"`
	ACL              []string       `json:"acl,omitempty"`
	HealthCheck      TCPHealthCheck `json:"health_check,omitempty"`
	Frontend         string         `json:"frontend,omitempty"`
}

//Type imprments the ServiceConfig interface
func (tcpp TCPConfig) Type() ServiceType {
	return TCP
}

// Backend represents a backend in the loadbalancer configuration
type Backend struct {
	Host  string   `json:"host,omitempty"`
	Addrs []string `json:"addrs,omitempty"`
}

// TCPHealthCheck is a loadbalancer heath check for TCP services
type TCPHealthCheck struct {
	Port   uint16 `json:"port,omitempty"`
	Send   string `json:"send,omitempty"`
	Expect string `json:"expect,omitempty"`
}

// HTTPHealthCheck is a loadbalancer heath check for http services
type HTTPHealthCheck struct {
	URI        string `json:"uri,omitempty"`
	StatusCode int    `json:"status_code,omitempty"`
	Body       string `json:"body,omitempty"`
}

//SharedHTTPConfig represents the configuration of a TCP load balanced service, e.g.:
//"config": {
//	"names": ["site-a.example.com", "site-b.foo.org"],
//	"sticky_backends": false,
//	"backend_protocols": "both",
//	"http": {
//		"redirect_https": true,
//		"backend_port": 8080,
//		"health_check": {
//			"uri": "/",
//			 "status_code": 301
//		}
//	},
//	"https": {
//		 "private_key": "........",
//		 "certificate": "........",
//		 "backend_port": 8888,
//		 "health_check": {
//			 "uri": "/healthz",
//			 "status_code": 200,
//			 "body": "OK"
//			}
//	},
//	"backends": [
//		 "hostname1.example.com": {
//			 "addrs": ["10.3.2.1", "2001:700:f00d::4"]
//		 }
//	]
//}
type SharedHTTPConfig struct {
	Names            []string  `json:"names,omitempty"`
	StickyBackends   bool      `json:"sticky_backends"`
	BackendProtocols string    `json:"backend_protocols,omitempty"`
	HTTP             HTTP      `json:"http,omitempty"`
	HTTPS            HTTPS     `json:"https,omitempty"`
	Backends         []Backend `json:"backends,omitempty"`
}

//HTTP is the HTTP part of a a SharedHTTPConfig
type HTTP struct {
	RedirectHTTPS bool            `json:"redirect_https"`
	BackendPort   int             `json:"backend_port,omitempty"`
	HealthCheck   HTTPHealthCheck `json:"health_check,omitempty"`
}

//HTTPS is the HTTPS part of a a SharedHTTPConfig
type HTTPS struct {
	PrivateKey  string          `json:"private_key,omitempty"`
	Certificate string          `json:"certificate,omitempty"`
	BackendPort int             `json:"backend_port,omitempty"`
	HealthCheck HTTPHealthCheck `json:"health_check,omitempty"`
}

//Type imprments the ServiceConfig interface
func (shttp SharedHTTPConfig) Type() ServiceType {
	return SharedHTTP
}

//prepare the http request and marchal the object to send
func prepareRequest(obj interface{}, url, method string) (*http.Request, error) {
	data, err := json.Marshal(obj)
	if err != nil {
		return nil, errors.Wrap(err, "error marshalling Frontend")
	}
	buf := bytes.NewBuffer(data)

	req, err := http.NewRequest("PUT", url, buf)
	if err != nil {
		return nil, errors.Wrap(err, "error creatign http.Request")
	}
	req.Header.Set("Content-Type", jsonContent)
	return req, nil
}

//type of action to discriminate the http request when editing an object
type action int

const (
	replace action = iota
	reconfig
	delete
)

/*
type serviceType int

const (
	tcp serviceType = iota
	tcpProxyProtocol
	sharedHTTP
	mediasite
)
*/

// ServiceType represent the type of service offered by the load balancers
type ServiceType string

// known types of service
const (
	TCP              ServiceType = "tcp"
	TCPProxyProtocol ServiceType = "tcp_proxy_protocol"
	SharedHTTP       ServiceType = "shared-http"
	Mediasite        ServiceType = "mediasite"
)
