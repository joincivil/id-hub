package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"strconv"

	log "github.com/golang/glog"
)

// AnyValue is a GraphQL scalar to represent either a string or a map
type AnyValue struct {
	Value interface{}
}

// UnmarshalGQL implements the graphql.Unmarshaler interface
func (a *AnyValue) UnmarshalGQL(v interface{}) error {
	a.Value = v
	return nil
}

// MarshalGQL implements the graphql.Marshaler interface
func (a AnyValue) MarshalGQL(w io.Writer) {
	switch val := a.Value.(type) {
	case bool:
		_, err := fmt.Fprintf(w, "%v", strconv.FormatBool(val))
		if err != nil {
			log.Errorf("Error writing gql int: err %v", err)
		}
	case float64:
		_, err := fmt.Fprintf(w, "%v", val)
		if err != nil {
			log.Errorf("Error writing gql float: err %v", err)
		}
	case int:
		_, err := fmt.Fprintf(w, "%v", val)
		if err != nil {
			log.Errorf("Error writing gql int: err %v", err)
		}
	case string:
		_, err := fmt.Fprintf(w, "\"%v\"", val)
		if err != nil {
			log.Errorf("Error writing gql string: err %v", err)
		}
	default:
		bytes, err := json.Marshal(val)
		if err != nil {
			log.Errorf("Error marshaling map: err %v", err)
		}
		_, err = fmt.Fprintf(w, "%v", string(bytes))
		if err != nil {
			log.Errorf("Error writing gql map: err %v", err)
		}
	}
}
