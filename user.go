package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/go-xmlfmt/xmlfmt"
)

// Write DN and User ID to Cache
func writeUserToCache(DN string, ID string) {
	//logger(1, "Writting User to DN Cache: "+ID, false)
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
	hIF.SetParam("name", user.Account.Name)
	//-- Password is base64 encoded already in process_data
	hIF.SetParam("password", user.Account.Password)
	hIF.SetParam("userType", user.Account.UserType)
	if user.Account.FirstName != "" {
		hIF.SetParam("firstName", user.Account.FirstName)
	}
	if user.Account.LastName != "" {
		hIF.SetParam("lastName", user.Account.LastName)
	}
	if user.Account.JobTitle != "" {
		hIF.SetParam("jobTitle", user.Account.JobTitle)
	}
	if user.Account.Site != "" {
		hIF.SetParam("site", user.Account.Site)
	}
	if user.Account.Phone != "" {
		hIF.SetParam("phone", user.Account.Phone)
	}
	if user.Account.Email != "" {
		hIF.SetParam("email", user.Account.Email)
	}
	if user.Account.Mobile != "" {
		hIF.SetParam("mobile", user.Account.Mobile)
	}
	//hIF.SetParam("availabilityStatus", 1)
	if user.Account.AbsenceMessage != "" {
		hIF.SetParam("absenceMessage", user.Account.AbsenceMessage)
	}
	if user.Account.TimeZone != "" {
		hIF.SetParam("timeZone", user.Account.TimeZone)
	}
	if user.Account.Language != "" {
		hIF.SetParam("language", user.Account.Language)
	}
	if user.Account.DateTimeFormat != "" {
		hIF.SetParam("dateTimeFormat", user.Account.DateTimeFormat)
	}
	if user.Account.DateFormat != "" {
		hIF.SetParam("dateFormat", user.Account.DateFormat)
	}
	if user.Account.TimeFormat != "" {
		hIF.SetParam("timeFormat", user.Account.TimeFormat)
	}
	if user.Account.CurrencySymbol != "" {
		hIF.SetParam("currencySymbol", user.Account.CurrencySymbol)
	}
	if user.Account.CountryCode != "" {
		hIF.SetParam("countryCode", user.Account.CountryCode)
	}
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

	RespBody, xmlmcErr := hIF.Invoke("admin", "userCreate")
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
	if user.Account.FirstName != "" {
		hIF.SetParam("firstName", user.Account.FirstName)
	}
	if user.Account.LastName != "" {
		hIF.SetParam("lastName", user.Account.LastName)
	}
	if user.Account.JobTitle != "" {
		hIF.SetParam("jobTitle", user.Account.JobTitle)
	}
	if user.Account.Site != "" {
		hIF.SetParam("site", user.Account.Site)
	}
	if user.Account.Phone != "" {
		hIF.SetParam("phone", user.Account.Phone)
	}
	if user.Account.Email != "" {
		hIF.SetParam("email", user.Account.Email)
	}
	if user.Account.Mobile != "" {
		hIF.SetParam("mobile", user.Account.Mobile)
	}
	//hIF.SetParam("availabilityStatus", 1)
	if user.Account.AbsenceMessage != "" {
		hIF.SetParam("absenceMessage", user.Account.AbsenceMessage)
	}
	if user.Account.TimeZone != "" {
		hIF.SetParam("timeZone", user.Account.TimeZone)
	}
	if user.Account.Language != "" {
		hIF.SetParam("language", user.Account.Language)
	}
	if user.Account.DateTimeFormat != "" {
		hIF.SetParam("dateTimeFormat", user.Account.DateTimeFormat)
	}
	if user.Account.DateFormat != "" {
		hIF.SetParam("dateFormat", user.Account.DateFormat)
	}
	if user.Account.TimeFormat != "" {
		hIF.SetParam("timeFormat", user.Account.TimeFormat)
	}
	if user.Account.CurrencySymbol != "" {
		hIF.SetParam("currencySymbol", user.Account.CurrencySymbol)
	}
	if user.Account.CountryCode != "" {
		hIF.SetParam("countryCode", user.Account.CountryCode)
	}
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

func userProfileUpdate(user *userWorkingDataStruct) (bool, error) {
	logger(1, "User Update Profile: "+user.Account.UserID, false)

	hIF.OpenElement("profileData")

	//-- Set Params based on already processed params
	hIF.SetParam("userId", user.Account.UserID)
	if user.Profile.MiddleName != "" {
		hIF.SetParam("middleName", user.Profile.MiddleName)
	}
	if user.Profile.JobDescription != "" {
		hIF.SetParam("jobDescription", user.Profile.JobDescription)
	}
	if user.Profile.Manager != "" {
		hIF.SetParam("manager", user.Profile.Manager)
	}
	if user.Profile.WorkPhone != "" {
		hIF.SetParam("workPhone", user.Profile.WorkPhone)
	}
	if user.Profile.Qualifications != "" {
		hIF.SetParam("qualifications", user.Profile.Qualifications)
	}
	if user.Profile.Interests != "" {
		hIF.SetParam("interests", user.Profile.Interests)
	}
	if user.Profile.Expertise != "" {
		hIF.SetParam("expertise", user.Profile.Expertise)
	}
	if user.Profile.Gender != "" {
		hIF.SetParam("gender", user.Profile.Gender)
	}
	if user.Profile.Dob != "" {
		hIF.SetParam("dob", user.Profile.Dob)
	}
	if user.Profile.Nationality != "" {
		hIF.SetParam("nationality", user.Profile.Nationality)
	}
	if user.Profile.Religion != "" {
		hIF.SetParam("religion", user.Profile.Religion)
	}
	if user.Profile.HomeTelephone != "" {
		hIF.SetParam("homeTelephone", user.Profile.HomeTelephone)
	}
	if user.Profile.SocialNetworkA != "" {
		hIF.SetParam("socialNetworkA", user.Profile.SocialNetworkA)
	}
	if user.Profile.SocialNetworkB != "" {
		hIF.SetParam("socialNetworkB", user.Profile.SocialNetworkB)
	}
	if user.Profile.SocialNetworkC != "" {
		hIF.SetParam("socialNetworkC", user.Profile.SocialNetworkC)
	}
	if user.Profile.SocialNetworkD != "" {
		hIF.SetParam("socialNetworkD", user.Profile.SocialNetworkD)
	}
	if user.Profile.SocialNetworkE != "" {
		hIF.SetParam("socialNetworkE", user.Profile.SocialNetworkE)
	}
	if user.Profile.SocialNetworkF != "" {
		hIF.SetParam("socialNetworkF", user.Profile.SocialNetworkF)
	}
	if user.Profile.SocialNetworkG != "" {
		hIF.SetParam("socialNetworkG", user.Profile.SocialNetworkG)
	}
	if user.Profile.SocialNetworkH != "" {
		hIF.SetParam("socialNetworkH", user.Profile.SocialNetworkH)
	}
	if user.Profile.PersonalInterests != "" {
		hIF.SetParam("personalInterests", user.Profile.PersonalInterests)
	}
	if user.Profile.HomeAddress != "" {
		hIF.SetParam("homeAddress", user.Profile.HomeAddress)
	}
	if user.Profile.PersonalBlog != "" {
		hIF.SetParam("personalBlog", user.Profile.PersonalBlog)
	}
	if user.Profile.Attrib1 != "" {
		hIF.SetParam("attrib1", user.Profile.Attrib1)
	}
	if user.Profile.Attrib2 != "" {
		hIF.SetParam("attrib2", user.Profile.Attrib2)
	}
	if user.Profile.Attrib3 != "" {
		hIF.SetParam("attrib3", user.Profile.Attrib3)
	}
	if user.Profile.Attrib4 != "" {
		hIF.SetParam("attrib4", user.Profile.Attrib4)
	}
	if user.Profile.Attrib5 != "" {
		hIF.SetParam("attrib5", user.Profile.Attrib5)
	}
	if user.Profile.Attrib6 != "" {
		hIF.SetParam("attrib6", user.Profile.Attrib6)
	}
	if user.Profile.Attrib7 != "" {
		hIF.SetParam("attrib7", user.Profile.Attrib7)
	}
	if user.Profile.Attrib8 != "" {
		hIF.SetParam("attrib8", user.Profile.Attrib8)
	}

	hIF.CloseElement("profileData")
	//-- Dry Run
	if Flags.configDryRun == true {
		var XMLSTRING = hIF.GetParam()
		XMLSTRING = xmlfmt.FormatXML(XMLSTRING, "\t", "  ")
		logger(1, "User Create XML "+XMLSTRING, false)
		hIF.ClearParam()
		return true, nil
	}

	RespBody, xmlmcErr := hIF.Invoke("admin", "userProfileSet")
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
	logger(1, "User Update Profile Success: "+user.Account.UserID, false)
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

func userGroupsUpdate(user *userWorkingDataStruct) (bool, error) {

	for groupIndex := range user.Groups {
		group := user.Groups[groupIndex]
		logger(1, "Group Add User: "+user.Account.UserID+" Group: "+group.Name, false)

		hIF.SetParam("userId", user.Account.UserID)
		hIF.SetParam("groupId", group.Id)
		hIF.SetParam("memberRole", group.Membership)
		hIF.OpenElement("options")
		hIF.SetParam("tasksView", strconv.FormatBool(group.TasksView))
		hIF.SetParam("tasksAction", strconv.FormatBool(group.TasksAction))
		hIF.CloseElement("options")

		if Flags.configDryRun == true {
			var XMLSTRING = hIF.GetParam()
			XMLSTRING = xmlfmt.FormatXML(XMLSTRING, "", "")
			logger(1, "Group Add User XML "+XMLSTRING, false)
			hIF.ClearParam()
			return true, nil
		}

		RespBody, xmlmcErr := hIF.Invoke("admin", "userAddGroup")
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
		logger(1, "Group added to User: "+user.Account.UserID, false)
	}

	return true, nil
}

func userRolesUpdate(user *userWorkingDataStruct) (bool, error) {

	hIF.SetParam("userId", user.Account.UserID)
	for roleIndex := range user.Roles {
		role := user.Roles[roleIndex]
		logger(1, "User Add Role User: "+user.Account.UserID+" Role: "+role, false)
		hIF.SetParam("role", role)
	}
	if Flags.configDryRun == true {
		var XMLSTRING = hIF.GetParam()
		XMLSTRING = xmlfmt.FormatXML(XMLSTRING, "", "")
		logger(1, "User Add Role XML "+XMLSTRING, false)
		hIF.ClearParam()
		return true, nil
	}

	RespBody, xmlmcErr := hIF.Invoke("admin", "userAddRole")
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
	logger(1, "Role added to User: "+user.Account.UserID, false)
	return true, nil
}
