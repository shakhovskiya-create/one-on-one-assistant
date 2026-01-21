# Security Audit Report - EKF Hub
**Date:** 2026-01-21
**Auditor:** Principal Engineer (AI-assisted)
**Status:** CRITICAL ISSUES FOUND

---

## Executive Summary

This audit identified **multiple critical security vulnerabilities** that require immediate remediation before production deployment. The system stores passwords reversibly, exposes credentials in browser storage, and lacks essential security controls.

---

## CRITICAL FINDINGS (Must Fix Immediately)

### 1. Reversible Password Storage (AES instead of bcrypt)
- **Files:** `backend/internal/utils/crypto.go`, `backend/internal/handlers/connector.go:473-489`
- **Issue:** Passwords encrypted with AES-GCM using JWT_SECRET as key
- **Risk:** JWT_SECRET compromise = all passwords decrypted
- **CVSS:** 9.8 (Critical)

### 2. Credentials in Browser sessionStorage
- **File:** `frontend/src/lib/stores/auth.ts:101`
- **Code:** `sessionStorage.setItem('ews_credentials', JSON.stringify({ username, password }))`
- **Risk:** Any XSS vulnerability exposes all user credentials
- **CVSS:** 9.1 (Critical)

### 3. Hardcoded JWT Secret Default
- **File:** `backend/internal/config/config.go:60`
- **Code:** `JWTSecret: getEnv("JWT_SECRET", "change-me-in-production")`
- **Risk:** Forged JWT tokens if env var not set
- **CVSS:** 9.8 (Critical)

### 4. AD TLS Verification Disabled by Default
- **File:** `backend/internal/config/config.go:53`
- **Code:** `ADSkipVerify: getEnv("AD_SKIP_VERIFY", "true") == "true"`
- **Risk:** MITM attack on AD/LDAP connections
- **CVSS:** 8.1 (High)

### 5. Password in URL Query Parameters
- **File:** `backend/internal/handlers/connector.go:134-137`
- **Risk:** Passwords logged in nginx, browser history, proxies
- **CVSS:** 7.5 (High)

### 6. No CSRF Protection
- **File:** All state-changing API endpoints
- **Risk:** Cross-site request forgery attacks
- **CVSS:** 8.0 (High)

### 7. Unencrypted Database Connection
- **File:** `docker-compose.yml:24`
- **Code:** `sslmode=disable`
- **Risk:** Network sniffing of database traffic
- **CVSS:** 7.5 (High)

### 8. Predictable Token Generator
- **File:** `backend/internal/handlers/connector.go:706-713`
- **Issue:** Uses `time.Now().UnixNano()` for ALL characters
- **Risk:** Token prediction/brute force
- **CVSS:** 7.5 (High)

### 9. Debug Information in Auth Responses
- **File:** `backend/internal/handlers/connector.go:514-522`
- **Risk:** User enumeration, information disclosure
- **CVSS:** 5.3 (Medium)

### 10. No Rate Limiting on Authentication
- **File:** `backend/cmd/server/main.go`
- **Risk:** Brute force password attacks
- **CVSS:** 7.5 (High)

---

## HIGH SEVERITY FINDINGS

| # | Issue | Location | CVSS |
|---|-------|----------|------|
| 1 | WebSocket token in URL query | `connector.go:97` | 6.5 |
| 2 | No input validation | All handlers | 7.5 |
| 3 | Weak CSP (`'unsafe-inline'`) | `nginx-ssl.conf:71` | 6.1 |
| 4 | JWT in localStorage | `auth.ts:97` | 6.5 |
| 5 | No HTML sanitization | Frontend components | 6.1 |
| 6 | File upload without validation | `files.go` | 7.5 |
| 7 | Health endpoint info disclosure | `main.go` | 5.3 |
| 8 | No audit logging | Entire application | 5.0 |
| 9 | No automated tests | Entire codebase | N/A |
| 10 | IDOR vulnerabilities | Multiple endpoints | 6.5 |

---

## MEDIUM SEVERITY FINDINGS

- Missing HSTS preload
- No security.txt
- Permissive CORS configuration
- Missing Content-Type validation
- No request signing

---

## REMEDIATION STATUS

### Completed
- [ ] Remove password storage in sessionStorage
- [ ] Remove query params for credentials
- [ ] Change AD_SKIP_VERIFY default to false
- [ ] Panic on missing JWT_SECRET
- [ ] Add rate limiting on authentication
- [ ] Change sslmode to require
- [ ] Remove debug info from responses
- [ ] Fix random token generator
- [ ] Add CSRF middleware

### Pending
- [ ] Replace AES with bcrypt for passwords
- [ ] WebSocket auth via headers
- [ ] Add comprehensive input validation
- [ ] Fix CSP headers
- [ ] Add HTML sanitization
- [ ] Add file upload validation
- [ ] Add audit logging
- [ ] Write security tests

---

## Verdict

**❌ NOT READY FOR PRODUCTION**

The application contains multiple exploitable vulnerabilities that would result in:
- Complete compromise of user credentials
- Unauthorized access to user accounts
- Data breaches

All CRITICAL issues must be resolved before any production deployment.

---

## References

- OWASP Top 10 2021
- CWE-256: Plaintext Storage of a Password
- CWE-319: Cleartext Transmission of Sensitive Information
- CWE-352: Cross-Site Request Forgery
- CWE-798: Use of Hard-coded Credentials
