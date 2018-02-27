package main

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/hornbill/ldap"
)

//-- Store LDAP Usres in Map
func processLDAPUsers() {
	logger(1, "Processing LDAP User Data", true)

	//-- User Working Data
	HornbillCache.UsersWorking = make(map[string]*userWorkingDataStruct)
	HornbillCache.Managers = make(map[string]string)
	HornbillCache.DN = make(map[string]string)
	HornbillCache.Images = make(map[string]imageStruct)
	//-- Loop LDAP Users
	for user := range ldapUsers {
		// Process Params and return userId
		var userID = processUserParams(ldapUsers[user])
		if userID != "" {
			var userDN = processComplexFeild(ldapUsers[user], ldapImportConf.User.UserDN)
			//-- Write to Cache
			writeUserToCache(userDN, userID)
		}
	}

	logger(1, "LDAP User Data Processed \n", true)
}
func processData() {
	logger(1, "Processing User Data", true)

	for user := range HornbillCache.UsersWorking {

		currentUser := HornbillCache.UsersWorking[user]
		//-- Current UserID
		userID := strings.ToLower(currentUser.Account.UserID)
		hornbillUserData := HornbillCache.Users[userID]

		if userID == "" {
			continue
		}
		//-- Check Map no need to loop
		if strings.ToLower(hornbillUserData.HUserID) == userID {

			currentUser.Jobs.update = checkUserNeedsUpdate(currentUser, hornbillUserData)

			currentUser.Jobs.updateProfile = checkUserNeedsProfileUpdate(currentUser, hornbillUserData)

			currentUser.Jobs.updateType = checkUserNeedsTypeUpdate(currentUser, hornbillUserData)

			currentUser.Jobs.updateSite = checkUserNeedsSiteUpdate(currentUser, hornbillUserData)

			currentUser.Jobs.updateImage = checkUserNeedsImageUpdate(currentUser, hornbillUserData)

			checkUserNeedsOrgUpdate(currentUser, hornbillUserData)

			checkUserNeedsOrgRemoving(currentUser, hornbillUserData)

			checkUserNeedsRoleUpdate(currentUser, hornbillUserData)
		} else {
			//-- Check for Password
			setUserPasswordValueForCreate(currentUser)
			//-- Set Site ID Based on Config
			setUserSiteValueForCreate(currentUser, hornbillUserData)
			setUserRolesalueForCreate(currentUser, hornbillUserData)
			currentUser.Jobs.updateImage = checkUserNeedsImageCreate(currentUser, hornbillUserData)
			checkUserNeedsOrgCreate(currentUser, hornbillUserData)
			currentUser.Jobs.create = true
		}

		logger(1, "User: "+fmt.Sprintf("%s", userID)+" \n\tCreate: "+fmt.Sprintf("%t", currentUser.Jobs.create)+" \n\tUpdate: "+fmt.Sprintf("%t", currentUser.Jobs.update)+" \n\tUpdate Type: "+fmt.Sprintf("%t", currentUser.Jobs.updateType)+" \n\tUpdate Profile: "+fmt.Sprintf("%t", currentUser.Jobs.updateProfile)+" \n\tUpdate Site: "+fmt.Sprintf("%t", currentUser.Jobs.updateSite)+" \n\tRoles Count: "+fmt.Sprintf("%d", len(currentUser.Roles))+" \n\tUpdate Image: "+fmt.Sprintf("%t", currentUser.Jobs.updateImage)+" \n\tGroups: "+fmt.Sprintf("%d", len(currentUser.Groups))+"\n", false)
	}
	logger(1, "User Data Processed", true)
}
func checkUserNeedsOrgRemoving(importData *userWorkingDataStruct, currentData userAccountStruct) {
	for _, group := range importData.Groups {
		if group.OnlyOneGroupAssignment {
			var userExistingGroups = HornbillCache.UserGroups[strings.ToLower(importData.Account.UserID)]
			for index := range userExistingGroups {
				groupID := userExistingGroups[index]
				cacheGroup := HornbillCache.GroupsId[strings.ToLower(groupID)]
				//-- If not current group and its the same type then remove
				if groupID != group.ID && group.Type == cacheGroup.Type {
					//-- Push Group into list to remove
					importData.GroupsToRemove = append(importData.GroupsToRemove, groupID)
				}
			}
		}
	}
}
func setUserPasswordValueForCreate(importData *userWorkingDataStruct) {
	if importData.Account.Password == "" {
		//-- Generate Password
		importData.Account.Password = generatePasswordString(10)
		logger(1, "Auto Generated Password for: "+importData.Account.UserID+" - "+importData.Account.Password, false)
	}
	//-- Base64 Encode
	importData.Account.Password = base64.StdEncoding.EncodeToString([]byte(importData.Account.Password))
}
func checkUserNeedsOrgUpdate(importData *userWorkingDataStruct, currentData userAccountStruct) {
	if len(ldapImportConf.User.Org) > 0 {
		for orgIndex := range ldapImportConf.User.Org {
			orgAction := ldapImportConf.User.Org[orgIndex]
			if orgAction.Action == "Both" || orgAction.Action == "Update" {
				var GroupID = getOrgFromLookup(importData.LDAP, orgAction.Value)
				var userExistingGroups = HornbillCache.UserGroups[strings.ToLower(importData.Account.UserID)]
				//-- Is User Already a Memeber of the Group
				boolUserInGroup := false
				for index := range userExistingGroups {
					if strings.ToLower(GroupID) == strings.ToLower(userExistingGroups[index]) {
						boolUserInGroup = true
					}
				}
				if !boolUserInGroup && GroupID != "" {
					//-- Check User is a member of
					if orgAction.MemberOf != "" {
						if !isUserAMember(importData.LDAP, orgAction.MemberOf) {
							continue
						}
					}
					var group userGroupStruct
					group.ID = GroupID
					group.Name = orgAction.Value
					group.Type = orgAction.Options.Type
					group.Membership = orgAction.Options.Membership
					group.TasksView = orgAction.Options.TasksView
					group.TasksAction = orgAction.Options.TasksAction
					group.OnlyOneGroupAssignment = orgAction.Options.OnlyOneGroupAssignment

					importData.Groups = append(importData.Groups, group)
				}
			}
		}
	}
}
func checkUserNeedsOrgCreate(importData *userWorkingDataStruct, currentData userAccountStruct) {
	if len(ldapImportConf.User.Org) > 0 {
		for orgIndex := range ldapImportConf.User.Org {
			orgAction := ldapImportConf.User.Org[orgIndex]
			if orgAction.Action == "Both" || orgAction.Action == "Create" {

				var GroupID = getOrgFromLookup(importData.LDAP, orgAction.Value)

				if GroupID != "" && orgAction.MemberOf != "" {
					if !isUserAMember(importData.LDAP, orgAction.MemberOf) {
						continue
					}
				}
				var group userGroupStruct
				group.ID = GroupID
				group.Name = orgAction.Value
				group.Type = orgAction.Options.Type
				group.Membership = orgAction.Options.Membership
				group.TasksView = orgAction.Options.TasksView
				group.TasksAction = orgAction.Options.TasksAction
				group.OnlyOneGroupAssignment = orgAction.Options.OnlyOneGroupAssignment

				if GroupID != "" {
					importData.Groups = append(importData.Groups, group)
				}
			}
		}
	}
}
func setUserRolesalueForCreate(importData *userWorkingDataStruct, currentData userAccountStruct) {

	if ldapImportConf.User.Role.Action == "Both" || ldapImportConf.User.Role.Action == "Create" {
		importData.Roles = ldapImportConf.User.Role.Roles
	}
}
func checkUserNeedsRoleUpdate(importData *userWorkingDataStruct, currentData userAccountStruct) {

	if ldapImportConf.User.Role.Action == "Both" || ldapImportConf.User.Role.Action == "Update" {
		for index := range ldapImportConf.User.Role.Roles {
			roleName := ldapImportConf.User.Role.Roles[index]
			foundRole := false
			var userRoles = HornbillCache.UserRoles[strings.ToLower(importData.Account.UserID)]
			for index2 := range userRoles {
				if strings.ToLower(roleName) == strings.ToLower(userRoles[index2]) {
					foundRole = true
				}
			}
			if !foundRole {
				importData.Roles = append(importData.Roles, roleName)
			}
		}
	}
}
func checkUserNeedsImageCreate(importData *userWorkingDataStruct, currentData userAccountStruct) bool {
	//-- Is Type Enables for Update or both
	if ldapImportConf.User.Image.Action == "Both" || ldapImportConf.User.Image.Action == "Create" {

		//-- Check for Empty URI
		if ldapImportConf.User.Image.URI == "" {
			return false
		}
		image := getImage(importData)
		// check for changes
		if image.imageCheckSum != currentData.HIconChecksum {
			return true
		}
	}
	return false
}
func checkUserNeedsImageUpdate(importData *userWorkingDataStruct, currentData userAccountStruct) bool {
	//-- Is Type Enables for Update or both
	if ldapImportConf.User.Image.Action == "Both" || ldapImportConf.User.Image.Action == "Update" {

		//-- Check for Empty URI
		if ldapImportConf.User.Image.URI == "" {
			return false
		}
		image := getImage(importData)
		// check for changes
		if image.imageCheckSum != currentData.HIconChecksum {
			return true
		}
	}
	return false
}
func checkUserNeedsTypeUpdate(importData *userWorkingDataStruct, currentData userAccountStruct) bool {
	//-- Is Type Enables for Update or both
	if ldapImportConf.User.Type.Action == "Both" || ldapImportConf.User.Type.Action == "Update" {
		// -- 1 = user
		// -- 3 = basic
		switch importData.Account.UserType {
		case "user":
			if currentData.HClass != "1" {
				return true
			}
			break
		case "basic":
			if currentData.HClass != "3" {
				return true
			}
			break
		default:
			return false
		}
	} else {
		if currentData.HClass == "1" {
			importData.Account.UserType = "user"
		} else {
			importData.Account.UserType = "basic"
		}
	}
	return false
}
func setUserSiteValueForCreate(importData *userWorkingDataStruct, currentData userAccountStruct) bool {
	//-- Is Site Enables for Update or both
	if ldapImportConf.User.Site.Action == "Both" || ldapImportConf.User.Site.Action == "Create" {
		importData.Account.Site = getSiteFromLookup(importData.LDAP)
	}
	if importData.Account.Site != "" && importData.Account.Site != currentData.HSite {
		return true
	}
	return false
}
func checkUserNeedsSiteUpdate(importData *userWorkingDataStruct, currentData userAccountStruct) bool {
	//-- Is Site Enables for Update or both
	if ldapImportConf.User.Site.Action == "Both" || ldapImportConf.User.Site.Action == "Update" {
		importData.Account.Site = getSiteFromLookup(importData.LDAP)
	} else {
		//-- Else Default to current value
		importData.Account.Site = currentData.HSite
	}

	if importData.Account.Site != "" && importData.Account.Site != currentData.HSite {
		return true
	}
	return false
}
func checkUserNeedsUpdate(importData *userWorkingDataStruct, currentData userAccountStruct) bool {
	if importData.Account.Name != "" && importData.Account.Name != currentData.HName {
		logger(1, "Name: "+fmt.Sprintf("%s", importData.Account.Name)+" - "+fmt.Sprintf("%s", currentData.HName), true)
		return true
	}
	if importData.Account.FirstName != "" && importData.Account.FirstName != currentData.HFirstName {
		logger(1, "FirstName: "+fmt.Sprintf("%s", importData.Account.FirstName)+" - "+fmt.Sprintf("%s", currentData.HFirstName), true)
		return true
	}
	if importData.Account.LastName != "" && importData.Account.LastName != currentData.HLastName {
		logger(1, "LastName: "+fmt.Sprintf("%s", importData.Account.LastName)+" - "+fmt.Sprintf("%s", currentData.HLastName), true)
		return true
	}
	if importData.Account.JobTitle != "" && importData.Account.JobTitle != currentData.HJobTitle {
		logger(1, "JobTitle: "+fmt.Sprintf("%s", importData.Account.JobTitle)+" - "+fmt.Sprintf("%s", currentData.HJobTitle), true)
		return true
	}
	if importData.Account.Phone != "" && importData.Account.Phone != currentData.HPhone {
		logger(1, "Phone: "+fmt.Sprintf("%s", importData.Account.Phone)+" - "+fmt.Sprintf("%s", currentData.HPhone), true)
		return true
	}
	if importData.Account.Email != "" && importData.Account.Email != currentData.HEmail {
		logger(1, "Email: "+fmt.Sprintf("%s", importData.Account.Email)+" - "+fmt.Sprintf("%s", currentData.HEmail), true)
		return true
	}
	if importData.Account.Mobile != "" && importData.Account.Mobile != currentData.HMobile {
		logger(1, "Mobile: "+fmt.Sprintf("%s", importData.Account.Mobile)+" - "+fmt.Sprintf("%s", currentData.HMobile), true)
		return true
	}
	if importData.Account.AbsenceMessage != "" && importData.Account.AbsenceMessage != currentData.HAvailStatusMsg {
		logger(1, "AbsenceMessage: "+fmt.Sprintf("%s", importData.Account.AbsenceMessage)+" - "+fmt.Sprintf("%s", currentData.HAvailStatusMsg), true)
		return true
	}
	//-- If TimeZone mapping is empty then ignore as it defaults to a value
	if importData.Account.TimeZone != "" && importData.Account.TimeZone != currentData.HTimezone {
		logger(1, "TimeZone: "+fmt.Sprintf("%s", importData.Account.TimeZone)+" - "+fmt.Sprintf("%s", currentData.HTimezone), true)
		return true
	}
	//-- If Language mapping is empty then ignore as it defaults to a value
	if importData.Account.Language != "" && importData.Account.Language != currentData.HLanguage {
		logger(1, "Language: "+fmt.Sprintf("%s", importData.Account.Language)+" - "+fmt.Sprintf("%s", currentData.HLanguage), true)
		return true
	}
	//-- If DateTimeFormat mapping is empty then ignore as it defaults to a value
	if importData.Account.DateTimeFormat != "" && importData.Account.DateTimeFormat != currentData.HDateTimeFormat {
		logger(1, "DateTimeFormat: "+fmt.Sprintf("%s", importData.Account.DateTimeFormat)+" - "+fmt.Sprintf("%s", currentData.HDateTimeFormat), true)
		return true
	}
	//-- If DateFormat mapping is empty then ignore as it defaults to a value
	if importData.Account.DateFormat != "" && importData.Account.DateFormat != currentData.HDateFormat {
		logger(1, "DateFormat: "+fmt.Sprintf("%s", importData.Account.DateFormat)+" - "+fmt.Sprintf("%s", currentData.HDateFormat), true)
		return true
	}
	//-- If TimeFormat mapping is empty then ignore as it defaults to a value
	if importData.Account.TimeFormat != "" && importData.Account.TimeFormat != currentData.HTimeFormat {
		logger(1, "TimeFormat: "+fmt.Sprintf("%s", importData.Account.TimeFormat)+" - "+fmt.Sprintf("%s", currentData.HTimeFormat), true)
		return true
	}
	//-- If CurrencySymbol mapping is empty then ignore as it defaults to a value
	if importData.Account.CurrencySymbol != "" && importData.Account.CurrencySymbol != currentData.HCurrencySymbol {
		logger(1, "CurrencySymbol: "+fmt.Sprintf("%s", importData.Account.CurrencySymbol)+" - "+fmt.Sprintf("%s", currentData.HCurrencySymbol), true)
		return true
	}
	//-- If CountryCode mapping is empty then ignore as it defaults to a value
	if importData.Account.CountryCode != "" && importData.Account.CountryCode != currentData.HCountry {
		logger(1, "CountryCode: "+fmt.Sprintf("%s", importData.Account.CountryCode)+" - "+fmt.Sprintf("%s", currentData.HCountry), true)
		return true
	}

	return false
}
func checkUserNeedsProfileUpdate(importData *userWorkingDataStruct, currentData userAccountStruct) bool {

	if importData.Profile.MiddleName != "" && importData.Profile.MiddleName != currentData.HMiddleName {
		logger(1, "MiddleName: "+fmt.Sprintf("%s", importData.Profile.MiddleName)+" - "+fmt.Sprintf("%s", currentData.HMiddleName), true)
		return true
	}

	if importData.Profile.JobDescription != "" && importData.Profile.JobDescription != currentData.HSummary {
		logger(1, "JobDescription: "+fmt.Sprintf("%s", importData.Profile.JobDescription)+" - "+fmt.Sprintf("%s", currentData.HSummary), true)
		return true
	}
	if ldapImportConf.User.Manager.Action == "Both" || ldapImportConf.User.Manager.Action == "Update" {
		importData.Profile.Manager = getManager(importData, currentData)
	} else {
		//-- Use Current Value
		importData.Profile.Manager = currentData.HManager
	}
	if importData.Profile.Manager != "" && importData.Profile.Manager != currentData.HManager {
		logger(1, "Manager: "+fmt.Sprintf("%s", importData.Profile.Manager)+" - "+fmt.Sprintf("%s", currentData.HManager), true)
		return true
	}
	if importData.Profile.WorkPhone != "" && importData.Profile.WorkPhone != currentData.HPhone {
		logger(1, "WorkPhone: "+fmt.Sprintf("%s", importData.Profile.WorkPhone)+" - "+fmt.Sprintf("%s", currentData.HPhone), true)
		return true
	}
	if importData.Profile.Qualifications != "" && importData.Profile.Qualifications != currentData.HQualifications {
		logger(1, "Qualifications: "+fmt.Sprintf("%s", importData.Profile.Qualifications)+" - "+fmt.Sprintf("%s", currentData.HQualifications), true)
		return true
	}
	if importData.Profile.Interests != "" && importData.Profile.Interests != currentData.HInterests {
		logger(1, "Interests: "+fmt.Sprintf("%s", importData.Profile.Interests)+" - "+fmt.Sprintf("%s", currentData.HInterests), true)
		return true
	}
	if importData.Profile.Expertise != "" && importData.Profile.Expertise != currentData.HSkills {
		logger(1, "Expertise: "+fmt.Sprintf("%s", importData.Profile.Expertise)+" - "+fmt.Sprintf("%s", currentData.HSkills), true)
		return true
	}
	if importData.Profile.Gender != "" && importData.Profile.Gender != currentData.HGender {
		logger(1, "Gender: "+fmt.Sprintf("%s", importData.Profile.Gender)+" - "+fmt.Sprintf("%s", currentData.HGender), true)
		return true
	}
	if importData.Profile.Dob != "" && importData.Profile.Dob != currentData.HDob {
		logger(1, "Dob: "+fmt.Sprintf("%s", importData.Profile.Dob)+" - "+fmt.Sprintf("%s", currentData.HDob), true)
		return true
	}
	if importData.Profile.Nationality != "" && importData.Profile.Nationality != currentData.HNationality {
		logger(1, "Nationality: "+fmt.Sprintf("%s", importData.Profile.Nationality)+" - "+fmt.Sprintf("%s", currentData.HNationality), true)
		return true
	}
	if importData.Profile.Religion != "" && importData.Profile.Religion != currentData.HReligion {
		logger(1, "Religion: "+fmt.Sprintf("%s", importData.Profile.Religion)+" - "+fmt.Sprintf("%s", currentData.HReligion), true)
		return true
	}
	if importData.Profile.HomeTelephone != "" && importData.Profile.HomeTelephone != currentData.HHomeTelephoneNumber {
		logger(1, "HomeTelephone: "+fmt.Sprintf("%s", importData.Profile.HomeTelephone)+" - "+fmt.Sprintf("%s", currentData.HHomeTelephoneNumber), true)
		return true
	}
	if importData.Profile.SocialNetworkA != "" && importData.Profile.SocialNetworkA != currentData.HSnA {
		logger(1, "SocialNetworkA: "+fmt.Sprintf("%s", importData.Profile.SocialNetworkA)+" - "+fmt.Sprintf("%s", currentData.HSnA), true)
		return true
	}
	if importData.Profile.SocialNetworkB != "" && importData.Profile.SocialNetworkB != currentData.HSnB {
		logger(1, "SocialNetworkB: "+fmt.Sprintf("%s", importData.Profile.SocialNetworkB)+" - "+fmt.Sprintf("%s", currentData.HSnB), true)
		return true
	}
	if importData.Profile.SocialNetworkC != "" && importData.Profile.SocialNetworkC != currentData.HSnC {
		logger(1, "SocialNetworkC: "+fmt.Sprintf("%s", importData.Profile.SocialNetworkC)+" - "+fmt.Sprintf("%s", currentData.HSnC), true)
		return true
	}
	if importData.Profile.SocialNetworkD != "" && importData.Profile.SocialNetworkD != currentData.HSnD {
		logger(1, "SocialNetworkD: "+fmt.Sprintf("%s", importData.Profile.SocialNetworkD)+" - "+fmt.Sprintf("%s", currentData.HSnD), true)
		return true
	}
	if importData.Profile.SocialNetworkG != "" && importData.Profile.SocialNetworkG != currentData.HSnE {
		logger(1, "SocialNetworkE: "+fmt.Sprintf("%s", importData.Profile.SocialNetworkE)+" - "+fmt.Sprintf("%s", currentData.HSnE), true)
		return true
	}
	if importData.Profile.SocialNetworkG != "" && importData.Profile.SocialNetworkG != currentData.HSnF {
		logger(1, "SocialNetworkF: "+fmt.Sprintf("%s", importData.Profile.SocialNetworkF)+" - "+fmt.Sprintf("%s", currentData.HSnF), true)
		return true
	}
	if importData.Profile.SocialNetworkG != "" && importData.Profile.SocialNetworkG != currentData.HSnG {
		logger(1, "SocialNetworkG: "+fmt.Sprintf("%s", importData.Profile.SocialNetworkG)+" - "+fmt.Sprintf("%s", currentData.HSnG), true)
		return true
	}
	if importData.Profile.SocialNetworkH != "" && importData.Profile.SocialNetworkH != currentData.HSnH {
		logger(1, "SocialNetworkH: "+fmt.Sprintf("%s", importData.Profile.SocialNetworkH)+" - "+fmt.Sprintf("%s", currentData.HSnH), true)
		return true
	}
	if importData.Profile.PersonalInterests != "" && importData.Profile.PersonalInterests != currentData.HPersonalInterests {
		logger(1, "PersonalInterests: "+fmt.Sprintf("%s", importData.Profile.PersonalInterests)+" - "+fmt.Sprintf("%s", currentData.HPersonalInterests), true)
		return true
	}
	if importData.Profile.HomeAddress != "" && importData.Profile.HomeAddress != currentData.HHomeAddress {
		logger(1, "HomeAddress: "+fmt.Sprintf("%s", importData.Profile.HomeAddress)+" - "+fmt.Sprintf("%s", currentData.HHomeAddress), true)
		return true
	}
	if importData.Profile.PersonalBlog != "" && importData.Profile.PersonalBlog != currentData.HBlog {
		logger(1, "PersonalBlog: "+fmt.Sprintf("%s", importData.Profile.PersonalBlog)+" - "+fmt.Sprintf("%s", currentData.HBlog), true)
		return true
	}
	if importData.Profile.Attrib1 != "" && importData.Profile.Attrib1 != currentData.HAttrib1 {
		logger(1, "Attrib1: "+fmt.Sprintf("%s", importData.Profile.Attrib1)+" - "+fmt.Sprintf("%s", currentData.HAttrib1), true)
		return true
	}
	if importData.Profile.Attrib2 != "" && importData.Profile.Attrib2 != currentData.HAttrib2 {
		logger(1, "Attrib2: "+fmt.Sprintf("%s", importData.Profile.Attrib2)+" - "+fmt.Sprintf("%s", currentData.HAttrib2), true)
		return true
	}
	if importData.Profile.Attrib3 != "" && importData.Profile.Attrib3 != currentData.HAttrib3 {
		logger(1, "Attrib3: "+fmt.Sprintf("%s", importData.Profile.Attrib3)+" - "+fmt.Sprintf("%s", currentData.HAttrib3), true)
		return true
	}
	if importData.Profile.Attrib4 != "" && importData.Profile.Attrib4 != currentData.HAttrib4 {
		logger(1, "Attrib4: "+fmt.Sprintf("%s", importData.Profile.Attrib4)+" - "+fmt.Sprintf("%s", currentData.HAttrib4), true)
		return true
	}
	if importData.Profile.Attrib5 != "" && importData.Profile.Attrib5 != currentData.HAttrib5 {
		logger(1, "Attrib5: "+fmt.Sprintf("%s", importData.Profile.Attrib5)+" - "+fmt.Sprintf("%s", currentData.HAttrib5), true)
		return true
	}
	if importData.Profile.Attrib6 != "" && importData.Profile.Attrib6 != currentData.HAttrib6 {
		logger(1, "Attrib6: "+fmt.Sprintf("%s", importData.Profile.Attrib6)+" - "+fmt.Sprintf("%s", currentData.HAttrib6), true)
		return true
	}
	if importData.Profile.Attrib7 != "" && importData.Profile.Attrib7 != currentData.HAttrib7 {
		logger(1, "Attrib7: "+fmt.Sprintf("%s", importData.Profile.Attrib7)+" - "+fmt.Sprintf("%s", currentData.HAttrib7), true)
		return true
	}
	if importData.Profile.Attrib8 != "" && importData.Profile.Attrib8 != currentData.HAttrib8 {
		logger(1, "Attrib8: "+fmt.Sprintf("%s", importData.Profile.Attrib8)+" - "+fmt.Sprintf("%s", currentData.HAttrib8), true)
		return true
	}
	return false
}

