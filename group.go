package main

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"log"
	"strconv"
	"sync"

	"github.com/hornbill/goApiLib"
	"github.com/hornbill/ldap"
)

var (
	onceGroupSearch  sync.Once
	onceGroupList    sync.Once
	onceGroupRemove  sync.Once
	groupListAPI     *apiLib.XmlmcInstStruct
	groupSearchAPI   *apiLib.XmlmcInstStruct
	groupRemoveAPI   *apiLib.XmlmcInstStruct
	mutexGroupList   = &sync.Mutex{}
	mutexGroupSearch = &sync.Mutex{}
	mutexGroupRemove = &sync.Mutex{}
)

//-- Deal with adding a user to a group
func userAddGroup(u *ldap.Entry, buffer *bytes.Buffer, espXmlmc *apiLib.XmlmcInstStruct) bool {

	//-- Check if Site Attribute is set
	if ldapImportConf.OrgLookup.Attribute == "" {
		buffer.WriteString(loggerGen(1, "Org Lookup is Enabled but Attribute is not Defined"))
		return false
	}
	//-- Get Value of Attribute
	buffer.WriteString(loggerGen(1, "LDAP Attribute for Org Lookup: "+ldapImportConf.OrgLookup.Attribute))
	orgAttributeName := processComplexFeild(u, ldapImportConf.OrgLookup.Attribute, buffer)
	buffer.WriteString(loggerGen(1, "Looking Up Org "+orgAttributeName))
	orgIsInCache, orgID := groupInCache(orgAttributeName)
	//-- Check if we have Chached the site already
	if orgIsInCache {
		buffer.WriteString(loggerGen(1, "Found Org in Cache "+orgID))
		userAddGroupAsoc(u, orgID, buffer, espXmlmc)
		return true
	}

	//-- We Get here if not in cache
	orgIsOnInstance, orgID := searchGroup(orgAttributeName, buffer)
	if orgIsOnInstance {
		buffer.WriteString(loggerGen(1, "Org Lookup found Id "+orgID))
		userAddGroupAsoc(u, orgID, buffer, espXmlmc)
		return true
	}
	buffer.WriteString(loggerGen(1, "Unable to Find Organsiation "+orgAttributeName))
	return false
}
func removeUserGroupAssoc(user string, buffer *bytes.Buffer, strGroupName string) {
	if user == "" {
		return
	}
	mutexGroupList.Lock()
	defer mutexGroupList.Unlock()
	stringGroupType := ""
	switch ldapImportConf.OrgLookup.Type {
	case 0:
		stringGroupType = "general"
		break
	case 1:
		stringGroupType = "team"
		break
	case 2:
		stringGroupType = "department"
		break
	case 3:
		stringGroupType = "costcenter"
		break
	case 4:
		stringGroupType = "division"
		break
	case 5:
		stringGroupType = "company"
		break
	}
	buffer.WriteString(loggerGen(1, "Removing existing Group Assignments of Type: "+fmt.Sprintf("%s", stringGroupType)))
	//-- Get Existing User Assignments
	onceGroupList.Do(func() {
		groupListAPI = apiLib.NewXmlmcInstance(ldapImportConf.InstanceID)
		groupListAPI.SetAPIKey(ldapImportConf.APIKey)
		groupListAPI.SetTimeout(5)
	})
	groupListAPI.SetParam("userId", user)
	XMLUserGroupList, xmlmcErr := groupListAPI.Invoke("admin", "userGetGroupList")
	var xmlRespon xmlmcUserGroupListResponse
	if xmlmcErr != nil {
		log.Fatal(xmlmcErr)
		buffer.WriteString(loggerGen(4, "Unable to Get users Groups: "+fmt.Sprintf("%v", xmlmcErr)))
	}
	err := xml.Unmarshal([]byte(XMLUserGroupList), &xmlRespon)
	if err != nil {
		buffer.WriteString(loggerGen(4, "Unable to Get users Groups: "+fmt.Sprintf("%v", err)))
	} else {
		if xmlRespon.MethodResult == constOK {
			//-- Loop groups

			for _, group := range xmlRespon.Params.GroupItem {
				//-- If Type maches and not the group to assign then remove
				if group.Type == stringGroupType && group.GroupID != strGroupName {
					removeGroup(user, group.GroupID, buffer)
				}
			}
		}
	}
}
func removeGroup(user string, group string, buffer *bytes.Buffer) {
	buffer.WriteString(loggerGen(1, "Remove User: "+fmt.Sprintf("%s", user)+" From Group: "+fmt.Sprintf("%s", group)))
	mutexGroupRemove.Lock()
	defer mutexGroupRemove.Unlock()

	onceGroupRemove.Do(func() {
		groupRemoveAPI = apiLib.NewXmlmcInstance(ldapImportConf.InstanceID)
		groupRemoveAPI.SetAPIKey(ldapImportConf.APIKey)
		groupRemoveAPI.SetTimeout(5)
	})

	groupRemoveAPI.SetParam("userId", user)
	groupRemoveAPI.SetParam("groupId", group)

	XMLSiteSearch, xmlmcErr := groupRemoveAPI.Invoke("admin", "userDeleteGroup")
	var xmlRespon xmlmcGroupListResponse
	if xmlmcErr != nil {
		buffer.WriteString(loggerGen(4, "Unable to Remove User from Group: "+fmt.Sprintf("%v", xmlmcErr)))
	}
	err := xml.Unmarshal([]byte(XMLSiteSearch), &xmlRespon)
	if err != nil {
		buffer.WriteString(loggerGen(4, "Unable to Remove User from Group "+fmt.Sprintf("%v", err)))
	} else {
		if xmlRespon.MethodResult != constOK {
			buffer.WriteString(loggerGen(4, "Unable to Remove User from Group: "+xmlRespon.State.ErrorRet))
		} else {
			buffer.WriteString(loggerGen(1, "User: "+fmt.Sprintf("%s", user)+" Removed From Group: "+fmt.Sprintf("%s", group)))
		}
	}
}
func userAddGroupAsoc(u *ldap.Entry, orgID string, buffer *bytes.Buffer, espXmlmc *apiLib.XmlmcInstStruct) {
	//-- Get User iD
	UserID := getFeildValue(u, "UserID", buffer)
	//-- Check for exisiting Groups
	if ldapImportConf.OrgLookup.OnlyOneGroupAssignment {
		removeUserGroupAssoc(UserID, buffer, orgID)
	}

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
	mutexGroups.Lock()
	for _, group := range groups {
		if group.GroupName == groupName {
			boolReturn = true
			stringReturn = group.GroupID
			break
		}
	}
	mutexGroups.Unlock()
	return boolReturn, stringReturn
}

