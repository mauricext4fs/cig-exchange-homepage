package main

import (
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"

	"cig-exchange-homepage-backend/controllers"
	"fmt"
	"net/http"
	"os"
)

func main() {

	e := godotenv.Load()
	if e != nil {
		fmt.Print(e)
	}

	baseUri := os.Getenv("HOMEPAGE_BACKEND_BASE_URI")
	fmt.Println("Base URI set to " + baseUri)
	// For some god fucking reason using this does not work in router!!!!
	baseUri = "/invest/api/"

	router := mux.NewRouter()

	router.HandleFunc(baseUri+"ping", controllers.Ping).Methods("GET")
	router.HandleFunc(baseUri+"accounts", controllers.CreateAccount).Methods("POST")
	router.HandleFunc(baseUri+"accounts", controllers.GetAccount).Methods("GET")
	router.HandleFunc(baseUri+"verification_code", controllers.SendCode).Methods("POST")
	router.HandleFunc(baseUri+"verification_code", controllers.VerifyCode).Methods("GET")
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
