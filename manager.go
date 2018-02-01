package main

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"regexp"
	"strings"
	"sync"

	"github.com/hornbill/goApiLib"
	"github.com/hornbill/ldap"
)

var (
	onceManager  sync.Once
	managerAPI   *apiLib.XmlmcInstStruct
	mutexManager = &sync.Mutex{}
)

func getManagerFromLookup(u *ldap.Entry, buffer *bytes.Buffer) string {
	//-- Check if Manager Attribute is set
	if ldapImportConf.UserManagerMapping.Attribute == "" {
		buffer.WriteString(loggerGen(4, "Manager Lookup is Enabled but Attribute is not Defined"))
		return ""
	}

	//-- Get Value of Attribute
	buffer.WriteString(loggerGen(1, "LDAP Attribute for Manager Lookup: "+ldapImportConf.UserManagerMapping.Attribute))

	//-- Get Value of Attribute
	ManagerAttributeName := processComplexFeild(u, ldapImportConf.UserManagerMapping.Attribute, buffer)

	if ldapImportConf.UserManagerMapping.UseDNCacheFirst {
		buffer.WriteString(loggerGen(1, "Searching DN Cache for: "+ManagerAttributeName))
		managerID := getUserFromDNCache(ManagerAttributeName)
		if managerID != "" {
			buffer.WriteString(loggerGen(1, "Found Manager in DN Cache: "+managerID))
			return managerID
		}
		buffer.WriteString(loggerGen(1, "Unable to find Manager in DN Cache Coninuing search"))
	}
	//-- Dont Continue if we didn't get anything
	if ManagerAttributeName == "" {
		return ""
	}

	//-- Pull Data from Attriute using regext
	if ldapImportConf.UserManagerMapping.GetIDFromName {
		buffer.WriteString(loggerGen(1, "LDAP Manager String: "+ManagerAttributeName))
		ManagerAttributeName = getNameFromLDAPString(ManagerAttributeName, buffer)
	}
	//-- Is Search Enabled
	if ldapImportConf.UserManagerMapping.SearchforManager {
		buffer.WriteString(loggerGen(1, "Search for Manager is Enabled"))

		buffer.WriteString(loggerGen(1, "Looking Up Manager from Cache: "+ManagerAttributeName))
		managerIsInCache, ManagerIDCache := managerInCache(ManagerAttributeName)
	
		//-- Check if we have Chached the site already
		if managerIsInCache {
			buffer.WriteString(loggerGen(1, "Found Manager in Cache: "+ManagerIDCache))
			return ManagerIDCache
		}
		buffer.WriteString(loggerGen(1, "Manager Not In Cache Searching Hornbill"))
		ManagerIsOnInstance, ManagerIDInstance := searchManager(ManagerAttributeName, buffer)
		//-- If Returned set output
		if ManagerIsOnInstance {
			buffer.WriteString(loggerGen(1, "Manager Lookup found Id: "+ManagerIDInstance))
			return ManagerIDInstance
		}
	}else{
		buffer.WriteString(loggerGen(1, "Search for Manager is Disabled"))
		//-- Assume data is manager id
		buffer.WriteString(loggerGen(1, "Manager Id: "+ManagerAttributeName))
		return ManagerAttributeName
	}
	
	//else return empty
	return ""
}

