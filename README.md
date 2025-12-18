# HIMA TI e-Election

API Backend untuk sistem pemilihan elektronik (e-voting) HIMA TI yang dibangun menggunakan Go.

## 📋 Deskripsi

HIMA TI e-Election adalah sistem pemilihan elektronik berbasis REST API yang dirancang untuk memfasilitasi proses pemilihan umum organisasi mahasiswa HIMA TI. Sistem ini menyediakan fitur lengkap untuk manajemen pemilih, kandidat, dan proses voting real-time.

## ✨ Fitur Utama

- **Autentikasi & Autorisasi**
  - Login menggunakan NIM dan password
  - Session-based authentication
  - Role-based access control (Admin & Student)

- **Manajemen User**
  - Registrasi user individual
  - Import user massal via CSV
  - Generate dan kirim password otomatis via WhatsApp (Fonnte API)
  - Update dan hapus data user

- **Manajemen Kandidat**
  - CRUD kandidat lengkap dengan foto
  - Informasi visi dan misi kandidat
  - Upload foto kandidat ke S3-compatible storage

- **Sistem Voting**
  - One person, one vote
  - Real-time vote tracking via WebSocket
  - Vote logging untuk audit
  - Status voting per user

- **File Management**
  - Upload kandidat photo via presigned URL (S3)
  - Download vote logs via presigned URL

## 🛠 Tech Stack

- **Language**: Go 1.24.0
- **Web Framework**: `httprouter` (lightweight HTTP router)
- **Database**: PostgreSQL (via `pgx/v5`)
- **Storage**: AWS S3 / S3-compatible storage
- **Messaging**: Fonnte API (WhatsApp)
- **Validation**: `go-playground/validator/v10`
- **Logging**: `logrus`
- **Real-time**: WebSocket (`gorilla/websocket`)

### Dependencies

```
- github.com/aws/aws-sdk-go-v2 - AWS SDK untuk S3 integration
- github.com/jackc/pgx/v5 - PostgreSQL driver
- github.com/julienschmidt/httprouter - HTTP router
- github.com/go-playground/validator/v10 - Request validation
- github.com/gorilla/websocket - WebSocket untuk real-time updates
- github.com/sirupsen/logrus - Logging
- github.com/google/uuid - Generate UUID
- github.com/joho/godotenv - Environment variable management
- golang.org/x/crypto - Password hashing
```

## 📁 Struktur Proyek

```
.
├── api/                    # API specification (OpenAPI)
│   └── api-spec.json      # Dokumentasi API lengkap
├── cmd/                    # Application entrypoints
│   └── api/
│       └── main.go        # Main application
├── config/                 # Configuration files
│   ├── env_config.go      # Environment configuration
│   ├── logger.go          # Logger setup
│   └── validator.go       # Validator setup
├── controller/             # HTTP handlers
├── database/               # Database connection
│   └── connection.go
├── errors/                 # Custom error handling
├── helper/                 # Helper functions
├── log/                    # Application logs
├── middleware/             # HTTP middlewares
├── model/                  # Data models & DTOs
├── repository/             # Data access layer
├── service/                # Business logic layer
├── .env.example           # Environment variables example
├── go.mod                 # Go module definition
└── README.md              # Project documentation
```

## 🚀 Prerequisites

Sebelum menjalankan aplikasi, pastikan Anda telah menginstall:

- Go 1.24.0 atau lebih baru
- PostgreSQL 12 atau lebih baru
- AWS S3 atau S3-compatible storage (MinIO, DigitalOcean Spaces, dll)
- Akun Fonnte (untuk fitur WhatsApp)

## ⚙️ Instalasi

### 1. Clone Repository

```bash
git clone https://github.com/mhaatha/HIMA-TI-e-Election.git
cd HIMA-TI-e-Election
```

### 2. Install Dependencies

```bash
go mod download
```

### 3. Setup Environment Variables

Copy file `.env.example` ke `.env` dan sesuaikan dengan konfigurasi Anda:

```bash
cp .env.example .env
```

Edit file `.env`:

```env
# Database Configuration
DB_PORT=5432
DB_USER=your_db_user
DB_PASSWORD=your_db_password
DB_NAME=hima_ti_election
DB_HOST=localhost

# S3 Configuration
S3_ACCESS_KEY_ID=your_s3_access_key
S3_SECRET_ACCESS_KEY=your_s3_secret_key
S3_BUCKET=your_bucket_name
S3_URL=https://your-s3-endpoint.com

# Fonnte API (WhatsApp)
FONNTE_API_KEY=your_fonnte_api_key
```

