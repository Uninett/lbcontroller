package nlb

import (
	"net"
	"time"
)

//Frontend represent a frontend object for the load balancers
//{
//	"type": "frontend"
//	"metadata": ...
//	"config": {
//		 "addresses": ["10.40.50.23","2001:700:fffd::23"]
//	}
//}
type Frontend struct {
	Metadata Metadata `json:"metadata,omitempty"`
	//Config is the configuration of the frontend
	Config FrontendConfig `json:"config,omitempty"`
}

//FrontendConfig is the configuration of a Frontend object
type FrontendConfig struct {
	Addresses []net.IP `json:"addresses,omitempty"`
}

// TCPConfig represent the configuration of a TCP load balanced service
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
	Method           *string            `json:"method,omitempty"`
	Ports            []uint16           `json:"ports,omitempty"`
	Backends         map[string]Backend `json:"backends,omitempty"`
	UpstreamMaxConns *int               `json:"upstream_max_conns,omitempty"`
	ACL              []net.IPNet        `json:"acl,omitempty"`
	HealthCheck      *HealthCheck       `json:"health_check,omitempty"`
	Frontend         *Frontend          `json:"frontend,omitempty"`
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
