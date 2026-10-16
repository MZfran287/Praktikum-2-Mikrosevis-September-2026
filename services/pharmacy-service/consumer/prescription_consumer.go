package consumer

import (
	"context"
	"encoding/json"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type PrescriptionConsumer struct {
	conn *amqp.Connection
	db   *mongo.Database
}

func NewPrescriptionConsumer(conn *amqp.Connection, db *mongo.Database) *PrescriptionConsumer {
	return &PrescriptionConsumer{conn: conn, db: db}
}

func (c *PrescriptionConsumer) Start() {
	ch, err := c.conn.Channel()
	if err != nil {
		log.Fatalf("Gagal membuka channel: %v", err)
	}
	defer ch.Close()

	args := amqp.Table{
		"x-dead-letter-exchange": "medcore.deadletter.exchange",
	}
	q, err := ch.QueueDeclare(
		"pharmacy.prescription.order.q",
		true, false, false, false, args,
	)
	if err != nil {
		log.Fatalf("Gagal mendeklarasikan antrean: %v", err)
	}

	err = ch.QueueBind(
		q.Name,
		"medcore.event.clinical.consultation.completed",
		"medcore.clinical.events",
		false, nil,
	)
	if err != nil {
		log.Fatalf("Gagal melakukan binding: %v", err)
	}

	msgs, err := ch.Consume(
		q.Name, "", false, false, false, false, nil,
	)
	if err != nil {
		log.Fatalf("Gagal mendaftarkan consumer: %v", err)
	}

	log.Println("[*] Pharmacy Consumer berjalan. Menunggu pesanan resep...")

	collection := c.db.Collection("processed_events")

	for d := range msgs {
		var event struct {
			EventID   string `json:"event_id"`
			PatientID string `json:"patient_id"`
		}

		if err := json.Unmarshal(d.Body, &event); err != nil {
			log.Printf("Pesan korup terdeteksi! Mengirim ke DLQ... %v", err)
			d.Nack(false, false)
			continue
		}

		var existing bson.M
		err = collection.FindOne(context.TODO(), bson.M{"event_id": event.EventID}).Decode(&existing)
		if err == nil {
			log.Printf("Deduplikasi: Event %s sudah pernah diproses. Melewati...", event.EventID)
			d.Ack(false)
			continue
		}

		log.Printf("Memproses resep obat untuk Pasien ID: %s", event.PatientID)

		_, err = collection.InsertOne(context.TODO(), bson.M{"event_id": event.EventID})
		if err != nil {
			log.Printf("Gagal mencatat histori proses ke DB: %v", err)
			d.Nack(false, true)
			continue
		}

		d.Ack(false)
		log.Println("Pesan sukses diproses dan di-ACK.")
	}
}
