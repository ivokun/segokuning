package main

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/jmoiron/sqlx"
	"github.com/oklog/ulid/v2"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"time"
)

type User struct {
	ID         string     `json:"id,omitempty" db:"id"`
	Name       string     `json:"name" db:"name"`
	Email      string     `json:"email" db:"email"`
	Phone      string     `json:"phone" db:"phone"`
	Password   string     `json:"-" db:"password"`
	ImageURL   string     `json:"imageUrl" db:"image_url"`
	CreatedAt  *time.Time `json:"createdAt,omitempty" db:"created_at"`
	ModifiedAt *time.Time `json:"modifiedAt,omitempty" db:"modified_at"`
}

type LoggedInUser struct {
	Name     string `json:"name" db:"name"`
	Email    string `json:"email,omitempty" db:"email"`
	Phone    string `json:"phone,omitempty" db:"phone"`
	ImageURL string `json:"imageUrl,omitempty" db:"image_url"`
}

type UserRegisterRequest struct {
	User
	Password        string `json:"password"`
	CredentialType  string `json:"credentialType"`
	CredentialValue string `json:"credentialValue"`
}

type UserLoginRequest struct {
	Password        string `json:"password"`
	CredentialType  string `json:"credentialType"`
	CredentialValue string `json:"credentialValue"`
}

type UserWithToken struct {
	*LoggedInUser
	AccessToken string     `json:"accessToken"`
	CreatedAt   *time.Time `json:"-" db:"created_at"`
	ModifiedAt  *time.Time `json:"-" db:"modified_at"`
}

type UserRegisterResponse struct {
	BaseResponse
	Data UserWithToken `json:"data"`
}

// User utils
func isValidCredentialType(credentialType string) bool {
	return credentialType == "email" || credentialType == "phone"
}

func isValidPassword(password string) bool {
	return len(password) >= 5 && len(password) <= 15
}

// from: https://github.com/go-playground/validator/blob/a0f74b0fb2a7ae1750c0f0b0a49550d8b6e2e708/regexes.go#L18C37-L18C1316
func isValidEmail(email string) bool {
	emailRegexString := "^(?:(?:(?:(?:[a-zA-Z]|\\d|[!#\\$%&'\\*\\+\\-\\/=\\?\\^_`{\\|}~]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])+(?:\\.([a-zA-Z]|\\d|[!#\\$%&'\\*\\+\\-\\/=\\?\\^_`{\\|}~]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])+)*)|(?:(?:\\x22)(?:(?:(?:(?:\\x20|\\x09)*(?:\\x0d\\x0a))?(?:\\x20|\\x09)+)?(?:(?:[\\x01-\\x08\\x0b\\x0c\\x0e-\\x1f\\x7f]|\\x21|[\\x23-\\x5b]|[\\x5d-\\x7e]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])|(?:(?:[\\x01-\\x09\\x0b\\x0c\\x0d-\\x7f]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}]))))*(?:(?:(?:\\x20|\\x09)*(?:\\x0d\\x0a))?(\\x20|\\x09)+)?(?:\\x22))))@(?:(?:(?:[a-zA-Z]|\\d|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])|(?:(?:[a-zA-Z]|\\d|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])(?:[a-zA-Z]|\\d|-|\\.|~|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])*(?:[a-zA-Z]|\\d|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])))\\.)+(?:(?:[a-zA-Z]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])|(?:(?:[a-zA-Z]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])(?:[a-zA-Z]|\\d|-|\\.|~|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])*(?:[a-zA-Z]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])))\\.?$"
	emailRegex := regexp.MustCompile(emailRegexString)
	return emailRegex.MatchString(email)
}

func isValidPhoneNumber(username string) bool {
	//phone number should start with "international calling code" (including the "+" prefix)
	//with minLength=7 and maxLength=13 (including the "international calling code" with the "+" and only
	phoneRegexString := "^\\+[0-9]{7,13}$"
	phoneRegex := regexp.MustCompile(phoneRegexString)

	return phoneRegex.MatchString(username)
}

func NewUserRegisterResponse(data *UserWithToken) *UserRegisterResponse {
	resp := &UserRegisterResponse{
		BaseResponse: BaseResponse{
			Message: "User registered successfully",
		},
		Data: *data,
	}

	return resp
}

func ValidUserLoginResponse(data *UserWithToken) *UserRegisterResponse {
	resp := &UserRegisterResponse{
		BaseResponse: BaseResponse{
			Message: "User logged in successfully",
		},
		Data: *data,
	}

	return resp
}

func (rd *UserRegisterResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func generateToken(user *User) (string, error) {
	_, token, err := TokenAuth.Encode(map[string]interface{}{"user_id": user.ID})
	return token, err
}

func hashPassword(password string) (string, error) {
	saltLength, err := strconv.Atoi(os.Getenv("BCRYPT_SALT"))

	if err != nil {
		return "", err
	}

	if saltLength <= 0 {
		saltLength = 8
	}
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), saltLength)
	return string(bytes), err
}

