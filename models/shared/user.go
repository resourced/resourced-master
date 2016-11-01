package shared

type UserRow struct {
	ID                     int64  `db:"id"`
	Email                  string `db:"email"`
	Password               string `db:"password"`
	EmailVerificationToken string `db:"email_verification_token"`
	EmailVerified          bool   `db:"email_verified"`
}
