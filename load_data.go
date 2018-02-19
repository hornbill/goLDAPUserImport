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
	pageSize       uint64
)

func initXMLMC() {

	hornbillImport = apiLib.NewXmlmcInstance(ldapImportConf.InstanceID)
	hornbillImport.SetAPIKey(ldapImportConf.APIKey)
	hornbillImport.SetTimeout(5)
	hornbillImport.SetJSONResponse(true)
	pageSize = 100
}
func loadUsers() {
	//-- Init One connection to Hornbill to load all data
	initXMLMC()
	logger(1, "Loading Users from Hornbill", true)

	count := getCount("getUserAccountsList")
	getUserAccountList(count)

	logger(1, "Users Loaded: "+fmt.Sprintf("%d", len(HornbillCache.Users)), true)
}
func loadUsersRoles() {
	//-- Only Load if Enabled
	if ldapImportConf.User.Role.Action != "Create" && ldapImportConf.User.Role.Action != "Update" && ldapImportConf.User.Role.Action != "Both" {
		logger(1, "Skipping Loading Roles Due to Config", false)
		return
	}

	logger(1, "Loading Users Roles from Hornbill", true)

	count := getCount("getUserAccountsRolesList")
	getUserAccountsRolesList(count)

	logger(1, "Users Roles Loaded: "+fmt.Sprintf("%d", len(HornbillCache.UserRoles)), true)
}

