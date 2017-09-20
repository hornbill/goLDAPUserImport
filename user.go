package main

import (
	"bytes"
	"encoding/base64"
	"encoding/xml"
	"errors"
	"fmt"
	"strings"

	"github.com/hornbill/goApiLib"
	"github.com/hornbill/ldap"
)

//-- Does User Exist on Instance
func checkUserOnInstance(userID string, espXmlmc *apiLib.XmlmcInstStruct) (bool, error) {

	espXmlmc.SetParam("entity", "UserAccount")
	espXmlmc.SetParam("keyValue", userID)
	XMLCheckUser, xmlmcErr := espXmlmc.Invoke("data", "entityDoesRecordExist")
	var xmlRespon xmlmcCheckUserResponse
	if xmlmcErr != nil {
		return false, xmlmcErr
	}
	err := xml.Unmarshal([]byte(XMLCheckUser), &xmlRespon)
	if err != nil {
		stringError := err.Error()
		stringBody := string(XMLCheckUser)
		errWithBody := errors.New(stringError + " RESPONSE BODY: " + stringBody)
		return false, errWithBody
	}
	if xmlRespon.MethodResult != constOK {
		err := errors.New(xmlRespon.State.ErrorRet)
		return false, err
	}
	return xmlRespon.Params.RecordExist, nil
}

//-- Update User Record
func updateUser(u *ldap.Entry, buffer *bytes.Buffer, espXmlmc *apiLib.XmlmcInstStruct) (bool, error) {
	//-- Do we Lookup Site
	site := ""

	//-- Only use site lookup if enabled and not set to create only
	if ldapImportConf.SiteLookup.Enabled && ldapImportConf.SiteLookup.Action != createString {
		site = getSiteFromLookup(u, buffer)
	} else {
		site = getFeildValue(u, "Site", buffer)
	}

	userID := getFeildValue(u, "UserID", buffer)
	if userID != "" {
		espXmlmc.SetParam("userId", userID)
	}

	//-- Loop Through UserProfileMapping
	for key := range userUpdateArray {
		name := userUpdateArray[key]
		feild := userMappingMap[name]
		//-- Get Value From LDAP
		value := getFeildValue(u, name, buffer)

		if feild == "site" {
			value = site
		}

		//-- if we have Value then set it
		if value != "" {
			//-- handle User Type
			if feild == "userType" && getFeildValue(u, "UserType", buffer) != "" && ldapImportConf.UpdateUserType {
				espXmlmc.SetParam("userType", getFeildValue(u, "UserType", buffer))
			}
			//-- Skip Password on Update
			if feild != "password" && feild != "userType" {
				espXmlmc.SetParam(feild, value)
			}

		}
	}
	//-- Check for Dry Run
	if configDryRun != true {
		XMLUpdate, xmlmcErr := espXmlmc.Invoke("admin", "userUpdate")
		var xmlRespon xmlmcResponse
		if xmlmcErr != nil {
			return false, xmlmcErr
		}
		err := xml.Unmarshal([]byte(XMLUpdate), &xmlRespon)
		if err != nil {
			return false, err
		}

		if xmlRespon.MethodResult != constOK && xmlRespon.State.ErrorRet != noValuesToUpdate {
			err = errors.New(xmlRespon.State.ErrorRet)
			errorCountInc()
			return false, err

		}
		//-- Only use Org lookup if enabled and not set to create only
		if ldapImportConf.OrgLookup.Enabled && ldapImportConf.OrgLookup.Action != createString {
			userAddGroup(u, buffer, espXmlmc)
		}
		//-- Process User Status
		if ldapImportConf.UserAccountStatus.Enabled && ldapImportConf.UserAccountStatus.Action != createString {
			userSetStatus(userID, ldapImportConf.UserAccountStatus.Status, buffer)
		}

		//-- Add Roles
		if ldapImportConf.UserRoleAction != createString && len(ldapImportConf.Roles) > 0 {
			userAddRoles(userID, buffer, espXmlmc)
		}

		//-- Add Image
		if ldapImportConf.ImageLink.Enabled && ldapImportConf.ImageLink.Action != createString && ldapImportConf.ImageLink.URI != "" {
			userAddImage(u, buffer, espXmlmc)
		}

		//-- Process Profile Details
		boolUpdateProfile := userUpdateProfile(u, buffer, "Update")
		if boolUpdateProfile != true {
			err = errors.New("Error Updating User Profile")
			errorCountInc()
			return false, err
		}
		if xmlRespon.State.ErrorRet != noValuesToUpdate {
			buffer.WriteString(loggerGen(1, "User Update Success"))
			updateCountInc()
		} else {
			buffer.WriteString(loggerGen(1, "User Update Skipped"))
			updateSkippedCountInc()
		}

		return true, nil
	}
	//-- Process Profile Details as part of the dry run for testing
	boolUpdateProfile := userUpdateProfile(u, buffer, "Update")
	if boolUpdateProfile != true {
		err := errors.New("Error Updating User Profile")
		errorCountInc()
		return false, err
	}
	//-- Inc Counter
	updateSkippedCountInc()
	//-- DEBUG XML TO LOG FILE
	var XMLSTRING = espXmlmc.GetParam()
	buffer.WriteString(loggerGen(1, "User Update XML "+XMLSTRING))
	espXmlmc.ClearParam()

	return true, nil
}

