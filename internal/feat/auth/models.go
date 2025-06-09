package auth

import (
	"fmt"
	"strconv"

	"github.com/Nidal-Bakir/go-todo-backend/internal/apperr"
	"github.com/Nidal-Bakir/go-todo-backend/internal/database"
	"github.com/Nidal-Bakir/go-todo-backend/internal/utils"
	"github.com/Nidal-Bakir/go-todo-backend/internal/utils/emailvalidator"
	phonenumber "github.com/Nidal-Bakir/go-todo-backend/internal/utils/phone_number"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type LoginMethod string

const (
	LoginMethodEmail       LoginMethod = "email"
	LoginMethodPhoneNumber LoginMethod = "phone"
)

func (l LoginMethod) IsUsingEmail() bool {
	return l == LoginMethodEmail
}
func (l LoginMethod) IsUsingPhoneNumber() bool {
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

func (l *LoginMethod) Fold(onEmail func(), onPhone func()) {
	panicFn := func() {
		panic(fmt.Sprintf("Not supported login method %s", l.String()))
	}

	if l == nil {
		panicFn()
		return
	}

	l.FoldOr(
		onEmail,
		onPhone,
		panicFn,
	)
}

func (l *LoginMethod) FoldOr(onEmail func(), onPhone func(), orElse func()) {
	if l == nil {
		orElse()
		return
	}

	switch *l {
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
	Phone       phonenumber.PhoneNumber
	SentOTP     string
	Password    string
}

func (tu TempUser) ToMap() map[string]string {
	m := make(map[string]string, 8)
	m["id"] = tu.Id.String()
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

func (tu *TempUser) FromMap(m map[string]string) *TempUser {
	tu.Id = uuid.MustParse(m["id"])
	tu.Username = m["username"]
	tu.LoginMethod = LoginMethod(m["login_method"])
	tu.Fname = m["f_name"]
	tu.Lname = m["l_name"]
	tu.Email = m["email"]
	tu.Phone.Number = m["phone_number"]
	tu.Phone.CountryCode = m["phone_country_code"]
	tu.SentOTP = m["sent_otp"]
	tu.Password = m["password"]

	return tu
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
	ok = ok && len(tu.Password) >= PasswordRecommendedLength

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
		UserCreatedAt:    u.UserCreatedAt,
		UserUpdatedAt:    u.UserUpdatedAt,
		UserBlockedAt:    u.UserBlockedAt,
		UserDeletedAt:    u.UserDeletedAt,
		UserBlockedUntil: u.UserBlockedUntil,
		UserRoleID:       u.UserRoleID,

		SessionID:               u.SessionID,
		SessionToken:            u.SessionToken,
		SessionCreatedAt:        u.SessionCreatedAt,
		SessionUpdatedAt:        u.SessionUpdatedAt,
		SessionExpiresAt:        u.SessionExpiresAt,
		SessionDeletedAt:        u.SessionDeletedAt,
		SessionOriginatedFrom:   u.SessionOriginatedFrom,
		SessionUsedInstallation: u.SessionUsedInstallation,
	}
}

type ForgetPasswordTmpDataStore struct {
	Id            uuid.UUID // used as a key
	LoginOptionId int
	UserId        int
	SentOTP       string
}

func (f ForgetPasswordTmpDataStore) ToMap() map[string]string {
	m := make(map[string]string, 8)
	m["id"] = f.Id.String()
	m["login_option_id"] = strconv.Itoa(f.LoginOptionId)
	m["user_id"] = strconv.Itoa(f.UserId)
	m["sent_otp"] = f.SentOTP
	return m
}

func (f *ForgetPasswordTmpDataStore) FromMap(m map[string]string) *ForgetPasswordTmpDataStore {
	f.Id = uuid.MustParse(m["id"])
	f.LoginOptionId = utils.Must(strconv.Atoi(m["login_option_id"]))
	f.UserId = utils.Must(strconv.Atoi(m["user_id"]))
	f.SentOTP = m["sent_otp"]
	return f
}

type PasswordLoginAccessKey struct {
	Phone       phonenumber.PhoneNumber
	Email       string
	LoginMethod LoginMethod
}

func (p PasswordLoginAccessKey) accessKeyStr() string {
	a := ""
	p.LoginMethod.Fold(
		func() { a = p.Email },
		func() { a = p.Phone.ToAppStdForm() },
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
	ID          int32
	Phone       phonenumber.PhoneNumber
	Email       string
	LoginMethod LoginMethod
	IsVerified  bool
}
