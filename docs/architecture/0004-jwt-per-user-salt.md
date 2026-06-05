# ADR-0004: Per-User JWT Salt for Immediate Revocation

**Status**: Accepted
**Date**: 2026-06-05

## Context

Most JWT implementations sign tokens with a single static secret. The cost of a leaked secret is enormous: every previously-issued token must be considered valid until expiry. The mitigations are:
- Short expiry (poor UX)
- Token blacklist (extra round-trip per request)
- Refresh-token rotation (complex)

## Decision

We store a **per-user salt** in `users.salt`. The JWT signing input is `JWT_SECRET + user.salt`. The salt is generated with `crypto/rand` (32 bytes) at registration time.

```go
// pseudo
key := hmac(sha256, []byte(JWT_SECRET), []byte(user.Salt))
token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
signed, _ := token.SignedString(key)
```

### Revocation flow

1. Operator (admin or self) calls `POST /auth/revoke` (or `DELETE /users/:id/tokens` for admin)
2. The handler generates a new random salt for the user and updates the DB
3. **All previously-issued tokens become invalid** on the next request, because they were signed with the old (now-stale) salt
4. The user must re-authenticate to get a new token

This is **O(1)** in the number of revoked tokens: no blacklist to scan, no Redis lookup, no DB query on the hot path.

### Salt storage

- The salt is stored in `users.salt` (VARCHAR 64, base64 of 32 random bytes)
- The salt is **not** secret by itself — it is a per-user secret. The actual signing key is `HMAC_SHA256(JWT_SECRET, user.salt)`.
- Compromise of the salt alone is not enough to forge tokens; the attacker also needs `JWT_SECRET`.

### Trade-offs

- The user must hit the DB once per login to retrieve the salt. We mitigate this by short-circuiting: if the token signature verifies, the user is authenticated; the salt is only consulted on the login endpoint.
- Changing the salt invalidates ALL sessions for the user (including other devices). This is acceptable for the security win.

## Consequences

### Positive
- **O(1) revocation**: change a column, all tokens for that user die
- **No hot-path DB query**: token verification does not touch the DB
- **No blacklist storage**: no Redis set to maintain
- **Defense in depth**: even if `JWT_SECRET` leaks, the attacker must also know the user's salt to forge a token

### Negative
- **Re-auth on password change**: expected behavior
- **DB hit on login**: we cache the salt in-memory if needed; currently a single query, fast enough

## Security review

- The `alg confusion` attack is mitigated by the JWT library's strict algorithm allow-list (HS256 only, never `none` or `RS256` with a public key)
- The salt is generated with `crypto/rand`, never `math/rand`
- The salt length (32 bytes) gives 256 bits of entropy
- We **never** log the salt or the signing key
