package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/go-xmlfmt/xmlfmt"
)

// Write DN and User ID to Cache
func writeUserToCache(DN string, ID string) {
	logger(1, "Writting User to DN Cache: "+ID, false)
	_, found := HornbillCache.DN[DN]
	if !found {
		HornbillCache.DN[DN] = ID
	}
}

// Get User ID From Cache By DN
func getUserFromDNCache(DN string) string {
	_, found := HornbillCache.DN[DN]
	if found {
		return HornbillCache.DN[DN]
	}
	return ""
}

func userCreate(user *userWorkingDataStruct) (bool, error) {
	logger(1, "User Create: "+user.Account.UserID, false)
	//-- Set Params based on already processed params
	hIF.SetParam("userId", user.Account.UserID)
	hIF.SetParam("userType", user.Account.UserType)
	hIF.SetParam("name", user.Account.Name)
	//-- Password is base64 encoded already in process_data
	hIF.SetParam("password", user.Account.Password)
	hIF.SetParam("firstName", user.Account.FirstName)
	hIF.SetParam("lastName", user.Account.LastName)
	hIF.SetParam("jobTitle", user.Account.JobTitle)
	hIF.SetParam("site", user.Account.Site)
	hIF.SetParam("phone", user.Account.Phone)
	hIF.SetParam("email", user.Account.Email)
	hIF.SetParam("mobile", user.Account.Mobile)
	//hIF.SetParam("availabilityStatus", 1)
	hIF.SetParam("absenceMessage", user.Account.AbsenceMessage)
	hIF.SetParam("timeZone", user.Account.TimeZone)
	hIF.SetParam("language", user.Account.Language)
	hIF.SetParam("dateTimeFormat", user.Account.DateTimeFormat)
	hIF.SetParam("dateFormat", user.Account.DateFormat)
	hIF.SetParam("timeFormat", user.Account.TimeFormat)
	hIF.SetParam("currencySymbol", user.Account.CurrencySymbol)
	hIF.SetParam("countryCode", user.Account.CountryCode)
	//hIF.SetParam("notifyEmail", "")
	//hIF.SetParam("notifyTextMessage", "")

	//-- Dry Run
	if Flags.configDryRun == true {
		var XMLSTRING = hIF.GetParam()
		XMLSTRING = xmlfmt.FormatXML(XMLSTRING, "\t", "  ")
		logger(1, "User Create XML "+XMLSTRING, false)
		hIF.ClearParam()
		return true, nil
	}

	RespBody, xmlmcErr := hIF.Invoke("admin", "userUpdate")
	var JSONResp xmlmcResponse
	if xmlmcErr != nil {
		return false, xmlmcErr
	}
	err := json.Unmarshal([]byte(RespBody), &JSONResp)
	if err != nil {
		return false, err
	}
	if JSONResp.State.Error != "" {
		return false, errors.New(JSONResp.State.Error)
	}
	logger(1, "User Create Success: "+user.Account.UserID, false)
	return true, nil
}

