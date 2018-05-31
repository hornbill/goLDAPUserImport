package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/hornbill/goApiLib"
	"github.com/hornbill/pb"
)

var (
	hornbillImport *apiLib.XmlmcInstStruct
	pageSize       int
)

func initXMLMC() {

	hornbillImport = apiLib.NewXmlmcInstance(Flags.configInstanceID)
	hornbillImport.SetAPIKey(Flags.configAPIKey)
	hornbillImport.SetTimeout(Flags.configAPITimeout)
	hornbillImport.SetJSONResponse(true)

	pageSize = ldapImportConf.Advanced.PageSize

	if pageSize == 0 {
		pageSize = 100
	}
}
func loadUsers() {
	//-- Init One connection to Hornbill to load all data
	initXMLMC()
	logger(1, "Loading Users from Hornbill", false)

	count := getCount("getUserAccountsList")
	getUserAccountList(count)

	logger(1, "Users Loaded: "+fmt.Sprintf("%d", len(HornbillCache.Users)), false)
}
func loadUsersRoles() {
	//-- Only Load if Enabled
	if ldapImportConf.User.Role.Action != "Create" && ldapImportConf.User.Role.Action != "Update" && ldapImportConf.User.Role.Action != "Both" {
		logger(1, "Skipping Loading Roles Due to Config", false)
		return
	}

	logger(1, "Loading Users Roles from Hornbill", false)

	count := getCount("getUserAccountsRolesList")
	getUserAccountsRolesList(count)

	logger(1, "Users Roles Loaded: "+fmt.Sprintf("%d", len(HornbillCache.UserRoles)), false)
}

func loadSites() {
	//-- Only Load if Enabled
	if ldapImportConf.User.Site.Action != "Create" && ldapImportConf.User.Site.Action != "Update" && ldapImportConf.User.Site.Action != "Both" {
		logger(1, "Skipping Loading Sites Due to Config", false)
		return
	}

	logger(1, "Loading Sites from Hornbill", false)

	count := getCount("getSitesList")
	getSitesList(count)

	logger(1, "Sites Loaded: "+fmt.Sprintf("%d", len(HornbillCache.Sites)), false)
}
func loadGroups() {
	boolSkip := true
	for index := range ldapImportConf.User.Org {
		orgAction := ldapImportConf.User.Org[index]
		if orgAction.Action == "Create" || orgAction.Action == "Update" || orgAction.Action == "Both" {
			boolSkip = false
		}
	}
	if boolSkip {
		logger(1, "Skipping Loading Orgs Due to Config", false)
		return
	}
	//-- Only Load if Enabled
	logger(1, "Loading Orgs from Hornbill", false)

	count := getCount("getGroupsList")
	getGroupsList(count)

	logger(1, "Orgs Loaded: "+fmt.Sprintf("%d", len(HornbillCache.Groups)), false)
}
func loadUserGroups() {
	boolSkip := true
	for index := range ldapImportConf.User.Org {
		orgAction := ldapImportConf.User.Org[index]
		if orgAction.Action == "Create" || orgAction.Action == "Update" || orgAction.Action == "Both" {
			boolSkip = false
		}
	}
	if boolSkip {
		logger(1, "Skipping Loading User Orgs Due to Config", false)
		return
	}
	//-- Only Load if Enabled
	logger(1, "Loading User Orgs from Hornbill", false)

	count := getCount("getUserAccountsGroupsList")
	getUserAccountsGroupsList(count)

	logger(1, "User Orgs Loaded: "+fmt.Sprintf("%d", len(HornbillCache.UserGroups))+"\n", false)
}

//-- Check so that only data that relates to users in the LDAP data set are stored in the working set
func userIDExistsInLDAP(userID string) bool {
	userID = strings.ToLower(userID)
	_, present := HornbillCache.UsersWorking[userID]
	if present {
		return true
	}
	return false
}

