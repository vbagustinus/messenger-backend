# Ziyad Messenger Backend API Documentation

Daftar API yang tersedia di backend untuk dikonsumsi oleh client.

Semua service HTTP backend mendukung header `X-Request-ID`: jika client kirim, nilai dipropagasikan; jika tidak, service akan generate otomatis dan mengembalikannya di response.
Access log HTTP service ditulis dalam format JSON line agar langsung kompatibel dengan pipeline observability (ELK/Loki/OpenSearch).
Service non-admin (`auth`, `filetransfer`, `audit`, `cluster`) kini memakai hardening baseline: rate limit per-IP, body size limit, validasi input, dan timeout server HTTP.

Deploy via Portainer:
- Git mode default compose path: `docker-compose.yml` (root fallback untuk kompatibilitas Portainer).
- File sumber backend stack: `backend/deploy/docker-compose.yml`.
- Panduan: `backend/docs/operations/portainer-stack.md`.
- Catatan cepat: `backend/docker-compose.yml` memakai mode `build` (tanpa `BACKEND_IMAGE`).

---

## 🚀 Alur Komunikasi (Slack-Style)

### 1. Obrolan Grup (Group Chat)

1. **Daftarkan Channel**: Gunakan Admin Dashboard atau Admin API (`POST /admin/channels`).
2. **Kelola Anggota**: Tambahkan user ke channel (`POST /admin/channels/{id}/members`).
3. **Kirim/Terima**: Client konek ke WebSocket Messaging Service (`/ws?token=...`). Server hanya akan mengirim pesan grup ke user yang terdaftar sebagai anggota.

### 2. Pesan Langsung (Direct Message)

1. **Langsung Kirim**: Client mengirim pesan via WebSocket dengan `channel_id` berisi **UserID** tujuan.
2. **Otomatisasi**: Backend akan otomatis membuat channel privat `dm:userA:userB` dan mendaftarkan kedua user tersebut.
3. **Privasi**: Pesan DM hanya bisa dilihat dan diterima oleh kedua user tersebut.

### 3. Pengiriman File (File Transfer)

1. **Upload**: Client POST file ke `/upload` (Port 8082). Mendapatkan `file_id` dan `key`.
2. **Notifikasi**: Client mengirim pesan chat biasa dengan `type: 3` (File) dan isi pesan berupa JSON metadata file tersebut.
3. **Download**: Penerima mengambil file via `/download?id={file_id}`.

---

## Detail Service & Port... (lihat di bawah)

## 1. Admin API (Port 8090)

Digunakan khusus untuk Admin Dashboard.

### Cross-cutting

- Semua response Admin API menyertakan header `X-Request-ID` untuk tracing request lintas log.
- Set `ADMIN_ACCESS_LOG_DIR` untuk menyimpan access log JSON per hari (`access-YYYY-MM-DD.log`).

### Authentication & Admin

- `POST /admin/login`: Login admin
- `GET /admin/me`: Ambil profile admin yang login
- `POST /admin/admins`: Buat admin baru

### User Management

- `GET /admin/users`: List semua user
- `POST /admin/users`: Buat user baru
- `PUT /admin/users/{id}`: Edit user
- `DELETE /admin/users/{id}`: Hapus user
- `POST /admin/users/{id}/reset-password`: Reset password user

### Monitoring & System

- `GET /admin/monitoring/overview`: Ringkasan monitoring (network, users, messages, files, system)
- `GET /admin/system/health`: Cek kesehatan sistem
- `GET /admin/system/metrics`: Metrik HTTP per endpoint admin (count, error_count, avg/p95/p99/max latency)
- `GET /admin/cluster/status`: Status cluster node

### Example: Admin login

```bash
curl -s -X POST http://localhost:8090/admin/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin"}'
```

---

## 2. Auth Service (Port 8086)

Digunakan untuk registrasi dan login user umum.

- `POST /login`: Login user (Payload: `username`, `password`)
- `POST /register`: Registrasi user baru (Payload: `username`, `full_name`, `password`, `role`)
- `GET /health`: Cek status service

### Example: Register + login

```bash
curl -s -X POST http://localhost:8086/register \
  -H "Content-Type: application/json" \
  -d '{"username":"alice","full_name":"Alice","password":"password123","role":"user"}'

curl -s -X POST http://localhost:8086/login \
  -H "Content-Type: application/json" \
  -d '{"username":"alice","password":"password123"}'
```

---

## 3. Messaging Service (Port 8081)

Layanan utama untuk pengiriman pesan.

- `GET /ws?token={jwt}`: Koneksi WebSocket untuk real-time chat
- `GET /history?channel_id={id}`: Ambil riwayat pesan dalam channel
- `GET /channels`: List channel yang bisa diakses user (public + membership)
- `GET /channel-members?channel_id={id}`: List member channel (butuh akses)
- `POST /dm`: Buat/ambil DM channel (`{"target_user_id":"..."}`)
- `POST /send`: Kirim pesan via HTTP (Alternatif WebSocket)
- `GET /health`: Cek status service

### Authentication

Semua endpoint utama Messaging membutuhkan JWT user:

- Header: `Authorization: Bearer <token>`
- Atau query: `?token=<token>` (dipakai untuk `/ws`)

### Example: Connect WS + send message via HTTP

```bash
# 1) login dan ambil token
TOKEN="$(curl -s -X POST http://localhost:8086/login \
  -H "Content-Type: application/json" \
  -d '{"username":"alice","password":"password123"}' | jq -r .token)"

# 2) kirim pesan ke channel public 'general'
curl -s -X POST http://localhost:8081/send \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${TOKEN}" \
  -d '{"channel_id":"general","type":1,"content":"SGVsbG8="}'
```

---

## 4. File Transfer Service (Port 8082)

Layanan untuk upload dan download file.

- `POST /upload`: Upload file (Multipart form)
- `GET /download?id={file_id}`: Download file berdasarkan ID
- `GET /health`: Cek status service

### Example: Upload + download

```bash
curl -s -X POST http://localhost:8082/upload \
  -F "file=@./README.md"

# lalu download:
curl -L -o out.enc "http://localhost:8082/download?id=<file_id>"
```

---

## 5. Presence Service (Port 8083)

Layanan untuk status online/offline user.

- `POST /heartbeat`: Update status online user (Payload: `user_id`, `status`)
- `GET /status?user_id={id}`: Cek status terkini user
- `GET /health`: Cek status service

---

## 6. Audit Service (Port 8084)

Layanan pencatatan aktivitas sistem.

- `POST /log`: Mencatat event audit baru

Catatan: saat ini Audit service hanya menerima `POST /log` (belum ada endpoint read).

---

## 7. Cluster Service (Port 8085)

Layanan koordinasi antar node backend.

- `GET /join?node_id=...&addr=...`: Join request (stub)

Catatan: service ini masih stub, belum ada endpoint `GET /status`.
