# Expenses Tracker Go

A secure REST API built with Go for tracking financial transactions with JWT-based authentication, email verification, and advanced security features.

## Features

### 🔐 Authentication & Security

- **JWT-based Authentication**: Secure access and refresh tokens
- **Email Verification**: Required before account activation
- **Password Reset**: Secure token-based password reset flow
- **Rate Limiting**: Protection against brute force attacks
- **Account Lockout**: Automatic lockout after failed login attempts
- **Password Security**: bcrypt hashing with complexity requirements

### 💰 Transaction Management

- **User-specific Transactions**: Each user sees only their own transactions
- **Create Transactions**: POST `/transaction` (authenticated)
- **Retrieve Transactions**: GET `/transactions` (authenticated)
- **Database Integration**: PostgreSQL with proper indexing

### 📧 Email Integration

- **Resend Integration**: Professional email delivery
- **Email Templates**: Beautiful HTML email templates
- **Verification Emails**: Account activation emails
- **Password Reset Emails**: Secure reset link delivery

## Dependencies

### Core Dependencies

- **Go 1.24+**: Programming language and runtime
- **github.com/golang-jwt/jwt/v5**: JWT token handling
- **github.com/resend/resend-go/v2**: Email service integration
- **golang.org/x/crypto/bcrypt**: Password hashing
- **golang.org/x/time/rate**: Rate limiting
- **github.com/jackc/pgx/v5**: PostgreSQL driver (recommended by Supabase)
- **github.com/joho/godotenv**: Environment variable loading

## Prerequisites

Before running the application, ensure you have:

1. **Go 1.24 or higher** installed on your system

   ```bash
   go version
   ```

2. **PostgreSQL database** running and accessible

   - Install PostgreSQL locally or use a cloud service
   - Create a database for the application

