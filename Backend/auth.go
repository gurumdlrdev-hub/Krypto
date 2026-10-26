package main

import (
	"database/sql"
	"net/http"
	"strings"
	"time"

	"aidanwoods.dev/go-paseto"
	"golang.org/x/crypto/bcrypt"
)

var pasetoKey = paseto.NewV4SymmetricKey()

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func signupHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var creds Credentials
	if err := r.ParseForm(); err == nil && r.FormValue("username") != "" {
		creds.Username = r.FormValue("username")
		creds.Password = r.FormValue("password")
	}

	if creds.Username == "" || creds.Password == "" {
		http.Error(w, "Username and Password are required", http.StatusBadRequest)
		return
	}

	if strings.TrimSpace(creds.Username) == "" || len(creds.Password) < 6 {
		http.Error(w, "A password of at least 6 characters is required.", http.StatusBadRequest)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(creds.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Error processing the password", http.StatusInternalServerError)
		return
	}

	// SQL INSERT Query
	query := `INSERT INTO users (username, password_hash) VALUES (?, ?)`
	_, err = DB.Exec(query, creds.Username, string(hashedPassword))
	if err != nil {
		http.Error(w, "The username is already in use.", http.StatusConflict)
		return
	}

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var creds Credentials
	if err := r.ParseForm(); err == nil && r.FormValue("username") != "" {
		creds.Username = r.FormValue("username")
		creds.Password = r.FormValue("password")
	}

	if creds.Username == "" || creds.Password == "" {
		http.Error(w, "Username and Password are required", http.StatusBadRequest)
		return
	}

	var hashedPassword string
	query := `SELECT password_hash FROM users WHERE username = ?`
	err := DB.QueryRow(query, creds.Username).Scan(&hashedPassword)
	if err == sql.ErrNoRows {
		http.Error(w, "Incorrect Username or Password", http.StatusUnauthorized)
		return
	} else if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(creds.Password)); err != nil {
		http.Error(w, "Incorrect username or password", http.StatusUnauthorized)
		return
	}

	token := paseto.NewToken()
	token.SetIssuedAt(time.Now())
	token.SetExpiration(time.Now().Add(24 * time.Hour))
	token.SetString("username", creds.Username)

	encryptedToken := token.V4Encrypt(pasetoKey, nil)

	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    encryptedToken,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   86400,
	})

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
