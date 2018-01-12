package nlb

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/koki/json"
)

const testFrontendString = `{
	"type": "frontend",
	"metadata": {
		"name": "testservice"
	},
	"config": {
		"addresses": [
			"10.40.50.23",
			"2001:700:fffd::23"
		]
	}
}`

var testFrontendGo = Frontend{
	Type: "frontend",
	Metadata: Metadata{
		Name: "testservice",
	},
	Config: FrontendConfig{
		Addresses: []string{"10.40.50.23", "2001:700:fffd::23"},
	},
}

func TestUnmarshalFrontend(t *testing.T) {

	fronts := Frontend{}
	err := json.Unmarshal([]byte(testFrontendString), &fronts)
	if err != nil {
		t.Errorf("TestUnmarshalFrontend() error umarshalling frontend = %v", err)
		return
	}
	got := fronts
	want := testFrontendGo
	if !reflect.DeepEqual(got, want) {
		t.Errorf("TestUnmarshalFrontend()  = %+v, want %+v", got, want)
	}
}

func TestMarshalFrontend(t *testing.T) {

	j, err := json.Marshal(testFrontendGo)
	if err != nil {
		t.Errorf("TestMarshalFrontend() error marshalling = %v", err)
		return
	}
	var out bytes.Buffer
	json.Indent(&out, j, "", "\t")
	got := out.String()
	want := testFrontendString
	if want != got {
		t.Errorf("TestMarshalFrontend() = %+v, want %+v", got, want)
	}

}
