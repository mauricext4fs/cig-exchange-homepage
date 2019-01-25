package main

import (
	"strings"

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

	baseURI := os.Getenv("HOMEPAGE_BACKEND_BASE_URI")
	// cut double quote symbols
	baseURI = strings.Replace(baseURI, "\"", "", -1)
	fmt.Println("Base URI set to " + baseURI)

	router := mux.NewRouter()

	router.HandleFunc(baseURI+"ping", controllers.Ping).Methods("GET")
	router.HandleFunc(baseURI+"accounts/signup", controllers.CreateAccount).Methods("POST")
	router.HandleFunc(baseURI+"accounts/signin", controllers.GetAccount).Methods("POST")
	router.HandleFunc(baseURI+"accounts/send_otp", controllers.SendCode).Methods("POST")
	router.HandleFunc(baseURI+"contact_us", controllers.SendContactUsEmail).Methods("POST")

	//attach JWT auth middleware
	//router.Use(app.JwtAuthentication)

	//router.NotFoundHandler = app.NotFoundHandler

	// We always run in docker... for sack of convenience let's always use port 80
	port := "80"

	fmt.Println(port)

	err := http.ListenAndServe(":"+port, router) //Launch the app, visit localhost:80
	if err != nil {
		fmt.Print(err)
	}
}
