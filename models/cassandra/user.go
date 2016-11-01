package cassandra

import (
	"context"
	"errors"
	"fmt"

	"github.com/Sirupsen/logrus"
	"github.com/gocql/gocql"
	"golang.org/x/crypto/bcrypt"

	"github.com/resourced/resourced-master/contexthelper"
	"github.com/resourced/resourced-master/libstring"
	"github.com/resourced/resourced-master/models/shared"
)

func NewUser(ctx context.Context) *User {
	user := &User{}
	user.AppContext = ctx
	user.table = "users"
	user.hasID = true

	return user
}

type User struct {
	Base
}

func (u *User) GetCassandraSession() (*gocql.Session, error) {
	cassandradbs, err := contexthelper.GetCassandraDBConfig(u.AppContext)
	if err != nil {
		return nil, err
	}

	return cassandradbs.CoreSession, nil
}

// AllUsers returns all user rows.
func (u *User) AllUsers() ([]*shared.UserRow, error) {
	session, err := u.GetCassandraSession()
	if err != nil {
		return nil, err
	}

	users := []*shared.UserRow{}

	query := fmt.Sprintf(`SELECT id, email, password, email_verification_token, email_verified FROM %v`, u.table)

	var scannedID int64
	var scannedEmail, scannedPassword, scannedEmailVerificationToken string
	var scannedEmailVerified bool

	iter := session.Query(query).Iter()
	for iter.Scan(&scannedID, &scannedEmail, &scannedPassword, &scannedEmailVerificationToken, &scannedEmailVerified) {
		users = append(users, &shared.UserRow{
			ID:                     scannedID,
			Email:                  scannedEmail,
			Password:               scannedPassword,
			EmailVerificationToken: scannedEmailVerificationToken,
			EmailVerified:          scannedEmailVerified,
		})
	}
	if err := iter.Close(); err != nil {
		err = fmt.Errorf("%v. Query: %v", err.Error(), query)
		logrus.WithFields(logrus.Fields{"Method": "User.AllUsers"}).Error(err)

		return nil, err
	}
	return users, err
}

// GetByID returns record by id.
func (u *User) GetByID(id int64) (*shared.UserRow, error) {
	session, err := u.GetCassandraSession()
	if err != nil {
		return nil, err
	}

	query := fmt.Sprintf("SELECT id, email, password, email_verification_token, email_verified FROM %v WHERE id=? LIMIT 1", u.table)

	var scannedID int64
	var scannedEmail, scannedPassword, scannedEmailVerificationToken string
	var scannedEmailVerified bool

	err = session.Query(query, id).Scan(&scannedID, &scannedEmail, &scannedPassword, &scannedEmailVerificationToken, &scannedEmailVerified)
	if err != nil {
		return nil, err
	}

	user := &shared.UserRow{
		ID:                     scannedID,
		Email:                  scannedEmail,
		Password:               scannedPassword,
		EmailVerificationToken: scannedEmailVerificationToken,
		EmailVerified:          scannedEmailVerified,
	}

	return user, nil
}

// GetByEmail returns record by email.
func (u *User) GetByEmail(email string) (*shared.UserRow, error) {
	session, err := u.GetCassandraSession()
	if err != nil {
		return nil, err
	}

	query := fmt.Sprintf("SELECT id, email, password, email_verification_token, email_verified FROM %v WHERE email=? LIMIT 1", u.table)

	var scannedID int64
	var scannedEmail, scannedPassword, scannedEmailVerificationToken string
	var scannedEmailVerified bool

	err = session.Query(query, email).Scan(&scannedID, &scannedEmail, &scannedPassword, &scannedEmailVerificationToken, &scannedEmailVerified)
	if err != nil {
		return nil, err
	}

	user := &shared.UserRow{
		ID:                     scannedID,
		Email:                  scannedEmail,
		Password:               scannedPassword,
		EmailVerificationToken: scannedEmailVerificationToken,
		EmailVerified:          scannedEmailVerified,
	}

	return user, nil
}