func userUpdate(user *userWorkingDataStruct) (bool, error) {

	logger(1, "User Update: "+user.Account.UserID, false)
	//-- Set Params based on already processed params
	hIF.SetParam("userId", user.Account.UserID)
	hIF.SetParam("userType", user.Account.UserType)
	hIF.SetParam("name", user.Account.Name)
	//hIF.SetParam("password", user.Password)
	hIF.SetParam("firstName", user.Account.FirstName)
	hIF.SetParam("lastName", user.Account.LastName)
	hIF.SetParam("jobTitle", user.Account.JobTitle)
	hIF.SetParam("site", user.Account.Site)
	hIF.SetParam("phone", user.Account.Phone)
	hIF.SetParam("email", user.Account.Email)
	hIF.SetParam("mobile", user.Account.Mobile)
	//hIF.SetParam("availabilityStatus", 1)
	hIF.SetParam("absenceMessage", user.Account.AbsenceMessage)
	hIF.SetParam("timeZone", user.Account.TimeZone)
	hIF.SetParam("language", user.Account.Language)
	hIF.SetParam("dateTimeFormat", user.Account.DateTimeFormat)
	hIF.SetParam("dateFormat", user.Account.DateFormat)
	hIF.SetParam("timeFormat", user.Account.TimeFormat)
	hIF.SetParam("currencySymbol", user.Account.CurrencySymbol)
	hIF.SetParam("countryCode", user.Account.CountryCode)
	//hIF.SetParam("notifyEmail", "")
	//hIF.SetParam("notifyTextMessage", "")

	//-- Dry Run
	if Flags.configDryRun == true {
		var XMLSTRING = hIF.GetParam()
		XMLSTRING = xmlfmt.FormatXML(XMLSTRING, "", "")
		logger(1, "User Update XML "+XMLSTRING, false)
		hIF.ClearParam()
		return true, nil
	}

	RespBody, xmlmcErr := hIF.Invoke("admin", "userUpdate")
	var JSONResp xmlmcResponse
	if xmlmcErr != nil {
		return false, xmlmcErr
	}
	err := json.Unmarshal([]byte(RespBody), &JSONResp)
	if err != nil {
		return false, err
	}
	if JSONResp.State.Error != "" {
		return false, errors.New(JSONResp.State.Error)
	}
	logger(1, "User Update Success: "+user.Account.UserID, false)
	return true, nil
}

func userImageUpdate(user *userWorkingDataStruct) (bool, error) {
	//- Profile Images are already in cache as Bytes
	logger(1, "User Proflile Image Set: "+user.Account.UserID, false)
	//WebDAV upload
	image := HornbillCache.Images[user.ImageURI]
	value := ""
	relLink := "session/" + user.Account.UserID + "." + ldapImportConf.User.Image.ImageType
	strDAVurl := hIF.DavEndpoint + relLink

	strContentType := "image/jpeg"
	if ldapImportConf.User.Image.ImageType != "jpg" {
		strContentType = "image/png"
	}

	logger(1, "DAV Upload URL: "+fmt.Sprintf("%s", strDAVurl), false)

	if Flags.configDryRun != true {

		if len(image.imageBytes) > 0 {
			putbody := bytes.NewReader(image.imageBytes)
			req, Perr := http.NewRequest("PUT", strDAVurl, putbody)
			if Perr != nil {
				return false, Perr
			}
			req.Header.Set("Content-Type", strContentType)
			req.Header.Add("Authorization", "ESP-APIKEY "+Flags.configApiKey)
			req.Header.Set("User-Agent", "Go-http-client/1.1")
			response, Perr := client.Do(req)
			if Perr != nil {
				return false, Perr
			}
			defer response.Body.Close()
			_, _ = io.Copy(ioutil.Discard, response.Body)
			if response.StatusCode == 201 || response.StatusCode == 200 {
				value = "/" + relLink
			}
		} else {
			logger(1, "Unable to Uplaod Profile Image to Dav as its empty", false)
			return true, nil
		}
	}

	logger(1, "Profile Set Image URL: "+fmt.Sprintf("%s", value), false)
	hIF.SetParam("objectRef", "urn:sys:user:"+user.Account.UserID)
	hIF.SetParam("sourceImage", value)

	if Flags.configDryRun == true {
		var XMLSTRING = hIF.GetParam()
		XMLSTRING = xmlfmt.FormatXML(XMLSTRING, "", "")
		logger(1, "Profile Image Set XML "+XMLSTRING, false)
		hIF.ClearParam()
		return true, nil
	}

	RespBody, xmlmcErr := hIF.Invoke("activity", "profileImageSet")
	var JSONResp xmlmcResponse
	if xmlmcErr != nil {
		return false, xmlmcErr
	}
	err := json.Unmarshal([]byte(RespBody), &JSONResp)
	if err != nil {
		return false, err
	}
	if JSONResp.State.Error != "" {
		return false, errors.New(JSONResp.State.Error)
	}
	logger(1, "Image added to User: "+user.Account.UserID, false)
	return true, nil
}

func userProfileUpdate(user *userWorkingDataStruct) (bool, error) {
	return true, nil
}

func userGroupsUpdate(user *userWorkingDataStruct) (bool, error) {
	return true, nil
}

func userRolesUpdate(user *userWorkingDataStruct) (bool, error) {
	return true, nil
}
