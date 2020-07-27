package event

import (
	"github.com/pkg/errors"
)

//go:generate moq -out mock/marshaller.go -pkg mock . Marshaller

// AvroProducer of output events.
type AvroProducer struct {
	out        chan []byte
	marshaller Marshaller
}

// Marshaller marshals events into messages.
type Marshaller interface {
	Marshal(s interface{}) ([]byte, error)
}

// NewAvroProducer returns a new instance of AvroProducer.
func NewAvroProducer(outputChannel chan []byte, marshaller Marshaller) *AvroProducer {
	return &AvroProducer{
		out:        outputChannel,
		marshaller: marshaller,
	}
}

// Audit produces a new Audit event.
func (producer *AvroProducer) Audit(event *Audit) error {
	bytes, err := producer.Marshal(event)
	if err != nil {
		return err
	}
	producer.Send(bytes)
	return nil
}

// Marshal marshalls an Audit event and returns the corresponding byte array
func (producer *AvroProducer) Marshal(event *Audit) ([]byte, error) {
	if event == nil {
		return nil, errors.New("event required but was nil")
	}
	return producer.marshaller.Marshal(event)
}

// Send sends the byte array to the output channel
func (producer *AvroProducer) Send(bytes []byte) {
	producer.out <- bytes
}
