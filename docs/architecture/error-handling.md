# Error Handling Standard

> **Status**: Active (Enforced)
> **Last updated**: 2026-06-04

This document defines the project's error handling conventions. All code MUST follow these rules. Violations will be flagged in code review.

## Rules

### Rule 1: Domain errors are sentinel values

Domain errors (validation, business rule violations) MUST be declared as sentinel values using `errors.New`:

```go
// GOOD
var ErrInvalidEmail = errors.New("invalid email format")

// BAD - no context, breaks errors.Is
return fmt.Errorf("invalid email")  // ✗

// BAD - inline, can't be compared
return errors.New("invalid email format")  // ✗ for domain
```

### Rule 2: Naming convention

| Pattern | Example | Use for |
|---------|---------|---------|
| `Err<Subject><Predicate>` | `ErrInvalidEmail`, `ErrUserNotFound` | Domain validation, business rules |
| `Err<Resource>NotFound` | `ErrUserNotFound`, `ErrPropertyNotFound` | Lookup failures |
| `Err<Resource>AlreadyExists` | `ErrDuplicateEmail` | Uniqueness violations |
| `Err<Resource>Invalid<Field>` | `ErrInvalidName` | Field-level validation |

### Rule 3: Use `errors.Is` for comparison

Always use `errors.Is` (not `==`) to compare errors. This enables wrapping and composition:

```go
// GOOD
if errors.Is(err, domain.ErrInvalidEmail) { ... }

// BAD
if err == domain.ErrInvalidEmail { ... }
```

### Rule 4: Wrap errors with context

When returning an error from a lower layer, wrap it with context using `%w`:

```go
// GOOD
return fmt.Errorf("failed to save user: %w", err)

// BAD - loses error chain
return fmt.Errorf("failed to save user: %v", err)

// BAD - no context
return err  // acceptable only if no extra context to add
```

### Rule 5: Cross-cutting errors live in `internal/shared/errors`

For errors that span multiple bounded contexts:

| Error | Use for |
|-------|---------|
| `sharedErrors.ErrNotFound` | Repository lookup failed |
| `sharedErrors.ErrInvalidInput` | Generic invalid input |
| `sharedErrors.ErrUnauthorized` | Auth/authorization failure |
| `sharedErrors.ErrConflict` | Concurrency/data conflict |
| `sharedErrors.ErrInternal` | Unexpected internal error |

Bounded context errors are MORE SPECIFIC than shared errors. Prefer `domain.ErrInvalidEmail` over `sharedErrors.ErrInvalidInput` when possible.

### Rule 6: One declaration per line

Use one `var ErrXxx = ...` per line, not grouped in a `var ( ... )` block:

```go
// GOOD
var ErrInvalidEmail = errors.New("invalid email format")
var ErrUserNotFound = errors.New("user not found")

// ACCEPTABLE (for closely related errors)
var (
    ErrInvalidEmail = errors.New("invalid email format")
    ErrInvalidName  = errors.New("invalid name")
)
```

## Examples

### Validation error (sentinel)

```go
// internal/iam/domain/user.go
var ErrInvalidEmail = errors.New("invalid email format")

func NewUser(...) (*User, error) {
    if !isValidEmail(email) {
        return nil, ErrInvalidEmail
    }
    // ...
}
```

### Infrastructure error (wrapped)

```go
// internal/iam/adapter/postgres_repo.go
func (r *UserRepository) Save(ctx context.Context, user *User) error {
    _, err := r.db.Exec(ctx, "INSERT INTO users ...", ...)
    if err != nil {
        return fmt.Errorf("failed to insert user %s: %w", user.ID(), err)
    }
    return nil
}
```

### Handler comparison

```go
// internal/iam/adapter/http/auth_handler.go
if errors.Is(err, domain.ErrInvalidCredentials) {
    return c.JSON(http.StatusUnauthorized, ...)
}
```

## Anti-Patterns to Avoid

### Anti-Pattern 1: String matching

```go
// BAD
if strings.Contains(err.Error(), "not found") { ... }

// GOOD
if errors.Is(err, domain.ErrUserNotFound) { ... }
```

### Anti-Pattern 2: Generic errors in domain

```go
// BAD
return errors.New("invalid")  // too vague

// GOOD
return ErrInvalidEmail
```

### Anti-Pattern 3: Leaking infrastructure errors

```go
// BAD - exposes pgx error to caller
return err

// GOOD
return fmt.Errorf("failed to save: %w", err)
```

### Anti-Pattern 4: Inconsistent naming

```go
// BAD
ErrBadEmail
ErrEmailIsInvalid
ErrWrongEmail

// GOOD (all consistent)
ErrInvalidEmail
```

## Linting

This standard is enforced by the `errorlint` linter (enabled in `.golangci.yml`). Run `make lint` to check.
