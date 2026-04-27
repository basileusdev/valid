package valid

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
)

type testAAA struct{}

func (t testAAA) Validate(value any, field reflect.StructField) error {
	return ValidationError{
		Field:   field.Name,
		Rule:    "abc",
		Value:   value,
		Message: "aaa",
	}
}

func (t testAAA) Msg(message string) Rule {
	panic("implement me")
}

func TestCheck(t *testing.T) {
	type test2Struct struct {
		WWW string `valid:"abc"`
	}
	type testStruct struct {
		Abc string      `valid:"abc"`
		Cba string      `valid:"abc"`
		AAA test2Struct `valid:"abc"`
	}

	v := New()
	v.Rules["abc"] = testAAA{}

	err := v.Check(testStruct{
		Abc: "123",
		Cba: "456",
		AAA: test2Struct{
			WWW: "789",
		},
	})
	if err != nil {
		bytes, _ := json.MarshalIndent(err, "", "\t")
		t.Fatal(string(bytes))
	}

	fmt.Printf("OK")
}
