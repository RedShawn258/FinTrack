# JWT Refresh Token Lifecycle Documentation

## Overview

This document describes the JWT refresh token implementation, including the token lifecycle, security considerations, and API usage.

## Token Types

### Access Token
- **Lifetime**: 15 minutes (configurable via `ACCESS_TOKEN_EXPIRY`)
- **Purpose**: Short-lived token for API authentication
- **Storage**: Client-side (localStorage, memory, etc.)
- **Usage**: Sent in `Authorization: Bearer <token>` header for protected endpoints

### Refresh Token
- **Lifetime**: 7 days (configurable via `REFRESH_TOKEN_EXPIRY`)
- **Purpose**: Long-lived token for obtaining new access tokens
- **Storage**: Client-side (recommended: httpOnly cookie for web apps)
- **Usage**: Sent to `/api/v1/auth/refresh` endpoint to get new tokens

## Token Flow

### 1. Login Flow

```
Client → POST /api/v1/auth/login
         { identifier, password }
         
Server → Validates credentials
       → Generates access token (15min)
       → Generates refresh token (7 days)
       → Stores hashed refresh token in DB
       → Returns both tokens
       
Client ← { accessToken, refreshToken, expiresIn }
```

### 2. Access Token Usage

```
Client → GET /api/v1/protected-endpoint
         Authorization: Bearer <accessToken>
         
Server → Validates access token
       → Extracts user ID
       → Processes request
       
Client ← Response
```

### 3. Token Refresh Flow

```
Client → POST /api/v1/auth/refresh
         { refreshToken }
         
Server → Validates refresh token
       → Checks if token is expired
       → Checks if token has been replaced (reuse detection)
       → Generates new access token
       → Generates new refresh token
       → Marks old refresh token as replaced
       → Stores new refresh token
       → Returns new tokens
       
Client ← { accessToken, refreshToken, expiresIn }
```

### 4. Token Rotation

On each refresh:
- New refresh token is generated
- Old refresh token is marked as `replaced_by` in database
- Old token cannot be reused (security feature)
- Client must use the new refresh token for subsequent refreshes

## Security Features

### 1. Token Hashing
- Refresh tokens are hashed using SHA-256 before storage
- Original token is never stored in database
- Prevents token theft from database compromise

### 2. Token Rotation
- New refresh token issued on each refresh
- Old tokens are marked as replaced
- Prevents token reuse attacks

### 3. Reuse Detection
- If a replaced token is used, the system detects it
- All tokens for that user can be invalidated if needed
- Logs security events for monitoring

### 4. Short-Lived Access Tokens
- Access tokens expire in 15 minutes
- Limits exposure window if token is compromised
- Forces regular refresh token usage

### 5. Separate Secrets (Optional)
- Can use different secrets for access and refresh tokens
- Set `REFRESH_TOKEN_SECRET` environment variable
- If not set, uses `JWT_SECRET` for both

## API Endpoints

### POST /api/v1/auth/login

**Request:**
```json
{
  "identifier": "username or email",
  "password": "password"
}
```

**Response:**
```json
{
  "message": "Login successful",
  "accessToken": "eyJhbGc...",
  "refreshToken": "eyJhbGc...",
  "expiresIn": 900
}
```

### POST /api/v1/auth/refresh

**Request:**
```json
{
  "refreshToken": "eyJhbGc..."
}
```

**Response:**
```json
{
  "accessToken": "eyJhbGc...",
  "refreshToken": "eyJhbGc...",
  "expiresIn": 900
}
```

**Error Responses:**
- `400 Bad Request`: Invalid request format
- `401 Unauthorized`: Invalid or expired refresh token

## Configuration

### Environment Variables

```bash
# JWT Configuration
JWT_SECRET=your-secret-key-here
REFRESH_TOKEN_SECRET=optional-separate-secret  # Optional, defaults to JWT_SECRET

# Token Expiration
ACCESS_TOKEN_EXPIRY=15m      # 15 minutes
REFRESH_TOKEN_EXPIRY=168h    # 7 days (168 hours)
```

### Default Values

- `ACCESS_TOKEN_EXPIRY`: 15 minutes
- `REFRESH_TOKEN_EXPIRY`: 7 days (168 hours)
- `REFRESH_TOKEN_SECRET`: Uses `JWT_SECRET` if not provided

## Database Schema

### RefreshToken Table

```sql
CREATE TABLE refresh_tokens (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    user_id BIGINT UNSIGNED NOT NULL,
    token_hash VARCHAR(255) NOT NULL UNIQUE,
    expires_at DATETIME NOT NULL,
    replaced_by BIGINT UNSIGNED NULL,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    deleted_at DATETIME NULL,
    INDEX idx_user_id (user_id),
    INDEX idx_expires_at (expires_at),
    INDEX idx_replaced_by (replaced_by),
    FOREIGN KEY (user_id) REFERENCES users(id)
);
```

## Client Implementation Recommendations

### Web Applications

1. **Store access token in memory** (not localStorage)
2. **Store refresh token in httpOnly cookie** (if possible)
3. **Implement automatic token refresh** before expiration
4. **Handle 401 errors** by attempting refresh, then redirecting to login

### Mobile Applications

1. **Store tokens in secure storage** (Keychain/Keystore)
2. **Implement token refresh on app resume**
3. **Handle token expiration gracefully**

### Example Client Flow

```javascript
// On login
const response = await login(credentials);
localStorage.setItem('accessToken', response.accessToken);
localStorage.setItem('refreshToken', response.refreshToken);

// On API call
async function apiCall(url, options) {
  let token = localStorage.getItem('accessToken');
  
  // Try request
  let response = await fetch(url, {
    ...options,
    headers: {
      ...options.headers,
      'Authorization': `Bearer ${token}`
    }
  });
  
  // If 401, try refresh
  if (response.status === 401) {
    const newTokens = await refreshToken();
    localStorage.setItem('accessToken', newTokens.accessToken);
    localStorage.setItem('refreshToken', newTokens.refreshToken);
    
    // Retry original request
    response = await fetch(url, {
      ...options,
      headers: {
        ...options.headers,
        'Authorization': `Bearer ${newTokens.accessToken}`
      }
    });
  }
  
  return response;
}
```

## Security Best Practices

1. **Always use HTTPS** in production
2. **Set secure cookie flags** if using cookies (httpOnly, secure, sameSite)
3. **Implement rate limiting** on refresh endpoint
4. **Monitor for suspicious activity** (multiple refresh attempts, token reuse)
5. **Rotate secrets regularly** in production
6. **Implement logout** that invalidates refresh tokens
7. **Clean up expired tokens** periodically

## Troubleshooting

### Token Expired Errors

- **Access token expired**: Use refresh token to get new access token
- **Refresh token expired**: User must login again

### Token Reuse Detection

- If you see "refresh token has been revoked" error, the token was already used
- This is a security feature - old tokens cannot be reused
- User must login again

### Database Cleanup

Expired tokens are not automatically deleted. Consider implementing a cleanup job:

```go
// Run periodically (e.g., daily)
tokenRepo.DeleteExpiredTokens()
```

## Migration Notes

### From Old Token System

If migrating from the old single-token system:

1. Old tokens will stop working after their expiration
2. Users will need to login again to get new token pairs
3. No data migration needed - refresh tokens are new

### Breaking Changes

- Login response now returns `accessToken` and `refreshToken` instead of `token`
- Clients must implement refresh token logic
- Access tokens expire in 15 minutes (was 24 hours)

