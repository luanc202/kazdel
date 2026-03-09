---
trigger: always_on
---

---
paths: pkg/api/**/*.go
---

# API Endpoint Rules

When working with API endpoints:

- All endpoints must validate input
- All endpoints must have error handling and return an error message to the client
- Include OpenAPI documentation comments above each route handler
- Return proper HTTP status codes (200, 201, 400, 404, 500)