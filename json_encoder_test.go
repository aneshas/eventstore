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
	enc := eventstore.NewJsonEncoder(SomeEvent{}, AnotherEvent{})

	decodeEncode(t, enc, &SomeEvent{
		UserID: "some-user",
	})

	decodeEncode(t, enc, &AnotherEvent{
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
