package eventstore

import (
	"encoding/json"
	"reflect"
)

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

type JsonEncoder struct {
	types map[string]reflect.Type
}

func (e *JsonEncoder) Encode(evtData interface{}) (*EncodedEvt, error) {
	data, err := json.Marshal(evtData)
	if err != nil {
		return nil, err
	}

	return &EncodedEvt{
		Type: reflect.TypeOf(evtData).Elem().Name(),
		Data: string(data),
	}, nil
}

func (e *JsonEncoder) Decode(evt *EncodedEvt) (interface{}, error) {
	v := reflect.New(e.types[evt.Type])

	err := json.Unmarshal([]byte(evt.Data), v.Interface())
	if err != nil {
		return nil, err
	}

	return v.Interface(), nil
}
