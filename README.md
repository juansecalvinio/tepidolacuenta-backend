# Restaurant Call Backend

Backend API para el sistema de solicitud de cuenta en restaurantes.

## Stack
- **Language:** Go 1.25
- **Framework:** Gin
- **Database:** MongoDB
- **Real-time:** WebSockets
- **Deployment:** Docker + AWS EC2

## Prerequisites
- Go 1.21+
- MongoDB Atlas account (or local MongoDB)
- Git

## Setup Local

### 1. Clone the repository
```bash
git clone https://github.com/yourusername/restaurant-call-backend.git
cd restaurant-call-backend
```

### 2. Create `.env` file
```bash
cp .env.example .env
```

Edit `.env` and add your MongoDB URI and JWT secret:
```
MONGODB_URI=mongodb+srv://username:password@cluster.mongodb.net/restaurant-call
JWT_SECRET=your_secret_key_here
PORT=8080
GIN_MODE=debug
CORS_ALLOWED_ORIGINS=http://localhost:5173
```

### 3. Download dependencies
```bash
go mod download
```

### 4. Run locally
```bash
go run main.go
```

The server will start on `http://localhost:8080`

### 5. Test health endpoint
```bash
curl http://localhost:8080/health
```

## Development

### Run with auto-reload (optional, requires `air`)
```bash
# Install air
go install github.com/cosmtrek/air@latest

# Run with auto-reload
air
```

### Run tests
```bash
go test ./...
```

## Docker

### Build image
```bash
docker build -t restaurant-call-backend:latest .
```

### Run container
```bash
docker run -p 8080:8080 --env-file .env restaurant-call-backend:latest
```

## API Endpoints

### Health Check
- `GET /health` - Server health status

### Authentication
- `POST /api/v1/auth/register` - Register new restaurant
- `POST /api/v1/auth/login` - Login restaurant

### Tables (Protected)
- `POST /api/v1/restaurants/:restaurantId/tables` - Create table
- `GET /api/v1/restaurants/:restaurantId/tables` - Get tables

### Client
- `POST /api/v1/request-account` - Request account (client endpoint)
- `GET /api/v1/ws/:restaurantId` - WebSocket for real-time notifications

## Project Structure
```
├── main.go                 # Entry point
├── go.mod                  # Dependencies
├── Dockerfile             # Container config
├── .env.example          # Example env variables
├── models/               # Data models (WIP)
├── handlers/             # HTTP handlers (WIP)
├── services/             # Business logic (WIP)
├── middleware/           # Custom middleware (WIP)
└── db/                   # Database functions (WIP)
```

## Next Steps
- [ ] Implement MongoDB models
- [ ] Implement authentication handlers
- [ ] Implement WebSocket communication
- [ ] Add request validation
- [ ] Add error handling
- [ ] Add logging

## License
MIT