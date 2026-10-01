package main

import (
	"encoding/json"
	"fmt"
	// Sesuaikan path import ini dengan nama modul di go.mod Anda dan lokasi package proto/pb Anda
	// Contoh: pb "github.com/Kevinananda13/Praktikum-2-Mikrosevis-September-2026/proto"
	// "google.golang.org/protobuf/proto"
)

func main() {
	// 1. Data Sampel untuk payload JSON (REST API)
	type PatientJSON struct {
		ID      string `json:"id"`
		Name    string `json:"name"`
		Email   string `json:"email"`
		Status  string `json:"status"`
		Address string `json:"address"`
	}

	sampleData := PatientJSON{
		ID:      "PAT-99823",
		Name:    "Budi Santoso",
		Email:   "budi.santoso@example.com",
		Status:  "CONFIRMED",
		Address: "Jl. Raya Sesetan No. 45, Denpasar, Bali",
	}

	// 2. Hitung ukuran payload JSON
	jsonBytes, err := json.Marshal(sampleData)
	if err != nil {
		panic(err)
	}
	jsonSize := len(jsonBytes)

	// 3. Hitung ukuran payload Protobuf (gRPC)
	// Jika Anda ingin mengaktifkan perbandingan Protobuf secara nyata, 
	// uncomment baris di bawah ini dan sesuaikan dengan struct Protobuf yang ada di project Anda:
	/*
		protoMsg := &pb.PatientResponse{
			Id:      sampleData.ID,
			Name:    sampleData.Name,
			Email:   sampleData.Email,
			Status:  sampleData.Status,
			Address: sampleData.Address,
		}
		protoBytes, err := proto.Marshal(protoMsg)
		if err != nil {
			panic(err)
		}
		protoSize := len(protoBytes)
	*/

	// Untuk estimasi/simulasi jika struct protobuf belum di-import langsung:
	// Protobuf menggunakan enkripsi biner (Varint & Tag-Value) yang ukurannya jauh lebih ringkas dari key string JSON.
	// Sebagai ilustrasi umum, payload di atas dalam bentuk Protobuf biasanya sekitar 50-70 bytes.
	
	// 4. Tampilkan Hasil Analisis Wire-Size
	fmt.Println("==================================================")
	fmt.Println("       HASIL INSPEKSI WIRE-SIZE PAYLOAD         ")
	fmt.Println("==================================================")
	fmt.Printf("Ukuran JSON Payload (REST)     : %d bytes\n", jsonSize)
	
	// Jika sudah menggunakan pb.Marshal asli, ganti bagian ini dengan variabel protoSize:
	// fmt.Printf("Ukuran Protobuf Payload (gRPC) : %d bytes\n", protoSize)
	// reduction := (1 - float64(protoSize)/float64(jsonSize)) * 100
	// fmt.Printf("Efisiensi Reduksi Payload      : %.2f%%\n", reduction)
	
	fmt.Println("--------------------------------------------------")
	fmt.Println("Kesimpulan: Format biner Protobuf menghilangkan redundansi")
	fmt.Println("nama key (field name string) sehingga ukuran byte di")
	fmt.Println("jaringan jauh lebih kecil dibanding teks JSON.")
	fmt.Println("==================================================")
}