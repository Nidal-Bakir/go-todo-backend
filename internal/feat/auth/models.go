package auth

import (
	"fmt"
	"strconv"
	"time"

	"github.com/Nidal-Bakir/go-todo-backend/internal/apperr"
	"github.com/Nidal-Bakir/go-todo-backend/internal/database"
	"github.com/Nidal-Bakir/go-todo-backend/internal/utils"
	"github.com/Nidal-Bakir/go-todo-backend/internal/utils/emailvalidator"
	phonenumber "github.com/Nidal-Bakir/go-todo-backend/internal/utils/phone_number"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type LoginIdentityType string

const (
	LoginIdentityTypeEmail LoginIdentityType = "email"
	LoginIdentityTypePhone LoginIdentityType = "phone"
	LoginIdentityTypeOcid  LoginIdentityType = "ocid"
	LoginIdentityTypeGuest LoginIdentityType = "guest"
)

func (l LoginIdentityType) IsUsingEmail() bool {
	return l == LoginIdentityTypeEmail
}
func (l LoginIdentityType) IsUsingPhoneNumber() bool {
	return l == LoginIdentityTypePhone
}

func (l LoginIdentityType) SupportPassword() bool {
	var supportPassword bool
	l.FoldOr(
		LoginIdentityFoldActions{
			OnEmail: func() { supportPassword = true },
			OnPhone: func() { supportPassword = true },
		},
		func() {},
	)
	return supportPassword
}

func (l LoginIdentityType) String() string {
	return string(l)
}

func (l *LoginIdentityType) FromString(str string) (*LoginIdentityType, error) {
	switch {
	case LoginIdentityTypeEmail.String() == str:
		*l = LoginIdentityTypeEmail

	case LoginIdentityTypePhone.String() == str:
		*l = LoginIdentityTypePhone

	default:
		l = nil
		return l, apperr.ErrUnsupportedLoginIdentityType
	}

	return l, nil
}

func (l *LoginIdentityType) Fold(actions LoginIdentityFoldActions) {
	panicFn := func() {
		panic(fmt.Sprintf("Not supported login identity type %s", l.String()))
	}

	if l == nil {
		panicFn()
		return
	}

	l.FoldOr(actions, panicFn)
}

type LoginIdentityFoldActions struct {
	OnEmail func()
	OnPhone func()
	OnOcid  func()
	OnGuest func()
}

func (l *LoginIdentityType) FoldOr(actions LoginIdentityFoldActions, orElse func()) {
	if l == nil {
		orElse()
		return
	}

	actionOrElse := func(fn func()) func() {
		if fn == nil {
			return orElse
		}
		return fn
	}

	switch *l {
	case LoginIdentityTypeEmail:
		actionOrElse(actions.OnEmail)()

	case LoginIdentityTypePhone:
		actionOrElse(actions.OnPhone)()

	case LoginIdentityTypeOcid:
		actionOrElse(actions.OnOcid)()

	case LoginIdentityTypeGuest:
		actionOrElse(actions.OnGuest)()

	default:
		orElse()
	}
}

type TempPasswordUser struct {
	Id                uuid.UUID // used as a key
	Username          string    // will be the same as Id and will be updated with new value afterword, but initaly this should not be empty
	LoginIdentityType LoginIdentityType
	Fname             string
	Lname             string
	Email             string
	Phone             phonenumber.PhoneNumber
	SentOTP           string
	Password          string
}

func (tu TempPasswordUser) ToMap() map[string]string {
	m := make(map[string]string, 8)
	m["id"] = tu.Id.String()
	m["username"] = tu.Username
	m["login_identity_type"] = tu.LoginIdentityType.String()
	m["f_name"] = tu.Fname
	m["l_name"] = tu.Lname
	m["email"] = tu.Email
	m["phone_number"] = tu.Phone.Number
	m["phone_country_code"] = tu.Phone.CountryCode
	m["sent_otp"] = tu.SentOTP
	m["password"] = tu.Password
	return m
}

func (tu *TempPasswordUser) FromMap(m map[string]string) *TempPasswordUser {
	tu.Id = uuid.MustParse(m["id"])
	tu.Username = m["username"]
	tu.LoginIdentityType = LoginIdentityType(m["login_identity_type"])
	tu.Fname = m["f_name"]
	tu.Lname = m["l_name"]
	tu.Email = m["email"]
	tu.Phone.Number = m["phone_number"]
	tu.Phone.CountryCode = m["phone_country_code"]
	tu.SentOTP = m["sent_otp"]
	tu.Password = m["password"]

	return tu
}

func (tu TempPasswordUser) ValidateForStore() (ok bool) {
	ok = tu.Username == tu.Id.String()

	tu.LoginIdentityType.FoldOr(
		LoginIdentityFoldActions{
			OnEmail: func() { ok = ok && emailvalidator.IsValidEmail(tu.Email) },
			OnPhone: func() { ok = ok && tu.Phone.IsValid() },
		},
		func() { ok = false },
	)

	ok = ok && len(tu.Fname) != 0
	ok = ok && len(tu.Lname) != 0
	ok = ok && len(tu.Password) >= PasswordRecommendedLength

	return ok
}

type CreatePasswordUserArgs struct {
	Username          string
	Fname             string
	Lname             string
	LoginIdentityType LoginIdentityType
	Email             string
	Phone             string
	HashedPass        string
	PassSalt          string
	ProfileImagePath  string
	RoleID            *int32
	VerifiedAt        time.Time
}

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

type UserAndSession struct {
	UserID           int32              `json:"user_id"`
	UserUsername     string             `json:"user_username"`
	UserProfileImage pgtype.Text        `json:"user_profile_image"`
	UserFirstName    string             `json:"user_first_name"`
	UserMiddleName   pgtype.Text        `json:"user_middle_name"`
	UserLastName     pgtype.Text        `json:"user_last_name"`
	UserCreatedAt    pgtype.Timestamptz `json:"user_created_at"`
	UserUpdatedAt    pgtype.Timestamptz `json:"user_updated_at"`
	UserBlockedAt    pgtype.Timestamptz `json:"user_blocked_at"`
	UserBlockedUntil pgtype.Timestamptz `json:"user_blocked_until"`
	UserDeletedAt    pgtype.Timestamptz `json:"user_deleted_at"`
	UserRoleID       pgtype.Int4        `json:"user_role_id"`

	SessionID               int32              `json:"session_id"`
	SessionToken            string             `json:"session_token"`
	SessionCreatedAt        pgtype.Timestamptz `json:"session_created_at"`
	SessionUpdatedAt        pgtype.Timestamptz `json:"session_updated_at"`
	SessionExpiresAt        pgtype.Timestamptz `json:"session_expires_at"`
	SessionDeletedAt        pgtype.Timestamptz `json:"session_deleted_at"`
	SessionOriginatedFrom   int32              `json:"session_originated_from"`
	SessionUsedInstallation int32              `json:"session_used_installation"`
}

func NewUserAndSessionFromDatabaseUserAndSessionRow(u database.UsersGetUserAndSessionDataBySessionTokenRow) UserAndSession {
	return UserAndSession{
		UserID:           u.UserID,
		UserUsername:     u.UserUsername,
		UserProfileImage: u.UserProfileImage,
		UserFirstName:    u.UserFirstName,
		UserMiddleName:   u.UserMiddleName,
		UserLastName:     u.UserLastName,
		UserBlockedAt:    u.UserBlockedAt,
		UserBlockedUntil: u.UserBlockedUntil,
		UserRoleID:       u.UserRoleID,

		SessionID:               u.SessionID,
		SessionToken:            u.SessionToken,
		SessionOriginatedFrom:   u.SessionOriginatedFrom,
		SessionUsedInstallation: u.SessionUsedInstallation,
	}
}

type ForgetPasswordTmpDataStore struct {
	Id uuid.UUID // used as a key

	UserId  int
	SentOTP string
}

func (f ForgetPasswordTmpDataStore) ToMap() map[string]string {
	m := make(map[string]string, 8)
	m["id"] = f.Id.String()
	m["user_id"] = strconv.Itoa(f.UserId)
	m["sent_otp"] = f.SentOTP
	return m
}

func (f *ForgetPasswordTmpDataStore) FromMap(m map[string]string) *ForgetPasswordTmpDataStore {
	f.Id = uuid.MustParse(m["id"])
	f.UserId = utils.Must(strconv.Atoi(m["user_id"]))
	f.SentOTP = m["sent_otp"]
	return f
}

type PasswordLoginAccessKey struct {
	Phone             phonenumber.PhoneNumber
	Email             string
	LoginIdentityType LoginIdentityType
}

func (p PasswordLoginAccessKey) accessKeyStr() string {
	a := ""
	p.LoginIdentityType.Fold(
		LoginIdentityFoldActions{
			OnEmail: func() { a = p.Email },
			OnPhone: func() { a = p.Phone.ToAppStdForm() },
		},
	)
	return a
}

type CreateInstallationData struct {
	NotificationToken       string // e.g the FCM token
	Locale                  string // e.g: en-US ...
	TimezoneOffsetInMinutes int    // e.g: +180
	DeviceManufacturer      string // e.g: samsung
	DeviceOS                string // e.g: android
	DeviceOSVersion         string // e.g: 14
	AppVersion              string // e.g: 3.1.1
}

type UpdateInstallationData struct {
	NotificationToken       string // e.g the FCM token
	Locale                  string // e.g: en-US ...
	TimezoneOffsetInMinutes int    // e.g: +180
	AppVersion              string // e.g: 3.1.1
}

type PublicLoginOptionForProfile struct {
	ID                int32
	Phone             phonenumber.PhoneNumber
	Email             string
	LoginIdentityType LoginIdentityType
	IsVerified        bool
	OidcProvider      string
}
