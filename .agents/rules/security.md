---
trigger: always_on
---

# Security Requirements

Apply these rules across the entire codebase:

## Input Validation

- Validate all user input server-side (never trust client-side validation alone)
- Use type-safe validation
- Sanitize data before database queries

## Authentication

- Never store passwords in plain text (use bcrypt)
- Implement rate limiting on auth endpoints
- Use HTTP-only cookies for session tokens

## Secrets Management

- Never commit secrets to version control
- Use environment variables for API keys

## Database

- Use parameterized queries (never string concatenation)
- Implement row-level security where possible
- Encrypt sensitive data at rest