package service

import (
	"context"
	"testing"

	accountmodel "blog-server/internal/account/model"
	"blog-server/internal/authz"
	linkmodel "blog-server/internal/link/model"
	"blog-server/internal/link/repository"
	"blog-server/internal/shared/contextx"
	"blog-server/internal/shared/errx"
)

type deleteRepoStub struct {
	repository.LinkRepo
	ownerID uint
	admin   bool
}

func (r *deleteRepoStub) DeleteMine(_ context.Context, ownerID, _ uint, _ int64) error {
	r.ownerID = ownerID
	return nil
}

func (r *deleteRepoStub) DeleteAdmin(context.Context, uint, int64) error {
	r.admin = true
	return nil
}

type listRepoStub struct {
	repository.LinkRepo
	called bool
}

func (r *listRepoStub) ListAll(context.Context) ([]*linkmodel.Link, error) {
	r.called = true
	return []*linkmodel.Link{{ID: 1}}, nil
}

func TestAdminListRequiresAdmin(t *testing.T) {
	for _, tc := range []struct {
		name    string
		role    accountmodel.UserRole
		allowed bool
	}{
		{name: "reader", role: accountmodel.UserRoleReader},
		{name: "admin", role: accountmodel.UserRoleAdmin, allowed: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repo := &listRepoStub{}
			links, err := NewLinkService(repo, authz.NewAuthorizer()).AdminList(context.Background(), contextx.User{ID: 1, Role: tc.role})
			if tc.allowed {
				if err != nil || !repo.called || len(links) != 1 {
					t.Fatalf("admin list: links=%v called=%v err=%v", links, repo.called, err)
				}
			} else if err == nil || errx.ToAppError(err).Code != errx.CodeForbidden || repo.called {
				t.Fatalf("reader list: called=%v err=%v", repo.called, err)
			}
		})
	}
}

func TestDeleteLinkAuthorization(t *testing.T) {
	for _, tc := range []struct {
		name    string
		actor   contextx.User
		wantErr int
		ownerID uint
		admin   bool
	}{
		{name: "reader deletes own", actor: contextx.User{ID: 7, Role: accountmodel.UserRoleReader}, ownerID: 7},
		{name: "admin deletes", actor: contextx.User{ID: 8, Role: accountmodel.UserRoleAdmin}, admin: true},
		{name: "missing actor", actor: contextx.User{Role: accountmodel.UserRoleReader}, wantErr: errx.CodeUnauthorized},
		{name: "unknown role", actor: contextx.User{ID: 9}, wantErr: errx.CodeForbidden},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repo := &deleteRepoStub{}
			svc := NewLinkService(repo, authz.NewAuthorizer())
			err := svc.Delete(context.Background(), tc.actor, 1, 1)
			if tc.wantErr != 0 {
				if err == nil || errx.ToAppError(err).Code != tc.wantErr {
					t.Fatalf("want code %d, got %v", tc.wantErr, err)
				}
			} else if err != nil {
				t.Fatal(err)
			}
			if repo.ownerID != tc.ownerID || repo.admin != tc.admin {
				t.Fatalf("wrong delete target: owner=%d admin=%v", repo.ownerID, repo.admin)
			}
		})
	}
}
