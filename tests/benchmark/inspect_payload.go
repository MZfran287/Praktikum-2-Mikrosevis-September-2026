package main

import (
	"encoding/json"
	"fmt"
)

func main() {
	// Simulasi payload data yang sama persis
	respJSON := map[string]interface{}{
		"drug_code":     "MED-AMX-500",
		"is_available":  true,
		"current_stock": 500,
		"unit_price":    3500.0,
		"message":       "Stok obat mencukupi",
	}
	
	jsonBytes, _ := json.Marshal(respJSON)

	// Simulasi ukuran biner Protobuf (format biner wire-size gRPC)
	protoBytes := []byte{0x0a, 0x0b, 0x4d, 0x45, 0x44, 0x2d, 0x41, 0x4d, 0x58, 0x2d, 0x35, 0x30, 0x30, 0x10, 0x01, 0x18, 0xf4, 0x03, 0x21, 0x00, 0x00, 0x5c, 0x44, 0x2a, 0x13, 0x53, 0x74, 0x6f, 0x6b, 0x20, 0x6f, 0x62, 0x61, 0x74, 0x20, 0x6d, 0x65, 0x6e, 0x63, 0x75, 0x6b, 0x75, 0x70}

	fmt.Println("================ ANALISIS WIRE-SIZE PAYLOAD ================")
	fmt.Printf("Ukuran Payload JSON (HTTP/1.1) : %d Bytes\n", len(jsonBytes))
	fmt.Printf("Ukuran Payload Protobuf (HTTP/2) : %d Bytes\n", len(protoBytes))
	
	efficiency := (1.0 - float64(len(protoBytes))/float64(len(jsonBytes))) * 100.0
	fmt.Printf("Efisiensi Reduksi Ukuran Kawat : %.2f%%\n", efficiency)
	fmt.Println("===========================================================")
}