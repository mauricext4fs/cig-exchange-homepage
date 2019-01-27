# License


See LICENSE

If you like this please donate at mc at courtoisconsulting.com

If you have any custom wishes we can make it happens just contact us.

# Authors

This project is based on github.com/adigunhammedolalekan/go-contacts.git
and is owned and developed by the team at Courtois Consulting of Switzelrand 
for the CIG Exchange platform. 

For any licensing and / or support question please contact CIG Exchange directly at: 
https://cig-exchange.ch

# Info

CIG Exchange Homepage is the full backend use by https://www.cig-exchange.ch.

It provide the following functionality :

 - Set of REST API to register and login

You can literaly use this as the backend for any website to register and login users 
with JWT using no password and only 2nd Factor autentication see https://fidoalliance.org/.


# JWT Handling

The JWT is validated on the Reverse Proxy using Nginx custom module... this is not 
implemented inside this code. This repo is not JWT protected since it provide 
endpoints that aren't secure

# Authentication

Authentication is 2nd factor required... we don't use password. The user must 
enter a verification code sent into his Email or by SMS to be able to login

## Vision

For CIG Exchange Authentication we will do entirely away with the usual bad and 
insecure user / password combination and use purely 2 Factor authentication.

At the beginning we will be using email and sms but nothing stop us from integrating 
google authenticator or even some own custom Bank 3rd factor authentication.  

### How it works?

The login form will have only one field... the user will either choose to enter an 
email address or a mobile number. 

When the user enter his email / mobile we must verify that there is a match in the database 
and proceed immediately (if there is a match) to send him a one time code either by email or directly by SMS. In 
that message we must warn the user that if they did not ask to receive this to report it as a
it immediately. 

The user must then enter the correct verification. If the code match up we create the JWT 
and the user is considered authenticated. 


## Users type

For now there will be two type of user:

- Super Admin
- Platform

### Super Admin

The Super Admin can for the moment do the following:

- Create new user
- Send Invite Email to the new user that include a temporary jwt 
directly so it is logged in au automatically


### Platform user

The platform user can for the moment do the following:

- Change his information


## Organisation Model

- Organisation
    - uuid
    - info uuid

- Organisation Contacts
    - uuid
    - email

## User Model


- User
    - uuid
    - organisation_uuid
    - info uuid
    - contact uuid[]
    - login contact uuid

- Info
    - uuid
    - name (platform name)
    - lastname
    - homepage
    
- Login
    - uuid
    - contact uuid

- Contact
    - uuid
    - level (primary, secondary)
    - location (home, work)
    - type (telefon, address)
    - subtype (mobile, fix)
    - value1
    - value2
    - value3
    - value4
    - value5
    - value6