3. **Resend API Key** for email functionality

   - Sign up at [resend.com](https://resend.com)
   - Get your API key from the dashboard

4. **Environment configuration file** (`.env`)
   ```bash
   # Copy the example file
   cp env.example .env
   # Edit with your actual values
   ```

## Environment Variables

Create a `.env` file in the project root with the following variables:

```bash
# Database Configuration
DATABASE_URL=postgres://username:password@localhost:5432/expenses_tracker?sslmode=disable

# JWT Configuration
JWT_SECRET=your-super-secret-jwt-key-change-this-in-production
JWT_ACCESS_EXPIRY=15m
JWT_REFRESH_EXPIRY=168h

# Email Configuration (Resend)
RESEND_API_KEY=re_xxxxxxxxxxxxx
FROM_EMAIL=noreply@yourdomain.com

# Application Configuration
APP_URL=http://localhost:8080
```

## Database Setup

1. **Create the database**:

   ```sql
   CREATE DATABASE expenses_tracker;
   ```

2. **Run the migration**:

   ```bash
   # Connect to your PostgreSQL database and run:
   psql -d expenses_tracker -f migrations/001_auth_tables.sql
   ```

   Or manually execute the SQL from `migrations/001_auth_tables.sql`:

   ```sql
   -- Create users table
   CREATE TABLE users (
       id SERIAL PRIMARY KEY,
       email VARCHAR(255) UNIQUE NOT NULL,
       password_hash VARCHAR(255) NOT NULL,
       email_verified BOOLEAN DEFAULT FALSE,
       verification_token VARCHAR(255),
       verification_token_expires TIMESTAMP,
       reset_token VARCHAR(255),
       reset_token_expires TIMESTAMP,
       failed_login_attempts INTEGER DEFAULT 0,
       locked_until TIMESTAMP,
       created_at TIMESTAMP DEFAULT NOW(),
       updated_at TIMESTAMP DEFAULT NOW()
   );

   -- Create refresh_tokens table
   CREATE TABLE refresh_tokens (
       id SERIAL PRIMARY KEY,
       user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
       token VARCHAR(255) UNIQUE NOT NULL,
       expires_at TIMESTAMP NOT NULL,
       created_at TIMESTAMP DEFAULT NOW()
   );

   -- Add user_id to existing transaction table
   ALTER TABLE transaction ADD COLUMN user_id INTEGER REFERENCES users(id);

   -- Create indexes
   CREATE INDEX idx_users_email ON users(email);
   CREATE INDEX idx_refresh_tokens_token ON refresh_tokens(token);
   CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens(user_id);
   CREATE INDEX idx_transaction_user_id ON transaction(user_id);
   ```

## Running the Application

1. **Install dependencies**:

   ```bash
   go mod download
   ```

2. **Start the server**:

   ```bash
   go run cmd/server/main.go
   ```

   Or build and run:

   ```bash
   go build -o expenses-tracker cmd/server/main.go
   ./expenses-tracker
   ```

3. **Verify the application is running**:

   ```bash
   curl http://localhost:8080/status
   ```

   Expected response:

   ```json
   {
     "status": "live",
     "application": "Go Simple Web Server",
     "message": "Application is live and running!"
   }
   ```

## API Endpoints

### Public Endpoints (No Authentication Required)

#### Health Check

```bash
GET /status
```

#### User Registration

```bash
POST /auth/register
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "SecurePass123!"
}
```

#### User Login

```bash
POST /auth/login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "SecurePass123!"
}
```

#### Email Verification

```bash
GET /auth/verify-email?token=verification_token_here
```

#### Resend Verification Email

```bash
POST /auth/resend-verification
Content-Type: application/json

{
  "email": "user@example.com"
}
```

#### Forgot Password

```bash
POST /auth/forgot-password
Content-Type: application/json

{
  "email": "user@example.com"
}
```

#### Reset Password

```bash
POST /auth/reset-password
Content-Type: application/json

{
  "token": "reset_token_here",
  "new_password": "NewSecurePass123!"
}
```

#### Refresh Token

```bash
POST /auth/refresh
Content-Type: application/json

{
  "refresh_token": "refresh_token_here"
}
```

### Protected Endpoints (Authentication Required)

All protected endpoints require the `Authorization` header:

```
Authorization: Bearer your_access_token_here
```

#### Logout

```bash
POST /auth/logout
Authorization: Bearer your_access_token_here
Content-Type: application/json

{
  "refresh_token": "refresh_token_here"
}
```

#### Get Transactions

```bash
GET /transactions
Authorization: Bearer your_access_token_here
```

#### Create Transaction

```bash
POST /transaction
Authorization: Bearer your_access_token_here
Content-Type: application/json

{
  "amount": 25.50
}
```

## Security Features

### Password Requirements

- Minimum 8 characters
- At least one uppercase letter
- At least one lowercase letter
- At least one number
- At least one special character
- No more than 3 repeated characters in a row
- No common weak passwords

### Rate Limiting

- **Login**: 5 requests per minute per IP
- **Registration**: 3 requests per hour per IP
- **Password Reset**: 3 requests per hour per IP
- **Verification Resend**: 5 requests per hour per IP

### Account Security

- **Account Lockout**: 5 failed login attempts = 15 minute lockout
- **Email Verification**: Required before login
- **Token Expiry**: Access tokens expire in 15 minutes, refresh tokens in 7 days
- **Secure Storage**: All passwords hashed with bcrypt (cost factor 12)

## Example Usage

### Complete Registration and Login Flow

1. **Register a new user**:

   ```bash
   curl -X POST http://localhost:8080/auth/register \
     -H "Content-Type: application/json" \
     -d '{"email": "test@example.com", "password": "SecurePass123!"}'
   ```

2. **Check your email** for the verification link

3. **Verify your email** by clicking the link or using the token:

   ```bash
   curl "http://localhost:8080/auth/verify-email?token=your_verification_token"
   ```

4. **Login**:

   ```bash
   curl -X POST http://localhost:8080/auth/login \
     -H "Content-Type: application/json" \
     -d '{"email": "test@example.com", "password": "SecurePass123!"}'
   ```

5. **Use the access token** to create a transaction:

   ```bash
   curl -X POST http://localhost:8080/transaction \
     -H "Content-Type: application/json" \
     -H "Authorization: Bearer your_access_token_here" \
     -d '{"amount": 25.50}'
   ```

6. **Get your transactions**:
   ```bash
   curl -X GET http://localhost:8080/transactions \
     -H "Authorization: Bearer your_access_token_here"
   ```

## Development

### Project Structure

```
expenses-tracker-go/
├── cmd/
│   └── server/
│       └── main.go              # Application entry point
├── internal/
│   ├── auth/
│   │   ├── jwt.go              # JWT token management
│   │   ├── password.go         # Password hashing and validation
│   │   └── middleware.go       # Authentication middleware
│   ├── email/
│   │   └── resend.go           # Email service integration
│   ├── handlers/
│   │   ├── handlers.go         # Transaction handlers
│   │   ├── auth_handlers.go    # Authentication handlers
│   │   ├── auth_models.go      # Authentication models
│   │   └── models.go           # Transaction models
│   └── ratelimit/
│       └── ratelimit.go        # Rate limiting middleware
├── migrations/
│   └── 001_auth_tables.sql     # Database migration
├── env.example                 # Environment variables template
├── go.mod                      # Go module file
└── README.md                   # This file
```

### Adding New Features

1. Add new handlers in `internal/handlers/`
2. Define new models in `internal/handlers/models.go`
3. Register routes in `cmd/server/main.go`
4. Update database schema as needed
5. Add tests for new functionality

## Testing the Implementation

After implementation, test the complete flow:

1. ✅ Register a new user → Check email for verification link
2. ✅ Attempt login before verification → Should fail
3. ✅ Verify email → Click link or use token
4. ✅ Login → Receive access + refresh tokens
5. ✅ Access protected route with token
6. ✅ Test wrong password 5 times → Account locks
7. ✅ Request password reset → Check email
8. ✅ Reset password with token
9. ✅ Test refresh token endpoint
10. ✅ Logout → Token invalidated

## License

This project is open source. Please check the license file for more details.

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request
