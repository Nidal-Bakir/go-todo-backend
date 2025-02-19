package user

import (
	"github.com/Nidal-Bakir/go-todo-backend/internal/database"
	"github.com/Nidal-Bakir/go-todo-backend/internal/utils"
	"github.com/google/uuid"
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

func NewUserFromDatabaseUser(u database.User) User {
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

type UsernameType string

const (
	UserNameTypeEmail       UsernameType = "email"
	UserNameTypePhoneNumber UsernameType = "phone"
)

func (u UsernameType) isUsingEmail() bool {
	return u == UserNameTypeEmail
}
func (u UsernameType) isUsingPhoneNumber() bool {
	return u == UserNameTypePhoneNumber
}

type TempUser struct {
	Id           uuid.UUID // used as a key
	Username     string    // will be the same as Id and will be updated with new value afterword, but initaly this should not be empty
	UsernameType UsernameType
	Fname        string
	Lname        string
	Email        string
	Phone        utils.PhoneNumber
	SentOTP      string
	Password     string
}

func (tu TempUser) ToMap() map[string]string {
	m := make(map[string]string, 8)
	m["username"] = tu.Username
	m["user_name_type"] = string(tu.UsernameType)
	m["f_name"] = tu.Fname
	m["l_name"] = tu.Lname
	m["email"] = tu.Email
	m["phone_number"] = tu.Phone.Number
	m["phone_counter_code"] = tu.Phone.CounterCode
	m["sent_otp"] = tu.SentOTP
	m["password"] = tu.Password
	return m
}

func (tu *TempUser) FromMap(m map[string]string) {
	tu.Username = m["username"]
	tu.UsernameType = UsernameType(m["user_name_type"])
	tu.Fname = m["f_name"]
	tu.Lname = m["l_name"]
	tu.Email = m["email"]
	tu.Phone.Number = m["phone_number"]
	tu.Phone.CounterCode = m["phone_counter_code"]
	tu.SentOTP = m["sent_otp"]
	tu.Password = m["password"]
}

func (tu TempUser) ValidateForStore() (ok bool) {
	ok = tu.Username == tu.Id.String()

	switch tu.UsernameType {
	case UserNameTypeEmail:
		ok = ok && utils.IsValidEmail(tu.Email)
	case UserNameTypePhoneNumber:
		ok = ok && tu.Phone.IsValid()
	}

	ok = ok && len(tu.Fname) != 0
	ok = ok && len(tu.Lname) != 0
	ok = ok && len(tu.Password) >= 6

	return ok
}
