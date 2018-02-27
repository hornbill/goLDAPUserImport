package main

import (
	"crypto/rand"
	"encoding/xml"
	"fmt"
	"html"
	"log"
	"os"
	"reflect"
	"regexp"
	"strings"

	"github.com/fatih/color"
	"github.com/hornbill/goApiLib"
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
			//logger(3, "Unable to Load LDAP Attribute: "+v[1:len(v)-1]+" For Input Param: "+s, false)
			return LDAPAttributeValue
		}
		//-- TK UnescapeString to HTML entities are replaced
		s = html.UnescapeString(strings.Replace(s, v, LDAPAttributeValue, 1))
	}

	//-- Return Value
	return s
}

//-- Generate Password String
func generatePasswordString(n int) string {
	var arbytes = make([]byte, n)
	rand.Read(arbytes)
	for i, b := range arbytes {
		arbytes[i] = letterBytes[b%byte(len(letterBytes))]
	}
	return string(arbytes)
}
func loggerGen(t int, s string) string {
	//-- Ignore Logging level unless is 0
	if t < ldapImportConf.Advanced.LogLevel && t != 0 {
		return ""
	}

	var errorLogPrefix = ""
	//-- Create Log Entry
	switch t {
	case 1:
		errorLogPrefix = "[DEBUG] "
	case 2:
		errorLogPrefix = "[MESSAGE] "
	case 3:
		errorLogPrefix = "[WARN] "
	case 4:
		errorLogPrefix = "[ERROR] "
	}
	return errorLogPrefix + s + "\n\r"
}
func loggerWriteBuffer(s string) {
	if s != "" {
		logLines := strings.Split(s, "\n\r")
		for _, line := range logLines {
			if line != "" {
				logger(0, line, false)
			}
		}
	}
}

//-- Loggin function
func logger(t int, s string, outputtoCLI bool) {

	//-- Ignore Logging level unless is 0
	if t < ldapImportConf.Advanced.LogLevel && t != 0 {
		return
	}
	mutexLog.Lock()
	defer mutexLog.Unlock()

	onceLog.Do(func() {
		//-- Curreny WD
		cwd, _ := os.Getwd()
		//-- Log Folder
		logPath := cwd + "/log"
		//-- Log File
		logFileName := logPath + "/" + Flags.configLogPrefix + "LDAP_User_Import_" + Time.timeNow + ".log"
		//-- If Folder Does Not Exist then create it
		if _, err := os.Stat(logPath); os.IsNotExist(err) {
			err := os.Mkdir(logPath, 0777)
			if err != nil {
				fmt.Printf("Error Creating Log Folder %q: %s \r", logPath, err)
				os.Exit(101)
			}
		}

		//-- Open Log File
		var err error
		f, err = os.OpenFile(logFileName, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0777)
		if err != nil {
			fmt.Printf("Error Creating Log File %q: %s \n", logFileName, err)
			os.Exit(100)
		}
		log.SetOutput(f)

	})
	// don't forget to close it
	//defer f.Close()
	red := color.New(color.FgRed).PrintfFunc()
	orange := color.New(color.FgCyan).PrintfFunc()
	var errorLogPrefix = ""
	//-- Create Log Entry
	switch t {
	case 0:
	case 1:
		errorLogPrefix = "[DEBUG] "
	case 2:
		errorLogPrefix = "[MESSAGE] "
	case 3:
		errorLogPrefix = "[WARN] "
	case 4:
		errorLogPrefix = "[ERROR] "
	}
	if outputtoCLI {
		if t == 3 {
			orange(errorLogPrefix + s + "\n")
		} else if t == 4 {
			red(errorLogPrefix + s + "\n")
		} else {
			fmt.Printf(errorLogPrefix + s + "\n")
		}

	}
	log.Println(errorLogPrefix + s)
}

//-- Log to ESP
func espLogger(message string, severity string) bool {

	// We lock the whole function so we dont reuse the same connection for multiple logging attempts
	mutexLogger.Lock()
	defer mutexLogger.Unlock()

	// We initilaise the connection pool the first time the function is called and reuse it
	// This is reuse the connections rather than creating a pool each invocation
	once.Do(func() {

		loggerAPI = apiLib.NewXmlmcInstance(Flags.configInstanceID)
		loggerAPI.SetAPIKey(Flags.configAPIKey)
		loggerAPI.SetTimeout(5)
	})

	loggerAPI.SetParam("fileName", "LDAP_User_Import")
	loggerAPI.SetParam("group", "general")
	loggerAPI.SetParam("severity", severity)
	loggerAPI.SetParam("message", message)

	XMLLogger, xmlmcErr := loggerAPI.Invoke("system", "logMessage")
	var xmlRespon xmlmcLogMessageResponse
	if xmlmcErr != nil {
		logger(4, "Unable to write to log "+fmt.Sprintf("%s", xmlmcErr), true)
		return false
	}
	err := xml.Unmarshal([]byte(XMLLogger), &xmlRespon)
	if err != nil {
		logger(4, "Unable to write to log "+fmt.Sprintf("%s", err), true)
		return false
	}
	if xmlRespon.MethodResult != "ok" {
		logger(4, "Unable to write to log "+xmlRespon.State.ErrorRet, true)
		return false
	}

	return true
}
