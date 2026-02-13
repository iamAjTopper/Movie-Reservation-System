# 🎬 Movie Reservation System (Backend)

A role-based, transactional movie ticket reservation backend built using **Go (Gin)** and **PostgreSQL**.

This system supports user authentication, movie & showtime management, seat availability tracking, atomic reservations, cancellation logic, and admin-level reporting.

---

## 🚀 Tech Stack

- **Go** (Gin Framework)
- **PostgreSQL**
- **JWT Authentication**
- SQL Transactions
- Role-Based Access Control (RBAC)

---

## 🔐 Authentication & Authorization

- User registration & login
- Password hashing
- JWT-based authentication
- Role-based access control:
  - `USER`
  - `ADMIN`
- Admin promotion endpoint

---

## 🎬 Movie Management (Admin Only)

- Create movie
- Update movie
- Delete movie
- Public movie listing

Each movie includes:
- Title
- Description
- Genre
- Poster URL

---

## 🕒 Showtime Scheduling

- Admin can create showtimes
- Validations:
  - Start time must be in future
  - End time must be after start
  - No overlapping showtimes per movie
- Public showtime listing per movie

---

## 🎟 Seat Management

- Static seat layout
- Unique seat constraints
- Seat availability calculated per showtime
- Availability uses `NOT EXISTS` logic
- Proper isolation between showtimes

---

## 🎫 Reservation System

- Authenticated seat booking
- Multi-seat reservations
- Transaction-based booking
- Atomic seat assignment
- Double-booking protection
- Reservation cancellation
- Prevent booking or cancelling past showtimes

---

## 👤 User Features

- View personal reservations
- Cancel upcoming reservations
- Real-time seat availability updates

---

## 👑 Admin Features

- Promote users to admin
- View aggregated booking reports
- Revenue per showtime
- Seats sold per showtime

---

## 📊 Admin Report Endpoint

Returns:
- Movie title
- Showtime
- Seats sold
- Revenue

Uses SQL aggregation with conditional joins.

---

## 🧠 Design Highlights

- Proper relational schema with foreign keys
- ON DELETE CASCADE for integrity
- Transactional booking for atomic operations
- Time-based business rule enforcement
- Clean route separation
- Middleware-based authorization

---

## 📁 Project Structure

```
.
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   ├── db/
│   ├── models/
│   └── routes/
├── .env
├── go.mod
└── go.sum
```

```
## 🛠 Setup Instructions

1. Clone repository
2. Create a `.env` file with:
DB_HOST=
DB_PORT=
DB_USER=
DB_PASSWORD=
DB_NAME=
JWT_SECRET=
3. Setup PostgreSQL database
4. Run:
```

## 📌 Future Improvements

- DB migration files
- Docker support
- Pagination
- Payment integration
- Rate limiting
- CI/CD pipeline


## 🏁 Status

Core backend system complete with transactional integrity and role-based access.

Built as a portfolio-level backend system demonstrating database design, business logic enforcement, and concurrent-safe booking flows.