//-- Create Users
func createUser(u *ldap.Entry, buffer *bytes.Buffer, espXmlmc *apiLib.XmlmcInstStruct) (bool, error) {
	//-- Do we Lookup Site
	site := ""

	//-- Only use Site lookup if enabled and not set to Update only
	if ldapImportConf.SiteLookup.Enabled && ldapImportConf.OrgLookup.Action != updateString {
		site = getSiteFromLookup(u, buffer)
	} else {
		site = getFeildValue(u, "Site", buffer)
	}

	userID := getFeildValue(u, "UserID", buffer)
	if userID != "" {
		espXmlmc.SetParam("userId", userID)
	}

	//-- Loop Through UserProfileMapping
	for key := range userCreateArray {
		name := userCreateArray[key]
		feild := userMappingMap[name]
		//-- Get Value From LDAP
		value := getFeildValue(u, name, buffer)
		//-- Process Site
		if feild == "site" {
			value = site
		}
		//-- Process Password Feild
		if feild == "password" {
			buffer.WriteString(loggerGen(1, "password"))
			if value == "" {
				value = generatePasswordString(10)
				buffer.WriteString(loggerGen(1, "Auto Generated Password for: "+userID+" - "+value))
			}
			value = base64.StdEncoding.EncodeToString([]byte(value))
		}

		//-- if we have Value then set it
		if value != "" {
			espXmlmc.SetParam(feild, value)

		}
	}

	//-- Check for Dry Run
	if configDryRun != true {
		XMLCreate, xmlmcErr := espXmlmc.Invoke("admin", "userCreate")
		var xmlRespon xmlmcResponse
		if xmlmcErr != nil {
			errorCountInc()
			return false, xmlmcErr
		}
		err := xml.Unmarshal([]byte(XMLCreate), &xmlRespon)
		if err != nil {
			errorCountInc()
			return false, err
		}
		if xmlRespon.MethodResult != constOK {
			err = errors.New(xmlRespon.State.ErrorRet)
			errorCountInc()
			return false, err

		}
		buffer.WriteString(loggerGen(1, "User Create Success"))

		//-- Only use Org lookup if enabled and not set to Update only
		if ldapImportConf.OrgLookup.Enabled && ldapImportConf.OrgLookup.Action != updateString {
			userAddGroup(u, buffer, espXmlmc)
		}
		//-- Process Account Status
		if ldapImportConf.UserAccountStatus.Enabled && ldapImportConf.UserAccountStatus.Action != updateString {
			userSetStatus(userID, ldapImportConf.UserAccountStatus.Status, buffer)
		}

		if ldapImportConf.UserRoleAction != updateString && len(ldapImportConf.Roles) > 0 {
			userAddRoles(userID, buffer, espXmlmc)
		}

		//-- Add Image
		if ldapImportConf.ImageLink.Enabled && ldapImportConf.ImageLink.Action != updateString && ldapImportConf.ImageLink.URI != "" {
			userAddImage(u, buffer, espXmlmc)
		}

		//-- Process Profile Details
		boolUpdateProfile := userUpdateProfile(u, buffer, "Create")
		if boolUpdateProfile != true {
			err = errors.New("Error Updating User Profile")
			errorCountInc()
			return false, err
		}

		createCountInc()
		return true, nil
	}
	//-- Process Profile Details as part of the dry run for testing
	boolUpdateProfile := userUpdateProfile(u, buffer, "Create")
	if boolUpdateProfile != true {
		err := errors.New("Error Updating User Profile")
		errorCountInc()
		return false, err
	}
	//-- DEBUG XML TO LOG FILE
	var XMLSTRING = espXmlmc.GetParam()
	buffer.WriteString(loggerGen(1, "User Create XML "+XMLSTRING))
	createSkippedCountInc()
	espXmlmc.ClearParam()

	return true, nil
}

