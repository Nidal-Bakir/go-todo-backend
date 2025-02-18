package user

import (
	"github.com/Nidal-Bakir/go-todo-backend/internal/database"
	"github.com/Nidal-Bakir/go-todo-backend/internal/utils"
	"github.com/jackc/pgx/v5/pgtype"
)

type User struct {
	ID           int32              `json:"id"`
	Username     string             `json:"username"`
	ProfileImage pgtype.Text        `json:"profile_image"`
	FirstName    string             `json:"first_name"`
	MiddleName   pgtype.Text        `json:"middle_name"`
	LastName     pgtype.Text        `json:"last_name"`
	CreatedAt    pgtype.Timestamptz `json:"created_at"`
	UpdatedAt    pgtype.Timestamptz `json:"updated_at"`
	BlockedAt    pgtype.Timestamptz `json:"blocked_at"`
	DeletedAt    pgtype.Timestamptz `json:"deleted_at"`
	RoleID       pgtype.Int4        `json:"role_id"`
}

func NewUserFromDatabaseuser(u database.User) User {
	return User{
		ID:           u.ID,
		Username:     u.Username,
		ProfileImage: u.ProfileImage,
		FirstName:    u.FirstName,
		MiddleName:   u.MiddleName,
		LastName:     u.LastName,
		CreatedAt:    u.CreatedAt,
		UpdatedAt:    u.UpdatedAt,
		BlockedAt:    u.BlockedAt,
		DeletedAt:    u.DeletedAt,
		RoleID:       u.RoleID,
	}
}

type UserNameType string

const (
	UserNameTypeEmail       UserNameType = "email"
	UserNameTypePhoneNumber UserNameType = "phone"
)

type TempUser struct {
	Fname        string
	Lname        string
	UserNameType UserNameType
	phone        utils.PhoneNumber
	email        string
	otp          string
	password     string
}

func (tu TempUser) isUsingEmail() bool {
	return tu.UserNameType == UserNameTypeEmail
}

func (tu TempUser) isUsingPhoneNumber() bool {
	return tu.UserNameType == UserNameTypePhoneNumber
}
