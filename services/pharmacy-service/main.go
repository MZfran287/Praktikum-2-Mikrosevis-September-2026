package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	pb "pharmacy-service/pb"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type server struct {
	pb.UnimplementedPharmacyServiceServer
	mongoClient *mongo.Client
}

type DrugDoc struct {
	DrugCode   string  `bson:"drug_code"`
	Name       string  `bson:"name"`
	UnitPrice  float64 `bson:"unit_price"`
	TotalStock int32   `bson:"total_stock"`
}

func (s *server) CheckDrugAvailability(ctx context.Context, req *pb.CheckDrugRequest) (*pb.CheckDrugResponse, error) {
	log.Printf("[gRPC Server] Memeriksa stok obat: %s, jumlah diminta: %d", req.GetDrugCode(), req.GetQuantityNeeded())

	// === SIMULASI DATABASE BOTTLENECK (Langkah 3.4) ===
	// Server ditahan selama 2 detik untuk mensimulasikan latensi database
	time.Sleep(2 * time.Second)
	// ====================================================

	coll := s.mongoClient.Database("medcore_pharmacy_db").Collection("drugs")
	var drug DrugDoc
	err := coll.FindOne(ctx, bson.M{"drug_code": req.GetDrugCode()}).Decode(&drug)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, status.Errorf(codes.NotFound, "Obat dengan kode %s tidak ditemukan", req.GetDrugCode())
		}
		return nil, status.Errorf(codes.Internal, "Database error: %v", err)
	}

	if drug.TotalStock < req.GetQuantityNeeded() {
		return &pb.CheckDrugResponse{
			DrugCode:     drug.DrugCode,
			IsAvailable:  false,
			CurrentStock: drug.TotalStock,
			UnitPrice:    drug.UnitPrice,
			Message:      fmt.Sprintf("Stok tidak mencukupi. Tersedia: %d", drug.TotalStock),
		}, nil
	}

	return &pb.CheckDrugResponse{
		DrugCode:     drug.DrugCode,
		IsAvailable:  true,
		CurrentStock: drug.TotalStock,
		UnitPrice:    drug.UnitPrice,
		Message:      "Stok obat mencukupi",
	}, nil
}

// startRESTServer menjalankan HTTP/1.1 REST baseline sebagai pembanding gRPC
func startRESTServer(mongoClient *mongo.Client) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/drugs/check", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		drugCode := r.URL.Query().Get("drug_code")
		if drugCode == "" {
			http.Error(w, "Parameter drug_code wajib diisi", http.StatusBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		coll := mongoClient.Database("medcore_pharmacy_db").Collection("drugs")
		var drug DrugDoc
		err := coll.FindOne(ctx, bson.M{"drug_code": drugCode}).Decode(&drug)
		
		w.Header().Set("Content-Type", "application/json")
		if err != nil {
			if err == mongo.ErrNoDocuments {
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(map[string]string{"error": "Obat tidak ditemukan"})
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"drug_code":    drug.DrugCode,
			"is_available": drug.TotalStock >= 10,
			"current_stock": drug.TotalStock,
			"unit_price":   drug.UnitPrice,
			"message":      "Stok obat mencukupi",
		})
	})

	log.Println("MedCore Pharmacy REST Baseline aktif pada port :8081")
	if err := http.ListenAndServe(":8081", mux); err != nil {
		log.Fatalf("Gagal menjalankan REST server: %v", err)
	}
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	mongoURI := "mongodb://adm_pharmacy_svc:SecuredPassPharm2026!@localhost:27017/medcore_pharmacy_db?authSource=admin"
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatalf("Gagal terhubung ke MongoDB: %v", err)
	}

	// Jalankan REST server secara konkuren (background goroutine)
	go startRESTServer(client)

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Gagal membuka port 50051: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterPharmacyServiceServer(s, &server{mongoClient: client})

	log.Println("MedCore Pharmacy gRPC Service aktif pada port :50051")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Gagal menjalankan gRPC server: %v", err)
	}
}