# Go Event Management API

## Overview
RESTful API backend built with Gin framework, PostgreSQL + GORM, JWT authentication, and ImageKit integration for image hosting. Part of an event management platform with React/TypeScript frontend.

## Developer Commands
- `go run main.go`: Start development server on port 8080
- `go build`: Compile binary
- `go test ./...`: Run tests (if configured)

## Architecture & Code Conventions

### Project Structure
- **main.go**: Server entry point, Gin setup, route definitions, CORS configuration
- **config/db.go**: PostgreSQL connection via GORM, database auto-migration for User, Event, Booking models
- **models/**: Data structures (User, Event, Booking) with GORM relationships
- **controllers/**: HTTP handlers for users, events, bookings (user_controller.go, event_controller.go, booking_controller.go)
- **middlewares/**: Auth middleware (auth_middleware.go) using JWT Bearer tokens

### Database & ORM
- **ORM**: GORM with PostgreSQL driver
- **Database**: PostgreSQL (Neon cloud database)
- **Connection**: Configured via `DATABASE_URL` env var in `.env`
- **Auto-migration**: Runs on startup for User, Event, Booking models with relationships

### Authentication
- **Method**: JWT (Bearer token in Authorization header)
- **JWT Secret**: `JWT_SECRET` env var
- **Flow**: 
  - Login endpoint (`POST /api/auth/login`) returns JWT token
  - Protected routes require `Authorization: Bearer <token>` header
  - Middleware extracts and validates token, sets `userID` in context
  - Token claims use `sub` field for user ID

### Routes Structure
- **Public routes**:
  - `GET /api/events` - List all events
  - `GET /api/events/:id` - Get event details
  - `POST /api/auth/register` - User registration
  - `POST /api/auth/login` - User login (returns JWT)
  
- **Protected routes** (require valid JWT):
  - `GET /api/auth/me` - Get current user info
  - `GET /api/events/user` - List user's events
  - `POST /api/events` - Create event
  - `PUT /api/events/:id` - Update event
  - `DELETE /api/events/:id` - Delete event
  - `POST /api/booking` - Create booking
  - `GET /api/booking/user` - Get user's bookings
  - `DELETE /api/booking/:id` - Delete booking

### CORS Configuration
- **Allowed Origins**: `*` (all origins)
- **Allowed Methods**: GET, POST, PUT, DELETE, PATCH, OPTIONS
- **Allowed Headers**: Origin, Content-Type, Authorization

### External Services
- **Image Hosting**: ImageKit (credentials in `.env`: IMAGEKIT_PUBLIC_KEY, IMAGEKIT_PRIVATE_KEY, IMAGEKIT_URL_ENDPOINT)
- **Database**: Neon PostgreSQL (serverless, connection string in DATABASE_URL)

## Environment Setup
Configuration via `.env` file (not committed):
```env
DATABASE_URL=postgresql://...  # Neon database connection string
JWT_SECRET=your_secret_key     # Secret for JWT signing
IMAGEKIT_PUBLIC_KEY=...        # ImageKit public key
IMAGEKIT_PRIVATE_KEY=...       # ImageKit private key
IMAGEKIT_URL_ENDPOINT=...      # ImageKit URL endpoint
```

## Common Patterns
- **Error Handling**: Controllers return Gin JSON responses with error messages
- **GORM Relationships**: 
  - User has many Events (UserID foreign key in Event)
  - User has many Bookings (UserID foreign key in Booking)
  - Event has many Bookings (EventID foreign key in Booking)
- **Validation**: Gin's binding tags (e.g., `binding:"required"`) on model struct fields
- **User Context**: Protected handlers access current user ID via `c.Get("userID")`

## Frontend Integration
- Frontend expects API at `VITE_API_URI` (default `http://localhost:8080`)
- API returns JSON responses matching frontend's types (EventType, UserType in `src/types/types.ts`)
- Frontend persists JWT token in cookies (7-day expiry) after login
