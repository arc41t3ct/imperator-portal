package models

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base32"
	"errors"
	"net/http"
	"strings"
	"time"

	up "github.com/upper/db/v4"
)

type Token struct {
	ID        int       `db:"id,omitempty" json:"id"`
	UserID    int       `db:"user_id" json:"user_id"`
	FirstName string    `db:"first_name" json:"first_name"`
	Email     string    `db:"email" json:"email"`
	PlainText string    `db:"token" json:"token"`
	Hash      []byte    `db:"token_hash" json:"-"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
	Expires   time.Time `db:"expiry" json:"expiry"`
}

func (t *Token) Table() string {
	return "tokens"
}

func (t *Token) GetUserForToken(token string) (*User, error) {
	var u User
	var tok Token
	collection := upper.Collection(t.Table())
	res := collection.Find(up.Cond{"token": token})
	if err := res.One(&tok); err != nil {
		return nil, err
	}
	collection = upper.Collection(u.Table())
	res = collection.Find(up.Cond{"id": tok.UserID})
	if err := res.One(&u); err != nil {
		return nil, err
	}
	u.Token = tok
	return &u, nil
}

func (t *Token) GetTokensForUser(id int) ([]*Token, error) {
	var tokens []*Token
	collection := upper.Collection(t.Table())
	res := collection.Find(up.Cond{"user_id": id})
	if err := res.All(&tokens); err != nil {
		return nil, err
	}
	return tokens, nil
}

func (t *Token) Get(id int) (*Token, error) {
	var token Token
	collection := upper.Collection(t.Table())
	res := collection.Find(up.Cond{"id": id})
	if err := res.One(&token); err != nil {
		return nil, err
	}
	return &token, nil
}

func (t *Token) GetByToken(plainTextToken string) (*Token, error) {
	var token Token
	collection := upper.Collection(t.Table())
	res := collection.Find(up.Cond{"token": plainTextToken})
	if err := res.One(&token); err != nil {
		return nil, err
	}
	return &token, nil
}

func (t *Token) Delete(id int) error {
	collection := upper.Collection(t.Table())
	res := collection.Find(id)
	if err := res.Delete(); err != nil {
		return err
	}
	return nil
}

func (t *Token) DeleteByToken(plainTextToken string) error {
	collection := upper.Collection(t.Table())
	res := collection.Find(up.Cond{"token": plainTextToken})
	if err := res.Delete(); err != nil {
		return err
	}
	return nil
}

func (t *Token) Insert(token Token, user User) error {
	collection := upper.Collection(t.Table())
	res := collection.Find(up.Cond{"user_id": user.ID})
	if err := res.Delete(); err != nil {
		return err
	}
	token.CreatedAt = time.Now()
	token.UpdatedAt = time.Now()
	token.FirstName = user.FirstName
	token.Email = user.Email
	_, err := collection.Insert(token)
	if err != nil {
		return err
	}
	return nil
}

func (t *Token) GenerateToken(userID int, ttl time.Duration) (*Token, error) {
	token := &Token{
		UserID:  userID,
		Expires: time.Now().Add(ttl),
	}
	randomBytes := make([]byte, 16)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return nil, err
	}
	token.PlainText = base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(randomBytes)
	hash := sha256.Sum256([]byte(token.PlainText))
	token.Hash = hash[:]
	return token, nil
}

func (t *Token) AuthenticationToken(r *http.Request) (*User, error) {
	authorizationHeader := r.Header.Get("Authorization")
	if authorizationHeader == "" {
		return nil, errors.New("no authorization header")
	}
	headerParts := strings.Split(authorizationHeader, " ")
	if len(headerParts) != 2 || headerParts[0] != "Bearer" {
		return nil, errors.New("no authorization header received")
	}

	token := headerParts[1]
	if len(token) != 26 {
		return nil, errors.New("token malformed")
	}

	tok, err := t.GetByToken(token)
	if err != nil {
		return nil, errors.New("no matching token found")
	}

	if tok.Expires.Before(time.Now()) {
		return nil, errors.New("expired token")
	}

	user, err := tok.GetUserForToken(token)
	if err != nil {
		return nil, errors.New("no matching user found")
	}

	return user, nil
}

func (t *Token) ValidToken(token string) (bool, error) {
	user, err := t.GetUserForToken(token)
	if err != nil {
		return false, errors.New("no matching user found")
	}

	if user.Token.PlainText == "" {
		return false, errors.New("no matching token found")
	}

	if user.Token.Expires.Before(time.Now()) {
		return false, errors.New("expired token")
	}

	return true, nil
}