//-- For Each LDAP User Process Account And Mappings
func processUserParams(l *ldap.Entry) string {

	//-- Set User Account Attributes
	var data = new(userWorkingDataStruct)
	data.LDAP = l
	data.Account.UserID = getUserFeildValue(l, "UserID")
	data.Account.UserType = getUserFeildValue(l, "UserType")
	data.Account.Name = getUserFeildValue(l, "Name")
	data.Account.Password = getUserFeildValue(l, "Password")
	data.Account.FirstName = getUserFeildValue(l, "FirstName")
	data.Account.LastName = getUserFeildValue(l, "LastName")
	data.Account.JobTitle = getUserFeildValue(l, "JobTitle")
	data.Account.Site = getUserFeildValue(l, "Site")
	data.Account.Phone = getUserFeildValue(l, "Phone")
	data.Account.Email = getUserFeildValue(l, "Email")
	data.Account.Mobile = getUserFeildValue(l, "Mobile")
	data.Account.AbsenceMessage = getUserFeildValue(l, "AbsenceMessage")
	data.Account.TimeZone = getUserFeildValue(l, "TimeZone")
	data.Account.Language = getUserFeildValue(l, "Language")
	data.Account.DateTimeFormat = getUserFeildValue(l, "DateTimeFormat")
	data.Account.DateFormat = getUserFeildValue(l, "DateFormat")
	data.Account.TimeFormat = getUserFeildValue(l, "TimeFormat")
	data.Account.CurrencySymbol = getUserFeildValue(l, "CurrencySymbol")
	data.Account.CountryCode = getUserFeildValue(l, "CountryCode")

	data.Profile.MiddleName = getProfileFeildValue(l, "MiddleName")
	data.Profile.JobDescription = getProfileFeildValue(l, "JobDescription")
	data.Profile.Manager = getProfileFeildValue(l, "Manager")
	data.Profile.WorkPhone = getProfileFeildValue(l, "WorkPhone")
	data.Profile.Qualifications = getProfileFeildValue(l, "Qualifications")
	data.Profile.Interests = getProfileFeildValue(l, "Interests")
	data.Profile.Expertise = getProfileFeildValue(l, "Expertise")
	data.Profile.Gender = getProfileFeildValue(l, "Gender")
	data.Profile.Dob = getProfileFeildValue(l, "Dob")
	data.Profile.Nationality = getProfileFeildValue(l, "Nationality")
	data.Profile.Religion = getProfileFeildValue(l, "Religion")
	data.Profile.HomeTelephone = getProfileFeildValue(l, "HomeTelephone")
	data.Profile.SocialNetworkA = getProfileFeildValue(l, "SocialNetworkA")
	data.Profile.SocialNetworkB = getProfileFeildValue(l, "SocialNetworkB")
	data.Profile.SocialNetworkC = getProfileFeildValue(l, "SocialNetworkC")
	data.Profile.SocialNetworkD = getProfileFeildValue(l, "SocialNetworkD")
	data.Profile.SocialNetworkE = getProfileFeildValue(l, "SocialNetworkE")
	data.Profile.SocialNetworkF = getProfileFeildValue(l, "SocialNetworkF")
	data.Profile.SocialNetworkG = getProfileFeildValue(l, "SocialNetworkG")
	data.Profile.SocialNetworkH = getProfileFeildValue(l, "SocialNetworkH")
	data.Profile.PersonalInterests = getProfileFeildValue(l, "PersonalInterests")
	data.Profile.HomeAddress = getProfileFeildValue(l, "HomeAddress")
	data.Profile.PersonalBlog = getProfileFeildValue(l, "PersonalBlog")
	data.Profile.Attrib1 = getProfileFeildValue(l, "Attrib1")
	data.Profile.Attrib2 = getProfileFeildValue(l, "Attrib2")
	data.Profile.Attrib3 = getProfileFeildValue(l, "Attrib3")
	data.Profile.Attrib4 = getProfileFeildValue(l, "Attrib4")
	data.Profile.Attrib5 = getProfileFeildValue(l, "Attrib5")
	data.Profile.Attrib6 = getProfileFeildValue(l, "Attrib6")
	data.Profile.Attrib7 = getProfileFeildValue(l, "Attrib7")
	data.Profile.Attrib8 = getProfileFeildValue(l, "Attrib8")

	//-- Store Result in map of userid
	var userID = strings.ToLower(data.Account.UserID)
	HornbillCache.UsersWorking[userID] = data

	//-- Return User Id
	return userID

}
