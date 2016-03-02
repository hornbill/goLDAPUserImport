### LDAP Import Go - [GO](https://golang.org/) Import Script to Hornbill

### Quick links
- [Installation](#installation)
- [Config](#config)
    - [Instance Config](#InstanceConfig)
    - [LDAP Config](#LDAPConfig)
    - [LDAP Mapping](#LDAPMapping)
- [Execute](#execute)
- [Testing](testing)
- [Scheduling](#scheduling)
- [Logging](#logging)
- [Error Codes](#error codes)
- [Change Log](#change log)

# Installation

#### Windows
* Download the [https://github.com/hornbill/goLDAPUserImport/releases/download/v1.5.2/ldap_user_import_win_v1_6_1.zip  ldap_user_import_win_v1_56_1.zip]
* Extract zip into a folder you would like the application to run from e.g. `C:\LDAP_Import\`
* Open '''conf.json''' and add in the necessary configration
* Open Command Line Prompt as Administrator
* Change Directory to the folder with ldap_user_import.exe `C:\LDAP_Import\`
* Run the command ldap_user_import.exe -dryrun=true

# config

Example JSON File:

```json
{
    "UserName": "",
    "Password": "",
    "InstanceId": "",
    "UpdateUserType": false,
    "LDAPConf": {
        "Server": "",
        "UserName": "",
        "Password": "",
        "Port": 389,
        "ConnectionType": "",
        "InsecureSkipVerify": false,
        "Scope": 1,
		"DerefAliases": 1,
        "SizeLimit": 0,
		"TimeLimit": 0,
		"TypesOnly": false,
        "Filter": "(objectClass=user)",
        "DSN": ""
    },
    "LDAPMapping":{
        "UserId":"[sAMAccountName]",
        "UserType":"user",
        "Name":"[cn]",
        "Password":"",
        "FirstName":"[givenName]",
        "LastName":"[sn]",
        "JobTitle":"",
        "Site":"",
        "Phone":"[telephoneNumber]",
        "Email":"[userPrincipalName]",
        "Mobile":"[mobile]",
        "AbsenceMessage":"",
        "TimeZone":"",
        "Language":"",
        "DateTimeFormat":"",
        "DateFormat":"",
        "TimeFormat":"",
        "CurrencySymbol":"",
        "CountryCode":""
    },
    "LDAPAttirubutes":[
        "cn",
        "sn",
        "sAMAccountName",
        "userPrincipalName",
        "givenName",
        "description"
    ],
    "Roles":[
        "Collaboration Role"
    ],
    "SiteLookup":{
        "Enabled": false,
        "Attribute":""
    }
}
```
#### InstanceConfig
* "UserName" - Instance User Name with Create / Update User Rights
* "Password" - Instance Password for the above User
* "InstanceId" - Instance Id
* "UpdateUserType" - If set to True then the Type of User will be updated when the user account Update is triggered

#### LDAPConfig
* "Server" LDAP Server Address
* "UserName" LDAP User Name
* "Password" Password for Above User Name
* "Port" LDAP Port (389 for normal connections is default, 636 for SSL / TLS connection)
* "ConnectionType" Type of HTTP connection to use when communicating with the LDAP Server ("" = Normal HTTP, "SSL" = SSL Connection, "TLS" = TLS Connection)
* "InsecureSkipVerify" Used with SSL or TLS connection types will allow the verification of SSL Certifications to be disabled
* "Scope" Search Scope (ScopeBaseObject = 0, ScopeSingleLevel  = 1, ScopeWholeSubtree = 2) Default is 1
* "DerefAliases" dereference Aliases (NeverDerefAliases = 0, DerefInSearching = 1, DerefFindingBaseObj = 2, DerefAlways = 3) Default is 1
* "SizeLimit"  Size Limit for query 0 will disable
* "TimeLimit" Time Limit for query 0 will disable
* "TypesOnly" Return Attribute Descriptions
* "Filter" Search Filter I.e `(objectClass=user)`
* "DSN"  Search DSN I.e `DC=test,DC=hornbill,DC=com`

#### LDAPMapping
* Any value wrapped with [] will be treaded ad an LDAP field
* Any Other Value is treated literally as written example:
    * "Name":"[givenName] [sn]", - Both Variables are evaluated from LDAP and set to the Name param
    * "Password":"", - Auto Generated Password
    * "Site":"1" - The value of Site should be numeric
* If Password is left empty then a 10 character random string will be assigned so the user will need to recover the password using forgot my password functionality - The password will also be in the Log File
* "UserType" - This defines if a user is Co-Worker or Basic user and can have the value user or basic.

#### LDAPAttirubutes
* Array of Attributes to query from the LDAP Server, only Attributes specified here can be used in the LDAPMapping

#### Roles
This should contain an array of roles to be added to a user when they are created. If importing Basic Users then only the '''Basic User Role''' should be specified any role with a User Privilege level will be rejected

#### SiteLookup
In Hornbill the Site field against a user is the numeric Id of the site, as the Name of a users site from LDAP is likely the Name and not an Id specific to Hornbill  we provide the ability for the Import to look up the Name of the Site in Hornbill and use the Numeric Id when adding or updating a user.
The name of the Site in Hornbill must match the value of the Attribute in LDAP.
* Enabled - Turns on or off the Lookup of Sites
* Attribute - The LDAP Attribute to use for the name of the Site

# execute
Command Line Parameters
* file - Defaults to `conf.json` - Name of the Configuration file to load
* dryrun - Defaults to `false` - Set to True and the XMLMC for Create and Update users will not be called and instead the XML will be dumped to the log file, this is to aid in debugging the initial connection information.
* zone - Defaults to `eur` - Allows you to change the ZONE used for creating the XMLMC EndPoint URL https://{ZONE}api.hornbill.com/{INSTANCE}/

# Testing
If you run the application with the argument dryrun=true then no users will be created or updated, the XML used to create or update will be saved in the log file so you can ensure the LDAP mappings are correct before running the import.

'ldap_user_import.exe -dryrun=true'


# Scheduling

### Windows
You can schedule ldap_user_import.exe to run with any optional command line argument from Windows Task Scheduler.
* Ensure the user account running the task has rights to ldap_import.exe and the containing folder.
* Make sure the Start In parameter contains the folder where ldap_import.exe resides in otherwise it will not be able to pick up the correct path.

# logging
All Logging output is saved in the log directory in the same directory as the executable the file name contains the date and time the import was run 'LDAP_User_Import_2015-11-06T14-26-13Z.log'

# Error Codes
* `100` - Unable to create log File
* `101` - Unable to create log folder
* `102` - Unable to Load Configuration File
