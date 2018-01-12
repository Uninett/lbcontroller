package nlb

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/koki/json"
)

const (
	TestSharedHTTPServiceString string = `{
	"type": "shared-http",
	"metadata": {
		"name": "testservice"
	},
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
			"private_key": "5aNv4UxBlpIQIRsYKplIWWd+D",
			"certificate": "Df0tz2wBszL9sJYhPjOIAjk+a",
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

var (
	testTCPServiceGo = Service{
		Type: TCP,
		Metadata: Metadata{
			Name: "testservice",
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
			HealthCheck: TCPHealthCheck{
				Port:   1337,
				Send:   "healthz\n",
				Expect: "^OK$",
			},
			Frontend: "foobar",
		},
	}
	testSharedHTTPServiceGo = Service{
		Type: SharedHTTP,
		Metadata: Metadata{
			Name: "testservice",
		},
		Config: SharedHTTPConfig{
			Names:            []string{"site-a.example.com", "site-b.foo.org"},
			StickyBackends:   false,
			BackendProtocols: "both",
			HTTP: HTTP{
				RedirectHTTPS: true,
				BackendPort:   8080,
				HealthCheck: HTTPHealthCheck{
					URI:        "/",
					StatusCode: 301,
				},
			},
			HTTPS: HTTPS{
				PrivateKey:  "5aNv4UxBlpIQIRsYKplIWWd+D",
				Certificate: "Df0tz2wBszL9sJYhPjOIAjk+a",
				BackendPort: 8888,
				HealthCheck: HTTPHealthCheck{
					URI:        "/healthz",
					StatusCode: 200,
					Body:       "OK",
				},
			},
			Backends: []Backend{
				{
					Host:  "hostname1.example.com",
					Addrs: []string{"10.3.2.43", "2001:700:f00d::8"},
				},
			},
		},
	}
)

func TestUnmarshalTCPService(t *testing.T) {

	msg := new(Message)
	err := json.Unmarshal([]byte(TestTCPServiceString), msg)
	if err != nil {
		t.Errorf("TestUnmarshalTCPService() error umarshalling message = %v", err)
		return
	}

	got := msg.Type
	want := testTCPServiceGo.Type
	if !reflect.DeepEqual(got, want) {
		t.Errorf("TestUnmarshalTCPService() Type = %+v, want %+v", got, want)
	}

	got1 := msg.Metadata
	want1 := testTCPServiceGo.Metadata
	if !reflect.DeepEqual(got, want) {
		t.Errorf("TestUnmarshalTCPService() Metadata = %+v, want %+v", got1, want1)
	}

	tcpConf := new(TCPConfig)
	err = json.Unmarshal(msg.Config, tcpConf)
	if err != nil {
		t.Errorf("TestUnmarshalTCPService() error umarshalling config = %v", err)
		return
	}

	got2 := *tcpConf
	want2 := testTCPServiceGo.Config
	if !reflect.DeepEqual(got, want) {
		t.Errorf("TestUnmarshalTCPService()Config  = %+v, want %+v", got2, want2)
	}
}

//This might get a bit flaky...
func TestMarshalTCPService(t *testing.T) {

	j, err := json.Marshal(testTCPServiceGo)
	if err != nil {
		t.Errorf("TestMarshalTCPService() error marshalling = %v", err)
		return
	}
	var out bytes.Buffer
	json.Indent(&out, j, "", "\t")
	got := out.String()
	want := TestTCPServiceString
	if want != got {
		t.Errorf("TestMarshalTCPService() = %+v, want %+v", got, want)
	}

}

//This might get a bit flaky...
func TestMarshalSharedHTTPService(t *testing.T) {

	j, err := json.Marshal(testSharedHTTPServiceGo)
	if err != nil {
		t.Errorf("TestMarshalSharedHTTPService() error marshalling = %v", err)
		return
	}
	var out bytes.Buffer
	json.Indent(&out, j, "", "\t")
	got := out.String()
	want := TestSharedHTTPServiceString
	if want != got {
		t.Errorf("TestMarshalSharedHTTPService() = %+v, want %+v", got, want)
	}

}

func TestUnmarshalSharedHTTPService(t *testing.T) {

	msg := new(Message)
	err := json.Unmarshal([]byte(TestSharedHTTPServiceString), msg)
	if err != nil {
		t.Errorf("TestUnmarshalSharedHTTPService() error umarshalling message = %v", err)
		return
	}

	got := msg.Type
	want := testSharedHTTPServiceGo.Type
	if got != want {
		t.Errorf("TestUnmarshalSharedHTTPService()Type = %+v, want %+v", got, want)
		return
	}

	got1 := msg.Metadata
	want1 := testSharedHTTPServiceGo.Metadata
	if !reflect.DeepEqual(got, want) {
		t.Errorf("TestUnmarshalSharedHTTPService() Metadata = %+v, want %+v", got1, want1)
		return
	}

	tcpConf := new(SharedHTTPConfig)
	err = json.Unmarshal(msg.Config, tcpConf)
	if err != nil {
		t.Errorf("TestUnmarshalSharedHTTPService() error umarshalling config = %v", err)
		return
	}

	got2 := *tcpConf
	want2 := testSharedHTTPServiceGo.Config
	if !reflect.DeepEqual(got, want) {
		t.Errorf("TestUnmarshalSharedHTTPService() Config = %+v, want %+v", got2, want2)
	}
}
