package main

import (
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"

	"os"
	"cig-exchange-homepage-backend/controllers"
    "strings"
	"fmt"
	"net/http"
)

func main() {

    e := godotenv.Load()
	if e != nil {
		fmt.Print(e)
	}

	baseUri := os.Getenv("HOMEPAGE_BACKEND_BASE_URI")
    baseUri = strings.Replace(baseUri, "\"", "", -1)
	fmt.Println("Base URI set to " + baseUri)
    // For some god fucking reason using this does not work in router!!!!
    baseUri = "/invest/api/"

	router := mux.NewRouter()

	router.HandleFunc(baseUri+"ping", controllers.Ping).Methods("GET")
    fmt.Println("ping uri? " + baseUri+"ping")
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
