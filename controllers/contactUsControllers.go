package controllers

import (
	"cig-exchange-libs"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/keighl/mandrill"
)

// SendContactUsEmail handles POST api/contact_us endpoint
var SendContactUsEmail = func(w http.ResponseWriter, r *http.Request) {

	type contactUs struct {
		Name    string `json:"name"`
		Email   string `json:"email"`
		Message string `json:"message"`
	}
	contactInfo := &contactUs{}
	// decode contact us info from request body
	err := json.NewDecoder(r.Body).Decode(contactInfo)
	if err != nil {
		cigExchange.RespondWithError(w, 422, fmt.Errorf("Invalid request"))
		return
	}

	// check for empty request parameters
	if len(contactInfo.Name) == 0 || len(contactInfo.Email) == 0 || len(contactInfo.Message) == 0 {
		cigExchange.RespondWithError(w, 422, fmt.Errorf("Invalid request. %v", contactInfo))
		return
	}

	mandrillClient := cigExchange.GetMandrill()

	message := &mandrill.Message{}
	message.AddRecipient("info@cig-exchange.ch", "CIG Exchange team", "to")
	message.FromEmail = "info@cig-exchange.ch"
	message.FromName = "CIG Exchange contact us"
	message.Subject = "Contact Us message"
	message.Text = fmt.Sprintf("Name:%s\nEmail:%s\n\nMessage:\n%s", contactInfo.Name, contactInfo.Email, contactInfo.Message)

	resp, err := mandrillClient.MessagesSend(message)
	if err != nil {
		cigExchange.RespondWithError(w, 500, err)
		return
	}

	// we only have 1 recepient
	if len(resp) == 1 {
		if resp[0].Status == "rejected" {
			cigExchange.RespondWithError(w, 422, fmt.Errorf("Invalid request. %v", resp[0].RejectionReason))
			return
		}
	} else {
		cigExchange.RespondWithError(w, 500, fmt.Errorf("Unable to send email"))
		return
	}

	w.WriteHeader(204)
}
