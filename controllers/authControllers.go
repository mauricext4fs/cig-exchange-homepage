package controllers

import (
	"cig-exchange-libs/models"
	"encoding/json"
	"fmt"
	"github.com/mattbaird/gochimp"
	"github.com/mauricext4fs/cig-exchange-libs"
	"net/http"
	"time"
)

var CreateAccount = func(w http.ResponseWriter, r *http.Request) {

	account := &models.Account{}
	err := json.NewDecoder(r.Body).Decode(account) //decode the request body into struct and failed if any error occur
	if err != nil {
		cigExchange.Respond(w, cigExchange.Message(false, "Invalid request"))
		return
	}

	resp := account.Create() //Create account
	cigExchange.Respond(w, resp)
}

var Authenticate = func(w http.ResponseWriter, r *http.Request) {

	account := &models.Account{}
	err := json.NewDecoder(r.Body).Decode(account) //decode the request body into struct and failed if any error occur
	if err != nil {
		cigExchange.Respond(w, cigExchange.Message(false, "Invalid request"))
		return
	}

	resp := models.Login(account.Email)
	cigExchange.Respond(w, resp)
}

var SendVerificationCodeByEmail = func(w http.ResponseWriter, r *http.Request) {

	account := &models.Account{}
	err := json.NewDecoder(r.Body).Decode(account)

	code := &models.Code{}
	code.Code = "value"

	// Add random code in Redis
	rediskey := fmt.Sprintf("%s_code", account.Email)
	expiration := 5 * time.Minute
	err = cigExchange.GetRedis().Set(rediskey, "value", expiration).Err()
	if err != nil {
		panic(err)
	}

	if err != nil {
		cigExchange.Respond(w, cigExchange.Message(false, "Invalid request"))
		return
	}

	msg := fmt.Sprintf("Code Sent to: %s", account.Email)
	resp := cigExchange.Message(true, msg)
	cigExchange.Respond(w, resp)
}
