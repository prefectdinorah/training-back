# Healthy Lifestyle API Documentation

## Endpoints

### 1. Register User
**POST** `/api/register`

Create a new user account.

**Request:**
\`\`\`json
{
  "username": "john_doe",
  "email": "john@example.com",
  "password": "securepassword123",
  "first_name": "John",
  "last_name": "Doe",
  "phone": "+1234567890"
}
\`\`\`

**Response (201 Created):**
\`\`\`json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "username": "john_doe",
  "email": "john@example.com",
  "message": "User registered successfully"
}
\`\`\`

**Errors:**
- `400` - Invalid input or missing required fields
- `409` - Username or email already taken

---

### 2. Login
**POST** `/api/login`

Authenticate user and get JWT token.

**Request:**
\`\`\`json
{
  "email": "john@example.com",
  "password": "securepassword123"
}
\`\`\`

**Response (200 OK):**
\`\`\`json
{
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "email": "john@example.com",
  "message": "Login successful"
}
\`\`\`

**Errors:**
- `400` - Missing email or password
- `401` - Invalid credentials

---

### 3. Get Profile
**GET** `/api/profile`

Get authenticated user's profile information.

**Headers:**
\`\`\`
Authorization: Bearer <token>
\`\`\`

**Response (200 OK):**
\`\`\`json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "username": "john_doe",
  "email": "john@example.com",
  "first_name": "John",
  "last_name": "Doe",
  "phone": "+1234567890",
  "trainer_id": null,
  "is_trainer": false
}
\`\`\`

**Errors:**
- `401` - Missing or invalid token
- `404` - User not found

---

### 4. Update Profile
**PUT** `/api/update`

Update user profile information.

**Headers:**
\`\`\`
Authorization: Bearer <token>
Content-Type: application/json
\`\`\`

**Request:**
\`\`\`json
{
  "first_name": "Jonathan",
  "last_name": "Smith",
  "phone": "+9876543210",
  "trainer_id": "550e8400-e29b-41d4-a716-446655440001",
  "is_trainer": false
}
\`\`\`

**Response (200 OK):**
\`\`\`json
{
  "message": "Profile updated successfully"
}
\`\`\`

**Errors:**
- `400` - Invalid request body
- `401` - Missing or invalid token
- `500` - Database error

---

### 5. Delete Account
**DELETE** `/api/delete`

Permanently delete user account.

**Headers:**
\`\`\`
Authorization: Bearer <token>
\`\`\`

**Response (200 OK):**
\`\`\`json
{
  "message": "Account deleted successfully"
}
\`\`\`

**Errors:**
- `401` - Missing or invalid token
- `404` - User not found

---

## Database Schema

### Users Table

\`\`\`sql
CREATE TABLE users (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  username VARCHAR(255) UNIQUE NOT NULL,
  email VARCHAR(255) UNIQUE NOT NULL,
  password_hash VARCHAR(255) NOT NULL,
  first_name VARCHAR(255),
  last_name VARCHAR(255),
  phone VARCHAR(20),
  trainer_id UUID REFERENCES users(id) ON DELETE SET NULL,
  is_trainer BOOLEAN DEFAULT FALSE,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
\`\`\`

---

## Environment Variables

- `DATABASE_URL` - PostgreSQL connection string (provided by Supabase)
- `JWT_SECRET` - Secret key for JWT token signing (set in Vercel)

---

## Authentication

All protected endpoints require JWT token in the `Authorization` header:
\`\`\`
Authorization: Bearer <token>
\`\`\`

Token expires in 24 hours.

---

## CORS

All endpoints have CORS enabled and accept requests from any origin.
\`\`\`
Access-Control-Allow-Origin: *
Access-Control-Allow-Methods: GET, POST, PUT, DELETE, OPTIONS
Access-Control-Allow-Headers: Authorization, Content-Type
