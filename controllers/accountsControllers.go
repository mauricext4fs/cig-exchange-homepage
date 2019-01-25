package controllers

import (
	"cig-exchange-libs"
	"cig-exchange-libs/models"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/mattbaird/gochimp"
	uuid "github.com/satori/go.uuid"
)

type accountResponse struct {
	UUID string `json:"uuid"`
}

type verificationCodeRequest struct {
	UUID string `json:"uuid"`
	Type string `json:"type"`
	Code string `json:"code"`
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

// CreateAccount handles POST api/accounts/signup endpoint
var CreateAccount = func(w http.ResponseWriter, r *http.Request) {

	resp := &accountResponse{}
	resp.randomUUID()

	account := &models.Account{}
	// decode account object from request body
	err := json.NewDecoder(r.Body).Decode(account)
	if err != nil {
		fmt.Println("CreateAccount: body JSON decoding error:")
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

// GetAccount handles GET api/accounts/signin endpoint
var GetAccount = func(w http.ResponseWriter, r *http.Request) {

	resp := &accountResponse{}
	resp.randomUUID()

	account := &models.Account{}
	// decode account object from request body
	err := json.NewDecoder(r.Body).Decode(account)
	if err != nil {
		fmt.Println("GetAccount: body JSON decoding error:")
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

// SendCode handles POST api/accounts/send_otp endpoint
var SendCode = func(w http.ResponseWriter, r *http.Request) {

	w.WriteHeader(204)

	reqStruct := &verificationCodeRequest{}
	// decode verificationCodeRequest object from request body
	err := json.NewDecoder(r.Body).Decode(reqStruct)
	if err != nil {
		fmt.Println("SendCode: body JSON decoding error:")
		fmt.Println(err.Error())
		return
	}

	account, err := models.GetAccount(reqStruct.UUID)
	if err != nil {
		fmt.Println("SendCode: db Lookup error:")
		fmt.Println(err.Error())
		return
	}

	// send code to email or phone number
	if reqStruct.Type == "phone" {
		twilioClient := cigExchange.GetTwilio()
		_, err = twilioClient.ReceiveOTP(account.MobileCode, account.MobileNumber)
		if err != nil {
			fmt.Println("SendCode: twillio error:")
			fmt.Println(err.Error())
		}
	} else if reqStruct.Type == "email" {
		rediskey := cigExchange.GenerateRedisKey(reqStruct.UUID)
		expiration := 5 * time.Minute

		code := cigExchange.RandCode(6)
		err = cigExchange.GetRedis().Set(rediskey, code, expiration).Err()
		if err != nil {
			fmt.Println("SendCode: redis error:")
			fmt.Println(err.Error())
			return
		}
		sendCodeInEmail(code, account.Email)
	} else {
		fmt.Println("SendCode: Error: unsupported otp type")
	}
}

func sendCodeInEmail(code, email string) {

	mandrillClient := cigExchange.GetMandrill()

	templateName := "pin-code"
	templateContent, err := mandrillClient.TemplateInfo(templateName)
	if err != nil {
		fmt.Println("sendCodeInEmail: getting template error:")
		fmt.Println(err.Error())
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

	renderedTemplate, err := mandrillClient.TemplateRender(templateName, content, merge)
	if err != nil {
		fmt.Println("sendCodeInEmail: rendering template error:")
		fmt.Println(err.Error())
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

	_, err = mandrillClient.MessageSend(message, false)
	if err != nil {
		fmt.Println("sendCodeInEmail: send email error:")
		fmt.Println(err.Error())
	}
}
