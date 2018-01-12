package nlb

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/koki/json"
)

const (
	//TODO(gta): add metadata and certkey
	TestSharedHTTPServiceString string = `{
		"type": "shared-http",
		"config": {
		  "names": [
			"site-a.example.com",
			"site-b.foo.org"
		  ],
		  "sticky_backends": false,
		  "backend_protocols": "both",
		  "http": {
			"redirect_https": true,
			"backend_port": 8080,
			"health_check": {
			  "uri": "/",
			  "status_code": 301
			}
		  },
		  "https": {
			"private_key": "........",
			"certificate": "........",
			"backend_port": 8888,
			"health_check": {
			  "uri": "/healthz",
			  "status_code": 200,
			  "body": "OK"
			}
		  },
		  "backends": [
			{
				"host": "hostname1.example.com",
				"addrs": [
					"10.3.2.43",
					"2001:700:f00d::8"
				]
			}
		  ]
		}
	  }`
	TestTCPServiceString string = `{
		"type": "tcp",
		"metadata": {
			"name": "testservice"
		},
		"config": {
			"method": "least_conn",
			"ports": [
				80,
				443
			],
			"backends": [
				{
					"host": "hostname1.example.com",
					"addrs": [
						"10.3.2.43",
						"2001:700:f00d::8"
					]
				},
				{
					"host": "hostname2.example.com",
					"addrs": [
						"10.3.2.53",
						"2001:700:f00d::18"
					]
				}
			],
			"upstream_max_conns": 100,
			"acl": [
				"10.10.20.0/24",
				"2001:700:1337::/48"
			],
			"health_check": {
				"port": 1337,
				"send": "healthz\n",
				"expect": "^OK$"
			},
			"frontend": "foobar"
		}
	}`
)

var testSharedHTTPServiceGo = Service{
	Type: "tcp",
	Metadata: Metadata{
		Name: "testService",
	},
	Config: TCPConfig{
		Method: "least_conn",
		Ports:  []uint16{80, 443},
		Backends: []Backend{
			{
				Host:  "hostname1.example.com",
				Addrs: []string{"10.3.2.43", "2001:700:f00d::8"},
			},
			{
				Host:  "hostname2.example.com",
				Addrs: []string{"10.3.2.53", "2001:700:f00d::18"},
			},
		},
		UpstreamMaxConns: 100,
		ACL:              []string{"10.10.20.0/24", "2001:700:1337::/48"},
		HealthCheck: HealthCheck{
			Port:   1337,
			Send:   "healthz\n",
			Expect: "^OK$",
		},
		Frontend: "foobar",
	},
}

func TestUnmarshalTCPService(t *testing.T) {

	msg := new(Message)
	err := json.Unmarshal([]byte(TestTCPServiceString), msg)
	if err != nil {
		t.Errorf("TestUnmarshalTCPService() error umarshalling message = %v", err)
		return
	}

	got := msg.Type
	want := testSharedHTTPServiceGo.Type
	if !reflect.DeepEqual(got, want) {
		t.Errorf("TestUnmarshalTCPService() = %+v, want %+v", got, want)
	}

	got1 := msg.Metadata
	want1 := testSharedHTTPServiceGo.Metadata
	if !reflect.DeepEqual(got, want) {
		t.Errorf("TestUnmarshalTCPService() = %+v, want %+v", got1, want1)
	}

	tcpConf := new(TCPConfig)
	err = json.Unmarshal(msg.Config, tcpConf)
	if err != nil {
		t.Errorf("TestUnmarshalTCPService() error umarshalling config = %v", err)
		return
	}

	got2 := *tcpConf
	want2 := testSharedHTTPServiceGo.Config
	if !reflect.DeepEqual(got, want) {
		t.Errorf("TestUnmarshalTCPService() = %+v, want %+v", got2, want2)
	}
}

func TestMarshalTCPService(t *testing.T) {

	//msg := Message{
	//	Type:     "tcp",
	//	Metadata: Metadata{},
	//}
	msg := new(Message)
	err := json.Unmarshal([]byte(TestTCPServiceString), msg)
	if err != nil {
		t.Errorf("TestUnmarshalTCPService() error umarshalling message = %v", err)
		return
	}

	//tcpConf := TCPConfig{
	//	Method: "least_conn",
	//	Ports:  []uint16{80, 443},
	//}
	tcpConf := new(TCPConfig)
	err = json.Unmarshal(msg.Config, tcpConf)
	if err != nil {
		t.Errorf("TestUnmarshalTCPService() error umarshalling config = %v", err)
		return
	}
	fmt.Printf("%#v\n", tcpConf)

	//want := Service{}
	//if !reflect.DeepEqual(got, want) {
	//	t.Errorf("TestUnmarshalTCPService() = %v, want %v", got, want)
	//}
}
