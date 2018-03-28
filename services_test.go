package nlb

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/koki/json"
)

var TestServiceString string = `{
	"type": "tcp",
	"metadata": {
		"name": "testservice"
	},
	"config": {
		"method": "least_conn",
		"ports": {
			"80": 443
		},
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

var (
	testServiceGo = Service{
		Type: TCP,
		Metadata: Metadata{
			Name: "testservice",
		},
		Config: Config{
			Method: "least_conn",
			Ports:  map[string]int32{"80": 443},
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
)

func TestUnmarshalService(t *testing.T) {

	msg := new(Service)
	err := json.Unmarshal([]byte(TestServiceString), msg)
	if err != nil {
		t.Errorf("TestUnmarshalTCPService() error umarshalling message = %v", err)
		return
	}

	if !reflect.DeepEqual(*msg, testServiceGo) {
		t.Errorf("TestUnmarshalTCPService() Type = %+v, want %+v", msg, testServiceGo)
	}

}

//This might get a bit flaky...
func TestMarshalTCPService(t *testing.T) {

	j, err := json.Marshal(testServiceGo)
	if err != nil {
		t.Errorf("TestMarshalTCPService() error marshalling = %v", err)
		return
	}
	var out bytes.Buffer
	json.Indent(&out, j, "", "\t")
	got := out.String()
	want := TestServiceString
	if want != got {
		t.Errorf("TestMarshalTCPService() = %+v, want %+v", got, want)
	}

}
