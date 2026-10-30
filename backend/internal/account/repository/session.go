package repository

import (
	"context"
	"time"

	"blog-server/ent"
	"blog-server/ent/accountrefreshtoken"
	"blog-server/ent/accountsession"
	"blog-server/internal/platform/database"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
)

type SessionRepo struct {
	ds *database.DataStore
}

func NewSessionRepo(ds *database.DataStore) *SessionRepo {
	return &SessionRepo{
		ds: ds,
	}
}

func (r *SessionRepo) Create(ctx context.Context, id uuid.UUID, userID uint, hash string, expires time.Time) error {
	return r.ds.WithTx(ctx, func(ctx context.Context) error {
		client := r.ds.Client(ctx)

		if _, err := client.AccountSession.
			Create().
			SetID(id).
			SetUserID(userID).
			SetExpiresAt(expires).
			Save(ctx); err != nil {
			return err
		}

		if _, err := client.AccountRefreshToken.
			Create().
			SetID(hash).
			SetSessionID(id).
			SetExpiresAt(expires).
			Save(ctx); err != nil {
			return err
		}

		return nil
	})
}

// Rotate locks the session so concurrent use cannot consume the same token twice.
func (r *SessionRepo) Rotate(ctx context.Context, oldHash, newHash string, expires time.Time) (uint, uuid.UUID, bool, error) {
	var (
		userID    uint
		sessionID uuid.UUID
		rotated   bool
	)

	err := r.ds.WithTx(ctx, func(ctx context.Context) error {
		client := r.ds.Client(ctx)
		now := time.Now()

		token, err := client.AccountRefreshToken.
			Query().
			Where(
				accountrefreshtoken.IDEQ(oldHash),
				accountrefreshtoken.ExpiresAtGT(now),
			).
			ForUpdate().
			Only(ctx)
		if ent.IsNotFound(err) {
			return nil
		}
		if err != nil {
			return err
		}

		session, err := client.AccountSession.
			Query().
			Where(accountsession.IDEQ(token.SessionID)).
			ForUpdate().
			Only(ctx)
		if ent.IsNotFound(err) {
			return nil
		}
		if err != nil {
			return err
		}

		if session.RevokedAt != nil || !session.ExpiresAt.After(now) {
			return nil
		}

		if token.UsedAt != nil {
			_, err = client.AccountSession.
				UpdateOneID(session.ID).
				SetRevokedAt(now).
				Save(ctx)
			return err
		}

		if _, err = client.AccountRefreshToken.
			UpdateOneID(token.ID).
			SetUsedAt(now).
			Save(ctx); err != nil {
			return err
		}

		if _, err = client.AccountRefreshToken.
			Create().
			SetID(newHash).
			SetSessionID(session.ID).
			SetExpiresAt(expires).
			Save(ctx); err != nil {
			return err
		}

		if _, err = client.AccountSession.
			UpdateOneID(session.ID).
			SetExpiresAt(expires).
			Save(ctx); err != nil {
			return err
		}

		userID = session.UserID
		sessionID = session.ID
		rotated = true
		return nil
	})

	return userID, sessionID, rotated, err
}

// CleanupExpired bounds each deletion so cleanup does not create a large transaction.
func (r *SessionRepo) CleanupExpired(ctx context.Context) error {
	return r.ds.WithTx(ctx, func(ctx context.Context) error {
		client := r.ds.Client(ctx)
		now := time.Now()

		tokenIDs, err := client.AccountRefreshToken.
			Query().
			Where(
				accountrefreshtoken.ExpiresAtLTE(now),
			).
			Order(
				accountrefreshtoken.ByExpiresAt(),
				accountrefreshtoken.ByID(),
			).
			Limit(500).
			ForUpdate(
				sql.WithLockAction(sql.SkipLocked),
			).
			IDs(ctx)
		if err != nil {
			return err
		}

		if len(tokenIDs) > 0 {
			if _, err = client.AccountRefreshToken.
				Delete().
				Where(accountrefreshtoken.IDIn(tokenIDs...)).
				Exec(ctx); err != nil {
				return err
			}
		}

		sessionIDs, err := client.AccountSession.
			Query().
			Where(
				accountsession.Or(
					accountsession.ExpiresAtLTE(now),
					accountsession.RevokedAtNotNil(),
				),
				accountsession.Not(accountsession.HasRefreshTokens()),
			).
			Order(
				accountsession.ByExpiresAt(),
				accountsession.ByID(),
			).
			Limit(500).
			ForUpdate(
				sql.WithLockAction(sql.SkipLocked),
			).
			IDs(ctx)
		if err != nil {
			return err
		}

		if len(sessionIDs) == 0 {
			return nil
		}

		_, err = client.AccountSession.
			Delete().
			Where(accountsession.IDIn(sessionIDs...)).
			Exec(ctx)

		return err
	})
}

func (r *SessionRepo) Revoke(ctx context.Context, hash string) error {
	_, err := r.ds.Client(ctx).
		AccountSession.
		Update().
		Where(accountsession.HasRefreshTokensWith(
			accountrefreshtoken.IDEQ(hash),
		)).
		SetRevokedAt(time.Now()).
		Save(ctx)

	return err
}
