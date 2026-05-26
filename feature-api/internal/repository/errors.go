package repository

import "errors"

// Sentinel errors — use errors.Is() to match these in callers.
var (
	ErrNotFound      = errors.New("not found")
	ErrInvalidID     = errors.New("invalid id")
	ErrNoFields      = errors.New("no fields to update")
	ErrAlreadyExists = errors.New("already exists")
	ErrInvalidRules  = errors.New("invalid rules")
)
