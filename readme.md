# WhatsApp Web Multidevice API

A REST API for managing WhatsApp Web Multidevice connections with company and user management.

## Features

- Company and user management
- WhatsApp connection management
- Token-based authentication
- Admin dashboard with master token
- Company-specific dashboard
- User authentication and connection viewing

## Requirements

- Go 1.16 or higher
- PostgreSQL
- WhatsApp Web Multidevice client

## Installation

1. Clone the repository:
```bash
git clone https://github.com/yourusername/go-whatsapp.git
cd go-whatsapp
```

2. Install dependencies:
```bash
go mod download
```

3. Configure environment variables:
```bash
cp .env.example .env
# Edit .env with your configuration
```

4. Run the application:
```bash
go run src/main.go
```

## API Endpoints

### Admin Routes (Requires Master Token)

#### Company Management
- `GET /admin/companies` - List all companies
- `POST /admin/companies` - Create a new company
- `GET /admin/companies/:id` - Get company details
- `PUT /admin/companies/:id` - Update company
- `DELETE /admin/companies/:id` - Delete company

#### User Management
- `GET /admin/users` - List all users
- `POST /admin/users` - Create a new user
- `GET /admin/users/:id` - Get user details
- `PUT /admin/users/:id` - Update user
- `DELETE /admin/users/:id` - Delete user

### Public Routes

#### Authentication
- `POST /login` - Login with company token
- `POST /register` - Register new company

#### WhatsApp Operations
- `GET /api/whatsapp/status` - Get connection status
- `POST /api/whatsapp/login` - Login to WhatsApp
- `POST /api/whatsapp/logout` - Logout from WhatsApp
- `POST /api/whatsapp/send` - Send message
- `GET /api/whatsapp/groups` - List groups
- `GET /api/whatsapp/contacts` - List contacts

## Authentication

### Admin Authentication
Admin routes require a master token in the Authorization header:
```
Authorization: Bearer your-master-token
```

### Company Authentication
Company routes require a company token in the Authorization header:
```
Authorization: Bearer your-company-token
```

## Examples

### Create a Company (Admin)
```bash
curl -X POST http://localhost:3000/admin/companies \
  -H "Authorization: Bearer your-master-token" \
  -H "Content-Type: application/json" \
  -d '{"name": "Example Company"}'
```

### Create a User (Admin)
```bash
curl -X POST http://localhost:3000/admin/users \
  -H "Authorization: Bearer your-master-token" \
  -H "Content-Type: application/json" \
  -d '{
    "company_id": 1,
    "name": "John Doe",
    "email": "john@example.com",
    "password": "secret123"
  }'
```

### Send a Message
```bash
curl -X POST http://localhost:3000/api/whatsapp/send \
  -H "Authorization: Bearer your-company-token" \
  -H "Content-Type: application/json" \
  -d '{
    "phone": "1234567890",
    "message": "Hello, World!"
  }'
```

## License

MIT
