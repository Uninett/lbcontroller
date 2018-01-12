package nlb

import (
	"fmt"
	"testing"

	"github.com/koki/json"
)

func TestFrontendsUnmarshal(t *testing.T) {
	bytes := []byte(`
		{
        "type": "frontend",
        "metadata": {
			"name": "testservice",
			"created_at": "2012-04-23T18:25:43.511Z",
			"updated_at": "2016-12-23T18:25:43.511Z"
		},
        "config": {
			 "addresses": ["10.40.50.23","2001:700:fffd::23"]
		}
	}`)
	fronts := Frontend{}
	err := json.Unmarshal(bytes, &fronts)
	if err != nil {
		t.Errorf("%v\n", err)
	}
	fmt.Println(fronts)

}
