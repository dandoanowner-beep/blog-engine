package auth

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepository struct {
	db *pgxpool.Pool
}

func NewPostgresRepository(db *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) CreateUser(ctx context.Context, u *User) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO users (id, username, email, password_hash, google_id, role, verified, created_at, updated_at)
         VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		u.ID, u.Username, u.Email, u.PasswordHash, u.GoogleID, u.Role, u.Verified, u.CreatedAt, u.UpdatedAt,
	)
	return err
}

func (r *PostgresRepository) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	u := &User{}
	err := r.db.QueryRow(ctx,
		`SELECT id, username, email, password_hash, google_id, role, verified, login_attempts, locked_until, created_at, updated_at
         FROM users WHERE email=$1`, email,
	).Scan(&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.GoogleID, &u.Role, &u.Verified,
		&u.LoginAttempts, &u.LockedUntil, &u.CreatedAt, &u.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return u, err
}

func (r *PostgresRepository) GetUserByID(ctx context.Context, id uuid.UUID) (*User, error) {
	u := &User{}
	err := r.db.QueryRow(ctx,
		`SELECT id, username, email, password_hash, google_id, role, verified, login_attempts, locked_until, created_at, updated_at
         FROM users WHERE id=$1`, id,
	).Scan(&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.GoogleID, &u.Role, &u.Verified,
		&u.LoginAttempts, &u.LockedUntil, &u.CreatedAt, &u.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return u, err
}

func (r *PostgresRepository) GetUserByGoogleID(ctx context.Context, googleID string) (*User, error) {
	u := &User{}
	err := r.db.QueryRow(ctx,
		`SELECT id, username, email, role, verified, created_at, updated_at FROM users WHERE google_id=$1`, googleID,
	).Scan(&u.ID, &u.Username, &u.Email, &u.Role, &u.Verified, &u.CreatedAt, &u.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return u, err
}

func (r *PostgresRepository) UpdateUser(ctx context.Context, u *User) error {
	_, err := r.db.Exec(ctx,
		`UPDATE users SET username=$2, password_hash=$3, role=$4, verified=$5, updated_at=$6 WHERE id=$1`,
		u.ID, u.Username, u.PasswordHash, u.Role, u.Verified, u.UpdatedAt,
	)
	return err
}

func (r *PostgresRepository) SaveVerificationToken(ctx context.Context, userID uuid.UUID, token string, exp time.Time) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO email_verifications (id, user_id, token, expires_at) VALUES ($1,$2,$3,$4)`,
		uuid.New(), userID, token, exp,
	)
	return err
}

func (r *PostgresRepository) GetVerificationToken(ctx context.Context, token string) (*VerificationToken, error) {
	t := &VerificationToken{}
	err := r.db.QueryRow(ctx,
		`SELECT token, user_id, expires_at, used FROM email_verifications WHERE token=$1`, token,
	).Scan(&t.Token, &t.UserID, &t.ExpiresAt, &t.Used)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return t, err
}

func (r *PostgresRepository) MarkTokenUsed(ctx context.Context, token string) error {
	_, err := r.db.Exec(ctx, `UPDATE email_verifications SET used=true WHERE token=$1`, token)
	return err
}

func (r *PostgresRepository) SavePasswordReset(ctx context.Context, userID uuid.UUID, token string, exp time.Time) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO password_resets (id, user_id, token, expires_at) VALUES ($1,$2,$3,$4)`,
		uuid.New(), userID, token, exp,
	)
	return err
}

func (r *PostgresRepository) GetPasswordReset(ctx context.Context, token string) (*PasswordResetToken, error) {
	t := &PasswordResetToken{}
	err := r.db.QueryRow(ctx,
		`SELECT token, user_id, expires_at, used FROM password_resets WHERE token=$1`, token,
	).Scan(&t.Token, &t.UserID, &t.ExpiresAt, &t.Used)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return t, err
}

func (r *PostgresRepository) IncrementLoginAttempts(ctx context.Context, email string) error {
	_, err := r.db.Exec(ctx, `UPDATE users SET login_attempts=login_attempts+1 WHERE email=$1`, email)
	return err
}

func (r *PostgresRepository) LockAccount(ctx context.Context, email string, until time.Time) error {
	_, err := r.db.Exec(ctx, `UPDATE users SET locked_until=$2 WHERE email=$1`, email, until)
	return err
}

func (r *PostgresRepository) ResetLoginAttempts(ctx context.Context, email string) error {
	_, err := r.db.Exec(ctx, `UPDATE users SET login_attempts=0, locked_until=NULL WHERE email=$1`, email)
	return err
}

func (r *PostgresRepository) BlockUser(ctx context.Context, blockerID, blockedID uuid.UUID) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO blocks (blocker_id, blocked_id) VALUES ($1,$2) ON CONFLICT DO NOTHING`,
		blockerID, blockedID,
	)
	return err
}

func (r *PostgresRepository) UnblockUser(ctx context.Context, blockerID, blockedID uuid.UUID) error {
	_, err := r.db.Exec(ctx, `DELETE FROM blocks WHERE blocker_id=$1 AND blocked_id=$2`, blockerID, blockedID)
	return err
}
