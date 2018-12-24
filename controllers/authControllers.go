package controllers

import (
	"cig-exchange-libs"
	"cig-exchange-libs/models"
	"encoding/json"
	"fmt"
	"github.com/mattbaird/gochimp"
	"net/http"
	"os"
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

	templateName := "welcome email"
	contentVar := gochimp.Var{"main", "<h1>Welcome aboard!</h1>"}
	content := []gochimp.Var{contentVar}

	_, err = mandrillApi.TemplateAdd(templateName, fmt.Sprintf("%s", contentVar.Content), true)
	if err != nil {
		fmt.Println("Error adding template: %v", err)
		return
	}
	defer mandrillApi.TemplateDelete(templateName)
	renderedTemplate, err := mandrillApi.TemplateRender(templateName, content, nil)

	if err != nil {
		fmt.Println("Error rendering template: %v", err)
		return
	}

	recipients := []gochimp.Recipient{
		gochimp.Recipient{Email: email},
	}

	message := gochimp.Message{
		Html:      renderedTemplate,
		Subject:   "Welcome aboard!",
		FromEmail: email,
		FromName:  "Boss Man",
		To:        recipients,
	}

	_, err = mandrillApi.MessageSend(message, false)

	if err != nil {
		fmt.Println("Error sending message")
	}
}
