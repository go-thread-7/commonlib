package constants

import "errors"

const (
	ErrBadRequest         = "Bad request"
	ErrEmailAlreadyExists = "User with given email already exists"
	ErrNoSuchUser         = "User not found"
	ErrWrongCredentials   = "Wrong Credentials"
	ErrNotFound           = "Not Found"
	ErrUnauthorized       = "Unauthorized"
	ErrForbidden          = "Forbidden"
	ErrBadQueryParams     = "Invalid query params"
)

var (
	ErrorBadRequest            = errors.New("Bad request")
	ErrorWrongCredentials      = errors.New("Wrong Credentials")
	ErrorNotFound              = errors.New("Not Found")
	ErrorUnauthorized          = errors.New("Unauthorized")
	ErrorForbidden             = errors.New("Forbidden")
	ErrorPermissionDenied      = errors.New("Permission Denied")
	ErrorExpiredCSRFError      = errors.New("Expired CSRF token")
	ErrorWrongCSRFToken        = errors.New("Wrong CSRF token")
	ErrorCSRFNotPresented      = errors.New("CSRF not presented")
	ErrorNotRequiredFields     = errors.New("No such required fields")
	ErrorBadQueryParams        = errors.New("Invalid query params")
	ErrorInternalServerError   = errors.New("Internal Server Error")
	ErrorRequestTimeoutError   = errors.New("Request Timeout")
	ErrorExistsEmailError      = errors.New("User with given email already exists")
	ErrorInvalidJWTToken       = errors.New("Invalid JWT token")
	ErrorInvalidJWTClaims      = errors.New("Invalid JWT claims")
	ErrorNotAllowedImageHeader = errors.New("Not allowed image header")
	ErrorNoCookie              = errors.New("not found cookie header")
)
