# System vs Applied Code — Worked Example

Read when a concrete layering example clarifies a placement decision. Establish project evidence with the [Decision Protocol](../SKILL.md#decision-protocol) and apply the [code placement method](code-placement.md#the-decision-tree) first.

## Worked Example — System vs Applied Code

The principle behind [code-placement.md](code-placement.md). A reviewer should be able to read the domain file and say *"this is just business logic"* without knowing which database sits behind it.

**Mixed concerns — validation, intent, and persistence tangled; untestable without a live DB:**
```go
func RegisterUser(username, password string) error {
    if len(password) < 8 { return errors.New("password too short") }
    db, _ := sql.Open("postgres", "...")
    defer db.Close()
    _, err := db.Exec("INSERT INTO users ...", username, password)
    return err
}
```

This negative example also writes a raw password. Separating layers alone does not fix credential storage.

**Separated — domain is pure and unit-testable in milliseconds; the repository is the only place SQL lives:**
```go
// domain/user — Applied code, framework-free.
type User struct {
    Username     string
    PasswordHash string
}

var ErrPasswordTooShort = errors.New("password must be at least 8 characters")

func ValidatePassword(p string) error {
    if len(p) < 8 { return ErrPasswordTooShort }
    return nil
}

// app/register — ports and orchestration; concrete implementations are injected.
type PasswordHasher interface {
    Hash(ctx context.Context, password string) (string, error)
}

type UserStore interface {
    Save(ctx context.Context, u user.User) error
}

type RegisterUser struct {
    hasher PasswordHasher
    repo   UserStore
}

func (uc *RegisterUser) Execute(ctx context.Context, username, password string) error {
    if err := user.ValidatePassword(password); err != nil {
        return err
    }
    hash, err := uc.hasher.Hash(ctx, password)
    if err != nil {
        return err
    }
    return uc.repo.Save(ctx, user.User{Username: username, PasswordHash: hash})
}

// adapters/postgres — System code.
type UserRepository struct{ db *sql.DB }

func (r *UserRepository) Save(ctx context.Context, u user.User) error {
    _, err := r.db.ExecContext(ctx,
        "INSERT INTO users (username, password_hash) VALUES ($1, $2)",
        u.Username, u.PasswordHash)
    return err
}
```

The password rule is illustrative; follow the project's actual credential policy. The hasher adapter uses a vetted password-hashing implementation and its salt/work-factor configuration, following [OWASP password storage guidance](https://cheatsheetseries.owasp.org/cheatsheets/Password_Storage_Cheat_Sheet.html). The use case must hash before saving and stop on hashing failure; never persist or log the raw password. Test this orchestration with a fake hasher and store, and verify the concrete hashing adapter separately.

Each piece evolves and is tested independently. — *Clean Architecture* ch.20; *Implementing DDD* ch.4.

The examples in this skill use Go and a hexagonal layout because the layer boundaries are visible in the directory tree. The same reasoning applies in Ruby, TypeScript, Python, or Java; translate the layer names to the project's convention rather than importing Go's.
