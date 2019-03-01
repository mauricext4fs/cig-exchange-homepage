package controllers

import (
	"cig-exchange-libs"
	"cig-exchange-libs/models"
	"net/http"
)

type offeringsReponse struct {
	*models.Offering
	OrganisationName string `json:"organisation"`
	OrganisationURL  string `json:"organisation_website"`
}

func convertOffering(offering *models.Offering) *offeringsReponse {

	response := &offeringsReponse{}
	response.Offering = offering
	response.OrganisationName = offering.Organisation.Name
	response.OrganisationURL = offering.Organisation.Website
	return response
}

// GetOfferings handles GET api/offerings endpoint
var GetOfferings = func(w http.ResponseWriter, r *http.Request) {

	// query all offerings from db
	offerings, err := models.GetOfferings()
	if err != nil {
		cigExchange.RespondWithError(w, 500, err)
		return
	}

	// add organisation name to offerings structs
	respOfferings := make([]*offeringsReponse, 0)
	for _, offering := range offerings {
		if offering.IsVisible {
			respOfferings = append(respOfferings, convertOffering(offering))
		}
	}

	cigExchange.Respond(w, respOfferings)
}
