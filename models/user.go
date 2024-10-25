package models

import (
	"errors"
	"time"

	"github.com/arc41t3ct/imperator"
	up "github.com/upper/db/v4"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID        int       `db:"id,omitempty"`
	FirstName string    `db:"first_name"`
	LastName  string    `db:"last_name"`
	Active    int       `db:"user_active"`
	Email     string    `db:"email"`
	Password  string    `db:"password"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
	Token     Token     `db:"-"`
}

func (u *User) Table() string {
	return "users"
}

func (u *User) Validate(validator *imperator.Validation) {
	validator.Check(u.FirstName != "", "first_name", "First name is required")
	validator.Check(u.LastName != "", "last_name", "Last name is required")
	validator.Check(u.Email != "", "email", "Email is required")
	validator.IsEmail("email", u.Email)
}

func (u *User) GetAll() ([]*User, error) {
	collection := upper.Collection(u.Table())
	var all []*User
	res := collection.Find().OrderBy("created_at")
	if err := res.All(&all); err != nil {
		return nil, err
	}
	return all, nil
}

func (u *User) GetByEmail(email string) (*User, error) {
	var user *User
	collection := upper.Collection((u.Table()))
	res := collection.Find(up.Cond{"email =": email})
	if err := res.One(&user); err != nil {
		return nil, err
	}
	token, err := user.getToken()
	if err != nil {
		return nil, err
	}
	user.Token = token
	return user, nil
}

func (u *User) Get(id int) (*User, error) {
	var user *User
	collection := upper.Collection((u.Table()))
	res := collection.Find(up.Cond{"id =": id})
	if err := res.One(&user); err != nil {
		return nil, err
	}
	token, err := u.getToken()
	if err != nil {
		return nil, err
	}
	user.Token = token
	return user, nil
}

// Update updates a user based on the user model it is passed
func (u *User) Update(user User) error {
	user.UpdatedAt = time.Now()
	collection := upper.Collection(u.Table())
	res := collection.Find(user.ID)
	if err := res.Update(&user); err != nil {
		return err
	}
	return nil
}

// Delete deletes a user given the user's id
func (u *User) Delete(id int) error {
	collection := upper.Collection(u.Table())
	res := collection.Find(id)
	if err := res.Delete(); err != nil {
		return err
	}
	return nil
}

// Insert creates a new user given a user
func (u *User) Insert(user User) (int, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(user.Password), 12)
	if err != nil {
		return 0, err
	}
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	user.Password = string(hash)

	collection := upper.Collection(u.Table())
	res, err := collection.Insert(user)
	if err != nil {
		return 0, err
	}
	id := getInsertID(res.ID())
	return id, nil
}

// ResetPassword resets the password of a user given the id and new password
func (u *User) ResetPassword(id int, password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return err
	}
	user, err := u.Get(id)
	if err != nil {
		return err
	}

	u.UpdatedAt = time.Now()
	u.Password = string(hash)
	if err := user.Update(*u); err != nil {
		return err
	}
	return nil
}

// PasswordMatches check the supplied passwordInput and the hashed password to see if they match
func (u *User) PasswordMatches(passwordInput string) (bool, error) {
	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(passwordInput)); err != nil {
		switch {
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			return false, nil
		default:
			return false, err
		}
	}
	return true, nil
}

// getToken return a Token used which is used for authentiction
func (u *User) getToken() (Token, error) {
	var token Token
	collection := upper.Collection(token.Table())
	res := collection.Find(up.Cond{
		"user_id =": u.ID, "expiry >": time.Now(),
	}).OrderBy("created_at desc")
	err := res.One(&token)
	if err != nil {
		if err != up.ErrNilRecord && err != up.ErrNoMoreRows {
			return Token{}, err
		}
	}
	return token, nil
}

func (u *User) CheckForRememberToken(id int, token string) bool {
	var remeberToken RememberToken
	rt := RememberToken{}
	collection := upper.Collection(rt.Table())
	res := collection.Find(up.Cond{"user_id": id, "remeber_token": token})
	err := res.One(&remeberToken)
	return err == nil
}
