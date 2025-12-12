# FinTrack Backend Structure Summary

## Current Architecture

### Project Structure

```
backend/
├── cmd/
│   └── server/
│       └── main.go              # Application entry point
├── internal/
│   ├── config/
│   │   └── config.go            # Configuration management
│   ├── db/
│   │   ├── db.go                # Database connection & setup
│   │   └── migration.go         # Auto-migrations & seed data
│   ├── handlers/
│   │   ├── auth.go              # Authentication handlers
│   │   ├── budget.go            # Budget CRUD operations
│   │   ├── category.go          # Category management
│   │   ├── forecast.go          # Expense forecasting
│   │   ├── future_features.go   # Gamification & stubs
│   │   └── transaction.go       # Transaction CRUD
│   ├── handlers_test/           # Test files
│   ├── middlewares/
│   │   └── jwt.go               # JWT authentication middleware
│   ├── models/
│   │   ├── budget.go            # Budget model
│   │   ├── category.go          # Category model
│   │   ├── forecast.go          # Forecast models
│   │   ├── gamification.go      # Badge, UserBadge, UserPoints
│   │   ├── transaction.go       # Transaction model
│   │   └── user.go              # User model with profile fields
│   ├── routes/
│   │   └── routes.go            # Route definitions
│   └── utils/
│       └── mailer.go            # Email utility
└── test scripts                 # Various test shell scripts
```

### Technology Stack

- **Framework**: Gin (HTTP web framework)
- **ORM**: GORM (v2)
- **Database**: MySQL
- **Authentication**: JWT (golang-jwt/jwt/v4)
- **Logging**: Zap (uber-go/zap)
- **Password Hashing**: bcrypt (golang.org/x/crypto)
- **Email**: gomail (gopkg.in/gomail.v2)
- **CORS**: gin-contrib/cors

### Key Features Implemented

#### 1. Authentication & Authorization

- User registration with validation
- Login (username/email + password)
- JWT-based authentication
- Password reset flow (forgot/reset)
- Profile management with image upload

#### 2. Core Models

- **User**: Extended profile (firstName, lastName, profileImage, phoneNumber, currency, theme, notifications)
- **Category**: User-specific expense categories
- **Budget**: Category-specific or global budgets with date ranges
- **Transaction**: Expense tracking with category association
- **Gamification**: Badge, UserBadge, UserPoints models

#### 3. API Endpoints

**Public Routes** (`/api/v1/auth`):

- `POST /register` - User registration
- `POST /login` - User login
- `POST /forgot-password` - Request password reset
- `POST /reset-password` - Reset password with token

**Protected Routes** (`/api/v1` - JWT required):

- Profile: `GET`, `PUT /profile`, `POST /profile/image`
- Budgets: `POST`, `GET`, `PUT /budgets/:id`, `DELETE /budgets/:id`
- Categories: `POST`, `GET`, `DELETE /categories/:id`
- Transactions: `POST`, `GET`, `PUT /transactions/:id`, `DELETE /transactions/:id`
- Forecast: `POST /forecast/expenses`
- Future Features: `GET /features/gamification`, `/features/analytics`, `/features/notifications`

#### 4. Database Features

- Auto-migration on startup
- Soft deletes (GORM DeletedAt)
- Connection pooling (10 idle, 100 max open)
- Prepared statement cache
- UTC timezone handling
- Default badge seeding

#### 5. Business Logic

- Budget remaining amount calculation (recalculated on startup)
- Transaction-to-budget association
- Category-based or global budgets
- Expense forecasting
- Gamification system (badges, points, levels)

### Configuration

- Environment-based config (development/production)
- Database connection settings
- JWT secret management
- CORS configuration (localhost:3000, localhost:3001)
- Email settings (SMTP)

---

## Issues & Cleanup Steps

### 🔴 Critical Issues

1. **Missing `go.mod` file**

   - **Impact**: Project cannot be built or dependencies managed
   - **Action**: Initialize Go module with proper module path

2. **Missing `.env` file**
   - **Impact**: Application uses hardcoded defaults
   - **Action**: Create `.env` from `.env.example` template

### 🟡 Code Quality Issues

3. **Duplicate Middleware Setup**

   - **Location**: `main.go` (lines 67-71) and `routes.go` (lines 13-17)
   - **Impact**: Redundant code, logger/jwtSecret set twice per request
   - **Action**: Remove duplicate from `routes.go` (already set in `main.go`)

4. **Hardcoded CORS Origins**

   - **Location**: `main.go` line 59
   - **Impact**: Not configurable, requires code change for new origins
   - **Action**: Move to config/env variables

5. **Error Handling in Logger Initialization**

   - **Location**: `main.go` lines 33-36
   - **Impact**: Errors from `zap.NewProduction/Development` are ignored
   - **Action**: Properly handle logger initialization errors

6. **Missing Error Handling in Config**

   - **Location**: `config/config.go` line 24
   - **Impact**: `godotenv.Load()` error is ignored (duplicate call)
   - **Action**: Remove duplicate or handle properly

7. **Global Database Variable**

   - **Location**: `db/db.go` line 15
   - **Impact**: Makes testing harder, potential race conditions
   - **Action**: Consider dependency injection pattern (lower priority)

8. **Missing Input Validation**
   - **Location**: Various handlers
   - **Impact**: Potential security issues, data integrity
   - **Action**: Review and enhance validation (dates, amounts, etc.)

### 🟢 Code Organization

9. **Test Files Location**

   - **Location**: `handlers_test/` directory
   - **Impact**: Non-standard Go test location (should be `handlers/*_test.go`)
   - **Action**: Consider moving tests next to source files (optional)

10. **Missing Documentation**

    - **Impact**: Hard to understand API contracts
    - **Action**: Add Go doc comments, consider OpenAPI/Swagger

11. **Inconsistent Error Messages**

    - **Impact**: Poor developer experience
    - **Action**: Standardize error response format

12. **Magic Numbers/Strings**
    - **Location**: Various files (e.g., date formats, default values)
    - **Action**: Extract to constants

### 📋 Recommended Cleanup Priority

**Phase 1 (Critical - Do First)**:

1. Initialize `go.mod` file
2. Create `.env` file from example
3. Remove duplicate middleware setup
4. Fix logger error handling

**Phase 2 (Important - Do Next)**: 5. Move CORS origins to config 6. Fix duplicate `godotenv.Load()` calls 7. Add comprehensive input validation 8. Standardize error responses

**Phase 3 (Nice to Have)**: 9. Add API documentation 10. Extract magic values to constants 11. Consider dependency injection refactor 12. Move test files to standard locations

---

## Dependencies (Inferred from Imports)

Based on code analysis, the project likely needs:

- `github.com/gin-gonic/gin`
- `github.com/gin-contrib/cors`
- `github.com/joho/godotenv`
- `go.uber.org/zap`
- `gorm.io/gorm`
- `gorm.io/driver/mysql`
- `github.com/golang-jwt/jwt/v4`
- `golang.org/x/crypto/bcrypt`
- `gopkg.in/gomail.v2`
- `github.com/google/uuid`

---

## Next Steps

1. Run `go mod init` to create module file
2. Run `go mod tidy` to download dependencies
3. Create `.env` file with proper values
4. Fix duplicate middleware
5. Test build and run
