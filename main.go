package main

import (
	"cig-exchange-libs/auth"
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

	userAPI := auth.NewUserAPI(auth.PlatformTrading, baseURI)

	router.HandleFunc(baseURI+"ping", controllers.Ping).Methods("GET")
	router.HandleFunc(baseURI+"users/signup", userAPI.CreateUserHandler).Methods("POST")
	router.HandleFunc(baseURI+"users/signin", userAPI.GetUserHandler).Methods("POST")
	router.HandleFunc(baseURI+"users/send_otp", userAPI.SendCodeHandler).Methods("POST")
	router.HandleFunc(baseURI+"users/verify_otp", userAPI.VerifyCodeHandler).Methods("POST")
	router.HandleFunc(baseURI+"contact_us", controllers.SendContactUsEmail).Methods("POST")

	// attach JWT auth middleware
	router.Use(userAPI.JwtAuthenticationHandler)

	//router.NotFoundHandler = app.NotFoundHandler

	port := os.Getenv("DOCKER_LISTEN_DEFAULT_PORT")

	fmt.Println(port)

	err := http.ListenAndServe(":"+port, router) //Launch the app, visit localhost:80
	if err != nil {
		fmt.Print(err)
	}
}
