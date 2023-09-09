package models

import "time"

type JSONB map[string]interface{}

type Common struct {
	ID        int        `json:"id" yaml:"id"`
	CreatedAt *time.Time `json:"createdAt" yaml:"createdAt"`
}

// simplegen:settable-input
// simplegen:paginator
type User struct {
	Common
	FirstName string `json:"firstName" yaml:"firstName"`
	Email     string `json:"email" yaml:"email"`
	Age       int    `json:"age" yaml:"age"`
	Settings  JSONB  `json:"settings" yaml:"settings"`
}

// PaginateOptions describes pagination.
type PaginateOptions struct {
	Cursor       *string `json:"cursor"`
	Limit        int     `json:"limit"`
	NoPagination bool    `json:"noPagination"`
}

// IsPaginated returns if user requires pagination.
func (p PaginateOptions) IsPaginated() bool {
	return !p.NoPagination
}
