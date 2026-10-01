package main

import (
	"context"
	"log"
	"time"

	pb "appointment-service/pb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

func main() {
	log.Println("[Appointment Service] Menghubungkan ke gRPC Pharmacy Service...")

	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Tidak dapat terhubung ke Pharmacy Service: %v", err)
	}
	defer conn.Close()

	client := pb.NewPharmacyServiceClient(conn)

	// Permintaan cek stok obat (contoh: Amoxicillin 500mg, butuh 20)
	req := &pb.CheckDrugRequest{
		DrugCode:       "MED-AMX-500",
		QuantityNeeded: 20,
	}

	// === UBAH TIMEOUT MENJADI 100ms UNTUK MEMICU DEADLINE EXCEEDED ===
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	resp, err := client.CheckDrugAvailability(ctx, req)
	if err != nil {
		st, _ := status.FromError(err)
		log.Fatalf("RPC Gagal: Code=%s, Message=%s", st.Code(), st.Message())
	}

	log.Println("================== HASIL RESPON gRPC ==================")
	log.Printf("Kode Obat    : %s", resp.GetDrugCode())
	log.Printf("Tersedia     : %t", resp.GetIsAvailable())
	log.Printf("Stok Aktual  : %d", resp.GetCurrentStock())
	log.Printf("Harga Satuan : Rp %.2f", resp.GetUnitPrice())
	log.Printf("Pesan Sistem : %s", resp.GetMessage())
	log.Println("=======================================================")
}