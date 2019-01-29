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
	router.HandleFunc(baseURI+"users/signup", controllers.CreateUser).Methods("POST")
	router.HandleFunc(baseURI+"users/signin", controllers.GetUser).Methods("POST")
	router.HandleFunc(baseURI+"users/send_otp", controllers.SendCode).Methods("POST")
	router.HandleFunc(baseURI+"users/verify_otp", controllers.VerifyCode).Methods("POST")
	router.HandleFunc(baseURI+"contact_us", controllers.SendContactUsEmail).Methods("POST")

	//attach JWT auth middleware
	//router.Use(app.JwtAuthentication)

	//router.NotFoundHandler = app.NotFoundHandler

	port := os.Getenv("DOCKER_LISTEN_DEFAULT_PORT")

	fmt.Println(port)

	err := http.ListenAndServe(":"+port, router) //Launch the app, visit localhost:80
	if err != nil {
		fmt.Print(err)
	}
}
