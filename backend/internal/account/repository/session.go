package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"blog-server/internal/platform/database"
)

type Sessions struct{ ds *database.DataStore }

func NewSessions(ds *database.DataStore) *Sessions { return &Sessions{ds: ds} }
func (s *Sessions) Create(ctx context.Context, id string, userID uint, hash string, expires time.Time) error {
	tx, err := s.ds.DB().BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if _, err = tx.ExecContext(ctx, "INSERT INTO account_sessions(id,user_id,expires_at) VALUES($1,$2,$3)", id, userID, expires); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, "INSERT INTO account_refresh_tokens(hash,session_id,expires_at) VALUES($1,$2,$3)", hash, id, expires); err != nil {
		return err
	}
	return tx.Commit()
}

// Rotate locks the session so concurrent use cannot consume the same token twice.
func (s *Sessions) Rotate(ctx context.Context, oldHash, newHash string, expires time.Time) (uint, string, bool, error) {
	tx, err := s.ds.DB().BeginTx(ctx, nil)
	if err != nil {
		return 0, "", false, err
	}
	defer func() { _ = tx.Rollback() }()
	var userID uint
	var id string
	var revoked, used sql.NullTime
	var deadline time.Time
	err = tx.QueryRowContext(ctx, `SELECT s.user_id,s.id,s.expires_at,s.revoked_at,t.used_at FROM account_sessions s JOIN account_refresh_tokens t ON t.session_id=s.id WHERE t.hash=$1 AND t.expires_at > now() FOR UPDATE OF s`, oldHash).Scan(&userID, &id, &deadline, &revoked, &used)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, "", false, nil
	}
	if err != nil {
		return 0, "", false, err
	}
	if revoked.Valid || !deadline.After(time.Now()) {
		return 0, "", false, nil
	}
	if used.Valid {
		if _, err = tx.ExecContext(ctx, "UPDATE account_sessions SET revoked_at=now() WHERE id=$1", id); err != nil {
			return 0, "", false, err
		}
		return 0, "", false, tx.Commit()
	}
	// Read token usage again after obtaining the session lock: a concurrent query may have waited.
	if err = tx.QueryRowContext(ctx, "SELECT used_at FROM account_refresh_tokens WHERE hash=$1", oldHash).Scan(&used); err != nil {
		return 0, "", false, err
	}
	if used.Valid {
		if _, err = tx.ExecContext(ctx, "UPDATE account_sessions SET revoked_at=now() WHERE id=$1", id); err != nil {
			return 0, "", false, err
		}
		return 0, "", false, tx.Commit()
	}
	if _, err = tx.ExecContext(ctx, "UPDATE account_refresh_tokens SET used_at=now() WHERE hash=$1", oldHash); err != nil {
		return 0, "", false, err
	}
	if _, err = tx.ExecContext(ctx, "INSERT INTO account_refresh_tokens(hash,session_id,expires_at) VALUES($1,$2,$3)", newHash, id, expires); err != nil {
		return 0, "", false, err
	}
	if _, err = tx.ExecContext(ctx, "UPDATE account_sessions SET expires_at=$2 WHERE id=$1", id, expires); err != nil {
		return 0, "", false, err
	}
	if err = tx.Commit(); err != nil {
		return 0, "", false, err
	}
	return userID, id, true, nil
}

// CleanupExpired bounds each deletion so cleanup does not create a large transaction.
func (s *Sessions) CleanupExpired(ctx context.Context) error {
	if _, err := s.ds.DB().ExecContext(ctx, `DELETE FROM account_refresh_tokens WHERE hash IN (
		SELECT hash FROM account_refresh_tokens WHERE expires_at <= now()
		ORDER BY expires_at,hash LIMIT 500 FOR UPDATE SKIP LOCKED
	)`); err != nil {
		return err
	}
	_, err := s.ds.DB().ExecContext(ctx, `DELETE FROM account_sessions WHERE id IN (
		SELECT s.id FROM account_sessions s
		WHERE (s.expires_at <= now() OR s.revoked_at IS NOT NULL)
		AND NOT EXISTS (SELECT 1 FROM account_refresh_tokens t WHERE t.session_id=s.id)
		ORDER BY s.expires_at,s.id LIMIT 500 FOR UPDATE OF s SKIP LOCKED
	)`)
	return err
}

func (s *Sessions) Revoke(ctx context.Context, hash string) error {
	_, err := s.ds.DB().ExecContext(ctx, `UPDATE account_sessions SET revoked_at=now() WHERE id=(SELECT session_id FROM account_refresh_tokens WHERE hash=$1)`, hash)
	return err
}