//-- Search Manager on Instance
func searchManager(managerName string, buffer *bytes.Buffer) (bool, string) {
	boolReturn := false
	strReturn := ""
	//-- ESP Query for site
	if managerName == "" {
		return boolReturn, strReturn
	}
	// We lock the whole function so we dont reuse the same connection for multiple logging attempts
	mutexManager.Lock()
	defer mutexManager.Unlock()

	// We initilaise the connection pool the first time the function is called and reuse it
	// This is reuse the connections rather than creating a pool each invocation
	onceManager.Do(func() {
		managerAPI = apiLib.NewXmlmcInstance(ldapImportConf.InstanceID)
		managerAPI.SetAPIKey(ldapImportConf.APIKey)
		managerAPI.SetTimeout(5)
	})
	//-- Add support for Search Feild configuration
	strSearchField := "h_name"
	if ldapImportConf.UserManagerMapping.ManagerSearchField != "" {
		strSearchField = ldapImportConf.UserManagerMapping.ManagerSearchField
	}

	buffer.WriteString(loggerGen(1, "Manager Search: "+fmt.Sprintf("%s", strSearchField)+" - "+fmt.Sprintf("%s", managerName)))
	managerAPI.SetParam("entity", "UserAccount")
	managerAPI.SetParam("matchScope", "all")
	managerAPI.OpenElement("searchFilter")
	managerAPI.SetParam(strSearchField, managerName)
	managerAPI.CloseElement("searchFilter")
	managerAPI.SetParam("maxResults", "1")
	XMLUserSearch, xmlmcErr := managerAPI.Invoke("data", "entityBrowseRecords")
	var xmlRespon xmlmcUserListResponse
	if xmlmcErr != nil {
		buffer.WriteString(loggerGen(4, "Unable to Search for Manager: "+fmt.Sprintf("%v", xmlmcErr)))
	}
	err := xml.Unmarshal([]byte(XMLUserSearch), &xmlRespon)
	if err != nil {
		stringError := err.Error()
		stringBody := string(XMLUserSearch)
		buffer.WriteString(loggerGen(4, "Unable to Search for Manager: "+fmt.Sprintf("%v", stringError+" RESPONSE BODY: "+stringBody)))
	} else {
		if xmlRespon.MethodResult != constOK {
			buffer.WriteString(loggerGen(4, "Unable to Search for Manager: "+xmlRespon.State.ErrorRet))
		} else {
			//-- Check Response
			if xmlRespon.Params.RowData.Row.UserName != "" {
				if strings.ToLower(xmlRespon.Params.RowData.Row.UserName) == strings.ToLower(managerName) {

					strReturn = xmlRespon.Params.RowData.Row.UserID
					boolReturn = true
					//-- Add Site to Cache
					mutexManagers.Lock()
					var newManagerForCache managerListStruct
					newManagerForCache.UserID = strReturn
					newManagerForCache.UserName = managerName
					name := []managerListStruct{newManagerForCache}
					managers = append(managers, name...)
					mutexManagers.Unlock()
				}
			}
		}
	}
	return boolReturn, strReturn
}

//-- Check if Manager in Cache
func managerInCache(managerName string) (bool, string) {
	boolReturn := false
	stringReturn := ""
	//-- Check if in Cache
	mutexManagers.Lock()
	for _, manager := range managers {
		if strings.ToLower(manager.UserName) == strings.ToLower(managerName) {
			boolReturn = true
			stringReturn = manager.UserID
		}
	}
	mutexManagers.Unlock()
	return boolReturn, stringReturn
}

//-- Takes a string based on a LDAP DN and returns to the CN String Name
func getNameFromLDAPString(feild string, buffer *bytes.Buffer) string {

	regex := ldapImportConf.UserManagerMapping.Regex
	reverse := ldapImportConf.UserManagerMapping.Reverse
	stringReturn := ""

	//-- Match $variables from String
	re1, err := regexp.Compile(regex)
	if err != nil {
		fmt.Printf("%v \n", err)

	}
	//-- Get Array of all Matched max 100
	result := re1.FindAllString(feild, 100)

	//-- Loop Matches
	for _, v := range result {
		//-- String LDAP String Chars Out from match
		v = strings.Replace(v, "CN=", "", -1)
		v = strings.Replace(v, "OU=", "", -1)
		v = strings.Replace(v, "DC=", "", -1)
		v = strings.Replace(v, "\\", "", -1)
		nameArray := strings.Split(v, ",")

		for _, n := range nameArray {
			n = strings.Trim(n, " ")
			if n != "" {
				if reverse {
					stringReturn = n + " " + stringReturn
				} else {
					stringReturn = stringReturn + " " + n
				}
			}

		}

	}
	stringReturn = strings.Trim(stringReturn, " ")
	return stringReturn
}
