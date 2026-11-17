package models

import (
	"database/sql"
	"errors"
	"time"

	"github.com/lib/pq"
	"github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

type UserModelInterface interface {
	Insert(name, email, password string) error
	Authenticate(email, password string) (int, error)
	Get(id int) (User, error)
	Exists(id int) (bool, error)
	PasswordUpdate(id int, currentPassword, newPassword string) error
}

type User struct {
	ID             string
	Name           string
	Email          string
	HashedPassword []byte
	Created        time.Time
}

type UserModel struct {
	DB      *sql.DB
	Dialect string
}

func (m *UserModel) PasswordUpdate(id int, currentPassword, newPassword string) error {
	var user User

	stmt := bindQuery(`SELECT hashed_password FROM users WHERE id = ?`, m.Dialect)
	m.DB.QueryRow(stmt, id).Scan(&user.HashedPassword)
	hash := &user.HashedPassword

	err := bcrypt.CompareHashAndPassword(*hash, []byte(currentPassword))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return ErrInvalidCredentials
		} else {
			return err
		}
	}

	newPasswordhash, err := bcrypt.GenerateFromPassword([]byte(newPassword), 12)
	if err != nil {
		return err
	}

	stmt = bindQuery(`UPDATE users SET hashed_password = ? WHERE id = ?`, m.Dialect)

	_, err = m.DB.Exec(stmt, newPasswordhash, id)
	return err
}

func (m *UserModel) Get(id int) (User, error) {
	var user User

	_, err := m.Exists(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return User{}, ErrNoRecord
		}
		return User{}, err
	}

	stmt := bindQuery(`SELECT id, name, email, created FROM users WHERE id = ?`, m.Dialect)
	m.DB.QueryRow(stmt, id).Scan(&user.ID, &user.Name, &user.Email, &user.Created)
	return user, nil
}

func (m *UserModel) Insert(name, email, password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return err
	}

	stmt := bindQuery(`INSERT INTO users (name, email, hashed_password, created)
    VALUES(?, ?, ?, ?)`, m.Dialect)

	_, err = m.DB.Exec(stmt, name, email, hashedPassword, time.Now().UTC())
	if err != nil {
		if isDuplicateEmailError(err) {
			return ErrDuplicateEmail
		}
		return err
	}
	return nil
}

func isDuplicateEmailError(err error) bool {
	var postgresError *pq.Error
	if errors.As(err, &postgresError) {
		return postgresError.Code == "23505" && postgresError.Constraint == "users_email_key"
	}

	var sqliteError sqlite3.Error
	return errors.As(err, &sqliteError) && sqliteError.ExtendedCode == sqlite3.ErrConstraintUnique
}

func (m *UserModel) Authenticate(email string, password string) (int, error) {
	var id int
	var hashedPassword []byte

	row := m.DB.QueryRow(bindQuery("SELECT id, hashed_password FROM users WHERE email = ?", m.Dialect), email)
	err := row.Scan(&id, &hashedPassword)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, ErrInvalidCredentials
		} else {
			return 0, err
		}
	}

	err = bcrypt.CompareHashAndPassword(hashedPassword, []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return 0, ErrInvalidCredentials
		} else {
			return 0, err
		}
	}
	return id, nil
}

func (m *UserModel) Exists(id int) (bool, error) {
	var exists bool
	stmt := bindQuery(`SELECT EXISTS(SELECT true FROM users WHERE id = ?)`, m.Dialect)
	err := m.DB.QueryRow(stmt, id).Scan(&exists)
	return exists, err
}
