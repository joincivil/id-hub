package utils_test

import (
	"bytes"

	"testing"

	"github.com/joincivil/id-hub/pkg/utils"
)

func TestAnyValue(t *testing.T) {
	val := &utils.AnyValue{Value: "stringtest"}
	buf := bytes.NewBufferString("")
	// gql marshalling string
	val.MarshalGQL(buf)
	if buf.String() != "\"stringtest\"" {
		t.Errorf("Incorrect string value: %v", buf.String())
	}

	val = &utils.AnyValue{Value: 10}
	buf = bytes.NewBufferString("")
	// gql marshalling string
	val.MarshalGQL(buf)
	if buf.String() != "10" {
		t.Errorf("Incorrect int value")
	}

	val = &utils.AnyValue{Value: false}
	buf = bytes.NewBufferString("")
	// gql marshalling string
	val.MarshalGQL(buf)
	if buf.String() != "false" {
		t.Errorf("Incorrect bool value")
	}

	val = &utils.AnyValue{Value: float64(10.1)}
	buf = bytes.NewBufferString("")
	// gql marshalling string
	val.MarshalGQL(buf)
	if buf.String() != "10.1" {
		t.Errorf("Incorrect int value")
	}

	val = &utils.AnyValue{Value: map[string]int{"test": 1, "test2": 100}}
	buf = bytes.NewBufferString("")
	// gql marshalling string
	val.MarshalGQL(buf)
	if buf.String() != "{\"test\":1,\"test2\":100}" {
		t.Errorf("Incorrect int value")
	}

	val = &utils.AnyValue{}
	_ = val.UnmarshalGQL("{\"test\":1,\"test2\":100}")

}
