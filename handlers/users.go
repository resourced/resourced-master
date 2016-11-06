package handlers

import (
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"github.com/gorilla/sessions"
	"github.com/pressly/chi"

	"github.com/resourced/resourced-master/contexthelper"
	"github.com/resourced/resourced-master/libhttp"
	"github.com/resourced/resourced-master/mailer"
	"github.com/resourced/resourced-master/models/cassandra"
	"github.com/resourced/resourced-master/models/shared"
)

func GetSignup(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	qParams := r.URL.Query()
	email := qParams.Get("email")
	token := qParams.Get("token")

	data := struct {
		Email                  string
		EmailVerificationToken string
	}{
		email,
		token,
	}

	tmpl, err := template.ParseFiles("templates/users/login-signup.html.tmpl", "templates/users/signup.html.tmpl")
	if err != nil {
		libhttp.HandleErrorHTML(w, err, 500)
		return
	}

	tmpl.Execute(w, data)
}

func PostSignup(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	generalConfig, err := contexthelper.GetGeneralConfig(r.Context())
	if err != nil {
		libhttp.HandleErrorHTML(w, err, 500)
		return
	}

	email := r.FormValue("Email")
	password := r.FormValue("Password")
	passwordAgain := r.FormValue("PasswordAgain")
	emailVerificationToken := r.FormValue("EmailVerificationToken")

	emailValidated := false

	userRow, err := cassandra.NewUser(r.Context()).GetByEmail(email)

	if err != nil && err.Error() == "sql: no rows in result set" {
		// There's no existing user in the database, create a new one.
		userRow, err = cassandra.NewUser(r.Context()).Signup(email, password, passwordAgain)
		if err != nil {
			libhttp.HandleErrorHTML(w, err, 500)
			return
		}

		// Create a default cluster
		clusterRow, err := cassandra.NewCluster(r.Context()).Create(userRow, "Default")
		if err != nil {
			libhttp.HandleErrorHTML(w, err, 500)
			return
		}

		// Create a default access-token
		_, err = cassandra.NewAccessToken(r.Context()).Create(userRow.ID, clusterRow.ID, "write")
		if err != nil {
			libhttp.HandleErrorHTML(w, err, 500)
			return
		}

	} else if userRow != nil {
		if userRow.EmailVerificationToken != emailVerificationToken {
			libhttp.HandleErrorHTML(w, errors.New("Mismatch token"), 500)
			return
		}

		emailValidated = true

		// There's an existing user in the database, update email and password info.
		userRow, err = cassandra.NewUser(r.Context()).UpdateEmailAndPasswordByID(userRow.ID, email, password, passwordAgain)
		if err != nil {
			libhttp.HandleErrorHTML(w, err, 500)
			return
		}

		// Verified that emailVerificationToken works.
		_, err = cassandra.NewUser(r.Context()).UpdateEmailVerification(emailVerificationToken)
		if err != nil {
			libhttp.HandleErrorHTML(w, err, 500)
			return
		}
	}

	// Send email verification if needed
	if !emailValidated {
		go func(userRow *shared.UserRow) {
			if userRow.EmailVerificationToken != "" {
				mailer := r.Context().Value("mailer.GeneralConfig").(*mailer.Mailer)

				url := fmt.Sprintf("%v://%v/users/email-verification/%v", generalConfig.VIPProtocol, generalConfig.VIPAddr, userRow.EmailVerificationToken)

				body := fmt.Sprintf("Click the following link to verify your email address:\n\n%v", url)

				mailer.Send(userRow.Email, "Email Verification", body)
			}
		}(userRow)
	}

	// Perform login
	PostLogin(w, r)
}

func GetLoginWithoutSession(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	tmpl, err := template.ParseFiles("templates/users/login-signup.html.tmpl", "templates/users/login.html.tmpl")
	if err != nil {
		libhttp.HandleErrorHTML(w, err, 500)
		return
	}

	tmpl.Execute(w, nil)
}

// GetLogin get login page.
func GetLogin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	cookieStore := r.Context().Value("CookieStore").(*sessions.CookieStore)

	session, _ := cookieStore.Get(r, "resourcedmaster-session")

	currentUserInterface := session.Values["user"]
	if currentUserInterface != nil {
		http.Redirect(w, r, "/", 301)
		return
	}

	GetLoginWithoutSession(w, r)
}

// PostLogin performs login.
func PostLogin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	cookieStore := r.Context().Value("CookieStore").(*sessions.CookieStore)

	email := r.FormValue("Email")
	password := r.FormValue("Password")

	user, err := cassandra.NewUser(r.Context()).GetUserByEmailAndPassword(email, password)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	session, _ := cookieStore.Get(r, "resourcedmaster-session")
	session.Values["user"] = user

	err = session.Save(r, w)
	if err != nil {
		libhttp.HandleErrorHTML(w, err, 500)
		return
	}

	http.Redirect(w, r, "/", 301)
}

func PostPutDeleteUsersID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	method := r.FormValue("_method")
	if method == "" || strings.ToLower(method) == "post" || strings.ToLower(method) == "put" {
		PutUsersID(w, r)
	} else if strings.ToLower(method) == "delete" {
		DeleteUsersID(w, r)
	}
}

func PutUsersID(w http.ResponseWriter, r *http.Request) {
	userId, err := getInt64SlugFromPath(w, r, "id")
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	cookieStore := r.Context().Value("CookieStore").(*sessions.CookieStore)

	session, _ := cookieStore.Get(r, "resourcedmaster-session")

	currentUser := session.Values["user"].(*shared.UserRow)

	if currentUser.ID != userId {
		err := errors.New("Modifying other user is not allowed.")
		libhttp.HandleErrorJson(w, err)
		return
	}

	email := r.FormValue("Email")
	password := r.FormValue("Password")
	passwordAgain := r.FormValue("PasswordAgain")

	currentUser, err = cassandra.NewUser(r.Context()).UpdateEmailAndPasswordByID(currentUser.ID, email, password, passwordAgain)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	// Update currentUser stored in session.
	session.Values["user"] = currentUser
	err = session.Save(r, w)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	http.Redirect(w, r, "/", 301)
}

func DeleteUsersID(w http.ResponseWriter, r *http.Request) {
	err := errors.New("DELETE method is not implemented.")
	libhttp.HandleErrorJson(w, err)
	return
}

// GetUsersEmailVerificationToken verifies user email.
func GetUsersEmailVerificationToken(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	emailVerificationToken := chi.URLParam(r, "token")

	_, err := cassandra.NewUser(r.Context()).UpdateEmailVerification(emailVerificationToken)
	if err != nil {
		libhttp.HandleErrorHTML(w, err, 500)
		return
	}

	http.Redirect(w, r, "/login", 301)
}
