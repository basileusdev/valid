package valid

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
)

type testAAA struct{}

func (t testAAA) Validate(value reflect.Value) (string, error) {
	return "aaa", nil
}

func (t testAAA) Msg(message string) Rule {
	panic("implement me")
}

func TestCheck(t *testing.T) {
	type test2Struct struct {
		WWW string `validate:"required,min=3,(max=4,len=3)"`
	}
	type testStruct struct {
		Abc string        `validate:"abc"`
		Cba string        `validate:"abc"`
		AAA []test2Struct `validate:"abc"`
	}

	v := New()
	v.Rules["abc"] = testAAA{}

	err := v.Check(testStruct{
		Abc: "123",
		Cba: "456",
		AAA: []test2Struct{
			{
				WWW: "789",
			},
			{
				WWW: "123",
			},
		},
	})
	if err != nil {
		bytes, _ := json.MarshalIndent(err, "", "\t")
		t.Fatal(string(bytes))
	}

	fmt.Printf("OK")
}
