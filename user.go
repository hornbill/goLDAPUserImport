package main

import (
	"bytes"
	"encoding/base64"
	"encoding/xml"
	"errors"
	"fmt"

	"github.com/hornbill/goApiLib"
	"github.com/hornbill/ldap"
)

//-- Does User Exist on Instance
func checkUserOnInstance(userID string) (bool, error) {
	espXmlmc := apiLib.NewXmlmcInstance(ldapImportConf.URL)
	espXmlmc.SetAPIKey(ldapImportConf.APIKey)
	espXmlmc.SetParam("entity", "UserAccount")
	espXmlmc.SetParam("keyValue", userID)
	XMLCheckUser, xmlmcErr := espXmlmc.Invoke("data", "entityDoesRecordExist")
	var xmlRespon xmlmcCheckUserResponse
	if xmlmcErr != nil {
		return false, xmlmcErr
	}
	err := xml.Unmarshal([]byte(XMLCheckUser), &xmlRespon)
	if err != nil {
		return false, err
	}
	if xmlRespon.MethodResult != constOK {
		err := errors.New(xmlRespon.State.ErrorRet)
		return false, err
	}
	return xmlRespon.Params.RecordExist, nil
}

//-- Update User Record
func updateUser(u *ldap.Entry, buffer *bytes.Buffer) (bool, error) {
	//-- Do we Lookup Site
	site := ""

	//-- Only use site lookup if enabled and not set to create only
	if ldapImportConf.SiteLookup.Enabled && ldapImportConf.SiteLookup.Action != createString {
		site = getSiteFromLookup(u, buffer)
	} else {
		site = getFeildValue(u, "Site", buffer)
	}

	espXmlmc := apiLib.NewXmlmcInstance(ldapImportConf.URL)
	espXmlmc.SetAPIKey(ldapImportConf.APIKey)

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
			errorCount++
			return false, err

		}
		//-- Only use Org lookup if enabled and not set to create only
		if ldapImportConf.OrgLookup.Enabled && ldapImportConf.OrgLookup.Action != createString {
			userAddGroup(u, buffer)
		}
		//-- Process User Status
		if ldapImportConf.UserAccountStatus.Enabled && ldapImportConf.UserAccountStatus.Action != createString {
			userSetStatus(userID, ldapImportConf.UserAccountStatus.Status, buffer)
		}

		//-- Process Profile Details
		boolUpdateProfile := userUpdateProfile(u, buffer)
		if boolUpdateProfile != true {
			err = errors.New("Error Updating User Profile")
			errorCount++
			return false, err
		}
		if xmlRespon.State.ErrorRet != noValuesToUpdate {
			buffer.WriteString(loggerGen(1, "User Update Success"))
			counters.updated++
		} else {
			counters.updatedSkipped++
		}

		return true, nil
	}
	//-- Inc Counter
	counters.updatedSkipped++
	//-- DEBUG XML TO LOG FILE
	var XMLSTRING = espXmlmc.GetParam()
	buffer.WriteString(loggerGen(1, "User Update XML "+fmt.Sprintf("%s", XMLSTRING)))
	espXmlmc.ClearParam()

	//-- Process Profile Details
	userUpdateProfile(u, buffer)

	return true, nil
}

//-- Create Users
func createUser(u *ldap.Entry, buffer *bytes.Buffer) (bool, error) {
	//-- Do we Lookup Site
	site := ""

	//-- Only use Site lookup if enabled and not set to Update only
	if ldapImportConf.SiteLookup.Enabled && ldapImportConf.OrgLookup.Action != updateString {
		site = getSiteFromLookup(u, buffer)
	} else {
		site = getFeildValue(u, "Site", buffer)
	}
	espXmlmc := apiLib.NewXmlmcInstance(ldapImportConf.URL)
	espXmlmc.SetAPIKey(ldapImportConf.APIKey)

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
			errorCount++
			return false, xmlmcErr
		}
		err := xml.Unmarshal([]byte(XMLCreate), &xmlRespon)
		if err != nil {
			errorCount++
			return false, err
		}
		if xmlRespon.MethodResult != constOK {
			err = errors.New(xmlRespon.State.ErrorRet)
			errorCount++
			return false, err

		}
		buffer.WriteString(loggerGen(1, "User Create Success"))

		//-- Only use Org lookup if enabled and not set to Update only
		if ldapImportConf.OrgLookup.Enabled && ldapImportConf.OrgLookup.Action != updateString {
			userAddGroup(u, buffer)
		}
		//-- Process Account Status
		if ldapImportConf.UserAccountStatus.Enabled && ldapImportConf.UserAccountStatus.Action != updateString {
			userSetStatus(userID, ldapImportConf.UserAccountStatus.Status, buffer)
		}

		if len(ldapImportConf.Roles) > 0 {
			userAddRoles(userID, buffer)
		}
		//-- Process Profile Details
		boolUpdateProfile := userUpdateProfile(u, buffer)
		if boolUpdateProfile != true {
			err = errors.New("Error Updating User Profile")
			errorCount++
			return false, err
		}

		counters.created++
		return true, nil
	}
	//-- DEBUG XML TO LOG FILE
	var XMLSTRING = espXmlmc.GetParam()
	buffer.WriteString(loggerGen(1, "User Create XML "+fmt.Sprintf("%s", XMLSTRING)))
	counters.createskipped++
	espXmlmc.ClearParam()

	//-- Process Profile Details
	userUpdateProfile(u, buffer)

	return true, nil
}

func userAddRoles(userID string, buffer *bytes.Buffer) bool {
	espXmlmc := apiLib.NewXmlmcInstance(ldapImportConf.URL)
	espXmlmc.SetAPIKey(ldapImportConf.APIKey)
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
		espLogger("Unable to Assign Role to User: "+xmlRespon.State.ErrorRet, "error")
		return false
	}
	buffer.WriteString(loggerGen(1, "Roles Added Successfully"))
	return true
}

func userSetStatus(userID string, status string, buffer *bytes.Buffer) bool {
	buffer.WriteString(loggerGen(1, "Set Status for User: "+fmt.Sprintf("%s", userID)+" Status:"+fmt.Sprintf("%s", status)))

	espXmlmc := apiLib.NewXmlmcInstance(ldapImportConf.URL)
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
			buffer.WriteString(loggerGen(4, "Unable to Set User Status 111: "+xmlRespon.State.ErrorRet))
			espLogger("Unable to Set User Status: "+xmlRespon.State.ErrorRet, "error")
			return false
		}
		buffer.WriteString(loggerGen(1, "User Status Already Set to: "+fmt.Sprintf("%s", status)))
		return true
	}
	buffer.WriteString(loggerGen(1, "User Status Set Successfully"))
	return true

	return true
}
