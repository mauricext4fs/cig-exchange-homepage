package controllers

import (
	"cig-exchange-libs"
	"cig-exchange-libs/models"
	"encoding/json"
	"net/http"

	uuid "github.com/satori/go.uuid"
)

type accountResponse struct {
	UUID string `json:"uuid"`
}

func (resp *accountResponse) randomUUID() {
	UUID, err := uuid.NewV4()
	if err != nil {
		resp.UUID = "fdb283d4-7341-4517-b501-371d22d27cfc"
		return
	}
	resp.UUID = UUID.String()
}

// CreateAccount handles POST api/accounts endpoint
var CreateAccount = func(w http.ResponseWriter, r *http.Request) {

	resp := &accountResponse{}
	resp.randomUUID()

	account := &models.Account{}
	// decode account object from request body
	err := json.NewDecoder(r.Body).Decode(account)
	if err != nil {
		cigExchange.Respond(w, resp)
		return
	}

	// try to create account
	err = account.Create()
	if err != nil {
		cigExchange.Respond(w, resp)
		return
	}
	resp.UUID = account.ID
	cigExchange.Respond(w, resp)
}

// GetAccount handles GET api/accounts endpoint
var GetAccount = func(w http.ResponseWriter, r *http.Request) {

	resp := &accountResponse{}
	resp.randomUUID()

	account := &models.Account{}
	// decode account object from request body
	err := json.NewDecoder(r.Body).Decode(account)
	if err != nil {
		cigExchange.Respond(w, resp)
		return
	}

	// login using email or phone number
	if len(account.Email) > 0 {
		account, err = models.GetAccountByEmail(account.Email)
	} else if len(account.MobileCode) > 0 && len(account.MobileNumber) > 0 {
		account, err = models.GetAccountByMobile(account.MobileCode, account.MobileNumber)
	} else {
		cigExchange.Respond(w, resp)
		return
	}

	if err != nil {
		cigExchange.Respond(w, resp)
		return
	}
	resp.UUID = account.ID
	cigExchange.Respond(w, resp)
}

// SendCode handles POST api/verification_code endpoint
var SendCode = func(w http.ResponseWriter, r *http.Request) {

	type verificationRequest struct {
		UUID string `json:"uuid"`
		Type string `json:"type"`
	}

	reqStruct := &verificationRequest{}
	// decode verificationRequest object from request body
	err := json.NewDecoder(r.Body).Decode(reqStruct)
	if err != nil {
		w.WriteHeader(204)
		return
	}

	account, err := models.GetAccount(reqStruct.UUID)
	if err != nil {
		w.WriteHeader(204)
		return
	}

	// send code to email or phone number
	if reqStruct.Type == "phone" {
		twilioClient := cigExchange.GetTwilio()
		twilioClient.ReceiveOTP(account.MobileCode, account.MobileNumber)
	} else if reqStruct.Type == "email" {
		mandrillClient := cigExchange.GetMandrill()

	}

	w.WriteHeader(204)
}
