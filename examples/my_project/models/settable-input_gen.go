// Code generated by github.com/AlwxSin/simplegen, DO NOT EDIT.
package models

import (
	"time"
)

// Settable acts like sql.NullString, sql.NullInt64 but generic.
// It allows to define was value set or it's zero value.
type Settable[T any] struct {
	Value T
	IsSet bool
}

// NewSettable returns set value.
func NewSettable[T any](value T) Settable[T] {
	return Settable[T]{
		Value: value,
		IsSet: true,
	}
}

// UserSettable allows to use User with Settable fields
type UserSettable struct {
	ID        Settable[int]        `json:&#34;id&#34; yaml:&#34;id&#34;`
	CreatedAt Settable[*time.Time] `json:&#34;createdAt&#34; yaml:&#34;createdAt&#34;`
	FirstName Settable[string]     `json:&#34;firstName&#34; yaml:&#34;firstName&#34;`
	Email     Settable[string]     `json:&#34;email&#34; yaml:&#34;email&#34;`
	Age       Settable[int]        `json:&#34;age&#34; yaml:&#34;age&#34;`
	Settings  Settable[JSONB]      `json:&#34;settings&#34; yaml:&#34;settings&#34;`
}

func (inp *User) ToSettable(inputFields map[string]interface{}) *UserSettable {
	settable := &UserSettable{}

	if _, ok := inputFields["id"]; ok {
		settable.ID = NewSettable(inp.ID)
	}

	if _, ok := inputFields["createdAt"]; ok {
		settable.CreatedAt = NewSettable(inp.CreatedAt)
	}

	if _, ok := inputFields["firstName"]; ok {
		settable.FirstName = NewSettable(inp.FirstName)
	}

	if _, ok := inputFields["email"]; ok {
		settable.Email = NewSettable(inp.Email)
	}

	if _, ok := inputFields["age"]; ok {
		settable.Age = NewSettable(inp.Age)
	}

	if _, ok := inputFields["settings"]; ok {
		settable.Settings = NewSettable(inp.Settings)
	}

	return settable
}