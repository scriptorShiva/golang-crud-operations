package response

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
)

type CommonErrRes struct {
	// we are using tags/annotations to convert to json and format in lower case
	Status string `json:"status"`
	Error  string `json:"error"`
}

func WriteJson(w http.ResponseWriter, status int, data interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	
	return json.NewEncoder(w).Encode(data)
}


func CommonError(err error) CommonErrRes {
	return CommonErrRes{
		Status: "Error",
		Error:  err.Error(),
	}
}

func ValidationError(errs validator.ValidationErrors) CommonErrRes {
	var errMsgs []string

	for _, err := range errs {
		switch err.ActualTag() {
			case "required":
				errMsgs = append(errMsgs, fmt.Sprintf("%s is required", err.Field()))
			default :
				errMsgs = append(errMsgs, err.Error())
		}
	}

	return CommonErrRes{
		Status: "Error",
		Error:  strings.Join(errMsgs, ", "),
	}
}