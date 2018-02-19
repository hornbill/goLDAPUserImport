package main

import (
	"fmt"
	"html"
	"reflect"
	"regexp"
	"strings"

	"github.com/hornbill/ldap"
)

func getUserFeildValue(u *ldap.Entry, s string) string {
	//-- Dyniamicly Grab Mapped Value
	r := reflect.ValueOf(ldapImportConf.User.AccountMapping)
	f := reflect.Indirect(r).FieldByName(s)
	//-- Get Mapped Value
	var UserMapping = f.String()
	return processComplexFeild(u, UserMapping)
}

//-- Get XMLMC Feild from mapping via profile Object
func getProfileFeildValue(u *ldap.Entry, s string) string {
	//-- Dyniamicly Grab Mapped Value
	r := reflect.ValueOf(ldapImportConf.User.ProfileMapping)

	f := reflect.Indirect(r).FieldByName(s)

	//-- Get Mapped Value
	var UserProfileMapping = f.String()
	return processComplexFeild(u, UserProfileMapping)
}
func processComplexFeild(u *ldap.Entry, s string) string {
	//-- Match $variables from String
	re1, err := regexp.Compile(`\[(.*?)\]`)
	if err != nil {
		logger(4, "Regex Error: "+fmt.Sprintf("%v", err), false)
	}
	//-- Get Array of all Matched max 100
	result := re1.FindAllString(s, 100)

	//-- Loop Matches
	for _, v := range result {
		//-- Grab LDAP Mapping value from result set
		var LDAPAttributeValue = u.GetAttributeValue(v[1 : len(v)-1])
		//-- Check for Invalid Value
		if LDAPAttributeValue == "" {
			logger(3, "Unable to Load LDAP Attribute: "+v[1:len(v)-1]+" For Input Param: "+s, false)
			return LDAPAttributeValue
		}
		//-- TK UnescapeString to HTML entities are replaced
		s = html.UnescapeString(strings.Replace(s, v, LDAPAttributeValue, 1))
	}

	//-- Return Value
	return s
}
