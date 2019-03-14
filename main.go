package main

import (
	"cig-exchange-homepage-backend/controllers"
	"cig-exchange-libs"
	"cig-exchange-libs/auth"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
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

	// List of endpoints that doesn't require auth
	skipJWT := []string{
		baseURI + "ping",
		baseURI + "users/signup",
		baseURI + "users/signin",
		baseURI + "users/send_otp",
		baseURI + "users/verify_otp",
		baseURI + "offerings",
		baseURI + "contact_us",
	}
	userAPI := auth.UserAPI{
		SkipJWT: skipJWT,
	}

	router.HandleFunc(baseURI+"ping", controllers.Ping).Methods("GET")
	router.HandleFunc(baseURI+"users/signup", userAPI.CreateUserHandler).Methods("POST")
	router.HandleFunc(baseURI+"users/signin", userAPI.GetUserHandler).Methods("POST")
	router.HandleFunc(baseURI+"users/send_otp", userAPI.SendCodeHandler).Methods("POST")
	router.HandleFunc(baseURI+"users/verify_otp", userAPI.VerifyCodeHandler).Methods("POST")
	router.HandleFunc(baseURI+"offerings", controllers.GetAllOfferings).Methods("GET")
	router.HandleFunc(baseURI+"contact_us", controllers.SendContactUsEmail).Methods("POST")

	// dev environment api call to get signup code
	if cigExchange.IsDevEnv() {
		fmt.Println("DEV ENVIRONMENT!!! users/code api call enabled")
		router.HandleFunc(baseURI+"users/code", func(w http.ResponseWriter, r *http.Request) {
			type input struct {
				UUID string `json:"uuid"`
			}
			var in input
			err := json.NewDecoder(r.Body).Decode(&in)
			if err != nil {
				cigExchange.RespondWithError(w, 422, fmt.Errorf("Invalid request"))
				return
			}
			if len(in.UUID) == 0 {
				cigExchange.RespondWithError(w, 422, fmt.Errorf("Empty uuid parameter"))
				return
			}

			rediskey := cigExchange.GenerateRedisKey(in.UUID)
			redisCmd := cigExchange.GetRedis().Get(rediskey)
			if redisCmd.Err() != nil {
				cigExchange.RespondWithError(w, 422, fmt.Errorf("Redis error"))
				return
			}
			resp := make(map[string]string, 0)
			resp["code"] = redisCmd.Val()
			cigExchange.Respond(w, resp)
		}).Methods("POST")
	}

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
