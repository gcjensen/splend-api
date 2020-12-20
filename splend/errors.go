package splend

import "errors"

// Errors associated with users and outgoings
var (
	ErrOutgoingUnknown = errors.New("unknown outgoing")
	ErrUserNotInCouple = errors.New("user does not have a couple_id")
	ErrUserUnknown     = errors.New("unknown user")
)

// Errors associated with linked accounts
var (
	ErrAlreadyExists             = errors.New("amex transaction already exists")
	ErrMonzoAccountAlreadyLinked = errors.New("monzo account already linked")
)
