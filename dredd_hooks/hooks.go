package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/go-redis/redis"
	"github.com/joho/godotenv"
	"github.com/snikch/goodman/hooks"
	trans "github.com/snikch/goodman/transaction"
)

func main() {

	h := hooks.NewHooks()
	server := hooks.NewServer(hooks.NewHooksRunner(h))

	err := godotenv.Load()
	if err != nil {
		fmt.Print(err)
	}

	redisHost := os.Getenv("REDIS_HOST")
	redisPort := os.Getenv("REDIS_PORT")

	client := redis.NewClient(&redis.Options{
		Addr:     redisHost + ":" + redisPort,
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	_, err = client.Ping().Result()
	if err != nil {
		log.Print(err.Error())
	}

	// save created user UUID and JWT
	userUUID := ""
	userJWT := ""

	h.After("Users > invest/api/users/signin > Signin", func(t *trans.Transaction) {
		// happens when api is down
		if t.Real == nil {
			return
		}

		body := map[string]interface{}{}

		err := json.Unmarshal([]byte(t.Real.Body), &body)
		if err != nil {
			log.Printf("Signin: Error: %v", err.Error())
			return
		}

		createdUUIDInterface, ok := body["uuid"]
		if !ok {
			log.Printf("Signin: Error: Unable to save signin uuid")
			return
		}
		userUUID = createdUUIDInterface.(string)
	})

	h.Before("Users > invest/api/users/send_otp > Send OTP", func(t *trans.Transaction) {

		if len(userUUID) == 0 || t.Request == nil {
			return
		}

		body := map[string]interface{}{}

		err := json.Unmarshal([]byte(t.Request.Body), &body)
		if err != nil {
			log.Printf("Send OTP: Error: %v", err.Error())
			return
		}

		body["uuid"] = userUUID
		b, err := json.Marshal(body)
		if err != nil {
			log.Printf("Send OTP: Error: %v", err.Error())
			return
		}

		t.Request.Body = string(b)
	})

	h.Before("Users > invest/api/users/verify_otp > Verify OTP", func(t *trans.Transaction) {

		if len(userUUID) == 0 || t.Request == nil {
			return
		}

		body := map[string]interface{}{}

		err := json.Unmarshal([]byte(t.Request.Body), &body)
		if err != nil {
			log.Printf("Verify OTP: Error: %v", err.Error())
			return
		}

		rediskey := fmt.Sprintf("%s_signup_key", userUUID)
		redisCmd := client.Get(rediskey)
		if redisCmd.Err() != nil {
			log.Printf("Verify OTP: Error: %v", redisCmd.Err().Error())
			return
		}

		body["uuid"] = userUUID
		body["code"] = redisCmd.Val()
		b, err := json.Marshal(body)
		if err != nil {
			log.Printf("Verify OTP: Error: %v", err.Error())
			return
		}

		t.Request.Body = string(b)
	})

	h.After("Users > invest/api/users/verify_otp > Verify OTP", func(t *trans.Transaction) {

		// happens when api is down
		if t.Real == nil {
			return
		}

		body := map[string]interface{}{}

		err := json.Unmarshal([]byte(t.Real.Body), &body)
		if err != nil {
			log.Printf("Verify OTP: Error: %v", err.Error())
			return
		}
		userJWTInterface, ok := body["jwt"]
		if !ok {
			log.Printf("Verify OTP: Error: Unable to save user jwt")
			return
		}
		userJWT = userJWTInterface.(string)
		log.Printf("jwt: %v", userJWT)

		// save jwt in redis for p2p backend tests
		expiration := 5 * time.Minute
		err = client.Set("jwt", userJWT, expiration).Err()
		if err != nil {
			fmt.Println("Verify OTP: redis error:")
			fmt.Println(err.Error())
			return
		}
	})

	server.Serve()
	defer server.Listener.Close()
}
