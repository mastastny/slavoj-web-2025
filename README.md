# Web app for TJ Slavoj Loštice #

## Required environment variables

The application fails to start (`internal/config/config.go`) if these are not set:

| Variable | Description |
|---|---|
| `LINK_SECRET_KEY` | Secret key for signing links |
| `ADMIN_EMAIL` | Administrator email |
| `LOCKER_LOCK_ID` | Lock ID (TTLock) |
| `LOCKER_CLIENT_ID` | Client ID for the TTLock API |
| `LOCKER_CLIENT_SECRET` | Client secret for the TTLock API |
| `LOCKER_REFRESH_TOKEN` | Refresh token for the TTLock API |
| `LOCKER_TOKEN_LIFESPAN` | Lock token lifespan, in days |
| `LOCKER_REFRESH_WINDOW` | Lock token refresh window, in days |
| `SENDGRID_API_KEY` | SendGrid API key |
| `SENDGRID_FROM_EMAIL` | SendGrid sender email |
| `RESEND_API_KEY` | Resend API key |
| `JWT_SECRET_KEY` | Secret key for signing JWTs |


