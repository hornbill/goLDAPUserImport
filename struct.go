package main

import (
	"sync"
	"time"

	"github.com/hornbill/ldap"
)

//----- Constants -----
const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const version = "2.1.0"
const constOK = "ok"
const updateString = "Update"
const createString = "Create"

//-- MUTEX
var mutexSites = &sync.Mutex{}
var mutexGroups = &sync.Mutex{}
var mutexManagers = &sync.Mutex{}
var mutexCounters = &sync.Mutex{}
var logFileMutex = &sync.Mutex{}
var bufferMutex = &sync.Mutex{}

var userProfileMappingMap = map[string]string{
	"MiddleName":        "middleName",
	"JobDescription":    "jobDescription",
	"Manager":           "manager",
	"WorkPhone":         "workPhone",
	"Qualifications":    "qualifications",
	"Interests":         "interests",
	"Expertise":         "expertise",
	"Gender":            "gender",
	"Dob":               "dob",
	"Nationality":       "nationality",
	"Religion":          "religion",
	"HomeTelephone":     "homeTelephone",
	"SocialNetworkA":    "socialNetworkA",
	"SocialNetworkB":    "socialNetworkB",
	"SocialNetworkC":    "socialNetworkC",
	"SocialNetworkD":    "socialNetworkD",
	"SocialNetworkE":    "socialNetworkE",
	"SocialNetworkF":    "socialNetworkF",
	"SocialNetworkG":    "socialNetworkG",
	"SocialNetworkH":    "socialNetworkH",
	"PersonalInterests": "personalInterests",
	"HomeAddress":       "homeAddress",
	"PersonalBlog":      "personalBlog",
	"Attrib1":           "attrib1",
	"Attrib2":           "attrib2",
	"Attrib3":           "attrib3",
	"Attrib4":           "attrib4",
	"Attrib5":           "attrib5",
	"Attrib6":           "attrib6",
	"Attrib7":           "attrib7",
	"Attrib8":           "attrib8"}
var userProfileArray = []string{
	"MiddleName",
	"JobDescription",
	"Manager",
	"WorkPhone",
	"Qualifications",
	"Interests",
	"Expertise",
	"Gender",
	"Dob",
	"Nationality",
	"Religion",
	"HomeTelephone",
	"SocialNetworkA",
	"SocialNetworkB",
	"SocialNetworkC",
	"SocialNetworkD",
	"SocialNetworkE",
	"SocialNetworkF",
	"SocialNetworkG",
	"SocialNetworkH",
	"PersonalInterests",
	"HomeAddress",
	"PersonalBlog",
	"Attrib1",
	"Attrib2",
	"Attrib3",
	"Attrib4",
	"Attrib5",
	"Attrib6",
	"Attrib7",
	"Attrib8"}

var userMappingMap = map[string]string{
	"Name":           "name",
	"Password":       "password",
	"UserType":       "userType",
	"FirstName":      "firstName",
	"LastName":       "lastName",
	"JobTitle":       "jobTitle",
	"Site":           "site",
	"Phone":          "phone",
	"Email":          "email",
	"Mobile":         "mobile",
	"AbsenceMessage": "absenceMessage",
	"TimeZone":       "timeZone",
	"Language":       "language",
	"DateTimeFormat": "dateTimeFormat",
	"DateFormat":     "dateFormat",
	"TimeFormat":     "timeFormat",
	"CurrencySymbol": "currencySymbol",
	"CountryCode":    "countryCode"}
var userUpdateArray = []string{
	"UserType",
	"Name",
	"Password",
	"FirstName",
	"LastName",
	"JobTitle",
	"Site",
	"Phone",
	"Email",
	"Mobile",
	"AbsenceMessage",
	"TimeZone",
	"Language",
	"DateTimeFormat",
	"DateFormat",
	"TimeFormat",
	"CurrencySymbol",
	"CountryCode"}
var userCreateArray = []string{
	"Name",
	"Password",
	"UserType",
	"FirstName",
	"LastName",
	"JobTitle",
	"Site",
	"Phone",
	"Email",
	"Mobile",
	"AbsenceMessage",
	"TimeZone",
	"Language",
	"DateTimeFormat",
	"DateFormat",
	"TimeFormat",
	"CurrencySymbol",
	"CountryCode"}

//----- Variables -----
var ldapImportConf ldapImportConfStruct
var xmlmcInstanceConfig xmlmcConfig
var ldapUsers []*ldap.Entry
var xmlmcUsers []userListItemStruct
var sites []siteListStruct
var managers []managerListStruct
var groups []groupListStruct
var counters counterTypeStruct
var configFileName string
var configZone string
var configLogPrefix string
var configDryRun bool
var configVersion bool
var configWorkers int
var timeNow string
var startTime time.Time
var endTime time.Duration
var errorCount uint64
var noValuesToUpdate = "There are no values to update"

//----- Structures -----
type siteListStruct struct {
	SiteName string
	SiteID   int
}
type managerListStruct struct {
	UserName string
	UserID   string
}
type groupListStruct struct {
	GroupName string
	GroupID   string
}

type xmlmcConfig struct {
	instance string
	zone     string
	url      string
}

