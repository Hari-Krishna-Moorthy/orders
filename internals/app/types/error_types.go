package types

import "errors"

var (
	ALREADY_EXISTS_ERROR              = errors.New("exists")
	NOT_FOUND_ERROR                   = errors.New("not_found")
	TOPIC_NOT_FOUND_ERROR             = errors.New("topic_not_found")
	EMAIL_AND_PASSWORD_REQUIRED_ERROR = errors.New("email and password required")
	EMAIL_ALREADY_REGISTERED_ERROR    = errors.New("email already registered")
	INVALID_CREDENTIALS_ERROR         = errors.New("invalid credentials")
)
