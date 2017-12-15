package nlb

import (
	"encoding/json"
	"net"
	"time"
)

//Message from/to the API endpoint, e.g.
//{
//	"type": "frontend"
//	"metadata": ...
//	"config": {
//		 "addresses": ["10.40.50.23","2001:700:fffd::23"]
//	}
//}
//different types comes with different configurations.
type Message struct {
	Type     string          `json:"type,omitempty"`
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
	Addresses []net.IP `json:"addresses,omitempty"`
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
	Method           string             `json:"method,omitempty"`
	Ports            []uint16           `json:"ports,omitempty"`
	Backends         map[string]Backend `json:"backends,omitempty"`
	UpstreamMaxConns int                `json:"upstream_max_conns,omitempty"`
	ACL              []net.IPNet        `json:"acl,omitempty"`
	HealthCheck      HealthCheck        `json:"health_check,omitempty"`
	Frontend         string             `json:"frontend,omitempty"`
}

// Backend represents a backend in the loadbalancer configuration
type Backend struct {
	Addrs []net.IP
}

// HealthCheck is a loadbalancer heath check
type HealthCheck struct {
	Port   uint16 `json:"port,omitempty"`
	Send   string `json:"send,omitempty"`
	Expect string `json:"expect,omitempty"`
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
	Names            []string
	StickyBackends   bool
	BackendProtocols string
	HTTP             json.RawMessage
	HTTPS            json.RawMessage
	Backends         []Backend
}
