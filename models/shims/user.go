package shims

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"

	"github.com/resourced/resourced-master/contexthelper"
	"github.com/resourced/resourced-master/models/cassandra"
	"github.com/resourced/resourced-master/models/pg"
)

func NewUser(ctx context.Context) *User {
	u := &User{}
	u.AppContext = ctx
	return u
}

type User struct {
	Base
}

func (u *User) GetDBType() string {
	generalConfig, err := contexthelper.GetGeneralConfig(u.AppContext)
	if err != nil {
		return ""
	}

	return generalConfig.GetCoreDBType()
}

func (u *User) GetPGDB() (*sqlx.DB, error) {
	pgdbs, err := contexthelper.GetPGDBConfig(u.AppContext)
	if err != nil {
		return nil, err
	}

	return pgdbs.Core, nil
}

func (u *User) AllUsers() (interface{}, error) {
	if u.GetDBType() == "pg" {
		return pg.NewUser(u.AppContext).AllUsers(nil)

	} else if u.GetDBType() == "cassandra" {
		return cassandra.NewUser(u.AppContext).AllUsers()
	}

	return nil, fmt.Errorf("Unrecognized DBType, valid options are: pg or cassandra")
}

func (u *User) GetByID(id int64) (interface{}, error) {
	if u.GetDBType() == "pg" {
		return pg.NewUser(u.AppContext).GetByID(nil, id)

	} else if u.GetDBType() == "cassandra" {
		return cassandra.NewUser(u.AppContext).GetByID(id)
	}

	return nil, fmt.Errorf("Unrecognized DBType, valid options are: pg or cassandra")
}

func (u *User) GetByEmail(email string) (interface{}, error) {
	if u.GetDBType() == "pg" {
		return pg.NewUser(u.AppContext).GetByEmail(nil, email)

	} else if u.GetDBType() == "cassandra" {
		return cassandra.NewUser(u.AppContext).GetByEmail(email)
	}

	return nil, fmt.Errorf("Unrecognized DBType, valid options are: pg or cassandra")
}

func (u *User) GetByEmailVerificationToken(emailVerificationToken string) (interface{}, error) {
	if u.GetDBType() == "pg" {
		return pg.NewUser(u.AppContext).GetByEmailVerificationToken(nil, emailVerificationToken)

	} else if u.GetDBType() == "cassandra" {
		return cassandra.NewUser(u.AppContext).GetByEmailVerificationToken(emailVerificationToken)
	}

	return nil, fmt.Errorf("Unrecognized DBType, valid options are: pg or cassandra")
}

func (u *User) GetUserByEmailAndPassword(email, password string) (interface{}, error) {
	if u.GetDBType() == "pg" {
		return pg.NewUser(u.AppContext).GetUserByEmailAndPassword(nil, email, password)

	} else if u.GetDBType() == "cassandra" {
		return cassandra.NewUser(u.AppContext).GetUserByEmailAndPassword(email, password)
	}

	return nil, fmt.Errorf("Unrecognized DBType, valid options are: pg or cassandra")
}

func (u *User) SignupRandomPassword(email string) (interface{}, error) {
	if u.GetDBType() == "pg" {
		return pg.NewUser(u.AppContext).SignupRandomPassword(nil, email)

	} else if u.GetDBType() == "cassandra" {
		return cassandra.NewUser(u.AppContext).SignupRandomPassword(email)
	}

	return nil, fmt.Errorf("Unrecognized DBType, valid options are: pg or cassandra")
}

func (u *User) Signup(email, password, passwordAgain string) (interface{}, error) {
	if u.GetDBType() == "pg" {
		return pg.NewUser(u.AppContext).Signup(nil, email, password, passwordAgain)

	} else if u.GetDBType() == "cassandra" {
		return cassandra.NewUser(u.AppContext).Signup(email, password, passwordAgain)
	}

	return nil, fmt.Errorf("Unrecognized DBType, valid options are: pg or cassandra")
}

func (u *User) UpdateEmailAndPasswordByID(id int64, email, password, passwordAgain string) (interface{}, error) {
	if u.GetDBType() == "pg" {
		return pg.NewUser(u.AppContext).UpdateEmailAndPasswordByID(nil, id, email, password, passwordAgain)

	} else if u.GetDBType() == "cassandra" {
		return cassandra.NewUser(u.AppContext).UpdateEmailAndPasswordByID(id, email, password, passwordAgain)
	}

	return nil, fmt.Errorf("Unrecognized DBType, valid options are: pg or cassandra")
}

func (u *User) UpdateEmailVerification(emailVerificationToken string) (interface{}, error) {
	if u.GetDBType() == "pg" {
		return pg.NewUser(u.AppContext).UpdateEmailVerification(nil, emailVerificationToken)

	} else if u.GetDBType() == "cassandra" {
		return cassandra.NewUser(u.AppContext).UpdateEmailVerification(emailVerificationToken)
	}

	return nil, fmt.Errorf("Unrecognized DBType, valid options are: pg or cassandra")
}
