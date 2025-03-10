package server

import (
	"github.com/Nidal-Bakir/go-todo-backend/internal/feat/auth"
	"github.com/Nidal-Bakir/go-todo-backend/internal/utils/password_hasher"
)

func (s *Server) NewAuthRepository() auth.Repository {
	return auth.NewRepository(
		auth.NewDataSource(s.db, s.rdb),
		s.gatewaysProvider,
		password_hasher.NewPasswordHasher(password_hasher.BcryptPasswordHash),
	)
}
