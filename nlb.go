package nlb

import (
	"net"
	"time"
)

//Frontend represent a frontend object fore the load balancers
type Frontend struct {
	Metadata Metadata
	//Config is the configuration of the frontend
	Config FrontendConfig
}

//FrontendConfig is the configuration of a Frontend object
type FrontendConfig struct {
	Addresses []net.IP
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
	Port   uint16
	Send   string
	Expect string
}

//Metadata of messages sent to the API
//"metadata": {
//	"name": "testservice",
//	"created_at": "......",
//	"updated_at": "......",
//	...
//}
type Metadata struct {
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}
