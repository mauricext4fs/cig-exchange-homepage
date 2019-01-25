package controllers

import (
	"cig-exchange-libs"
	"cig-exchange-libs/models"
	"encoding/json"
	"fmt"
	"net/http"

	uuid "github.com/satori/go.uuid"
)

type accountResponse struct {
	UUID string `json:"uuid"`
}

func (resp *accountResponse) randomUUID() {
	UUID, err := uuid.NewV4()
	if err != nil {
		// uuid for an unlikely event of NewV4 failure
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
		fmt.Println("CreateAccount: body JSON encoding error:")
		fmt.Println(err.Error())
		cigExchange.Respond(w, resp)
		return
	}

	// try to create account
	err = account.Create()
	if err != nil {
		fmt.Println("CreateAccount: db Create error:")
		fmt.Println(err.Error())
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
		fmt.Println("GetAccount: body JSON encoding error:")
		fmt.Println(err.Error())
		cigExchange.Respond(w, resp)
		return
	}

	// login using email or phone number
	if len(account.Email) > 0 {
		account, err = models.GetAccountByEmail(account.Email)
	} else if len(account.MobileCode) > 0 && len(account.MobileNumber) > 0 {
		account, err = models.GetAccountByMobile(account.MobileCode, account.MobileNumber)
	} else {
		fmt.Println("GetAccount: neither email or mobile number specified in post body")
		cigExchange.Respond(w, resp)
		return
	}

	if err != nil {
		fmt.Println("GetAccount: db Lookup error:")
		fmt.Println(err.Error())
		cigExchange.Respond(w, resp)
		return
	}
	resp.UUID = account.ID
	cigExchange.Respond(w, resp)
}
