package main

import (
	"reflect"
	"testing"
)

func TestMessageKey(t *testing.T) {
	m := Message{
		ContentID: 3,
		ClientID:  2,
		Timestamp: 946692184,
		Payload:   []byte("{\"test\": 1}"),
	}

	key := m.Key()

	if key != "/chat/1970-01-12/content_logs_1970-01-12_2" {
		t.Error("Message S3 object key does not follow convention")
	}
}

func TestNewWorker (t *testing.T) {
	w := NewWorker(10)

	if reflect.TypeOf(w).String() != "*main.Worker" {
		t.Error("NewWorker should create Worker instance")
	}
}
