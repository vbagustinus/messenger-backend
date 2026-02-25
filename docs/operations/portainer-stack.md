# Portainer Stack Deployment (Backend)

Dokumen ini untuk deploy backend langsung dari Portainer (tanpa `docker compose` manual).

## 1) Siapkan image backend

Stack Portainer ini memakai satu image yang berisi semua binary service backend.

Contoh build + push:

```bash
# dari root repo backend
docker build -f deploy/Dockerfile -t ghcr.io/ziyadbooks/lan-chat-backend:latest .
docker push ghcr.io/ziyadbooks/lan-chat-backend:latest
```

Jika registry private, pastikan endpoint Portainer sudah login ke registry tersebut.

## 2) Upload stack di Portainer

1. Buka `Portainer -> Stacks -> Add stack`.
2. Isi nama stack, contoh: `lan-chat-backend`.
3. Pilih salah satu:
   - **Repository/Git mode (otomatis)**: pakai compose path default `docker-compose.yml` (root repo).
   - **Web editor / Upload mode**: paste file `backend/deploy/docker-compose.yml`.
4. Tambahkan Environment variables (copy dari `backend/deploy/portainer.env.example`):
   - `BACKEND_IMAGE`
   - `ADMIN_JWT_SECRET`
   - `MESSAGING_JWT_SECRET`
   - `NODE_ID` (optional)
5. Klik `Deploy the stack`.

## 3) Port yang diekspos

- `8080` discovery
- `8081` messaging
- `8082` filetransfer
- `8083` presence
- `8084` audit
- `8085` cluster
- `8086` auth
- `8090` admin-api
- `5354/udp` discovery multicast bridge

## 4) Verifikasi cepat

```bash
curl -i http://<host>:8086/health
curl -i http://<host>:8081/health
curl -i http://<host>:8083/health
curl -i http://<host>:8090/health
curl -i "http://<host>:8085/join?node_id=node-2&addr=127.0.0.1:10001"
curl -i -X POST http://<host>:8084/log -H "Content-Type: application/json" -d '{"actor_id":"u1","action":"test.event","target_resource":"/","details":""}'
```

Target minimal: HTTP `200` untuk health endpoint dan `201` untuk `POST /log`.

## 5) Persistensi data

Stack ini menggunakan named volumes:

- `shared_data` -> DB bersama (`platform.db`) untuk auth/admin/messaging/presence
- `files_data` -> file terenkripsi filetransfer
- `audit_data` -> audit log file
- `raft_data` -> data raft cluster

Untuk backup, snapshot volume-volume ini secara berkala.

## 6) Catatan operasional

- Untuk produksi, ganti secret default di env.
- Jika port bentrok di host, mapping port bisa diubah di file stack.
- Untuk upgrade versi:
  1. push image baru (`BACKEND_IMAGE` tag baru),
  2. update env di stack,
  3. redeploy.