func getUserAccountsGroupsList(count uint64) {
	var loopCount uint64

	//-- Init Map
	HornbillCache.UserGroups = make(map[string][]string)
	bar := pb.StartNew(int(count))
	//-- Load Results in pages of pageSize
	for loopCount < count {
		logger(1, "Loading User Accounts Orgs List Offset: "+fmt.Sprintf("%d", loopCount), false)

		hornbillImport.SetParam("application", "com.hornbill.core")
		hornbillImport.SetParam("queryName", "getUserAccountsGroupsList")
		hornbillImport.OpenElement("queryParams")
		hornbillImport.SetParam("rowstart", strconv.FormatUint(loopCount, 10))
		hornbillImport.SetParam("limit", strconv.Itoa(pageSize))
		hornbillImport.CloseElement("queryParams")
		RespBody, xmlmcErr := hornbillImport.Invoke("data", "queryExec")

		var JSONResp xmlmcUserGroupListResponse
		if xmlmcErr != nil {
			logger(4, "Unable to Query Accounts Orgs List "+fmt.Sprintf("%s", xmlmcErr), false)
		}
		err := json.Unmarshal([]byte(RespBody), &JSONResp)
		if err != nil {
			logger(4, "Unable to Query Accounts Orgs  List "+fmt.Sprintf("%s", err), false)
		}
		if JSONResp.State.Error != "" {
			logger(4, "Unable to Query Accounts Orgs  List "+JSONResp.State.Error, false)
		}

		//-- Push into Map of slices to userId = array of roles
		for index := range JSONResp.Params.RowData.Row {
			if userIDExistsInLDAP(JSONResp.Params.RowData.Row[index].HUserID) {
				HornbillCache.UserGroups[strings.ToLower(JSONResp.Params.RowData.Row[index].HUserID)] = append(HornbillCache.UserGroups[strings.ToLower(JSONResp.Params.RowData.Row[index].HUserID)], JSONResp.Params.RowData.Row[index].HGroupID)
			}
		}
		// Add 100
		loopCount += uint64(pageSize)
		bar.Add(len(JSONResp.Params.RowData.Row))
	}
	bar.FinishPrint("Account Orgs Loaded \n")

}
func getGroupsList(count uint64) {
	var loopCount uint64
	//-- Init Map
	HornbillCache.Groups = make(map[string]userGroupStruct)
	HornbillCache.GroupsID = make(map[string]userGroupStruct)
	//-- Load Results in pages of pageSize
	bar := pb.StartNew(int(count))
	for loopCount < count {
		logger(1, "Loading Orgs List Offset: "+fmt.Sprintf("%d", loopCount), false)

		hornbillImport.SetParam("application", "com.hornbill.core")
		hornbillImport.SetParam("queryName", "getGroupsList")
		hornbillImport.OpenElement("queryParams")
		hornbillImport.SetParam("rowstart", strconv.FormatUint(loopCount, 10))
		hornbillImport.SetParam("limit", strconv.Itoa(pageSize))
		hornbillImport.CloseElement("queryParams")
		RespBody, xmlmcErr := hornbillImport.Invoke("data", "queryExec")

		var JSONResp xmlmcGroupListResponse
		if xmlmcErr != nil {
			logger(4, "Unable to Query Orgs List "+fmt.Sprintf("%s", xmlmcErr), false)
		}
		err := json.Unmarshal([]byte(RespBody), &JSONResp)
		if err != nil {
			logger(4, "Unable to Query Orgs List "+fmt.Sprintf("%s", err), false)
		}
		if JSONResp.State.Error != "" {
			logger(4, "Unable to Query Orgs List "+JSONResp.State.Error, false)
		}

		//-- Push into Map
		for _, rec := range JSONResp.Params.RowData.Row {
			var group userGroupStruct
			group.ID = rec.HID
			group.Name = rec.HName
			group.Type, _ = strconv.Atoi(rec.HType)

			//-- List of group names to group object for name to id lookup
			HornbillCache.Groups[strings.ToLower(rec.HName)] = group
			//-- List of group id to group objects for id to type lookup
			HornbillCache.GroupsID[strings.ToLower(rec.HID)] = group
		}
		// Add 100
		loopCount += uint64(pageSize)
		bar.Add(len(JSONResp.Params.RowData.Row))
	}
	bar.FinishPrint("Orgs Loaded  \n")
}

