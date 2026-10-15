package publisher

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
)

type ConsultationCompletedEvent struct {
	EventID        string    `json:"event_id"`
	ConsultationID string    `json:"consultation_id"`
	PatientID      string    `json:"patient_id"`
	DoctorID       string    `json:"doctor_id"`
	Diagnosis      string    `json:"diagnosis"`
	CompletedAt    time.Time `json:"completed_at"`
}

type EventPublisher struct {
	conn    *amqp.Connection
	channel *amqp.Channel
}

func NewEventPublisher(amqpURL string) (*EventPublisher, error) {
	conn, err := amqp.Dial(amqpURL)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, err
	}

	// Deklarasikan Topic Exchange sesuai standar topologi
	err = ch.ExchangeDeclare(
		"medcore.clinical.events", // name
		"topic",                   // type
		true,                      // durable
		false,                     // auto-deleted
		false,                     // internal
		false,                     // no-wait
		nil,                       // arguments
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, err
	}

	return &EventPublisher{conn: conn, channel: ch}, nil
}

func (p *EventPublisher) PublishConsultationCompleted(ctx context.Context, consultationID, patientID, doctorID, diagnosis string) error {
	event := ConsultationCompletedEvent{
		EventID:        uuid.New().String(),
		ConsultationID: consultationID,
		PatientID:      patientID,
		DoctorID:       doctorID,
		Diagnosis:      diagnosis,
		CompletedAt:    time.Now(),
	}

	body, err := json.Marshal(event)
	if err != nil {
		return err
	}

	routingKey := "clinical.consultation.completed"

	err = p.ChannelPublishWithContext(ctx, "medcore.clinical.events", routingKey, body)
	if err != nil {
		log.Printf("[Publisher Error] Gagal mengirim event: %v", err)
		return err
	}

	log.Printf("[Publisher Success] Event Consultation Completed dikirim dengan ID: %s", event.EventID)
	return nil
}

func (p *EventPublisher) ChannelPublishWithContext(ctx context.Context, exchange, routingKey string, body []byte) error {
	return p.channel.PublishWithContext(
		ctx,
		exchange,
		routingKey,
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
}

func (p *EventPublisher) Close() {
	if p.channel != nil {
		p.channel.Close()
	}
	if p.conn != nil {
		p.conn.Close()
	}
}