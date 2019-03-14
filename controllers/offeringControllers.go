package controllers

import (
	"cig-exchange-libs"
	"cig-exchange-libs/models"
	"net/http"
)

// GetAllOfferings handles GET api/offerings endpoint
// does not perform JWT based organisation filtering
var GetAllOfferings = func(w http.ResponseWriter, r *http.Request) {

	// query all offerings from db
	offerings, err := models.GetOfferings()
	if err != nil {
		cigExchange.RespondWithError(w, 500, err)
		return
	}

	// extended response with organisation and org website
	type offeringsReponse struct {
		*models.Offering
		OrganisationName string `json:"organisation"`
		OrganisationURL  string `json:"organisation_website"`
	}

	// add organisation name to offerings structs
	respOfferings := make([]*offeringsReponse, 0)
	for _, offering := range offerings {
		if offering.IsVisible {
			respOffering := &offeringsReponse{}
			respOffering.Offering = offering
			respOffering.OrganisationName = offering.Organisation.Name
			respOffering.OrganisationURL = offering.Organisation.Website
			respOfferings = append(respOfferings, respOffering)
		}
	}

	cigExchange.Respond(w, respOfferings)
}
