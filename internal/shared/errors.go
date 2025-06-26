package shared

import (
	"errors"
	"net/http"
)

var ErrInvalidInputs = errors.New("invalid inputs")
var ErrUserAlreadyExist = errors.New("user already exist")

func WriteInternalError(w http.ResponseWriter, msg string, err error) {
	http.Error(w, "Internal Server Error", http.StatusInternalServerError)
}
