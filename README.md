# Christ API

REST API backend dibangun dengan **Go 1.25** + **Fiber v2**. Fokus: cepat dikembangkan, mudah dibaca, dan siap dikembangkan lebih lanjut.

**Version:** 1.2.0+ | **Status:** Active Development

**Features:**
- 🔐 Authentication (Email/Password + Google Sign-In)
- ✉️ OTP Verification & Email notifications
- 👥 Role Management (Admin, User, etc.)
- 📇 Contact Management
- 📰 News & Articles
- ⭐ Points System
- 🛡️ JWT + Admin Middleware

---

## � Kamu Baru Clone Project? Mulai Di Sini!

```
git clone <repo-url>
cd christ-api
          ↓
    👇 Pilih yang cocok 👇

┌─────────────────────────────────────┐
│ WINDOWS + Punya Docker Desktop?     │
├─────────────────────────────────────┤
│ ✅ Recommend: Run 1 command setup   │
│                                     │
│ powershell -ExecutionPolicy Bypass  │
│   -File .\dalamNamaTuhan.ps1        │
│                                     │
│ ⏱️  Selesai dalam ~15 detik        │
│                                     │
│ 📖 Detail: [SETUP.md](./SETUP.md)  │
└─────────────────────────────────────┘

┌─────────────────────────────────────┐
│ LINUX/macOS/Prefer Local Dev?       │
├─────────────────────────────────────┤
│ 📖 Read: [SETUP.md](./SETUP.md)     │
│                                     │
│ Pilih 2 cara:                       │
│ • Docker (recommended untuk team)   │
│ • Local (direct Go + PostgreSQL)    │
└─────────────────────────────────────┘
```

---

## 🚀 Quick Start

### Windows (1 Command):
```powershell
powershell -ExecutionPolicy Bypass -File .\dalamNamaTuhan.ps1
```
**API ready at:** http://localhost:3001

Kalau container sudah pernah dibuat dan kamu cuma mau menjalankan lagi tanpa build ulang:
```powershell
powershell -ExecutionPolicy Bypass -File .\dalamNamaTuhan.ps1 -NoBuild -NoMigrate
```

Kalau kamu cuma mau apply migration baru ke database existing:
```powershell
powershell -ExecutionPolicy Bypass -File .\dalamNamaTuhan.ps1 -MigrateOnly
```

### Atau baca [SETUP.md](./SETUP.md) untuk:
- ✅ Penjelasan lengkap setiap step
- ✅ Mode full setup dan mode run-only
- ✅ Troubleshooting tips
- ✅ Local setup (non-Docker)
- ✅ Database access commands

Catatan singkat: `.env.local` dipakai untuk `go run`, sedangkan `.env.docker` dipakai Docker Compose.

---

## 📚 Dokumentasi Utama

| File | Untuk Apa |
|------|-----------|
| **[SETUP.md](./SETUP.md)** | 👈 **Baca pertama!** Panduan setup lengkap |
| [QUICKSTART.md](./QUICKSTART.md) | Cheat sheet commands |
| [DOCKER.md](./DOCKER.md) | Docker detail & best practices |
| [CHECKLIST.md](./CHECKLIST.md) | Development workflow |
| [docs/schema.sql](./docs/schema.sql) | Database schema |

---

## 🏗️ Struktur Proyek

```
ChristAPI/
├── cmd/server/           → Entry point (main.go)
├── internal/             → Feature modules
│   ├── auth/            → Authentication (Email, OTP, Google Sign-In)
│   ├── role/            → Role Management (Admin, User, etc.)
│   ├── email/           → Email notifications & OTP
│   ├── contacts/        → Contact management
│   ├── news/            → News & articles
│   ├── points/          → Points/reward system
│   ├── sites/           → Site management
│   └── middleware/      → Admin, JWT, etc.
├── pkg/                 → Reusable packages
│   ├── database/        → DB connection & context
│   ├── jwt/             → JWT utilities
│   └── response/        → Response formatting
├── routes/              → API endpoints registration
├── migrations/          → SQL migrations (auto-run on setup)
├── docs/                → API documentation & schema
├── .githooks/           → Git hooks (pre-commit format check)
├── SETUP.md             → 👈 Start here!
├── QUICKSTART.md        → Common commands
├── DOCKER.md            → Docker guide
├── CHECKLIST.md         → Development workflow
└── docker-compose.yml   → Container orchestration
```

### Prinsip Struktur:
Setiap fitur di `internal/<feature>/` punya 4 file utama:
- **handler.go** — Parse request, return response (thin layer)
- **service.go** — Business logic, no DB queries
- **repository.go** — Database queries (parameterized, context-aware)
- **model.go** — Data structures & constants

---

## 🔐 Authentication & Authorization

**Supported Methods:**
- Email + Password (with OTP verification)
- Google Sign-In (OAuth 2.0)
- Admin user seeding on first setup

**Features:**
- OTP via email with enhanced HTML templates
- JWT-based authentication
- Admin middleware for protected routes
- User approval workflow
- Profile completion flow (post-Google sign-in)
- Automatic contact creation on login

**Key Endpoints:**
```
POST   /api/auth/login              → Email/password login
POST   /api/auth/register           → Register new user
POST   /api/auth/google             → Google OAuth flow
POST   /api/auth/verify-otp         → Verify OTP
POST   /api/auth/logout             → Logout
GET    /api/auth/me                 → Get current user
```

---

## 👥 Role Management

**Built-in Roles:**
- `admin` — Full system access
- `user` — Standard user permissions

