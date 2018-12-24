package controllers

import (
	u "cig-exchange-libs/utils"
	"net/http"
)

var Ping = func(w http.ResponseWriter, r *http.Request) {

	resp := u.Message(true, "Pong!")
	u.Respond(w, resp)
}
