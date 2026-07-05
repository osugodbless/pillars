# PILLARS COOPERATIVE

A cooperative management web app for tracking attendance, dues, fines, events, and member onboarding.

## Current status
The initial implementation includes:
- a simple Go server entrypoint
- a domain store with attendance, dues, fine, event, contribution, and onboarding models
- initial tests for attendance fine creation and member balance calculation

## Run locally
```bash
go test ./...
go run ./cmd/server
```

Then open http://localhost:8080/ and http://localhost:8080/health
