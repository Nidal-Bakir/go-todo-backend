# Go Light Framework

A batteries-included Go backend framework with PostgreSQL, Redis, Docker Compose, SQLC, Goose migrations, JWT auth, OAuth flows, role-based permissions, caching, and advanced rate limiting. 

This project demonstrates a complete production-grade backend architecture with authentication, OAuth, rate limiting, validation, caching, and a fully modular internal design.

---

## ✨ Features

### **Authentication & Authorization**
- Password login
- Google OAuth 2.0 login
- Guest login
- JWT-based authentication (access + refresh tokens)
- Role-based access control (User / Admin)
- CSRF protection
- CORS configuration

### **User & Identity System**
- Account creation via phone or email
- Google libphonenumber for validating phone numbers
- Verification codes (OTP)
- Password reset flows
- Change password (logged-in users)
- Full profile endpoint (`/auth/me`)

### **Installation Tracking**
Used for mobile/web clients:
- Device OS, version
- Locale & timezone
- Push notification token
- Client type (mobile / web)

### **Todo Module**
A small example module demonstrating CRUD operations with:
- Pagination
- Ownership checks
- Status updates

### **Settings API**
Simple key/value storage for internal configuration.

### **Database & Storage**
- PostgreSQL
- Migrations via **Goose**
- Type-safe SQL using **SQLC**
- Seeders for:
  - Roles
  - Admin user
  - Default settings

### **Caching & Rate Limiting**
- Redis caching
- 3 rate limiting modes:
  - Fixed Window
  - Sliding Window
  - Token Bucket

### **Postman Collection**
A comprehensive Postman collection is included:

📁 [go-light-framework.postman_collection.json](./go-light-framework.postman_collection.json)

It covers:
- Auth flows (password, OTP, OAuth)
- Settings API
- Todo CRUD
- Installations
- Profile

---

## 🚀 Getting Started

### Clone the repository
```
git clone https://github.com/Nidal-Bakir/go-light-framework
cd go-light-framework
```

### Create and configure `.env`
Rename:
```
cp .env.example .env
```
Fill DB, Redis, OAuth, and RSA key paths.

---

## 🐳 Running with Docker

Start:
```
make docker-up
```

Rebuild + start:
```
make docker-up-build
```

Development mode:
```
make docker-dev-up-build
```

Shutdown:
```
make docker-down
```

---

## 🛠️ Migrations & Seeders

Run latest migrations:
```
make goose-up
```

Reset + fresh + flush Redis:
```
make fresh-server
```

Generate SQLC code:
```
make sqlc-gen
```

---

## 🧪 Tests
```
make test
```

Linting:
```
make chk
```

---

## 📡 API Endpoints Overview

### **Installation**
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/installation/create` | Register new installation |
| POST | `/installation/update` | Update installation metadata |

---

### **Auth & Identity**
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/auth/create-account` | Create user (phone/email) |
| POST | `/auth/verify-account` | Verify account with OTP |
| POST | `/auth/login` | Login using password |
| POST | `/auth/oauth/google` | Login using Google OAuth |
| POST | `/auth/logout` | Logout |
| POST | `/auth/change-password` | Change password for authenticated users |
| POST | `/auth/forget-password` | Request password reset code |
| POST | `/auth/reset-password` | Reset password |

---

### **Profile**
| Method | Endpoint |
|--------|----------|
| GET | `/auth/me` |

---

### **Todo**
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/todo` | List todos (paginated) |
| GET | `/todo/{id}` | Read todo |
| POST | `/todo` | Create todo |
| PATCH | `/todo/{id}` | Update todo |
| DELETE | `/todo/{id}` | Delete todo |

---

### **Settings**
| Method | Endpoint |
|--------|----------|
| GET | `/settings/{label}` |
| POST | `/settings` |
| DELETE | `/settings/{label}` |

---

## 🧭 Project Structure
```
bin/
cmd/
data/
internal/
  appenv/
  apperr/
  database/
    migrations/
    database_queries/
  feat/
  gateway/
  l10n  
  logger  
  middleware  
  redis_db  
  server  
  tracker  
  utils
l10n/
public/
```
Additional modules:
- Installation tracking
- Settings manager
- Phone number validation service (Google libphonenumber)

---

## 🧱 Build Locally
```
make build
make run
```

---

## 📦 Technologies
- Go
- PostgreSQL
- Redis
- Docker Compose
- Goose
- SQLC
- JWT
- Google OAuth
- libphonenumber

---