func loadSites() {
	//-- Only Load if Enabled
	if ldapImportConf.User.Site.Action != "Create" && ldapImportConf.User.Site.Action != "Update" && ldapImportConf.User.Site.Action != "Both" {
		logger(1, "Skipping Loading Sites Due to Config", false)
		return
	}

	logger(1, "Loading Sites from Hornbill", true)

	count := getCount("getSitesList")
	getSitesList(count)

	logger(1, "Sites Loaded: "+fmt.Sprintf("%d", len(HornbillCache.Sites)), true)
}
func loadGroups() {
	//-- Only Load if Enabled
	if ldapImportConf.User.Org.Action != "Create" && ldapImportConf.User.Org.Action != "Update" && ldapImportConf.User.Org.Action != "Both" {
		logger(1, "Skipping Loading Orgs Due to Config", false)
		return
	}

	logger(1, "Loading Orgs from Hornbill", true)

	count := getCount("getGroupsList")
	getGroupsList(count)

	logger(1, "Orgs Loaded: "+fmt.Sprintf("%d", len(HornbillCache.Groups)), true)
}
func loadUserGroups() {
	//-- Only Load if Enabled
	if ldapImportConf.User.Org.Action != "Create" && ldapImportConf.User.Org.Action != "Update" && ldapImportConf.User.Org.Action != "Both" {
		logger(1, "Skipping Loading User Orgs Due to Config", false)
		return
	}

	logger(1, "Loading User Orgs from Hornbill", true)

	count := getCount("getUserAccountsGroupsList")
	getUserAccountsGroupsList(count)

	logger(1, "User Orgs Loaded: "+fmt.Sprintf("%d", len(HornbillCache.UserGroups))+"\n", true)
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
		logger(1, "Loading User Accounts Groups List Offset: "+fmt.Sprintf("%d", loopCount), false)

		hornbillImport.SetParam("application", "com.hornbill.core")
		hornbillImport.SetParam("queryName", "getUserAccountsGroupsList")
		hornbillImport.OpenElement("queryParams")
		hornbillImport.SetParam("rowstart", strconv.FormatUint(loopCount, 10))
		hornbillImport.SetParam("limit", strconv.FormatUint(pageSize, 10))
		hornbillImport.CloseElement("queryParams")
		RespBody, xmlmcErr := hornbillImport.Invoke("data", "queryExec")

		var JSONResp xmlmcUserGroupListResponse
		if xmlmcErr != nil {
			logger(4, "Unable to Query Accounts Groups List "+fmt.Sprintf("%s", xmlmcErr), true)
		}
		err := json.Unmarshal([]byte(RespBody), &JSONResp)
		if err != nil {
			logger(4, "Unable to Query Accounts Groups  List "+fmt.Sprintf("%s", err), true)
		}
		if JSONResp.State.Error != "" {
			logger(4, "Unable to Query Accounts Groups  List "+JSONResp.State.Error, true)
		}

		//-- Push into Map of slices to userId = array of roles
		for index := range JSONResp.Params.RowData.Row {
			if userIDExistsInLDAP(JSONResp.Params.RowData.Row[index].HUserID) {
				HornbillCache.UserGroups[strings.ToLower(JSONResp.Params.RowData.Row[index].HUserID)] = append(HornbillCache.UserGroups[JSONResp.Params.RowData.Row[index].HUserID], JSONResp.Params.RowData.Row[index].HGroupID)
			}
		}
		// Add 100
		loopCount += pageSize
		bar.Add(len(JSONResp.Params.RowData.Row))
	}
	bar.FinishPrint("Acount Groups Loaded")

}
func getGroupsList(count uint64) {
	var loopCount uint64
	//-- Init Map
	HornbillCache.Groups = make(map[string]string)
	//-- Load Results in pages of pageSize
	bar := pb.StartNew(int(count))
	for loopCount < count {
		logger(1, "Loading Orgs List Offset: "+fmt.Sprintf("%d", loopCount), false)

		hornbillImport.SetParam("application", "com.hornbill.core")
		hornbillImport.SetParam("queryName", "getGroupsList")
		hornbillImport.OpenElement("queryParams")
		hornbillImport.SetParam("rowstart", strconv.FormatUint(loopCount, 10))
		hornbillImport.SetParam("limit", strconv.FormatUint(pageSize, 10))
		hornbillImport.CloseElement("queryParams")
		RespBody, xmlmcErr := hornbillImport.Invoke("data", "queryExec")

		var JSONResp xmlmcGroupListResponse
		if xmlmcErr != nil {
			logger(4, "Unable to Query Orgs List "+fmt.Sprintf("%s", xmlmcErr), true)
		}
		err := json.Unmarshal([]byte(RespBody), &JSONResp)
		if err != nil {
			logger(4, "Unable to Query Orgs List "+fmt.Sprintf("%s", err), true)
		}
		if JSONResp.State.Error != "" {
			logger(4, "Unable to Query Orgs List "+JSONResp.State.Error, true)
		}

		//-- Push into Map
		for index := range JSONResp.Params.RowData.Row {
			HornbillCache.Groups[strings.ToLower(JSONResp.Params.RowData.Row[index].HName)] = JSONResp.Params.RowData.Row[index].HID
		}
		// Add 100
		loopCount += pageSize
		bar.Add(len(JSONResp.Params.RowData.Row))
	}
	bar.FinishPrint("Orgs Loaded")
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
		hornbillImport.SetParam("limit", strconv.FormatUint(pageSize, 10))
		hornbillImport.CloseElement("queryParams")
		RespBody, xmlmcErr := hornbillImport.Invoke("data", "queryExec")

		var JSONResp xmlmcUserRolesListResponse
		if xmlmcErr != nil {
			logger(4, "Unable to Query Accounts Roles List "+fmt.Sprintf("%s", xmlmcErr), true)
		}
		err := json.Unmarshal([]byte(RespBody), &JSONResp)
		if err != nil {
			logger(4, "Unable to Query Accounts Roles  List "+fmt.Sprintf("%s", err), true)
		}
		if JSONResp.State.Error != "" {
			logger(4, "Unable to Query Accounts Roles  List "+JSONResp.State.Error, true)
		}

		//-- Push into Map of slices to userId = array of roles
		for index := range JSONResp.Params.RowData.Row {
			if userIDExistsInLDAP(JSONResp.Params.RowData.Row[index].HUserID) {
				HornbillCache.UserRoles[strings.ToLower(JSONResp.Params.RowData.Row[index].HUserID)] = append(HornbillCache.UserRoles[JSONResp.Params.RowData.Row[index].HUserID], JSONResp.Params.RowData.Row[index].HRole)
			}
		}
		// Add 100
		loopCount += pageSize
		bar.Add(len(JSONResp.Params.RowData.Row))
	}
	bar.FinishPrint("Account Roles Loaded")
}
func getUserAccountList(count uint64) {
	var loopCount uint64
	//-- Init Map
	HornbillCache.Users = make(map[string]userAccountStruct)
	//-- Load Results in pages of pageSize
	bar := pb.StartNew(int(count))
	for loopCount < count {
		logger(1, "Loading User Accounts List Offset: "+fmt.Sprintf("%d", loopCount), false)

		hornbillImport.SetParam("application", "com.hornbill.core")
		hornbillImport.SetParam("queryName", "getUserAccountsList")
		hornbillImport.OpenElement("queryParams")
		hornbillImport.SetParam("rowstart", strconv.FormatUint(loopCount, 10))
		hornbillImport.SetParam("limit", strconv.FormatUint(pageSize, 10))
		hornbillImport.CloseElement("queryParams")
		RespBody, xmlmcErr := hornbillImport.Invoke("data", "queryExec")

		var JSONResp xmlmcUserListResponse
		if xmlmcErr != nil {
			logger(4, "Unable to Query Accounts List "+fmt.Sprintf("%s", xmlmcErr), true)
		}
		err := json.Unmarshal([]byte(RespBody), &JSONResp)
		if err != nil {
			logger(4, "Unable to Query Accounts List "+fmt.Sprintf("%s", err), true)
		}
		if JSONResp.State.Error != "" {
			logger(4, "Unable to Query Accounts List "+JSONResp.State.Error, true)
		}
		//-- Push into Map
		for index := range JSONResp.Params.RowData.Row {
			//-- Store All Users so we can search later for manager on HName
			//-- This is better than calling back to the instance
			HornbillCache.Users[strings.ToLower(JSONResp.Params.RowData.Row[index].HUserID)] = JSONResp.Params.RowData.Row[index]
		}

		// Add 100
		loopCount += pageSize
		bar.Add(len(JSONResp.Params.RowData.Row))
	}
	bar.FinishPrint("Accounts Loaded")
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
		hornbillImport.SetParam("limit", strconv.FormatUint(pageSize, 10))
		hornbillImport.CloseElement("queryParams")
		RespBody, xmlmcErr := hornbillImport.Invoke("data", "queryExec")

		var JSONResp xmlmcSiteListResponse
		if xmlmcErr != nil {
			logger(4, "Unable to Query Site List "+fmt.Sprintf("%s", xmlmcErr), true)
		}
		err := json.Unmarshal([]byte(RespBody), &JSONResp)
		if err != nil {
			logger(4, "Unable to Query Site List "+fmt.Sprintf("%s", err), true)
		}
		if JSONResp.State.Error != "" {
			logger(4, "Unable to Query Site List "+JSONResp.State.Error, true)
		}

		//-- Push into Map
		for index := range JSONResp.Params.RowData.Row {
			HornbillCache.Sites[strings.ToLower(JSONResp.Params.RowData.Row[index].HSiteName)] = JSONResp.Params.RowData.Row[index]
		}
		// Add 100
		loopCount += pageSize
		bar.Add(len(JSONResp.Params.RowData.Row))
	}
	bar.FinishPrint("Sites Loaded")

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
		logger(4, "Unable to run Query ["+query+"] "+fmt.Sprintf("%s", xmlmcErr), true)
		return 0
	}
	err := json.Unmarshal([]byte(RespBody), &JSONResp)
	if err != nil {
		logger(4, "Unable to run Query ["+query+"] "+fmt.Sprintf("%s", err), true)
		return 0
	}
	if JSONResp.State.Error != "" {
		logger(4, "Unable to run Query ["+query+"] "+JSONResp.State.Error, true)
		return 0
	}

	//-- return Count
	count, _ := strconv.ParseUint(JSONResp.Params.RowData.Row[0].Count, 10, 16)
	return count
}
