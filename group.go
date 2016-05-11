package main

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"log"
	"strconv"

	"github.com/hornbill/goApiLib"
	"github.com/hornbill/ldap"
)

//-- Deal with adding a user to a group
func userAddGroup(u *ldap.Entry, buffer *bytes.Buffer) bool {

	//-- Check if Site Attribute is set
	if ldapImportConf.OrgLookup.Attribute == "" {
		buffer.WriteString(loggerGen(2, "Org Lookup is Enabled but Attribute is not Defined"))
		return false
	}
	//-- Get Value of Attribute
	buffer.WriteString(loggerGen(2, "LDAP Attribute for Org Lookup: "+ldapImportConf.OrgLookup.Attribute))
	orgAttributeName := processComplexFeild(u, ldapImportConf.OrgLookup.Attribute, buffer)
	buffer.WriteString(loggerGen(2, "Looking Up Org "+orgAttributeName))
	orgIsInCache, orgID := groupInCache(orgAttributeName)
	//-- Check if we have Chached the site already
	if orgIsInCache {
		buffer.WriteString(loggerGen(1, "Found Org in Cache "+orgID))
		userAddGroupAsoc(u, orgID, buffer)
		return true
	}

	//-- We Get here if not in cache
	orgIsOnInstance, orgID := searchGroup(orgAttributeName, buffer)
	if orgIsOnInstance {
		buffer.WriteString(loggerGen(1, "Org Lookup found Id "+orgID))
		userAddGroupAsoc(u, orgID, buffer)
		return true
	}
	buffer.WriteString(loggerGen(1, "Unable to Find Organsiation "+orgAttributeName))
	return false
}

func userAddGroupAsoc(u *ldap.Entry, orgID string, buffer *bytes.Buffer) {
	UserID := getFeildValue(u, "UserID", buffer)
	espXmlmc := apiLib.NewXmlmcInstance(ldapImportConf.URL)
	espXmlmc.SetAPIKey(ldapImportConf.APIKey)
	espXmlmc.SetParam("userId", UserID)
	espXmlmc.SetParam("groupId", orgID)
	espXmlmc.SetParam("memberRole", ldapImportConf.OrgLookup.Membership)
	espXmlmc.OpenElement("options")
	espXmlmc.SetParam("tasksView", strconv.FormatBool(ldapImportConf.OrgLookup.TasksView))
	espXmlmc.SetParam("tasksAction", strconv.FormatBool(ldapImportConf.OrgLookup.TasksAction))
	espXmlmc.CloseElement("options")

	XMLSiteSearch, xmlmcErr := espXmlmc.Invoke("admin", "userAddGroup")
	var xmlRespon xmlmcuserSetGroupOptionsResponse
	if xmlmcErr != nil {
		log.Fatal(xmlmcErr)
		buffer.WriteString(loggerGen(4, "Unable to Associate User To Group: "+fmt.Sprintf("%v", xmlmcErr)))
	}
	err := xml.Unmarshal([]byte(XMLSiteSearch), &xmlRespon)
	if err != nil {
		buffer.WriteString(loggerGen(4, "Unable to Associate User To Group: "+fmt.Sprintf("%v", err)))
	} else {
		if xmlRespon.MethodResult != constOK {
			if xmlRespon.State.ErrorRet != "The specified user ["+UserID+"] already belongs to ["+orgID+"] group" {
				buffer.WriteString(loggerGen(4, "Unable to Associate User To Organsiation: "+xmlRespon.State.ErrorRet))
			} else {
				buffer.WriteString(loggerGen(1, "User: "+UserID+" Already Added to Organsiation: "+orgID))
			}

		} else {
			buffer.WriteString(loggerGen(1, "User: "+UserID+" Added to Organsiation: "+orgID))
		}
	}

}

//-- Function to Check if in Cache
func groupInCache(groupName string) (bool, string) {
	boolReturn := false
	stringReturn := ""
	//-- Check if in Cache
	for _, group := range groups {
		if group.GroupName == groupName {
			boolReturn = true
			stringReturn = group.GroupID
		}
	}

	return boolReturn, stringReturn
}

//-- Function to Check if site is on the instance
func searchGroup(orgName string, buffer *bytes.Buffer) (bool, string) {
	boolReturn := false
	strReturn := ""
	//-- ESP Query for site
	espXmlmc := apiLib.NewXmlmcInstance(ldapImportConf.URL)
	espXmlmc.SetAPIKey(ldapImportConf.APIKey)
	if orgName == "" {
		return boolReturn, strReturn
	}
	espXmlmc.SetParam("application", "com.hornbill.core")
	espXmlmc.SetParam("queryName", "GetGroupByName")
	espXmlmc.OpenElement("queryParams")
	espXmlmc.SetParam("h_name", orgName)
	espXmlmc.SetParam("h_type", strconv.Itoa(ldapImportConf.OrgLookup.Type))
	espXmlmc.CloseElement("queryParams")

	XMLSiteSearch, xmlmcErr := espXmlmc.Invoke("data", "queryExec")
	var xmlRespon xmlmcGroupListResponse
	if xmlmcErr != nil {
		buffer.WriteString(loggerGen(4, "Unable to Search for Group: "+fmt.Sprintf("%v", xmlmcErr)))
	}
	err := xml.Unmarshal([]byte(XMLSiteSearch), &xmlRespon)
	if err != nil {
		buffer.WriteString(loggerGen(4, "Unable to Search for Group: "+fmt.Sprintf("%v", err)))
	} else {
		if xmlRespon.MethodResult != constOK {
			buffer.WriteString(loggerGen(4, "Unable to Search for Group: "+xmlRespon.State.ErrorRet))
		} else {
			//-- Check Response
			if xmlRespon.Params.RowData.Row.GroupID != "" {
				strReturn = xmlRespon.Params.RowData.Row.GroupID
				boolReturn = true
				//-- Add Group to Cache
				var newgroupForCache groupListStruct
				newgroupForCache.GroupID = strReturn
				newgroupForCache.GroupName = orgName
				name := []groupListStruct{newgroupForCache}
				groups = append(groups, name...)

			}
		}
	}

	return boolReturn, strReturn
}
