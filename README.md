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
* Download the [x64 Binary](https://github.com/hornbill/goLDAPUserImport/releases/download/2.4.2/ldap_user_import_win_x64_v2_4_2.zip) or [x86 Binary](https://github.com/hornbill/goLDAPUserImport/releases/download/2.4.2/ldap_user_import_win_x86_v2_4_2.zip)
* Extract zip into a folder you would like the application to run from e.g. `C:\LDAP_Import\`
* Open '''conf.json''' and add in the necessary configration
* Open Command Line Prompt as Administrator
* Change Directory to the folder with ldap_user_import.exe `C:\LDAP_Import\`
* Run the command ldap_user_import.exe -dryrun=true

# config

Example JSON File:

```json
{
    "APIKey": "",
    "InstanceId": "",
    "UpdateUserType": false,
    "UserRoleAction": "Create",
    "LDAPServerConf": {
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
        "DSN": "",
        "Debug": false
    },
    "UserMapping":{
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
        "CountryCode":"",
        "UserDNCache":"[distinguishedName]"
    },
    "UserAccountStatus":{
        "Action":"Update",
        "Enabled": false,
        "Status":"active"
    },
    "UserProfileMapping":{
        "MiddleName":"",
    	"JobDescription":"",
    	"Manager":"",
    	"WorkPhone":"",
    	"Qualifications":"",
    	"Interests":"",
    	"Expertise":"",
    	"Gender":"",
    	"Dob":"",
    	"Nationality":"",
    	"Religion":"",
    	"HomeTelephone":"",
    	"SocialNetworkA":"",
    	"SocialNetworkB":"",
    	"SocialNetworkC":"",
    	"SocialNetworkD":"",
    	"SocialNetworkE":"",
    	"SocialNetworkF":"",
    	"SocialNetworkG":"",
    	"SocialNetworkH":"",
    	"PersonalInterests":"",
    	"homeAddress":"",
    	"PersonalBlog":"",
        "Attrib1":"",
    	"Attrib2":"",
    	"Attrib3":"",
    	"Attrib4":"",
    	"Attrib5":"",
    	"Attrib6":"",
    	"Attrib7":"",
    	"Attrib8":""
    },
    "UserManagerMapping":{
        "Action":"Create",
        "Enabled":true,
        "Attribute":"[manager]",
        "GetIDFromName":true,
        "SearchforManager":true,
        "Regex":"CN=(.*?)(?:,[A-Z]+=|$)",
        "Reverse":true,
        "ManagerSearchField":"h_name",
        "UseDNCacheFirst":false
    },
    "LDAPAttributes":[
        "cn",
        "distinguishedName",
        "sn",
        "sAMAccountName",
        "userPrincipalName",
        "givenName",
        "description",
        "manager",
        "thumbnailPhoto"
    ],
    "Roles":[
        "Collaboration Role"
    ],
    "ImageLink":{
        "Action":"Both"
        , "Enabled": true
        , "UploadType": "AD"
        , "ImageType": "jpg"
        , "URI": "[thumbnailPhoto]"
    },
    "SiteLookup":{
        "Action":"Both",
        "Enabled": false,
        "Attribute":""
    },
    "OrgLookup":{
        "Action":"Both",
        "Enabled":false,
        "Attribute":"[sAMAccountName]",
        "Type":2,
        "Membership":"member",
        "TasksView":false,
        "TasksAction":false,
        "OnlyOneGroupAssignment":false
    }
}
```
#### InstanceConfig
* "APIKey" - A Valid API Assigned to a user with enough rights to process the import
* "InstanceId" - Instance Id
* "UpdateUserType" - If set to True then the Type of User will be updated when the user account Update is triggered
* "UserRoleAction" - (Both | Update | Create) - When to Set controls what action will assign roles ro a user Create, On Update or Both

#### LDAPServerConf
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
* "Debug"  Enable LDAP Connection Debugging, should only ever be enabled to troubleshoot connection issues.

#### UserMapping
* Any value wrapped with [] will be treaded as an LDAP field
* Do not try and add any new properties here they will be ignored
* Any Other Value is treated literally as written example:
    * "Name":"[givenName] [sn]", - Both Variables are evaluated from LDAP and set to the Name param
    * "Password":"", - Auto Generated Password
    * "Site":"1" - The value of Site should be numeric
* If Password is left empty then a 10 character random string will be assigned so the user will need to recover the password using forgot my password functionality - The password will also be in the Log File
* "UserType" - This defines if a user is Co-Worker or Basic user and can have the value user or basic.


#### UserAccountStatus
* Action - (Both | Update | Create) - When to Set the User Account Status On Create, On Update or Both
* Enabled - Turns on or off the Status update
* Status - Can be one of the following strings (active | suspended | archived)

#### UserProfileMapping
* Works in the same way as UserMapping
* Do not try and add any new properties here they will be ignored

#### UserManagerMapping
* Action - (Both | Update | Create) - When to Set the User Manager On Create, On Update or Both
* Enabled - Turns on or off the Manager Import
* Attribute - The LDAP Attribute to use for the name of the Manager ,Any value wrapped with [] will be treaded as an LDAP field
* GetIDFromName - Pull data from the Attribute using the bellow Regex (true | false)
* SearchforManager - Use data extracted from Attribute to search for the name of the manager this must match the name as stored in hornbill. If set to false then the data pulled from the attribute using the Regex match is used as the manager id (true | false)
* Regex - Optional Regex String to Match the Name from an DSN String when using GetIDFromName
* Reverse - Reverse the Name String Matched from the Regex (true | false) when using GetIDFromName
* ManagerSearchField - Change the User profile feild that is being searched on, default is h_name
* UseDNCacheFirst - When enabled the manager look up will match based on a cache of users and their DN matching the Attribute specified, if nothing is found it will try the search using the Attribute and Regex

#### LDAPAttributes
* Array of Attributes to query from the LDAP Server, only Attributes specified here can be used in the LDAPMapping
* Please note that thumbnailPhoto will need to be included IF you are planning to use AD stored images

#### Roles
This should contain an array of roles to be added to a user when they are created. If importing Basic Users then only the '''Basic User Role''' should be specified any role with a User Privilege level will be rejected

#### ImageLink
The ability to upload images for User profiles in Hornbill. Only jpg/jpeg and png formats are accepted.

* Action - (Both | Update | Create) - When to associate an Image On Create, On Update or Both
* Enabled - Turns on or off the Image association
* UploadType - (AD | URL | URI) - what TYPE of image upload sequence we are using
** AD - using data stored in AD - set the "URI" as JUST the ldap field with brackets
** URL - find the image at end of "URI" below - assuming that the URL is visible by our Hornbill servers (eg http://whatever.com/[userPrincipalName].jpg)
** URI - find the image at end of "URI" below - from a LOCAL server (not fully tested; eg http://localserver/[userPrincipalName].jpg)
* ImageType - (jpg | png) type of image as stored in AD
* InsecureSkipVerify - Will allow the verification of SSL Certifications to be disabled - For use with internal URI and self signed certificates
* URI - referencing the image data - Any value wrapped with [] will be treaded as an LDAP field

#### SiteLookup
In Hornbill the Site field against a user is the numeric Id of the site, as the Name of a users site from LDAP is likely the Name and not an Id specific to Hornbill  we provide the ability for the Import to look up the Name of the Site in Hornbill and use the Numeric Id when adding or updating a user.
The name of the Site in Hornbill must match the value of the Attribute in LDAP.

* Action - (Both | Update | Create) - When to Associate Sites On Create, On Update or Both
* Enabled - Turns on or off the Lookup of Sites
* Attribute - The LDAP Attribute to use for the name of the Site ,Any value wrapped with [] will be treaded as an LDAP field

#### OrgLookup
The name of the Organization in Hornbill must match the value of the Attribute in LDAP.

* Action - (Both | Update | Create) - When to Associate Organisation On Create, On Update or Both
* Enabled - Turns on or off the Lookup of Orgnisations
* Attribute - The LDAP Attribute to use for the name of the Site ,Any value wrapped with [] will be treaded as an LDAP field
* Type - The Organisation Type (0=general ,1=team ,2=department ,3=costcenter ,4=division ,5=company)
* Membership - The Organisation Membership the users will be added with (member,teamLeader,manager)
* TasksView - If set true, then the user can view tasks assigned to this group
* TasksAction - If set true, then the user can action tasks assigned to this group.
* OnlyOneGroupAssignment - If set to try then any existing group assignment of the specified type will be removed.

# execute
Command Line Parameters
* file - Defaults to `conf.json` - Name of the Configuration file to load
* dryrun - Defaults to `false` - Set to True and the XMLMC for Create and Update users will not be called and instead the XML will be dumped to the log file, this is to aid in debugging the initial connection information.
* logprefix - Default to `` - Allows you to define a string to prepend to the name of the log file generated
* workers - Defaults to `10` - Allows you to change the number of worker threads used to process the import, this can improve performance on slow import but using too many workers have a detriment to performance of your Hornbill instance.

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
