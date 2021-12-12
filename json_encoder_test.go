package eventstore_test

import (
	"reflect"
	"testing"

	"github.com/aneshas/eventstore"
)

type AnotherEvent struct {
	Smth string
}

func TestShouldDecodeEncodedEvent(t *testing.T) {
	enc := eventstore.NewJSONEncoder(SomeEvent{}, AnotherEvent{})

	decodeEncode(t, enc, SomeEvent{
		UserID: "some-user",
	})

	decodeEncode(t, enc, AnotherEvent{
		Smth: "foo",
	})
}

func decodeEncode(t *testing.T, enc eventstore.Encoder, e interface{}) {
	encoded, err := enc.Encode(e)
	if err != nil {
		t.Fatalf("%v", err)
	}

	decoded, err := enc.Decode(encoded)
	if err != nil {
		t.Fatalf("%v", err)
	}

	if !reflect.DeepEqual(e, decoded) {
		t.Fatalf("event not decoded. want: %#v, got: %#v", e, decoded)
	}
}

func TestShouldErrorOutIfMalformedJSON(t *testing.T) {
	enc := eventstore.NewJSONEncoder(SomeEvent{}, AnotherEvent{})

	_, err := enc.Decode(&eventstore.EncodedEvt{
		Data: "malformed-json",
		Type: "SomeEvent",
	})
	if err == nil {
		t.Fatal("should error out")
	}
}

func TestShouldErrorOutIfEvtTypeUnknown(t *testing.T) {
	enc := eventstore.NewJSONEncoder(SomeEvent{}, AnotherEvent{})

	_, err := enc.Decode(&eventstore.EncodedEvt{
		Data: `{"userId": "123"}`,
		Type: "unknown",
	})
	if err == nil {
		t.Fatal("should error out")
	}
}