### 4. Setup Database

Buat database PostgreSQL:

```sql
CREATE DATABASE hima_ti_election;
```

Jalankan migrasi database (sesuaikan dengan skema yang diperlukan).

### 5. Run Application

```bash
# Development
go run cmd/api/main.go

# Production (build binary)
go build -o bin/election-api cmd/api/main.go
./bin/election-api
```

Server akan berjalan di `http://localhost:8080` (atau sesuai PORT yang di-set di environment).

## 📚 API Documentation

Dokumentasi API lengkap tersedia dalam format OpenAPI 3.0 di file `api/api-spec.json`.

### Base URL

```
Production: https://api-hima-ti-e-election.sgp.dom.my.id
Local: http://localhost:8080
```

### Main Endpoints

#### Authentication
- `POST /api/auth/login` - Login user
- `POST /api/auth/logout` - Logout user

#### Users
- `POST /api/users` - Create user (Admin only)
- `GET /api/users` - Get all users (Admin only)
- `GET /api/users/current` - Get current user
- `PATCH /api/users/:userId` - Update user (Admin only)
- `DELETE /api/users/:userId` - Delete user (Admin only)
- `POST /api/users/bulk` - Bulk create via CSV (Admin only)
- `POST /api/users/generate-passwords` - Generate & send passwords (Admin only)

#### Candidates
- `POST /api/candidates` - Create candidate (Admin only)
- `GET /api/candidates` - Get all candidates
- `GET /api/candidates/:candidateId` - Get candidate by ID (Admin only)
- `PATCH /api/candidates/:candidateId` - Update candidate (Admin only)
- `DELETE /api/candidates/:candidateId` - Delete candidate (Admin only)

#### Voting
- `POST /api/votes` - Cast vote
- `GET /api/votes/:candidateId` - Get vote count (Admin only)
- `GET /ws/votes` - Real-time vote updates via WebSocket (Admin only)
- `GET /api/user/vote-status` - Check if user has voted

#### File Upload/Download
- `GET /api/upload/candidates/presigned-url` - Get presigned URL for upload (Admin only)
- `GET /api/download/logs/vote` - Download vote logs (Admin only)

### Authentication

API menggunakan session-based authentication. Setelah login, session cookie akan disimpan dan dikirim pada setiap request berikutnya.

```bash
# Example: Login
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "nim": "1234567890",
    "password": "yourpassword"
  }'
```

## 🔒 Security Features

- Password hashing menggunakan bcrypt
- Session-based authentication
- Role-based access control (RBAC)
- Input validation menggunakan validator
- SQL injection prevention (menggunakan parameterized queries)
- CORS middleware
- Request logging untuk audit trail

## 🌐 WebSocket Real-time Updates

Endpoint WebSocket untuk monitoring hasil voting real-time:

```javascript
// Connect to WebSocket
const ws = new WebSocket('wss://api-hima-ti-e-election.sgp.dom.my.id/ws/votes?session=YOUR_SESSION_COOKIE');

// Listen for updates
ws.onmessage = function(event) {
  const voteData = JSON.parse(event.data);
  console.log('Real-time vote update:', voteData);
};
```

## 📝 Import Users via CSV

Format CSV untuk bulk import users:

```csv
nim,full_name,study_program,phone_number,role
1234567890,John Doe,Informatika,081234567890,student
0987654321,Jane Smith,Sistem Informasi,081298765432,student
```

## 🤝 Contributing

1. Fork repository ini
2. Buat branch baru (`git checkout -b feature/AmazingFeature`)
3. Commit perubahan (`git commit -m 'Add some AmazingFeature'`)
4. Push ke branch (`git push origin feature/AmazingFeature`)
5. Buat Pull Request

## 📄 License

Project ini dibuat untuk keperluan internal HIMA TI.

## 👥 Authors

- [@mhaatha](https://github.com/mhaatha)

## 🐛 Bug Reports & Feature Requests

Jika menemukan bug atau ingin request fitur baru, silakan buat issue di repository ini.

## 📞 Support

Untuk pertanyaan lebih lanjut, hubungi tim developer HIMA TI.

---

**Note**: Pastikan untuk tidak meng-commit file `.env` ke repository untuk menjaga keamanan kredensial.
