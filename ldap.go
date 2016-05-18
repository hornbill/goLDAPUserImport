package main

import (
	"crypto/tls"
	"fmt"

	"github.com/hornbill/ldap"
)

func connectLDAP() *ldap.LDAPConnection {
	TLSconfig := &tls.Config{
		ServerName:         ldapImportConf.LDAPServerConf.Server,
		InsecureSkipVerify: ldapImportConf.LDAPServerConf.InsecureSkipVerify,
	}
	//-- Based on Connection Type Normal | TLS | SSL
	logger(1, "Attempting Connection to LDAP... \nServer: "+ldapImportConf.LDAPServerConf.Server+"\nPort: "+fmt.Sprintf("%d", ldapImportConf.LDAPServerConf.Port)+"\nType: "+ldapImportConf.LDAPServerConf.ConnectionType+"\nSkip Verify: "+fmt.Sprintf("%t", ldapImportConf.LDAPServerConf.InsecureSkipVerify)+"\nDebug: "+fmt.Sprintf("%t", ldapImportConf.LDAPServerConf.Debug), true)
	t := ldapImportConf.LDAPServerConf.ConnectionType
	switch t {
	case "":
		//-- Normal
		logger(1, "Creating LDAP Connection", false)
		l := ldap.NewLDAPConnection(ldapImportConf.LDAPServerConf.Server, ldapImportConf.LDAPServerConf.Port)
		l.Debug = ldapImportConf.LDAPServerConf.Debug
		return l
	case "TLS":
		//-- TLS
		logger(1, "Creating LDAP Connection (TLS)", false)
		l := ldap.NewLDAPTLSConnection(ldapImportConf.LDAPServerConf.Server, ldapImportConf.LDAPServerConf.Port, TLSconfig)
		l.Debug = ldapImportConf.LDAPServerConf.Debug
		return l
	case "SSL":
		//-- SSL
		logger(1, "Creating LDAP Connection (SSL)", false)
		l := ldap.NewLDAPSSLConnection(ldapImportConf.LDAPServerConf.Server, ldapImportConf.LDAPServerConf.Port, TLSconfig)
		l.Debug = ldapImportConf.LDAPServerConf.Debug
		return l
	}

	return nil
}

//-- Query LDAP
func queryLdap() bool {
	//-- Create LDAP Connection
	l := connectLDAP()
	conErr := l.Connect()
	if conErr != nil {
		logger(4, "Connecting Error: "+fmt.Sprintf("%v", conErr), true)
		return false
	}
	defer l.Close()

	//-- Bind
	bindErr := l.Bind(ldapImportConf.LDAPServerConf.UserName, ldapImportConf.LDAPServerConf.Password)
	if bindErr != nil {
		logger(4, "Bind Error: "+fmt.Sprintf("%v", bindErr), true)
		return false
	}
	logger(1, "LDAP Search Query \n"+fmt.Sprintf("%+v", ldapImportConf.LDAPServerConf)+" ----", false)
	//-- Build Search Request
	searchRequest := ldap.NewSearchRequest(
		ldapImportConf.LDAPServerConf.DSN,
		ldapImportConf.LDAPServerConf.Scope,
		ldapImportConf.LDAPServerConf.DerefAliases,
		ldapImportConf.LDAPServerConf.SizeLimit,
		ldapImportConf.LDAPServerConf.TimeLimit,
		ldapImportConf.LDAPServerConf.TypesOnly,
		ldapImportConf.LDAPServerConf.Filter,
		ldapImportConf.LDAPAttributes,
		nil)

	//-- Search Request with 1000 limit pagaing
	results, searchErr := l.SearchWithPaging(searchRequest, 1000)
	if searchErr != nil {
		logger(4, "Search Error: "+fmt.Sprintf("%v", searchErr), true)
		return false
	}

	logger(1, "LDAP Results: "+fmt.Sprintf("%d", len(results.Entries)), true)
	//-- Catch zero results
	if len(results.Entries) == 0 {
		logger(4, "[LDAP] [SEARCH] No Users Found ", true)
		return false
	}
	ldapUsers = results.Entries
	return true
}
