# Expenses Tracker Go

A simple REST API built with Go for tracking financial transactions. This application provides endpoints to create and retrieve transactions stored in a PostgreSQL database.

## Features

- **Health Check**: `/status` endpoint to verify the application is running
- **Transaction Management**:
  - Create new transactions via POST `/transaction`
  - Retrieve all transactions via GET `/transactions`
- **Database Integration**: PostgreSQL database with connection pooling
- **Environment Configuration**: Uses `.env` file for database configuration

## Dependencies

### Core Dependencies

- **Go 1.19+**: Programming language and runtime
- **github.com/lib/pq**: PostgreSQL driver for Go's database/sql package
- **github.com/joho/godotenv**: Load environment variables from `.env` file

### Standard Library Packages Used

- `net/http`: HTTP server and client functionality
- `database/sql`: Generic SQL database interface
- `encoding/json`: JSON encoding and decoding
- `fmt`: Formatted I/O operations
- `log`: Logging functionality
- `os`: Operating system interface (environment variables)
- `time`: Time-related functionality

## Prerequisites

Before running the application, ensure you have:

1. **Go 1.19 or higher** installed on your system

   ```bash
   go version
   ```

2. **PostgreSQL database** running and accessible

   - Install PostgreSQL locally or use a cloud service
   - Create a database for the application

3. **Environment configuration file** (`.env`)
   ```bash
   # Create a .env file in the project root
   DATABASE_URL=postgres://username:password@localhost:5432/database_name?sslmode=disable
   ```

## Running the Application

1. **Set up the database**:

   ```sql
   -- Connect to your PostgreSQL database and run:
   CREATE TABLE transaction (
       id SERIAL PRIMARY KEY,
       amount DECIMAL(10,2) NOT NULL
   );
   ```

2. **Configure environment variables**:

   ```bash
   # Create .env file in the project root
   echo "DATABASE_URL=postgres://username:password@localhost:5432/your_database?sslmode=disable" > .env
   ```

3. **Install dependencies**:

   ```bash
   go mod download
   ```

4. **Start the server**:

   ```bash
   go run cmd/server/main.go
   ```

   Or build and run:

   ```bash
   go build -o expenses-tracker cmd/server/main.go
   ./expenses-tracker
   ```

5. **Verify the application is running**:

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

### Environment Variables

The application uses the following environment variable:

- `DATABASE_URL`: PostgreSQL connection string
  - Format: `postgres://username:password@host:port/database?sslmode=mode`
  - Example: `postgres://user:pass@localhost:5432/expenses?sslmode=disable`

### Server Configuration

The HTTP server is configured with the following timeouts:

- **Read Timeout**: 5 seconds
- **Write Timeout**: 10 seconds
- **Idle Timeout**: 120 seconds
- **Port**: 8080

## Development

### Adding New Features

1. Add new handlers in `internal/handlers/handlers.go`
2. Define new models in `internal/handlers/models.go`
3. Register routes in `cmd/server/main.go`
4. Update database schema as needed

## License

This project is open source. Please check the license file for more details.

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request
