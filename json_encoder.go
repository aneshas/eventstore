package eventstore

import (
	"encoding/json"
	"fmt"
	"reflect"
)

// NewJSONEncoder constructs json encoder
// It receives a slice of event types it should be able to encode/decode
func NewJSONEncoder(events ...any) *JsonEncoder {
	enc := JsonEncoder{
		types: make(map[string]reflect.Type),
	}

	for _, evt := range events {
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
func (e *JsonEncoder) Encode(evt any) (*EncodedEvt, error) {
	data, err := json.Marshal(evt)
	if err != nil {
		return nil, err
	}

	return &EncodedEvt{
		Type: reflect.TypeOf(evt).Name(),
		Data: string(data),
	}, nil
}

// Decode decodes incoming event to it's corresponding go type
func (e *JsonEncoder) Decode(evt *EncodedEvt) (any, error) {
	t, ok := e.types[evt.Type]
	if !ok {
		return nil, fmt.Errorf("event not registered via NewJSONEncoder")
	}

	v := reflect.New(t)

	err := json.Unmarshal([]byte(evt.Data), v.Interface())
	if err != nil {
		return nil, err
	}

	return v.Elem().Interface(), nil
}
