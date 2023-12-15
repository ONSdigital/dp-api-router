package event

import (
	"context"

	kafka "github.com/ONSdigital/dp-kafka/v4"
	"github.com/pkg/errors"
)

//go:generate moq -out mock/marshaller.go -pkg mock . Marshaller

// AvroProducer of output events.
type AvroProducer struct {
	out        chan kafka.BytesMessage
	marshaller Marshaller
}

// Marshaller marshals events into messages.
type Marshaller interface {
	Marshal(s interface{}) ([]byte, error)
}

// NewAvroProducer returns a new instance of AvroProducer.
func NewAvroProducer(outputChannel chan kafka.BytesMessage, marshaller Marshaller) *AvroProducer {
	return &AvroProducer{
		out:        outputChannel,
		marshaller: marshaller,
	}
}

// Audit produces a new Audit event.
func (producer *AvroProducer) Audit(ctx context.Context, event *Audit) error {
	bytes, err := producer.Marshal(event)
	if err != nil {
		return err
	}
	producer.Send(ctx, bytes)
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
func (producer *AvroProducer) Send(ctx context.Context, bytes []byte) {
	producer.out <- kafka.BytesMessage{Value: bytes, Context: ctx}
}
