package http

import "errors"

var (
	ErrOutgoingIDExpected      = errors.New("outgoingID parameter expected")
	ErrUnregisteredWebhookType = errors.New("unregistered webhook type")
)
