package handlers

import (
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/jmoiron/sqlx"
	"github.com/resourced/resourced-master/dal"
	"github.com/resourced/resourced-master/libhttp"
	"github.com/resourced/resourced-master/mailer"
)

func GetSignup(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	tmpl, err := template.ParseFiles("templates/users/login-signup.html.tmpl", "templates/users/signup.html.tmpl")
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	tmpl.Execute(w, nil)
}

func PostSignup(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	db := context.Get(r, "db.Core").(*sqlx.DB)

	email := r.FormValue("Email")
	password := r.FormValue("Password")
	passwordAgain := r.FormValue("PasswordAgain")

	userRow, err := dal.NewUser(db).Signup(nil, email, password, passwordAgain)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	// Create a default cluster
	clusterRow, err := dal.NewCluster(db).Create(nil, userRow.ID, "Default")
	if err != nil {
		libhttp.HandleErrorHTML(w, err, 500)
		return
	}

	// Create a default access-token
	_, err = dal.NewAccessToken(db).Create(nil, userRow.ID, clusterRow.ID, "execute")
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	go func(userRow *dal.UserRow) {
		if userRow.EmailVerificationToken.String != "" {
			mailer := context.Get(r, "mailer.GeneralConfig").(*mailer.Mailer)

			url := fmt.Sprintf("http://localhost:55655/users/email-verification/%v", userRow.EmailVerificationToken.String)

			body := fmt.Sprintf("Click the following link to verify your email address:\n\n%v", url)

			mailer.Send(userRow.Email, "Email Verification", body)
		}
	}(userRow)

	// Perform login
	PostLogin(w, r)
}

func GetLoginWithoutSession(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	tmpl, err := template.ParseFiles("templates/users/login-signup.html.tmpl", "templates/users/login.html.tmpl")
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	tmpl.Execute(w, nil)
}

// GetLogin get login page.
func GetLogin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	cookieStore := context.Get(r, "cookieStore").(*sessions.CookieStore)

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

	db := context.Get(r, "db.Core").(*sqlx.DB)
	cookieStore := context.Get(r, "cookieStore").(*sessions.CookieStore)

	email := r.FormValue("Email")
	password := r.FormValue("Password")

	u := dal.NewUser(db)

	user, err := u.GetUserByEmailAndPassword(nil, email, password)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	session, _ := cookieStore.Get(r, "resourcedmaster-session")
	session.Values["user"] = user

	err = session.Save(r, w)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
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

	db := context.Get(r, "db.Core").(*sqlx.DB)

	cookieStore := context.Get(r, "cookieStore").(*sessions.CookieStore)

	session, _ := cookieStore.Get(r, "resourcedmaster-session")

	currentUser := session.Values["user"].(*dal.UserRow)

	if currentUser.ID != userId {
		err := errors.New("Modifying other user is not allowed.")
		libhttp.HandleErrorJson(w, err)
		return
	}

	email := r.FormValue("Email")
	password := r.FormValue("Password")
	passwordAgain := r.FormValue("PasswordAgain")

	u := dal.NewUser(db)

	currentUser, err = u.UpdateEmailAndPasswordById(nil, currentUser.ID, email, password, passwordAgain)
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

	db := context.Get(r, "db.Core").(*sqlx.DB)

	emailVerificationToken := mux.Vars(r)["token"]

	_, err := dal.NewUser(db).UpdateEmailVerification(nil, emailVerificationToken)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	http.Redirect(w, r, "/login", 301)
}
