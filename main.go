// main.go
package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/sessions"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

var (
	tpl   *template.Template
	db    *sql.DB
	store *sessions.CookieStore
)

func init() {
	// Read secret from file
	secret, err := os.ReadFile("session_key.txt")
	if err != nil {
		log.Fatalf("Unable to read secret key: %v", err)
	}

	// Parse templates
	tpl = template.Must(template.ParseGlob("templates/*.html"))

	// Initialize session store
	store = sessions.NewCookieStore(secret)
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7, // 7 days
		HttpOnly: true,
	}

	// Database connection verification
	go func() {
		var err error
		db, err = sql.Open("sqlite3", "./db/users.db")
		if err != nil {
			log.Fatal(err)
		}

		err = db.Ping()
		if err != nil {
			log.Fatal(err)
		}
		log.Println("Database connection successful")
	}()

	// Test password hashing
	testPassword := "yourpassword"
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(testPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Test hash generation failed: %v", err)
	}

	err = bcrypt.CompareHashAndPassword(hashedPassword, []byte(testPassword))
	if err != nil {
		log.Printf("Test hash comparison failed: %v", err)
	} else {
		log.Printf("Password hashing test successful")
	}
}

func main() {
	var err error
	db, err = sql.Open("sqlite3", "./db/users.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Routes
	http.HandleFunc("/", loginHandler)
	http.HandleFunc("/register", registerHandler)
	http.HandleFunc("/main", authMiddleware(mainHandler))
	http.HandleFunc("/logout", logoutHandler)

	// Serve static files
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	fmt.Println("Server started at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		tpl.ExecuteTemplate(w, "login.html", nil)
		return
	}

	email := r.FormValue("email")
	password := r.FormValue("password")

	log.Printf("Login attempt for email: %s", email)
	log.Printf("Submitted password length: %d", len(password))

	var hashedPassword string
	var id int
	err := db.QueryRow("SELECT id, password FROM users WHERE email = ?", email).Scan(&id, &hashedPassword)
	if err != nil {
		log.Printf("Database query error: %v", err)
		tpl.ExecuteTemplate(w, "login.html", "Invalid email or password")
		return
	}

	log.Printf("Found user with ID: %d", id)

	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		log.Printf("Password comparison failed: %v", err)
		tpl.ExecuteTemplate(w, "login.html", "Invalid email or password")
		return
	}

	log.Printf("Password comparison successful")

	// Create new session
	session, err := store.New(r, "session")
	if err != nil {
		log.Printf("Session creation error: %v", err)
		tpl.ExecuteTemplate(w, "login.html", "Session error")
		return
	}

	// Set session values
	session.Values["authenticated"] = true
	session.Values["email"] = email
	session.Values["user_id"] = id

	// Save session
	err = session.Save(r, w)
	if err != nil {
		log.Printf("Session save error: %v", err)
		tpl.ExecuteTemplate(w, "login.html", "Session save error")
		return
	}

	log.Printf("Session created successfully, redirecting to /main")
	http.Redirect(w, r, "/main", http.StatusSeeOther)
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		tpl.ExecuteTemplate(w, "register.html", nil)
		return
	}

	fullname := r.FormValue("fullname")
	email := r.FormValue("email")
	password := r.FormValue("password")

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		tpl.ExecuteTemplate(w, "register.html", "Error creating account")
		return
	}

	_, err = db.Exec("INSERT INTO users (fullname, email, password) VALUES (?, ?, ?)",
		fullname, email, hashedPassword)
	if err != nil {
		tpl.ExecuteTemplate(w, "register.html", "Email already registered")
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func mainHandler(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, "session")
	if err != nil {
		log.Printf("Main handler - session error: %v", err)
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	email, ok := session.Values["email"].(string)
	if !ok {
		log.Printf("Main handler - email not found in session")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	var fullname string
	err = db.QueryRow("SELECT fullname FROM users WHERE email = ?", email).Scan(&fullname)
	if err != nil {
		log.Printf("Main handler - database error: %v", err)
		http.Error(w, "User not found", http.StatusInternalServerError)
		return
	}

	log.Printf("Main handler - rendering page for user: %s", fullname)
	err = tpl.ExecuteTemplate(w, "main.html", fullname)
	if err != nil {
		log.Printf("Main handler - template error: %v", err)
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session")
	session.Options.MaxAge = -1
	session.Save(r, w)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, err := store.Get(r, "session")
		if err != nil {
			log.Printf("Auth middleware - session error: %v", err)
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		// Check if the session exists and is authenticated
		auth, ok := session.Values["authenticated"].(bool)
		email, emailOk := session.Values["email"].(string)

		log.Printf("Auth middleware - Session values: authenticated=%v, ok=%v, email=%v, emailOk=%v",
			auth, ok, email, emailOk)

		if !ok || !auth || !emailOk {
			log.Printf("Auth middleware - authentication failed")
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		log.Printf("Auth middleware - authentication successful for user: %v", email)
		next.ServeHTTP(w, r)
	}
}