// GetByEmailVerificationToken returns record by email_verification_token.
func (u *User) GetByEmailVerificationToken(emailVerificationToken string) (*shared.UserRow, error) {
	session, err := u.GetCassandraSession()
	if err != nil {
		return nil, err
	}

	query := fmt.Sprintf("SELECT id, email, password, email_verification_token, email_verified FROM %v WHERE email_verification_token=? LIMIT 1", u.table)

	var scannedID int64
	var scannedEmail, scannedPassword, scannedEmailVerificationToken string
	var scannedEmailVerified bool

	err = session.Query(query, emailVerificationToken).Scan(&scannedID, &scannedEmail, &scannedPassword, &scannedEmailVerificationToken, &scannedEmailVerified)
	if err != nil {
		return nil, err
	}

	user := &shared.UserRow{
		ID:                     scannedID,
		Email:                  scannedEmail,
		Password:               scannedPassword,
		EmailVerificationToken: scannedEmailVerificationToken,
		EmailVerified:          scannedEmailVerified,
	}

	return user, nil
}

// GetByEmail returns record by email but checks password first.
func (u *User) GetUserByEmailAndPassword(email, password string) (*shared.UserRow, error) {
	user, err := u.GetByEmail(email)
	if err != nil {
		return nil, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return nil, err
	}

	return user, err
}

// SignupRandomPassword create a new record of user with random password.
func (u *User) SignupRandomPassword(email string) (*shared.UserRow, error) {
	password, _ := libstring.GeneratePassword(32)
	passwordAgain := password

	return u.Signup(email, password, passwordAgain)
}

// Signup create a new record of user.
func (u *User) Signup(email, password, passwordAgain string) (*shared.UserRow, error) {
	session, err := u.GetCassandraSession()
	if err != nil {
		return nil, err
	}

	if email == "" {
		return nil, errors.New("Email cannot be blank.")
	}
	if password == "" {
		return nil, errors.New("Password cannot be blank.")
	}
	if password != passwordAgain {
		return nil, errors.New("Password is invalid.")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 5)
	if err != nil {
		return nil, err
	}

	emailVerificationToken, err := libstring.GeneratePassword(32)
	if err != nil {
		return nil, err
	}

	id := NewExplicitID()

	query := fmt.Sprintf("INSERT INTO %v (id, email, password, email_verification_token) VALUES (?, ?, ?, ?)", u.table)

	err = session.Query(query, id, email, hashedPassword, emailVerificationToken).Exec()
	if err != nil {
		return nil, err
	}

	return u.GetByID(id)
}

// UpdateEmailAndPasswordByID updates user email and password.
func (u *User) UpdateEmailAndPasswordByID(id int64, email, password, passwordAgain string) (*shared.UserRow, error) {
	session, err := u.GetCassandraSession()
	if err != nil {
		return nil, err
	}

	query := fmt.Sprintf("UPDATE %v SET email=?, password=? WHERE id=? IF EXISTS", u.table)

	var hashedPassword string

	if password != "" && passwordAgain != "" && password == passwordAgain {
		hashedPasswordBytes, err := bcrypt.GenerateFromPassword([]byte(password), 5)
		if err != nil {
			return nil, err
		}
		hashedPassword = string(hashedPasswordBytes)
	}

	err = session.Query(query, email, hashedPassword, id).Exec()
	if err != nil {
		return nil, err
	}

	return u.GetByID(id)
}

// UpdateEmailVerification acknowledge email verification.
func (u *User) UpdateEmailVerification(emailVerificationToken string) (*shared.UserRow, error) {
	session, err := u.GetCassandraSession()
	if err != nil {
		return nil, err
	}

	if emailVerificationToken == "" {
		return nil, errors.New("Token cannot be empty")
	}

	existingUser, err := u.GetByEmailVerificationToken(emailVerificationToken)
	if err != nil {
		return nil, err
	}

	query := fmt.Sprintf("UPDATE %v SET email_verification_token='', email_verified=true WHERE email_verification_token=? IF EXISTS", u.table)

	err = session.Query(query, emailVerificationToken).Exec()
	if err != nil {
		return nil, err
	}

	existingUser.EmailVerificationToken = ""
	existingUser.EmailVerified = true

	return existingUser, nil
}
