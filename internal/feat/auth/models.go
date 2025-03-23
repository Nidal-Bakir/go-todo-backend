package auth

import (
	"fmt"

	"github.com/Nidal-Bakir/go-todo-backend/internal/apperr"
	"github.com/Nidal-Bakir/go-todo-backend/internal/database"
	"github.com/Nidal-Bakir/go-todo-backend/internal/utils"
	"github.com/Nidal-Bakir/go-todo-backend/internal/utils/emailvalidator"
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

type LoginMethod string

const (
	LoginMethodEmail       LoginMethod = "email"
	LoginMethodPhoneNumber LoginMethod = "phone"
)

func (l LoginMethod) isUsingEmail() bool {
	return l == LoginMethodEmail
}
func (l LoginMethod) isUsingPhoneNumber() bool {
	return l == LoginMethodPhoneNumber
}

func (l LoginMethod) SupportPassword() bool {
	var supportPassword bool

	l.Fold(
		func() { supportPassword = true },
		func() { supportPassword = true },
	)

	return supportPassword
}

func (l LoginMethod) String() string {
	return string(l)
}

func (l *LoginMethod) FromString(str string) (*LoginMethod, error) {
	switch {
	case LoginMethodEmail.String() == str:
		*l = LoginMethodEmail

	case LoginMethodPhoneNumber.String() == str:
		*l = LoginMethodPhoneNumber

	default:
		l = nil
		return l, apperr.ErrUnsupportedLoginMethod
	}

	return l, nil
}

func (l LoginMethod) Fold(onEmail func(), onPhone func()) {
	l.FoldOr(
		onEmail,
		onPhone,
		func() {
			panic(fmt.Sprintf("Not supported login method %s", l.String()))
		},
	)
}

func (l LoginMethod) FoldOr(onEmail func(), onPhone func(), orElse func()) {
	switch l {
	case LoginMethodEmail:
		onEmail()
	case LoginMethodPhoneNumber:
		onPhone()
	default:
		orElse()
	}
}

type TempUser struct {
	Id          uuid.UUID // used as a key
	Username    string    // will be the same as Id and will be updated with new value afterword, but initaly this should not be empty
	LoginMethod LoginMethod
	Fname       string
	Lname       string
	Email       string
	Phone       utils.PhoneNumber
	SentOTP     string
	Password    string
}

func (tu TempUser) ToMap() map[string]string {
	m := make(map[string]string, 8)
	m["username"] = tu.Username
	m["login_method"] = tu.LoginMethod.String()
	m["f_name"] = tu.Fname
	m["l_name"] = tu.Lname
	m["email"] = tu.Email
	m["phone_number"] = tu.Phone.Number
	m["phone_country_code"] = tu.Phone.CountryCode
	m["sent_otp"] = tu.SentOTP
	m["password"] = tu.Password
	return m
}

func (tu *TempUser) FromMap(m map[string]string) {
	tu.Username = m["username"]
	tu.LoginMethod = LoginMethod(m["login_method"])
	tu.Fname = m["f_name"]
	tu.Lname = m["l_name"]
	tu.Email = m["email"]
	tu.Phone.Number = m["phone_number"]
	tu.Phone.CountryCode = m["phone_country_code"]
	tu.SentOTP = m["sent_otp"]
	tu.Password = m["password"]
}

func (tu TempUser) ValidateForStore() (ok bool) {
	ok = tu.Username == tu.Id.String()

	tu.LoginMethod.FoldOr(
		func() { ok = ok && emailvalidator.IsValidEmail(tu.Email) },
		func() { ok = ok && tu.Phone.IsValid() },
		func() { ok = false },
	)

	ok = ok && len(tu.Fname) != 0
	ok = ok && len(tu.Lname) != 0
	ok = ok && len(tu.Password) >= 6

	return ok
}

type CreateUserArgs struct {
	Username         string
	Fname            string
	Lname            string
	LoginMethod      LoginMethod
	AccessKey        string
	HashedPass       string
	PassSalt         string
	ProfileImagePath string
	RoleID           int32
}
