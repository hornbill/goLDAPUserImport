package main

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"sync"

	"github.com/hornbill/goApiLib"
	"github.com/hornbill/ldap"
)

var (
	onceProfile  sync.Once
	profileAPI   *apiLib.XmlmcInstStruct
	mutexProfile = &sync.Mutex{}
)

//-- Deal with User Profile Data
func userUpdateProfile(u *ldap.Entry, buffer *bytes.Buffer, updateType string) bool {
	// We lock the whole function so we dont reuse the same connection for multiple logging attempts
	mutexProfile.Lock()
	defer mutexProfile.Unlock()

	// We initilaise the connection pool the first time the function is called and reuse it
	// This is reuse the connections rather than creating a pool each invocation
	onceProfile.Do(func() {
		profileAPI = apiLib.NewXmlmcInstance(ldapImportConf.InstanceID)
		profileAPI.SetAPIKey(ldapImportConf.APIKey)
		profileAPI.SetTimeout(5)
	})

	UserID := getFeildValue(u, "UserID", buffer)
	buffer.WriteString(loggerGen(1, "Processing User Profile Data "+UserID))

	profileAPI.OpenElement("profileData")
	profileAPI.SetParam("userID", UserID)
	value := ""
	//-- Loop Through UserProfileMapping
	for key := range userProfileArray {
		name := userProfileArray[key]
		feild := userProfileMappingMap[name]

		if feild == "manager" {
			//-- Process User manager if enabled
			if ldapImportConf.UserManagerMapping.Enabled {
				//-- Action is Update
				if updateType == "Update" && ldapImportConf.UserManagerMapping.Action != createString {
					value = getManagerFromLookup(u, buffer)
				}
				//-- Action is Create
				if updateType == "Create" && ldapImportConf.UserManagerMapping.Action != updateString {
					value = getManagerFromLookup(u, buffer)
				}

			} else {
				//-- Get Value From LDAP
				value = getFeildValueProfile(u, name, buffer)
			}

		} else {
			//-- Get Value From LDAP
			value = getFeildValueProfile(u, name, buffer)
		}

		//-- if we have Value then set it
		if value != "" {
			profileAPI.SetParam(feild, value)
		}
	}

	profileAPI.CloseElement("profileData")
	//-- Check for Dry Run
	if configDryRun != true {
		XMLCreate, xmlmcErr := profileAPI.Invoke("admin", "userProfileSet")
		var xmlRespon xmlmcResponse
		if xmlmcErr != nil {
			buffer.WriteString(loggerGen(4, "Unable to Update User Profile: "+fmt.Sprintf("%v", xmlmcErr)))
			return false
		}
		err := xml.Unmarshal([]byte(XMLCreate), &xmlRespon)
		if err != nil {
			buffer.WriteString(loggerGen(4, "Unable to Update User Profile: "+fmt.Sprintf("%v", err)))

			return false
		}
		if xmlRespon.MethodResult != constOK {
			profileSkippedCountInc()
			if xmlRespon.State.ErrorRet == noValuesToUpdate {
				return true
			}
			err := errors.New(xmlRespon.State.ErrorRet)
			buffer.WriteString(loggerGen(4, "Unable to Update User Profile: "+fmt.Sprintf("%v", err)))
			return false
		}
		profileCountInc()
		buffer.WriteString(loggerGen(1, "User Profile Update Success"))
		return true

	}
	//-- DEBUG XML TO LOG FILE
	var XMLSTRING = profileAPI.GetParam()
	buffer.WriteString(loggerGen(1, "User Profile Update XML "+XMLSTRING))
	profileSkippedCountInc()
	profileAPI.ClearParam()
	return true

}
