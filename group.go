package main

import (
	"strings"
)

func getOrgFromLookup(importDate *userWorkingDataStruct) string {

	//-- Check if Site Attribute is set
	if ldapImportConf.User.Org.Value == "" {
		logger(1, "Org Lookup is Enabled but Attribute is not Defined", false)
		return ""
	}
	//-- Get Value of Attribute
	logger(1, "LDAP Attribute for Org Lookup: "+ldapImportConf.User.Org.Value, false)
	orgAttributeName := processComplexFeild(importDate.LDAP, ldapImportConf.User.Org.Value)
	logger(1, "Looking Up Org "+orgAttributeName, false)
	_, found := HornbillCache.Groups[strings.ToLower(orgAttributeName)]
	if found {
		logger(1, "Org Lookup found Id "+HornbillCache.Groups[strings.ToLower(orgAttributeName)], false)
		return HornbillCache.Groups[strings.ToLower(orgAttributeName)]
	}
	logger(1, "Unable to Find Organsiation "+orgAttributeName, false)
	return ""
}
