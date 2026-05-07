package valid

import (
	"encoding/json"
	"fmt"
	"testing"
)

type testAAA struct{}

func (t testAAA) Validate(value any) (*Violation, error) {
	return &Violation{
		Value: value,
		Msg:   "aaa",
	}, nil
}

func (t testAAA) Msg(message string) Rule {
	panic("implement me")
}

func TestCheck(t *testing.T) {
	type test2Struct struct {
		WWW string `validate:"abc"`
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