type counterTypeStruct struct {
	updated        uint16
	created        uint16
	profileUpdated uint16
	updatedSkipped uint16
	createskipped  uint16
	profileSkipped uint16
}
type ldapImportConfStruct struct {
	APIKey             string
	InstanceID         string
	UpdateUserType     bool
	UserRoleAction     string
	URL                string
	DAVURL             string
	LDAPServerConf     ldapServerConfStruct
	UserMapping        userMappingStruct
	UserAccountStatus  userAccountStatusStruct
	UserProfileMapping userProfileMappingStruct
	UserManagerMapping userManagerStruct
	LDAPAttributes     []string
	Roles              []string
	ImageLink          imageLinkStruct
	SiteLookup         siteLookupStruct
	OrgLookup          orgLookupStruct
}
type userMappingStruct struct {
	UserID         string
	UserType       string
	Name           string
	Password       string
	FirstName      string
	LastName       string
	JobTitle       string
	Site           string
	Phone          string
	Email          string
	Mobile         string
	AbsenceMessage string
	TimeZone       string
	Language       string
	DateTimeFormat string
	DateFormat     string
	TimeFormat     string
	CurrencySymbol string
	CountryCode    string
}
type userAccountStatusStruct struct {
	Action  string
	Enabled bool
	Status  string
}
type userProfileMappingStruct struct {
	MiddleName        string
	JobDescription    string
	Manager           string
	WorkPhone         string
	Qualifications    string
	Interests         string
	Expertise         string
	Gender            string
	Dob               string
	Nationality       string
	Religion          string
	HomeTelephone     string
	SocialNetworkA    string
	SocialNetworkB    string
	SocialNetworkC    string
	SocialNetworkD    string
	SocialNetworkE    string
	SocialNetworkF    string
	SocialNetworkG    string
	SocialNetworkH    string
	PersonalInterests string
	HomeAddress       string
	PersonalBlog      string
	Attrib1           string
	Attrib2           string
	Attrib3           string
	Attrib4           string
	Attrib5           string
	Attrib6           string
	Attrib7           string
	Attrib8           string
}
type userManagerStruct struct {
	Action        string
	Enabled       bool
	Attribute     string
	GetIDFromName bool
	Regex         string
	Reverse       bool
}

type ldapServerConfStruct struct {
	Server             string
	UserName           string
	Password           string
	Port               uint16
	ConnectionType     string
	InsecureSkipVerify bool
	Scope              int
	DerefAliases       int
	SizeLimit          int
	TimeLimit          int
	TypesOnly          bool
	Filter             string
	DSN                string
	Debug              bool
}
type siteLookupStruct struct {
	Action    string
	Enabled   bool
	Attribute string
}
type orgLookupStruct struct {
	Action      string
	Enabled     bool
	Attribute   string
	Type        int
	Membership  string
	TasksView   bool
	TasksAction bool
}
type xmlmcResponse struct {
	MethodResult string       `xml:"status,attr"`
	Params       paramsStruct `xml:"params"`
	State        stateStruct  `xml:"state"`
}
type xmlmcCheckUserResponse struct {
	MethodResult string                 `xml:"status,attr"`
	Params       paramsCheckUsersStruct `xml:"params"`
	State        stateStruct            `xml:"state"`
}
type xmlmcUserListResponse struct {
	MethodResult string                     `xml:"status,attr"`
	Params       paramsUserSearchListStruct `xml:"params"`
	State        stateStruct                `xml:"state"`
}
type xmlmcuserSetGroupOptionsResponse struct {
	MethodResult string      `xml:"status,attr"`
	State        stateStruct `xml:"state"`
}
type paramsUserSearchListStruct struct {
	RowData paramsUserRowDataListStruct `xml:"rowData"`
}
type paramsUserRowDataListStruct struct {
	Row userObjectStruct `xml:"row"`
}
type userObjectStruct struct {
	UserID   string `xml:"h_user_id"`
	UserName string `xml:"h_name"`
}

type xmlmcSiteListResponse struct {
	MethodResult string               `xml:"status,attr"`
	Params       paramsSiteListStruct `xml:"params"`
	State        stateStruct          `xml:"state"`
}
type paramsSiteListStruct struct {
	RowData paramsSiteRowDataListStruct `xml:"rowData"`
}
type paramsSiteRowDataListStruct struct {
	Row siteObjectStruct `xml:"row"`
}
type siteObjectStruct struct {
	SiteID      int    `xml:"h_id"`
	SiteName    string `xml:"h_site_name"`
	SiteCountry string `xml:"h_country"`
}

type xmlmcGroupListResponse struct {
	MethodResult string                `xml:"status,attr"`
	Params       paramsGroupListStruct `xml:"params"`
	State        stateStruct           `xml:"state"`
}

type paramsGroupListStruct struct {
	RowData paramsGroupRowDataListStruct `xml:"rowData"`
}

type paramsGroupRowDataListStruct struct {
	Row groupObjectStruct `xml:"row"`
}

type groupObjectStruct struct {
	GroupID   string `xml:"h_id"`
	GroupName string `xml:"h_name"`
}
type stateStruct struct {
	Code     string `xml:"code"`
	ErrorRet string `xml:"error"`
}
type paramsCheckUsersStruct struct {
	RecordExist bool `xml:"recordExist"`
}
type paramsStruct struct {
	SessionID string `xml:"sessionId"`
}
type paramsUserListStruct struct {
	UserListItem []userListItemStruct `xml:"userListItem"`
}
type userListItemStruct struct {
	UserID string `xml:"userId"`
	Name   string `xml:"name"`
}
type imageLinkStruct struct {
	Action     string
	Enabled    bool
	UploadType string
	ImageType  string
	URI        string
}
type xmlmcprofileSetImageResponse struct {
	MethodResult string                `xml:"status,attr"`
	Params       paramsGroupListStruct `xml:"params"`
	State        stateStruct           `xml:"state"`
}
