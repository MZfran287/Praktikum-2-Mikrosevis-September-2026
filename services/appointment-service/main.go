package main

import (
	"context"
	"log"
	"time"

	pb "appointment-service/pb"
	"appointment-service/publisher" // Pastikan sesuai nama modul go.mod Anda

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

func main() {
	log.Println("[Appointment Service] Memulai layanan...")

	// 1. Integrasi gRPC ke Pharmacy Service (Non-Fatal Error Handling)
	log.Println("[Appointment Service] Menghubungkan ke gRPC Pharmacy Service...")
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("Peringatan: Tidak dapat terhubung ke Pharmacy Service: %v", err)
	} else {
		defer conn.Close()

		client := pb.NewPharmacyServiceClient(conn)

		req := &pb.CheckDrugRequest{
			DrugCode:       "MED-AMX-500",
			QuantityNeeded: 20,
		}

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		resp, err := client.CheckDrugAvailability(ctx, req)
		if err != nil {
			st, _ := status.FromError(err)
			log.Printf("gRPC RPC Gagal: Code=%s, Message=%s", st.Code(), st.Message())
		} else {
			log.Println("================== HASIL RESPON gRPC ==================")
			log.Printf("Kode Obat    : %s", resp.GetDrugCode())
			log.Printf("Tersedia     : %t", resp.GetIsAvailable())
			log.Printf("Stok Aktual  : %d", resp.GetCurrentStock())
			log.Printf("Harga Satuan : Rp %.2f", resp.GetUnitPrice())
			log.Printf("Pesan Sistem : %s", resp.GetMessage())
			log.Println("=======================================================")
		}
	}

	// 2. Inisialisasi RabbitMQ Event Publisher (Tugas Mahasiswa B)
	amqpURL := "amqp://admin_medcore:SecuredPassRabbit2026!@127.0.0.1:5672/%2Fmedcore_vhost"
	pub, err := publisher.NewEventPublisher(amqpURL)
	if err != nil {
		log.Fatalf("Gagal terhubung ke RabbitMQ Broker: %v", err)
	}
	defer pub.Close()

	log.Println("[Appointment Service] RabbitMQ Publisher aktif. Mengirim event konsultasi selesai...")

	pubCtx, pubCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer pubCancel()

	// Memublikasikan event konsultasi selesai ke broker
	err = pub.PublishConsultationCompleted(
		pubCtx,
		"CONS-2026-001",
		"PAT-9988",
		"DOC-1122",
		"Hypertension Stage 1",
	)
	if err != nil {
		log.Printf("Gagal memublikasikan event RabbitMQ: %v", err)
	} else {
		log.Println("[Appointment Service] Event konsultasi selesai berhasil dikirim via RabbitMQ!")
	}
}