func userAddRoles(userID string, buffer *bytes.Buffer, espXmlmc *apiLib.XmlmcInstStruct) bool {

	espXmlmc.SetParam("userId", userID)
	for _, role := range ldapImportConf.Roles {
		espXmlmc.SetParam("role", role)
		buffer.WriteString(loggerGen(1, "Add Role to User: "+role))
	}
	XMLCreate, xmlmcErr := espXmlmc.Invoke("admin", "userAddRole")
	var xmlRespon xmlmcResponse
	if xmlmcErr != nil {
		logger(4, "Unable to Assign Role to User: "+fmt.Sprintf("%s", xmlmcErr), true)

	}
	err := xml.Unmarshal([]byte(XMLCreate), &xmlRespon)
	if err != nil {
		buffer.WriteString(loggerGen(4, "Unable to Assign Role to User: "+fmt.Sprintf("%s", err)))
		return false
	}
	if xmlRespon.MethodResult != constOK {
		buffer.WriteString(loggerGen(4, "Unable to Assign Role to User: "+xmlRespon.State.ErrorRet))
		return false
	}
	buffer.WriteString(loggerGen(1, "Roles Added Successfully"))
	return true
}

func userSetStatus(userID string, status string, buffer *bytes.Buffer) bool {
	buffer.WriteString(loggerGen(1, "Set Status for User: "+userID+" Status:"+status))

	espXmlmc := apiLib.NewXmlmcInstance(ldapImportConf.InstanceID)
	espXmlmc.SetAPIKey(ldapImportConf.APIKey)
	espXmlmc.SetParam("userId", userID)
	espXmlmc.SetParam("accountStatus", status)

	XMLCreate, xmlmcErr := espXmlmc.Invoke("admin", "userSetAccountStatus")
	var xmlRespon xmlmcResponse
	if xmlmcErr != nil {
		logger(4, "Unable to Set User Status: "+fmt.Sprintf("%s", xmlmcErr), true)
	}
	err := xml.Unmarshal([]byte(XMLCreate), &xmlRespon)
	if err != nil {
		buffer.WriteString(loggerGen(4, "Unable to Set User Status "+fmt.Sprintf("%s", err)))
		return false
	}
	if xmlRespon.MethodResult != constOK {
		if xmlRespon.State.ErrorRet != "Failed to update account status (target and the current status is the same)." {
			buffer.WriteString(loggerGen(4, "Unable to Set User Status: "+xmlRespon.State.ErrorRet))
			return false
		}
		buffer.WriteString(loggerGen(1, "User Status Already Set to: "+status))
		return true
	}
	buffer.WriteString(loggerGen(1, "User Status Set Successfully"))
	return true
}

// Write DN and User ID to Cache
func writeUserToCache(DN string, ID string, buffer *bytes.Buffer) {

	//-- Check if ID or DN is empty then just return
	if DN == "" {
		return
	}
	if ID == "" {
		return
	}

	buffer.WriteString(loggerGen(1, "Write to DN Cache: "+DN+" ID:"+ID))
	mutexUsersDN.Lock()
	var newuserForCache usersDNStruct
	newuserForCache.DN = DN
	newuserForCache.UserID = ID
	name := []usersDNStruct{newuserForCache}
	usersDN = append(usersDN, name...)
	mutexUsersDN.Unlock()
}

// Get User ID From Cache By DN
func getUserFromDNCache(DN string) string {
	stringReturn := ""
	//-- Check if in Cache
	mutexUsersDN.Lock()
	for _, user := range usersDN {
		if strings.ToLower(user.DN) == strings.ToLower(DN) {
			stringReturn = user.UserID
		}
	}
	mutexUsersDN.Unlock()
	return stringReturn
}
