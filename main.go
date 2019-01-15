package main

import (
	"github.com/gorilla/mux"

	"cig-exchange-homepage-backend/controllers"
	"fmt"
	"net/http"
)

func main() {

	baseUri := "/invest/api/"

	router := mux.NewRouter()

	router.HandleFunc(baseUri+"ping", controllers.Ping).Methods("GET")
	router.HandleFunc(baseUri+"user/new", controllers.CreateAccount).Methods("POST")
	router.HandleFunc(baseUri+"user/login", controllers.Authenticate).Methods("POST")
	router.HandleFunc(baseUri+"sendcode", controllers.SendVerificationCodeByEmail).Methods("POST")
	router.HandleFunc(baseUri+"contacts/new", controllers.CreateContact).Methods("POST")
	router.HandleFunc(baseUri+"me/contacts", controllers.GetContactsFor).Methods("GET") //  user/2/contacts
	router.HandleFunc(baseUri+"contact_us", controllers.SendContactUsEmail).Methods("POST")

	//attach JWT auth middleware
	//router.Use(app.JwtAuthentication)

	//router.NotFoundHandler = app.NotFoundHandler

	// We always run in docker... for sack of convenience let's always use port 80
	port := "80"

	fmt.Println(port)

	err := http.ListenAndServe(":"+port, router) //Launch the app, visit localhost:8000/api
	if err != nil {
		fmt.Print(err)
	}
}