//-- Function to Check if site is on the instance
func searchGroup(orgName string, buffer *bytes.Buffer) (bool, string) {
	boolReturn := false
	strReturn := ""
	//-- ESP Query for site
	if orgName == "" {
		return boolReturn, strReturn
	}
	mutexGroupSearch.Lock()
	defer mutexGroupSearch.Unlock()

	onceGroupSearch.Do(func() {
		groupSearchAPI = apiLib.NewXmlmcInstance(ldapImportConf.InstanceID)
		groupSearchAPI.SetAPIKey(ldapImportConf.APIKey)
		groupSearchAPI.SetTimeout(5)
	})

	groupSearchAPI.SetParam("application", "com.hornbill.core")
	groupSearchAPI.SetParam("queryName", "GetGroupByName")
	groupSearchAPI.OpenElement("queryParams")
	groupSearchAPI.SetParam("h_name", orgName)
	groupSearchAPI.SetParam("h_type", strconv.Itoa(ldapImportConf.OrgLookup.Type))
	groupSearchAPI.CloseElement("queryParams")

	XMLSiteSearch, xmlmcErr := groupSearchAPI.Invoke("data", "queryExec")
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
				mutexGroups.Lock()
				var newgroupForCache groupListStruct
				newgroupForCache.GroupID = strReturn
				newgroupForCache.GroupName = orgName
				name := []groupListStruct{newgroupForCache}
				groups = append(groups, name...)
				mutexGroups.Unlock()
			}
		}
	}

	return boolReturn, strReturn
}