func comparePassword(hashedPassword string, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

// User database operations
func getUserByPhone(db *sqlx.DB, phoneNumber string) (*User, error) {
	query := `SELECT id, name, password, email, phone, image_url  FROM users WHERE phone = $1`
	data := &User{}
	err := db.Get(data, query, phoneNumber)
	return data, err
}

func getUserByEmail(db *sqlx.DB, email string) (*User, error) {
	query := `SELECT id, name, password, email, phone, image_url  FROM users WHERE email = $1`
	data := &User{}
	err := db.Get(data, query, email)
	return data, err
}

func getUser(db *sqlx.DB, payload *UserLoginRequest) (*User, error) {
	if payload.CredentialType == "phone" {
		return getUserByPhone(db, payload.CredentialValue)
	}
	return getUserByEmail(db, payload.CredentialValue)
}

func generateUserWithPhone(payload *UserRegisterRequest) (*User, error) {
	hashedPassword, err := hashPassword(payload.Password)

	if err != nil {
		return nil, err
	}

	return &User{
		ID:       ulid.Make().String(),
		Name:     payload.Name,
		Phone:    payload.CredentialValue,
		Password: hashedPassword,
	}, nil
}

func generateUserWithEmail(payload *UserRegisterRequest) (*User, error) {
	hashedPassword, err := hashPassword(payload.Password)

	if err != nil {
		return nil, err
	}

	return &User{
		ID:       ulid.Make().String(),
		Name:     payload.Name,
		Email:    payload.CredentialValue,
		Password: hashedPassword,
	}, nil
}

func generateUser(payload *UserRegisterRequest) (*User, error) {
	if payload.CredentialType == "phone" {
		return generateUserWithPhone(payload)
	}
	return generateUserWithEmail(payload)
}

func createUser(db *sqlx.DB, user *User) error {
	_, err := db.NamedExec(`INSERT INTO users (id, name, email, phone, image_url, password) VALUES (:id, :name, :email, :phone, :image_url, :password)`, user)
	return err
}

// User router
func UserRouter(db *sqlx.DB) chi.Router {
	r := chi.NewRouter()
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok gas, ok gas"))
	})
	r.Post("/register", func(w http.ResponseWriter, r *http.Request) {
		UserRegistrationHandler(w, r, db)
	})
	r.Post("/login", func(w http.ResponseWriter, r *http.Request) {
		UserLoginHandler(w, r, db)
	})
	return r
}

// User handlers
func UserRegistrationHandler(w http.ResponseWriter, r *http.Request, db *sqlx.DB) {
	payload := &UserRegisterRequest{}

	if err := render.Decode(r, payload); err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	if !isValidCredentialType(payload.CredentialType) {
		render.Render(w, r, ErrInvalidRequest(fmt.Errorf("Invalid credential type")))
		return
	}

	if payload.CredentialType == "email" && !isValidEmail(payload.CredentialValue) {
		render.Render(w, r, ErrInvalidRequest(fmt.Errorf("Invalid email")))
		return
	}

	if payload.CredentialType == "phone" && !isValidPhoneNumber(payload.CredentialValue) {
		render.Render(w, r, ErrInvalidRequest(fmt.Errorf("Invalid phone number")))
		return
	}

	if !isValidPassword(payload.Password) {
		render.Render(w, r, ErrInvalidRequest(fmt.Errorf("Password must be between 5 and 15 characters")))
		return
	}

	user, err := generateUser(payload)

	if err != nil {
		render.Render(w, r, ErrServer(fmt.Errorf("Something wrong, please try again."), http.StatusInternalServerError))
		return
	}

	err = createUser(db, user)

	if err != nil {
		render.Render(w, r, ErrServer(ParseDBErrorMessage(err)))
		return
	}

	token, err := generateToken(user)

	if err != nil {
		render.Render(w, r, ErrServer(fmt.Errorf("Error generating token, please try again."), http.StatusInternalServerError))
		return
	}

	data := &UserWithToken{
		LoggedInUser: &LoggedInUser{
			Name:     user.Name,
			Email:    user.Email,
			Phone:    user.Phone,
			ImageURL: user.ImageURL,
		},
		AccessToken: token,
	}

	render.Status(r, http.StatusCreated)
	render.Render(w, r, NewUserRegisterResponse(data))
}

func UserLoginHandler(w http.ResponseWriter, r *http.Request, db *sqlx.DB) {
	payload := &UserLoginRequest{}

	if err := render.Decode(r, payload); err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	if !isValidCredentialType(payload.CredentialType) {
		render.Render(w, r, ErrInvalidRequest(fmt.Errorf("Invalid credential type")))
		return
	}

	if payload.CredentialType == "email" && !isValidEmail(payload.CredentialValue) {
		render.Render(w, r, ErrInvalidRequest(fmt.Errorf("Invalid email/phone number or password")))
		return
	}

	if payload.CredentialType == "phone" && !isValidPhoneNumber(payload.CredentialValue) {
		render.Render(w, r, ErrInvalidRequest(fmt.Errorf("Invalid email/phone number or password")))
		return
	}
	if !isValidPassword(payload.Password) {
		render.Render(w, r, ErrInvalidRequest(fmt.Errorf("Invalid email/phone number or password")))
		return
	}

	user, err := getUser(db, payload)

	if err != nil {
		render.Render(w, r, ErrServer(ParseDBErrorMessage(err)))
		return
	}

	if !comparePassword(user.Password, payload.Password) {
		render.Render(w, r, ErrInvalidRequest(fmt.Errorf("Invalid username or password")))
		return
	}

	token, err := generateToken(user)

	if err != nil {
		render.Render(w, r, ErrServer(fmt.Errorf("Error generating token, please try again."), http.StatusInternalServerError))
		return
	}

	data := &UserWithToken{
		LoggedInUser: &LoggedInUser{
			Name:     user.Name,
			Email:    user.Email,
			Phone:    user.Phone,
			ImageURL: user.ImageURL,
		},
		AccessToken: token,
	}

	render.Status(r, http.StatusOK)
	render.Render(w, r, ValidUserLoginResponse(data))
}