**Role Operations:**
```
GET    /api/admin/roles             → List all roles (admin-only)
POST   /api/admin/roles             → Create role (admin-only)
PUT    /api/admin/roles/:id         → Update role (admin-only)
DELETE /api/admin/roles/:id         → Delete role (admin-only)
```

Each role has:
- `id` — Numeric ID
- `code` — Unique code (e.g., "admin", "user")
- `name` — Display name
- `description` — Role description
- `created_at`, `updated_at` — Timestamps

---

## ✉️ Email & Notifications

**Email Features:**
- OTP delivery with styled HTML templates
- Parameterized SMTP configuration
- GoMail integration

**Template System:**
- Dynamic content injection
- Professional HTML formatting
- Responsive design

---

## 📊 Additional Features

**Contact Management**
- Create & manage contacts
- Link to user profiles

**News & Articles**
- CRUD operations
- Site-based organization

**Points System**
- User points tracking
- Reward management

---

## ⚙️ Development Workflow

**Adding a New Endpoint:**

1. Create feature folder
```bash
mkdir -p internal/<feature>
```

2. Define data structure in `model.go`
```go
type Model struct {
  ID        int64     `json:"id"`
  UUID      string    `json:"uuid"`
  Name      string    `json:"name"`
  CreatedAt time.Time `json:"created_at"`
  UpdatedAt time.Time `json:"updated_at"`
}
```

3. Define repository interface in `repository.go`
```go
type Repository interface {
  FindByID(ctx context.Context, id int64) (*Model, error)
  List(ctx context.Context, limit, offset int) ([]Model, error)
  Create(ctx context.Context, m *Model) (*Model, error)
}
```

4. Implement business logic in `service.go`
```go
type Service interface {
  Get(ctx context.Context, id int64) (*Model, error)
  Create(ctx context.Context, m *Model) (*Model, error)
}
```

5. Create HTTP handlers in `handler.go`
```go
func (h *Handler) RegisterRoutes(app *fiber.App) {
  g := app.Group("/api/<feature>")
  g.Get("/:id", h.Get)
  g.Post("/", h.Create)
}
```

6. Wire up in `routes/routes.go`
```go
repo := feature.NewRepository(db)
svc := feature.NewService(repo)
handler := feature.NewHandler(svc)
handler.RegisterRoutes(app)
```

7. Add tests

**Core Principles:**
- Handlers: parse request, call service, return response (thin layer)
- Service: business logic only, no DB queries
- Repository: all data access with parameterized queries
- Use `context.Context` in all async operations
- Return errors, never panic

**Commit Messages:**
```
feat(role): add code field to roles and enhance management
fix(auth): handle nil DB connection during Google OAuth
refactor(contacts): improve validation logic
```

---

## 🧪 Testing

**Run all tests:**
```bash
go test ./...
```

**Strategy:**
- **Service tests:** Mock the Repository interface (unit tests)
- **Repository tests:** Use `github.com/DATA-DOG/go-sqlmock` to avoid needing a real database
- **Handler tests:** Use Fiber's app with `httptest` or run the server locally with Postman/curl

**Example requests:**
```bash
# Register
curl -X POST http://localhost:3001/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"securepass123"}'

# Login with OTP
curl -X POST http://localhost:3001/api/auth/verify-otp \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","otp":"123456"}'

# Get current user (requires JWT token in Authorization header)
curl -H "Authorization: Bearer <token>" \
  http://localhost:3001/api/auth/me

# Admin: List roles
curl -H "Authorization: Bearer <admin-token>" \
  http://localhost:3001/api/admin/roles
```

---

## 📦 Migrations

**Automatic setup:**
- Migrations run automatically on `dalamNamaTuhan.ps1` (Windows) or Docker startup
- Located in `migrations/` folder with sequential numbering (0001_, 0002_, etc.)

**Manual migration (if needed):**
```bash
psql -h $DB_HOST -U $DB_USER -d $DB_NAME -f migrations/0001_initial_schema.sql
```

**Adding a new migration:**
1. Create `migrations/NNNN_description.sql`
2. Use parameterized SQL only
3. Test locally with Docker: `docker-compose up --build`

---

## 🔍 Lint & Format

**Format code:**
```bash
gofmt -w .
```

**Vet for issues:**
```bash
go vet ./...
```

**Optional static analysis:**
```bash
staticcheck ./...
```

**Pre-commit hook** (recommended — blocks unformatted commits):
```bash
git config core.hooksPath .githooks
chmod +x .githooks/pre-commit
```

---

## ✅ Best Practices

- **No panics in repositories** — return errors always
- **Parameterized queries only** — prevents SQL injection
- **Dependency injection** — wire repos/services at startup, makes testing easier
- **Secrets in .env** — never commit them
- **Use context.Context** — for timeouts and cancellation in all I/O operations
- **Thin handlers** — parse, call service, return response
- **Service contains logic** — no DB queries in services
- **Repository does data access** — all queries parameterized, context-aware

---

## 📚 Additional Resources

- **[SETUP.md](./SETUP.md)** — Complete setup guide (Docker & local)
- **[QUICKSTART.md](./QUICKSTART.md)** — Common commands cheat sheet
- **[DOCKER.md](./DOCKER.md)** — Docker setup and troubleshooting
- **[CHECKLIST.md](./CHECKLIST.md)** — Pre-commit, CI, migration checklist
- **[docs/schema.sql](./docs/schema.sql)** — Database schema reference

---

**Questions?** Open an issue or check the documentation files above.


