package main

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"

	"github.com/hornbill/goApiLib"
	"github.com/hornbill/ldap"
)

//-- Deal with User Profile Data
func userUpdateProfile(u *ldap.Entry, buffer *bytes.Buffer, espXmlmc *apiLib.XmlmcInstStruct) bool {
	UserID := getFeildValue(u, "UserID", buffer)
	buffer.WriteString(loggerGen(1, "Processing User Profile Data "+UserID))

	espXmlmc.OpenElement("profileData")
	espXmlmc.SetParam("userID", UserID)
	value := ""
	//-- Loop Through UserProfileMapping
	for key := range userProfileArray {
		name := userProfileArray[key]
		feild := userProfileMappingMap[name]

		if feild == "manager" {
			//-- Process User manager
			if ldapImportConf.UserManagerMapping.Enabled && ldapImportConf.UserManagerMapping.Action != updateString {
				value = getManagerFromLookup(u, buffer)
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
			espXmlmc.SetParam(feild, value)
		}
	}

	espXmlmc.CloseElement("profileData")
	//-- Check for Dry Run
	if configDryRun != true {
		XMLCreate, xmlmcErr := espXmlmc.Invoke("admin", "userProfileSet")
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
	var XMLSTRING = espXmlmc.GetParam()
	buffer.WriteString(loggerGen(1, "User Profile Update XML "+fmt.Sprintf("%s", XMLSTRING)))
	profileSkippedCountInc()
	espXmlmc.ClearParam()
	return true

}