func getUserAccountsRolesList(count uint64) {
	var loopCount uint64

	//-- Init Map
	HornbillCache.UserRoles = make(map[string][]string)
	bar := pb.StartNew(int(count))
	//-- Load Results in pages of pageSize
	for loopCount < count {
		logger(1, "Loading User Accounts Roles List Offset: "+fmt.Sprintf("%d", loopCount), false)

		hornbillImport.SetParam("application", "com.hornbill.core")
		hornbillImport.SetParam("queryName", "getUserAccountsRolesList")
		hornbillImport.OpenElement("queryParams")
		hornbillImport.SetParam("rowstart", strconv.FormatUint(loopCount, 10))
		hornbillImport.SetParam("limit", strconv.Itoa(pageSize))
		hornbillImport.CloseElement("queryParams")
		RespBody, xmlmcErr := hornbillImport.Invoke("data", "queryExec")

		var JSONResp xmlmcUserRolesListResponse
		if xmlmcErr != nil {
			logger(4, "Unable to Query Accounts Roles List "+fmt.Sprintf("%s", xmlmcErr), false)
		}
		err := json.Unmarshal([]byte(RespBody), &JSONResp)
		if err != nil {
			logger(4, "Unable to Query Accounts Roles  List "+fmt.Sprintf("%s", err), false)
		}
		if JSONResp.State.Error != "" {
			logger(4, "Unable to Query Accounts Roles  List "+JSONResp.State.Error, false)
		}

		//-- Push into Map of slices to userId = array of roles
		for index := range JSONResp.Params.RowData.Row {
			if userIDExistsInLDAP(JSONResp.Params.RowData.Row[index].HUserID) {
				HornbillCache.UserRoles[strings.ToLower(JSONResp.Params.RowData.Row[index].HUserID)] = append(HornbillCache.UserRoles[strings.ToLower(JSONResp.Params.RowData.Row[index].HUserID)], JSONResp.Params.RowData.Row[index].HRole)
			} else {

			}
		}
		// Add 100
		loopCount += uint64(pageSize)
		bar.Add(len(JSONResp.Params.RowData.Row))
	}
	bar.FinishPrint("Account Roles Loaded  \n")
}
func getUserAccountList(count uint64) {
	var loopCount uint64
	//-- Init Map
	HornbillCache.Users = make(map[string]userAccountStruct)
	//-- Load Results in pages of pageSize
	bar := pb.StartNew(int(count))
	for loopCount < count {
		logger(1, "Loading User Accounts List Offset: "+fmt.Sprintf("%d", loopCount)+"\n", false)

		hornbillImport.SetParam("application", "com.hornbill.core")
		hornbillImport.SetParam("queryName", "getUserAccountsList")
		hornbillImport.OpenElement("queryParams")
		hornbillImport.SetParam("rowstart", strconv.FormatUint(loopCount, 10))
		hornbillImport.SetParam("limit", strconv.Itoa(pageSize))
		hornbillImport.CloseElement("queryParams")
		RespBody, xmlmcErr := hornbillImport.Invoke("data", "queryExec")

		var JSONResp xmlmcUserListResponse
		if xmlmcErr != nil {
			logger(4, "Unable to Query Accounts List "+fmt.Sprintf("%s", xmlmcErr), false)
		}
		err := json.Unmarshal([]byte(RespBody), &JSONResp)
		if err != nil {
			logger(4, "Unable to Query Accounts List "+fmt.Sprintf("%s", err), false)
		}
		if JSONResp.State.Error != "" {
			logger(4, "Unable to Query Accounts List "+JSONResp.State.Error, false)
		}
		//-- Push into Map
		for index := range JSONResp.Params.RowData.Row {
			//-- Store All Users so we can search later for manager on HName
			//-- This is better than calling back to the instance
			HornbillCache.Users[strings.ToLower(JSONResp.Params.RowData.Row[index].HUserID)] = JSONResp.Params.RowData.Row[index]
		}

		// Add 100
		loopCount += uint64(pageSize)
		bar.Add(len(JSONResp.Params.RowData.Row))
	}
	bar.FinishPrint("Accounts Loaded  \n")
}
func getSitesList(count uint64) {
	var loopCount uint64
	//-- Init Map
	HornbillCache.Sites = make(map[string]siteStruct)
	//-- Load Results in pages of pageSize
	bar := pb.StartNew(int(count))
	for loopCount < count {
		logger(1, "Loading Sites List Offset: "+fmt.Sprintf("%d", loopCount), false)

		hornbillImport.SetParam("application", "com.hornbill.core")
		hornbillImport.SetParam("queryName", "getSitesList")
		hornbillImport.OpenElement("queryParams")
		hornbillImport.SetParam("rowstart", strconv.FormatUint(loopCount, 10))
		hornbillImport.SetParam("limit", strconv.Itoa(pageSize))
		hornbillImport.CloseElement("queryParams")
		RespBody, xmlmcErr := hornbillImport.Invoke("data", "queryExec")

		var JSONResp xmlmcSiteListResponse
		if xmlmcErr != nil {
			logger(4, "Unable to Query Site List "+fmt.Sprintf("%s", xmlmcErr), false)
		}
		err := json.Unmarshal([]byte(RespBody), &JSONResp)
		if err != nil {
			logger(4, "Unable to Query Site List "+fmt.Sprintf("%s", err), false)
		}
		if JSONResp.State.Error != "" {
			logger(4, "Unable to Query Site List "+JSONResp.State.Error, false)
		}

		//-- Push into Map
		for index := range JSONResp.Params.RowData.Row {
			HornbillCache.Sites[strings.ToLower(JSONResp.Params.RowData.Row[index].HSiteName)] = JSONResp.Params.RowData.Row[index]
		}
		// Add 100
		loopCount += uint64(pageSize)
		bar.Add(len(JSONResp.Params.RowData.Row))
	}
	bar.FinishPrint("Sites Loaded  \n")

}
func getCount(query string) uint64 {

	hornbillImport.SetParam("application", "com.hornbill.core")
	hornbillImport.SetParam("queryName", query)
	hornbillImport.OpenElement("queryParams")
	hornbillImport.SetParam("getCount", "true")
	hornbillImport.CloseElement("queryParams")

	RespBody, xmlmcErr := hornbillImport.Invoke("data", "queryExec")

	var JSONResp xmlmcCountResponse
	if xmlmcErr != nil {
		logger(4, "Unable to run Query ["+query+"] "+fmt.Sprintf("%s", xmlmcErr), false)
		return 0
	}
	err := json.Unmarshal([]byte(RespBody), &JSONResp)
	if err != nil {
		logger(4, "Unable to run Query ["+query+"] "+fmt.Sprintf("%s", err), false)
		return 0
	}
	if JSONResp.State.Error != "" {
		logger(4, "Unable to run Query ["+query+"] "+JSONResp.State.Error, false)
		return 0
	}

	//-- return Count
	count, _ := strconv.ParseUint(JSONResp.Params.RowData.Row[0].Count, 10, 16)
	return count
}
