# Authentication & User Approval System

Sistem autentikasi ChristAPI mendukung dua metode pendaftaran dengan approval workflow admin yang ketat. Setiap user baru harus disetujui admin sebelum dapat mengakses API.

---

## 📋 Daftar Isi

1. [Alur Workflow](#alur-workflow)
2. [Setup Google OAuth](#setup-google-oauth)
3. [API Endpoints](#api-endpoints)
4. [Testing dengan HTML Demo](#testing-dengan-html-demo)
5. [Status & State Management](#status--state-management)
6. [Error Handling](#error-handling)

---

## 🔄 Alur Workflow

### Workflow A: Email & Password (dengan OTP Dummy)

```
User Input Email/Password
        ↓
Backend: Hash password, buat user dengan status "pending_otp"
        ↓
Generate OTP 6-digit (dummy), simpan ke DB, cetak ke console/return ke response
        ↓
User Input OTP
        ↓
Backend: Verifikasi OTP di DB
        ↓
Jika valid: Status user berubah ke "pending_approval", OTP dihapus
        ↓
Admin: Lihat daftar user pending approval
        ↓
Admin: Approve user → Status "approved", is_active = true
        ↓
User: Login dengan email & password → Dapatkan JWT token → Akses API
```

**Endpoint:**
- `POST /api/register` — Daftar + generate OTP
- `POST /api/verify-otp` — Verifikasi OTP
- `GET /api/admin/approvals` — List user pending (Protected)
- `POST /api/admin/approvals/:id/approve` — Approve user (Protected)
- `POST /api/login` — Login dengan email & password

---

### Workflow B: Google OAuth (Real Token Verification)

```
User: Klik tombol "Sign in with Google"
        ↓
Frontend: Google Sign-In library → User login & dapat ID token
        ↓
Frontend: Kirim ID token ke backend
        ↓
Backend: Verifikasi ID token dengan Google API menggunakan google-golang-org/api/idtoken
        ↓
Jika valid: Extract email & google_id (sub claim) dari token
        ↓
Backend: Cek apakah email sudah ada di DB
        ├─ Belum ada: Buat user baru dengan auth_provider="google", status="pending_username"
        └─ Sudah ada: Validasi auth_provider = "google"
        ↓
Jika status = "pending_username": Response minta user input username
        ↓
User: Input username
        ↓
Backend: Validasi username unique, update status ke "pending_approval"
        ↓
Admin: Approve user (sama seperti workflow A)
        ↓
User: Klik Google Sign-In lagi → Backend verifikasi token → Status "approved" → Return JWT token → Akses API
```

**Endpoint:**
- `POST /api/auth/google` — Google OAuth (verify token + register/login)
- `POST /api/auth/google/username` — Submit username (first-time Google users)
- `GET /api/admin/approvals` — List user pending (Protected)
- `POST /api/admin/approvals/:id/approve` — Approve user (Protected)

---

### Admin Approval Flow

Setiap user baru (dari workflow A atau B) masuk ke state `pending_approval`. Admin harus menyetujui sebelum user bisa akses API.

```
Admin Panel:
  1. GET /api/admin/approvals → Lihat list semua user yang pending
  2. Inspect user: ID, Email, Username, Status Approval, Is Active
  3. POST /api/admin/approvals/:id/approve → Approve user
     - Status berubah: pending_approval → approved
     - Is Active berubah: false → true
  4. User sekarang bisa login & akses API
```

Atau jika reject:
```
  POST /api/admin/approvals/:id/reject → Reject user
  - Status berubah: pending_approval → rejected
  - User tidak bisa login (rejected status)
```

---

## 🔐 Setup Google OAuth

### Step 1: Setup di Google Cloud Console

1. Buka [Google Cloud Console](https://console.cloud.google.com)
2. Pilih project Anda (atau buat project baru)
3. Navigate ke **APIs & Services** → **Credentials**
4. Klik **Create Credentials** → **OAuth 2.0 Web Client**
5. Pilih **Web Application**
6. Di **Authorized JavaScript origins**, tambahkan:
   ```
   http://localhost:3000
   https://yourdomain.com (production)
   ```
7. Di **Authorized redirect URIs**, tambahkan:
   ```
   http://localhost:3000/
   https://yourdomain.com/ (production)
   ```
8. Copy **Client ID** dan **Client Secret**

### Step 2: Setup `.env.local` atau `.env`

```env
GOOGLE_CLIENT_ID=YOUR_CLIENT_ID.apps.googleusercontent.com
GOOGLE_CLIENT_SECRET=YOUR_CLIENT_SECRET
```

### Step 3: Backend Auto-Verifies Token

Backend secara otomatis:
1. Menerima `id_token` dari frontend
2. Verifikasi signature token dengan Google API menggunakan `google.golang.org/api/idtoken`
3. Extract email & google_id (subject claim)
4. Register atau login user

**Tidak perlu setup tambahan di backend** — verification sudah otomatis di handler `/api/auth/google`.

---

## 📡 API Endpoints

### Public Endpoints (Tidak perlu token)

#### 1. Register dengan Email & Password
```bash
POST /api/register
Content-Type: application/json

{
  "full_name": "John Doe",
  "email": "john@example.com",
  "password": "securepassword123",
  "phone": "+6281234567890",        # optional
  "address": "Jl. Merdeka 123",     # optional
  "role_id": null,                   # optional
  "site_id": null                    # optional
}

Response (201 Created):
{
  "success": true,
  "message": "User registered. Please verify using OTP sent.",
  "data": {
    "user": {
      "id": 1,
      "email": "john@example.com",
      "approval_status": "pending_otp",
      "is_active": false
    },
    "contact": { ... },
    "otp": "123456"  # Dummy OTP (dev mode)
  }
}
```

#### 2. Verifikasi OTP
```bash
POST /api/verify-otp
Content-Type: application/json

{
  "email": "john@example.com",
  "otp_code": "123456"
}

Response (200 OK):
{
  "success": true,
  "message": "OTP verified successfully. Your account is now pending admin approval."
}
```

#### 3. Google OAuth Login/Register
```bash
POST /api/auth/google
Content-Type: application/json

{
  "id_token": "eyJhbGciOiJSUzI1NiIsImtpZCI6IjEyMyIsInR5cCI6IkpXVCJ9..."
}

Response - First time Google user (200 OK):
{
  "success": true,
  "message": "Google registration successful. Please choose a username.",
  "data": {
    "status": "pending_username",
    "user": {
      "id": 2,
      "email": "user@gmail.com",
      "approval_status": "pending_username",
      "is_active": false
    }
  }
}

Response - Pending approval (200 OK):
{
  "success": true,
  "message": "Google login successful but status is pending/rejected.",
  "data": {
    "status": "pending_approval",
    "user": { ... }
  }
}

Response - Approved (200 OK):
{
  "success": true,
  "message": "Google Login berhasil",
  "data": {
    "user": {
      "id": 2,
      "name": "User Name",
      "email": "user@gmail.com",
      "username": "username_chosen",
      "role": "",
      "points": 0,
      "approval_status": "approved",
      "is_active": true
    },
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  }
}
```

#### 4. Submit Google Username (First-time users)
```bash
POST /api/auth/google/username
Content-Type: application/json

{
  "user_id": 2,
  "username": "johndoe"
}

Response (200 OK):
{
  "success": true,
  "message": "Username updated. Your account is now pending admin approval."
}
```

#### 5. Login dengan Email & Password
```bash
POST /api/login
Content-Type: application/json

{
  "email": "admin@christapi.dev",
  "password": "password",
  "site_id": null  # optional
}

Response (200 OK):
{
  "success": true,
  "message": "Login berhasil",
  "data": {
    "user": {
      "id": 1,
      "name": "Admin User",
      "email": "admin@christapi.dev",
      "username": "admin",
      "role": "admin",
      "points": 100,
      "approval_status": "approved",
      "is_active": true
    },
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  }
}
```

---

### Protected Endpoints (Perlu valid JWT token)

**Header:** `Authorization: Bearer <jwt_token>`

#### 1. Get Profile (Test Protected API)
```bash
GET /api/profile
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

Response (200 OK):
{
  "success": true,
  "message": "you are logged in",
  "data": { ... }
}
```

#### 2. List Pending Approvals (Admin)
```bash
GET /api/admin/approvals
Authorization: Bearer <admin_token>

Response (200 OK):
{
  "success": true,
  "message": "List pending approvals",
  "data": [
    {
      "id": 2,
      "email": "john@example.com",
      "username": "johndoe",
      "approval_status": "pending_approval",
      "is_active": false
    },
    {
      "id": 3,
      "email": "jane@gmail.com",
      "username": "janedoe",
      "approval_status": "pending_approval",
      "is_active": false
    }
  ]
}
```

#### 3. Approve User (Admin)
```bash
POST /api/admin/approvals/:id/approve
Authorization: Bearer <admin_token>

Response (200 OK):
{
  "success": true,
  "message": "User approved successfully."
}
```

#### 4. Reject User (Admin)
```bash
POST /api/admin/approvals/:id/reject
Authorization: Bearer <admin_token>

Response (200 OK):
{
  "success": true,
  "message": "User rejected successfully."
}
```

---

## 🧪 Testing dengan HTML Demo

Frontend testing sudah tersedia di `docs/auth_demo.html` — akses di browser:

```
http://localhost:3000/docs/auth_demo.html
```

### Features

1. **Login Panel** — Test login dengan credential yang sudah approved:
   - Email: `admin@christapi.dev`
   - Password: `password`

2. **Workflow A: Email & Password**
   - Isi form: Nama, Email, Password
   - Klik "1. Register Account"
   - Copy OTP dari response
   - Isi OTP, klik "Verify OTP"

3. **Workflow B: Google OAuth**
   - Klik tombol "Sign in with Google" (real Google Sign-In button)
   - Login dengan akun Google Anda
   - Backend verifikasi token
   - Jika first-time: input username → masuk antrean approval

4. **Admin Panel**
   - Klik "Refresh Antrean Approval"
   - Lihat list user yang pending
   - Klik "Approve" untuk approve user

5. **Test Protected API**
   - Klik "Test Protected API (/api/profile)"
   - Lihat response jika token valid atau ditolak

---

## 📊 Status & State Management

### User Status States

| Status | Deskripsi | Bisa Login? |
|--------|-----------|-----------|
| `pending_otp` | Menunggu verifikasi OTP (Email/Password flow) | ❌ No |
| `pending_username` | Menunggu input username (Google first-time) | ❌ No |
| `pending_approval` | Menunggu approval admin | ❌ No |
| `approved` | Sudah disetujui admin, aktif | ✅ Yes |
| `rejected` | Ditolak oleh admin | ❌ No |

### User Flags

| Flag | Default | Deskripsi |
|------|---------|-----------|
| `is_active` | `false` | User aktif atau tidak. Hanya `true` saat status `approved` |
| `auth_provider` | `credentials` | Metode auth: `credentials` (Email/Password) atau `google` |

---

## ⚠️ Error Handling

### Common Errors & Solutions

#### 1. "Invalid Google token"
```
❌ Google token verification failed
```
**Penyebab:** ID token tidak valid, expired, atau signature tidak sesuai  
**Solusi:** Pastikan token baru (belum expired), dan GOOGLE_CLIENT_ID benar di `.env`

#### 2. "User already exists"
```
❌ Email sudah terdaftar
```
**Penyebab:** Email sudah ada di database  
**Solusi:** Gunakan email berbeda atau login jika sudah punya akun

#### 3. "Invalid or expired OTP"
```
❌ OTP tidak valid atau sudah kedaluwarsa
```
**Penyebab:** OTP salah atau lebih dari 5 menit  
**Solusi:** Request OTP baru dengan register ulang

#### 4. "Your account is inactive or pending approval"
```
❌ 403 Forbidden (Protected API)
```
**Penyebab:** User belum disetujui admin  
**Solusi:** Admin harus approve user dulu

#### 5. "Invalid token" atau "missing token"
```
❌ 401 Unauthorized (Protected API)
```
**Penyebab:** Token tidak dikirim atau tidak valid  
**Solusi:** Pastikan header `Authorization: Bearer <token>` benar

#### 6. "CORS error" di browser
```
TypeError: NetworkError when attempting to fetch resource
```
**Penyebab:** Browser block request karena CORS  
**Solusi:** Pastikan server sudah jalan (CORS middleware sudah active)

#### 7. "Cannot GET /docs/auth_demo.html"
```
❌ 404 Not Found
```
**Penyebab:** Backend tidak serve static files  
**Solusi:** Restart server — sudah ada `app.Static("/docs", "./docs")` di main.go

---

## 🚀 Quick Start

### Development

```bash
# 1. Setup .env.local
export GOOGLE_CLIENT_ID="YOUR_CLIENT_ID"
export GOOGLE_CLIENT_SECRET="YOUR_CLIENT_SECRET"
export JWT_SECRET="your-secret-key"

# 2. Run migrations
docker exec -i postgre-chrisapi psql -U christ_user -d christ_db < migrations/4_auth_overhaul.up.sql

# 3. Start server
go run cmd/server/main.go

# 4. Open browser
# http://localhost:3000/docs/auth_demo.html
```

### Testing Email/Password Flow

```bash
# 1. Register
curl -X POST http://localhost:3000/api/register \
  -H "Content-Type: application/json" \
  -d '{
    "full_name": "Test User",
    "email": "test@example.com",
    "password": "password123"
  }'

# Save OTP dari response

# 2. Verify OTP
curl -X POST http://localhost:3000/api/verify-otp \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "otp_code": "123456"
  }'

# 3. Admin approve
curl -X POST http://localhost:3000/api/admin/approvals/2/approve \
  -H "Authorization: Bearer <admin_token>"

# 4. Login
curl -X POST http://localhost:3000/api/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123"
  }'

# Save token dari response

# 5. Test protected API
curl http://localhost:3000/api/profile \
  -H "Authorization: Bearer <your_token>"
```

---

## 📚 References

- [Google OAuth 2.0 Documentation](https://developers.google.com/identity/protocols/oauth2)
- [JWT Handbook](https://auth0.com/resources/ebooks/jwt-handbook)
- [OWASP Authentication Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Authentication_Cheat_Sheet.html)

---

**Dibuat:** 2026-08-03  
**Last Updated:** 2026-08-03
