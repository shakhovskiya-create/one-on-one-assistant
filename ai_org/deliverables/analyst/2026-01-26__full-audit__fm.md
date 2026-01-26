# Full Audit — Functional Model (Draft)

## Scope Model
- Actors: Employee, Manager, Admin, Connector Service, External Systems (AD, EWS, Confluence, GitHub, Camunda), Ops/SRE.
- Core flows: Auth (AD), Employee access, Meetings + AI processing, Tasks, Messenger (WS), Mail/Calendar (EWS), Service Desk, Improvement Requests, Resource Planning.

## Trust Boundaries
- User → nginx (public)
- nginx → frontend/backend (internal network)
- backend → PostgreSQL/Redis/MinIO (internal)
- backend ↔ connector (WS with API key)
- connector → AD/EWS (on-prem)

## State Ownership
- Backend owns domain state and validation.
- DB is passive storage.
- Frontend UI only.

## Data Sensitivity (non-exhaustive)
- AD/EWS credentials, JWT tokens, employee PII, meeting transcripts, mail bodies, files.

## Known Gaps
- RBAC/ABAC model not formalized in code.
- Test coverage absent.
- Input validation inconsistent.
