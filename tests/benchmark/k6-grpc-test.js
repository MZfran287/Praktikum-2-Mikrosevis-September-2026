import grpc from 'k6/net/grpc';
import { check, sleep } from 'k6';

const client = new grpc.Client();
client.load(['../../proto'], 'pharmacy.proto'); 

export default function () {
  client.connect('127.0.0.1:50051', {
    plaintext: true,
  });

  const response = client.invoke('medcore.pharmacy.v1.PharmacyService/CheckDrugAvailability', {
    drug_code: 'MED-AMX-500',
    quantity_needed: 1,
  });

  // Cetak respons asli dari server ke terminal
  console.log("ISI RESPON DARI SERVER:", JSON.stringify(response.message));

  check(response, {
    'status is OK': (r) => r && r.status === grpc.StatusOK,
    'has valid stock': (r) => {
      if (!r || r.status !== grpc.StatusOK) return false;
      const msg = r.message;
      // Gunakan camelCase sesuai respons server (isAvailable & currentStock)
      return msg && msg.isAvailable === true && msg.currentStock > 0;
    },
  });

  client.close();
  sleep(1);
}