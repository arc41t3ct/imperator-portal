package models

import (
	"time"

	up "github.com/upper/db/v4"
)

// Note - item.UpdatedAt is handled from the database

type RememberToken struct {
	ID            int       `db:"id,omitempty"`
	UserID        int       `db:"user_id"`
	RememberToken string    `db:"remember_token"`
	CreatedAt     time.Time `db:"created_at"`
	UpdatedAt     time.Time `db:"updated_at"`
}

// Table returns the table name for the RememberToken
func (m *RememberToken) Table() string {
	return "remember_tokens"
}

// Get gets a RememberToken from the database by passing the id
func (m *RememberToken) Get(id int) (*RememberToken, error) {
	var item *RememberToken
	collection := upper.Collection(m.Table())
	res := collection.Find(up.Cond{"id =": id})
	if err := res.One(&item); err != nil {
		return nil, err
	}
	return item, nil
}

// Delete deletes a RememberToken given the id
func (m *RememberToken) Delete(id int) error {
	collection := upper.Collection(m.Table())
	res := collection.Find(id)
	if err := res.Delete(); err != nil {
		return err
	}
	return nil
}

// Delete deletes a RememberToken given the id
func (m *RememberToken) DeleteByToken(token string) error {
	collection := upper.Collection(m.Table())
	res := collection.Find(up.Cond{"remember_token": token})
	if err := res.Delete(); err != nil {
		return err
	}
	return nil
}

// Insert creates a new RememberToken given the item
func (m *RememberToken) Insert(item RememberToken) (int, error) {
	item.CreatedAt = time.Now()

	collection := upper.Collection(m.Table())
	res, err := collection.Insert(item)
	if err != nil {
		return 0, err
	}
	id := getInsertID(res.ID())
	return id, nil
}
