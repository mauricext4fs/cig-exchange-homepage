package controllers

import (
	"cig-exchange-libs"
	"cig-exchange-libs/models"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/mattbaird/gochimp"
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

	sendCodePerEmail(account.Email, code.Code)

	msg := fmt.Sprintf("Code Sent to: %s", account.Email)
	resp := cigExchange.Message(true, msg)
	cigExchange.Respond(w, resp)
}

var sendCodePerEmail = func(email string, code string) {
	apiKey := os.Getenv("MANDRILL_KEY")
	mandrillApi, err := gochimp.NewMandrill(apiKey)

	if err != nil {
		fmt.Println("Error instantiating client")
	}

	templateName := "pin-code"
	templateContent, err := mandrillApi.TemplateInfo(templateName)
	if err != nil {
		fmt.Printf("Error getting template info: %v\n", err)
		return
	}

	contentVar := gochimp.Var{
		Name:    "pin-code",
		Content: templateContent,
	}
	content := []gochimp.Var{contentVar}

	mergeVar := gochimp.Var{
		Name:    "pincode",
		Content: code,
	}
	merge := []gochimp.Var{mergeVar}

	renderedTemplate, err := mandrillApi.TemplateRender(templateName, content, merge)

	if err != nil {
		fmt.Printf("Error rendering template: %v\n", err)
		return
	}

	recipients := []gochimp.Recipient{
		gochimp.Recipient{Email: email},
	}

	message := gochimp.Message{
		Html:      renderedTemplate,
		Subject:   "Welcome aboard!",
		FromEmail: "noreply@cig-exchange.ch",
		FromName:  "CIG Exchange",
		To:        recipients,
	}

	_, err = mandrillApi.MessageSend(message, false)

	if err != nil {
		fmt.Println("Error sending message")
	}

	fmt.Println("Message Sent to: ", email)

}
