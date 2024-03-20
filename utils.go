package main

import (
	"fmt"
	"github.com/go-chi/jwtauth/v5"
	"net/http"
)

var TokenAuth *jwtauth.JWTAuth

type BaseResponse struct {
	Message string `json:"message"`
}

func ParseDBErrorMessage(err error) (error, int) {
	errorMessage := err.Error()
	if errorMessage == "pq: duplicate key value violates unique constraint \"users_email_phone_key\"" {
		return fmt.Errorf("Email/Phone already registered"), http.StatusConflict
	}
	if errorMessage == "sql: no rows in result set" {
		return fmt.Errorf("Not found"), http.StatusNotFound
	}
	return fmt.Errorf("Something wrong with our server, please try again."), http.StatusInternalServerError
}
