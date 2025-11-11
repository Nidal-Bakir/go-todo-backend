package server

import (
	"github.com/Nidal-Bakir/go-todo-backend/internal/feat/auth"
	"github.com/Nidal-Bakir/go-todo-backend/internal/feat/perm"
	"github.com/Nidal-Bakir/go-todo-backend/internal/feat/settings"
	"github.com/Nidal-Bakir/go-todo-backend/internal/utils/appjwt"
	"github.com/Nidal-Bakir/go-todo-backend/internal/utils/password_hasher"
)

func (s *Server) NewAuthRepository() auth.Repository {
	return auth.NewRepository(
		auth.NewDataSource(s.db, s.rdb),
		s.gatewaysProvider,
		password_hasher.NewPasswordHasher(password_hasher.BcryptPasswordHash), // changing this value will break the auth system
		auth.NewAuthJWT(appjwt.NewAppJWT()),
	)
}

func (s *Server) NewSettingsRepository() settings.Repository {
	return settings.NewRepository(s.db, s.rdb, s.NewPermRepository())
}

func (s *Server) NewPermRepository() perm.Repository {
	return perm.NewRepository(s.db, s.rdb)
}
