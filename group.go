package main

import (
	"strings"

	"github.com/hornbill/ldap"
)

func getOrgFromLookup(l *ldap.Entry, orgValue string) string {

	//-- Check if Site Attribute is set
	if orgValue == "" {
		logger(1, "Org Lookup is Enabled but Attribute is not Defined", false)
		return ""
	}
	//-- Get Value of Attribute
	logger(1, "LDAP Attribute for Org Lookup: "+orgValue, false)
	orgAttributeName := processComplexFeild(l, orgValue)
	logger(1, "Looking Up Org "+orgAttributeName, false)
	_, found := HornbillCache.Groups[strings.ToLower(orgAttributeName)]
	if found {
		logger(1, "Org Lookup found Id "+HornbillCache.Groups[strings.ToLower(orgAttributeName)], false)
		return HornbillCache.Groups[strings.ToLower(orgAttributeName)]
	}
	logger(1, "Unable to Find Organsiation "+orgAttributeName, false)
	return ""
}
func isUserAMember(l *ldap.Entry, memberOf string) bool {
	userAdGroups := l.GetAttributeValues("memberof")
	for index := range userAdGroups {
		if userAdGroups[index] == memberOf {
			logger(1, "User is a Member of Ad Group: "+memberOf, false)
			return true
		}
	}
	logger(1, "User is not a Member of Ad Group: "+memberOf, false)
	return false
}
