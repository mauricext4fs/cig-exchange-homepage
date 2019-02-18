package main

import (
	"encoding/json"
	"fmt"
	"time"

	"cig-exchange-libs"
	"cig-exchange-libs/models"

	"github.com/lib/pq"
	"github.com/snikch/goodman/hooks"
	trans "github.com/snikch/goodman/transaction"
)

const dredd = "dredd"

func main() {

	h := hooks.NewHooks()
	server := hooks.NewServer(hooks.NewHooksRunner(h))

	redisClient := cigExchange.GetRedis()
	dbClient := cigExchange.GetDB()

	// save created user UUID and JWT and organization UUID
	userUUID := ""
	userJWT := ""
	orgUUID := ""

	// prepare the database:
	// 1. delete 'dredd' users if it exists (first name  = 'dredd')
	// 2. delete 'dredd' organisations if it exists (reference key = 'dredd')
	// 3. create 'dredd' organisation (user will be registered with it)
	// 4. create some offerings belonging to 'dredd' organization
	// 5. verify that created offerings are present in 'invest/offerings' api call

	// delete 'dredd' users if it exists (first name  = 'dredd')
	usersDelete := make([]models.User, 0)
	err := dbClient.Where(&models.User{Name: dredd}).Find(&usersDelete).Error
	if err == nil {
		for _, u := range usersDelete {
			dbClient.Delete(&u)
		}
	}

	// delete 'dredd' organisations if it exists (reference key = 'dredd')
	orgsDelete := make([]models.Organisation, 0)
	err = dbClient.Where(&models.Organisation{ReferenceKey: dredd}).Find(&orgsDelete).Error
	if err == nil {
		for _, o := range orgsDelete {
			dbClient.Delete(&o)
		}
	}

	// create 'dredd' organisation
	org := models.Organisation{
		Name:         dredd,
		ReferenceKey: dredd,
	}
	err = dbClient.Create(&org).Error
	if err != nil {
		fmt.Println("ERROR: prepareDatabase: create organisation:")
		fmt.Println(err.Error())
	}
	orgUUID = org.ID

	// create some offerings belonging to 'dredd' organization
	offering := models.Offering{
		Title:          dredd,
		OrganisationID: orgUUID,
		Type:           make(pq.StringArray, 0),
		IsVisible:      true,
	}
	err = dbClient.Create(&offering).Error
	if err != nil {
		fmt.Println("ERROR: prepareDatabase: create offering:")
		fmt.Println(err.Error())
	}

	h.After("Offerings > invest/api/offerings > Offerings", func(t *trans.Transaction) {

		// happens when api is down
		if t.Real == nil {
			return
		}

		// verify that created offerings are present in 'invest/offerings' api call
		offerings := make([]models.Offering, 0)
		err := json.Unmarshal([]byte(t.Real.Body), &offerings)
		if err != nil {
			t.Fail = fmt.Sprintf("Unable to parse response: %v", err.Error())
			return
		}

		for _, offering := range offerings {
			if offering.Title == dredd {
				// we found a match, api works fine
				return
			}
		}

		t.Fail = "Pre-created offering is missing"
	})

	h.Before("Users > invest/api/users/signup > Signup", func(t *trans.Transaction) {

		if t.Request == nil {
			return
		}
		if len(orgUUID) == 0 {
			t.Fail = "Organisation UUID missing"
			return
		}

		setBodyValue(&t.Request.Body, "reference_key", dredd)
		setBodyValue(&t.Request.Body, "name", dredd)
	})

	h.After("Users > invest/api/users/signin > Signin", func(t *trans.Transaction) {

		// happens when api is down
		if t.Real == nil {
			return
		}

		userUUID = getBodyValue(&t.Real.Body, "uuid")
		if len(userUUID) == 0 {
			t.Fail = "Unable to save user UUID"
			return
		}
	})

	h.Before("Users > invest/api/users/send_otp > Send OTP", func(t *trans.Transaction) {

		if t.Request == nil {
			return
		}
		if len(userUUID) == 0 {
			t.Fail = "User UUID missing"
			return
		}

		setBodyValue(&t.Request.Body, "uuid", userUUID)
	})

	h.Before("Users > invest/api/users/verify_otp > Verify OTP", func(t *trans.Transaction) {

		if t.Request == nil {
			return
		}
		if len(userUUID) == 0 {
			t.Fail = "User UUID missing"
			return
		}

		rediskey := fmt.Sprintf("%s_signup_key", userUUID)
		redisCmd := redisClient.Get(rediskey)
		if redisCmd.Err() != nil {
			t.Fail = fmt.Sprintf("Redis error: %v", redisCmd.Err().Error())
			return
		}

		setBodyValue(&t.Request.Body, "uuid", userUUID)
		setBodyValue(&t.Request.Body, "code", redisCmd.Val())
	})

	h.After("Users > invest/api/users/verify_otp > Verify OTP", func(t *trans.Transaction) {

		// happens when api is down
		if t.Real == nil {
			return
		}

		userJWT = getBodyValue(&t.Real.Body, "jwt")
		if len(userJWT) == 0 {
			t.Fail = "Unable to save user JWT"
			return
		}

		// save jwt and organization uuid in redis for p2p backend tests
		expiration := 5 * time.Minute
		err = redisClient.Set("jwt", userJWT, expiration).Err()
		if err != nil {
			t.Fail = fmt.Sprintf("Redis error: %v", err.Error())
			return
		}
		err = redisClient.Set("org", orgUUID, expiration).Err()
		if err != nil {
			t.Fail = fmt.Sprintf("Redis error: %v", err.Error())
			return
		}
	})

	server.Serve()
	defer server.Listener.Close()
}

func setBodyValue(body *string, key, value string) {

	if body == nil {
		return
	}

	bodyMap := map[string]interface{}{}

	err := json.Unmarshal([]byte(*body), &bodyMap)
	if err != nil {
		return
	}

	bodyMap[key] = value
	b, err := json.Marshal(bodyMap)
	if err != nil {
		return
	}

	*body = string(b)
}

func getBodyValue(body *string, key string) (value string) {

	if body == nil {
		return
	}

	bodyMap := map[string]interface{}{}

	err := json.Unmarshal([]byte(*body), &bodyMap)
	if err != nil {
		return
	}

	v, ok := bodyMap[key]
	if ok {
		// make sure it's a string
		vs, ok := v.(string)
		if ok {
			value = vs
		}
	}

	return
}
