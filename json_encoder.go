package eventstore

import (
	"encoding/json"
	"reflect"
)

// NewJsonEncoder constructs json encoder
func NewJsonEncoder(evts ...interface{}) *JsonEncoder {
	enc := JsonEncoder{
		types: make(map[string]reflect.Type),
	}

	for _, evt := range evts {
		t := reflect.TypeOf(evt)
		enc.types[t.Name()] = t
	}

	return &enc
}

// JsonEncoder provides default json Encoder implementation
// It will marshal and unmarshal events to/from json and store the type name
type JsonEncoder struct {
	types map[string]reflect.Type
}

// Encode marshals incoming event to it's json representation
func (e *JsonEncoder) Encode(evtData interface{}) (*EncodedEvt, error) {
	data, err := json.Marshal(evtData)
	if err != nil {
		return nil, err
	}

	return &EncodedEvt{
		Type: reflect.TypeOf(evtData).Name(),
		Data: string(data),
	}, nil
}

// Decode unmarshals incoming event to it's corresponding go type
func (e *JsonEncoder) Decode(evt *EncodedEvt) (interface{}, error) {
	v := reflect.New(e.types[evt.Type])

	err := json.Unmarshal([]byte(evt.Data), v.Interface())
	if err != nil {
		return nil, err
	}

	return v.Elem().Interface(), nil
}
