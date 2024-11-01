package constants

import "github.com/comfforts/errors"

var (
	ErrUserAccessDenied = errors.NewAppError("you do not have access to the requested resource")
	ErrNotFound         = errors.NewAppError("the requested resource not found")
	ErrTooManyRequests  = errors.NewAppError("you have exceeded throttle")
	ErrNilContext       = errors.NewAppError("context is nil")
)

const (
	ERROR_ENCODING_LAT_LONG        string = "error encoding lat/long"
	ERROR_DECODING_BOUNDS          string = "error decoding bounds"
	ERROR_ENCODING_ID              string = "error encoding store identifier"
	ERROR_NO_STORE_FOUND           string = "no store found"
	ERROR_NO_STORE_FOUND_FOR_ID    string = "no store found for id"
	ERROR_STORE_ID_ALREADY_EXISTS  string = "store already exists"
	ERROR_UNMARSHALLING_STORE_JSON string = "error unmarshalling store json to store"
	ERROR_MARSHALLING_RESULT       string = "error marshalling result to store json"
)
