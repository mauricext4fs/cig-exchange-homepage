package controllers

import (
	"cig-exchange-homepage-backend/app"
	"cig-exchange-libs"
	"cig-exchange-libs/models"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/mattbaird/gochimp"
	uuid "github.com/satori/go.uuid"
)

type userResponse struct {
	UUID string `json:"uuid"`
}

type verifyCodeResponse struct {
	JWT string `json:"jwt"`
}

type verificationCodeRequest struct {
	UUID string `json:"uuid"`
	Type string `json:"type"`
	Code string `json:"code"`
}

// UserRequest is a structure to represent the signup api request
type UserRequest struct {
	Sex              string `json:"sex"`
	Name             string `json:"name"`
	LastName         string `json:"lastname"`
	Email            string `json:"email"`
	PhoneCountryCode string `json:"phone_country_code"`
	PhoneNumber      string `json:"phone_number"`
}

func (user *UserRequest) convertRequestToUser() *models.User {
	mUser := &models.User{}

	mUser.Sex = user.Sex
	mUser.Role = "Platform"
	mUser.Name = user.Name
	mUser.LastName = user.LastName

	mUser.LoginEmail = models.Contact{Type: "email", Level: "primary", Value1: user.Email}
	mUser.LoginPhone = models.Contact{Type: "phone", Level: "secondary", Value1: user.PhoneCountryCode, Value2: user.PhoneNumber}

	return mUser
}

func (resp *userResponse) randomUUID() {
	UUID, err := uuid.NewV4()
	if err != nil {
		// uuid for an unlikely event of NewV4 failure
		resp.UUID = "fdb283d4-7341-4517-b501-371d22d27cfc"
		return
	}
	resp.UUID = UUID.String()
}

// CreateUser handles POST api/users/signup endpoint
var CreateUser = func(w http.ResponseWriter, r *http.Request) {

	resp := &userResponse{}
	resp.randomUUID()

	userReq := &UserRequest{}
	// decode user object from request body
	err := json.NewDecoder(r.Body).Decode(userReq)
	if err != nil {
		fmt.Println("CreateUser: body JSON decoding error:")
		fmt.Println(err.Error())
		cigExchange.Respond(w, resp)
		return
	}

	user := userReq.convertRequestToUser()

	// try to create user
	err = user.Create()
	if err != nil {
		fmt.Println("CreateUser: db Create error:")
		fmt.Println(err.Error())
		cigExchange.Respond(w, resp)
		return
	}
	resp.UUID = user.ID
	cigExchange.Respond(w, resp)
}

// GetUser handles GET api/users/signin endpoint
var GetUser = func(w http.ResponseWriter, r *http.Request) {

	resp := &userResponse{}
	resp.randomUUID()

	userReq := &UserRequest{}
	// decode user object from request body
	err := json.NewDecoder(r.Body).Decode(userReq)
	if err != nil {
		fmt.Println("GetUser: body JSON decoding error:")
		fmt.Println(err.Error())
		cigExchange.Respond(w, resp)
		return
	}

	user := &models.User{}
	// login using email or phone number
	if len(userReq.Email) > 0 {
		user, err = models.GetUserByEmail(userReq.Email)
	} else if len(userReq.PhoneCountryCode) > 0 && len(userReq.PhoneNumber) > 0 {
		user, err = models.GetUserByMobile(userReq.PhoneCountryCode, userReq.PhoneNumber)
	} else {
		fmt.Println("GetUser: neither email or mobile number specified in post body")
		cigExchange.Respond(w, resp)
		return
	}

	if err != nil {
		fmt.Println("GetUser: db Lookup error:")
		fmt.Println(err.Error())
		cigExchange.Respond(w, resp)
		return
	}
	resp.UUID = user.ID
	cigExchange.Respond(w, resp)
}

// SendCode handles POST api/users/send_otp endpoint
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

	user, err := models.GetUser(reqStruct.UUID)
	if err != nil {
		fmt.Println("SendCode: db Lookup error:")
		fmt.Println(err.Error())
		return
	}

	// send code to email or phone number
	if reqStruct.Type == "phone" {
		twilioClient := cigExchange.GetTwilio()
		_, err = twilioClient.ReceiveOTP(user.LoginPhone.Value1, user.LoginPhone.Value2)
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
		sendCodeInEmail(code, user.LoginEmail.Value1)
	} else {
		fmt.Println("SendCode: Error: unsupported otp type")
	}
}

// VerifyCode handles GET api/users/verify_otp endpoint
var VerifyCode = func(w http.ResponseWriter, r *http.Request) {

	retErr := fmt.Errorf("Invalid code")
	retCode := 401

	reqStruct := &verificationCodeRequest{}
	// decode verificationCodeRequest object from request body
	err := json.NewDecoder(r.Body).Decode(reqStruct)
	if err != nil {
		fmt.Println("VerifyCode: body JSON decoding error:")
		fmt.Println(err.Error())
		cigExchange.RespondWithError(w, retCode, retErr)
		return
	}

	user, err := models.GetUser(reqStruct.UUID)
	if err != nil {
		fmt.Println("VerifyCode: db Lookup error:")
		fmt.Println(err.Error())
		cigExchange.RespondWithError(w, retCode, retErr)
		return
	}

	// verify code
	if reqStruct.Type == "phone" {
		twilioClient := cigExchange.GetTwilio()
		_, err := twilioClient.VerifyOTP(reqStruct.Code, user.LoginPhone.Value1, user.LoginPhone.Value2)
		if err != nil {
			fmt.Println("VerifyCode: twillio error:")
			fmt.Println(err.Error())
			cigExchange.RespondWithError(w, retCode, retErr)
			return
		}

	} else if reqStruct.Type == "email" {
		rediskey := cigExchange.GenerateRedisKey(reqStruct.UUID)

		redisCmd := cigExchange.GetRedis().Get(rediskey)
		if redisCmd.Err() != nil {
			fmt.Println("VerifyCode: redis error:")
			fmt.Println(err.Error())
			cigExchange.RespondWithError(w, retCode, retErr)
			return
		}
		if redisCmd.Val() != reqStruct.Code {
			fmt.Println("VerifyCode: code mismatch, expecting " + redisCmd.Val())
			cigExchange.RespondWithError(w, retCode, retErr)
			return
		}
	} else {
		fmt.Println("VerifyCode: Error: unsupported otp type")
		cigExchange.RespondWithError(w, retCode, retErr)
		return
	}

	// verification passed, generate jwt and return it
	tk := &app.Token{UserUUID: user.ID}
	token := jwt.NewWithClaims(jwt.GetSigningMethod("HS256"), tk)
	tokenString, err := token.SignedString([]byte(os.Getenv("token_password")))
	if err != nil {
		fmt.Println("VerifyCode: jwt generation failed:")
		fmt.Println(err.Error())
		cigExchange.RespondWithError(w, retCode, retErr)
		return
	}

	resp := &verifyCodeResponse{JWT: tokenString}
	cigExchange.Respond(w, resp)
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
