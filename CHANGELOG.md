## 3.0.0 (N/A , 2018)

Features:

ADD: Ability to assign one or more Organisations to a user
ADD: Ability to filter user Organisation assignment based on the memberof attribtute in LDAP "MemberOf" added to the Organisation configuration that should match the name of the AD a User should be in before the Organisation is assinged.
ADD: LogLevel has been added which defaults to 2 (Message, Warning, Error), this can be changed to 1 (Debug, Message, Warning, Error) if additional debug logging is required
ADD: Dry Run XML Output is now formatted better in the log file

CHANGE: Import Configuration is now loaded from the Instance and not via a local JSON file, this is to provide a user freindly UI for generating the configuration.
CHANGE: The tool has been completely rewritten to be more efficient this will have a drematic effect on the time taken for the tool to run (In a Postitve way)
CHANGE: The log output has been updated to closer reflect how the tool is processing data the format is still (YYYY/MM/DD HH:MM:SS [LogLevel] Message)

## 2.4.2 (Feb 1 , 2018)

Features:

ADD: "SearchforManager" added to Manager Mapping configuration allowing you to disable the search for a managers userId if you already have the value from an LDAP Attribute.

## 2.4.1 (Sep 20 , 2017)

Bug Fixes:

FIX: Error setting User Status and Import Status following previous change


## 2.4.0 (Sep 11 , 2017)

Features:

ADD: It is now possible to overwrite group assignments to a user based on a specific type i.e users can only be assigned to one department and existing associations that do not match the updated associate that is to be made will be removed.

Bug Fixes:

FIX: Concurrency issue when assigning users to a group

## 2.3.1 (Sep 11 , 2017)

Bug Fixes:

- FIX: Manager Lookup Cache is build before users are processed.

## 2.3.0 (Sep 11 , 2017)

Features:

- ADD: Ability to Cache users and their DN to use for Manager lookups

## 2.2.2 (Sep 11 , 2017)

Bug Fixes:

- FIX: thumbnailPhoto in default configuration should now be [thumbnailPhoto] for Image Uploads

## 2.2.1 (Aug 30 , 2017)

Bug Fixes:

- FIX: Issue with Manager lookup following 2.2.0 changes

## 2.2.0 (Aug 30 , 2017)

Features:

- CHANGE: Remove -zone flag and look up correctly using instanceID
- CHANGE: -workers flag now default to 10 not 1 to improve performance

Bug Fixes:

- FIX: Profile Image Upload now correctly support URL (Public) and URI (Internal) Images as well as AD Binary images
- FIX: Manage Import can now correctly be earched on any user profile feild not just h_name

## 2.1.1 (Aug 30 , 2017)

Bug Fixes:

- FIX: Improve mutithreaded performance
- FIX: Improve http request pool
- FIX: Issue importing profile image from URI

## 2.1.0 (May 2 , 2017)

Features:

- ADD: Profile images

## 2.0.6 (Jan 5 , 2017)

Features:

- ADD: Further debugging for LDAP Manager Lookups

Bug Fixes:

- Fix: Issue where if the Manager Import action was set to Update then the import would not process Managers

## 2.0.5 (Oct 17 , 2016)

Features:

- ADD: Validate LDAP Server ConnectionType, an invalid connection type would cause a panic in the LDAP Library.

Bug Fixes:

- Fix: Performance Issue with number of HTTP Connections being created

## 2.0.4 (Aug 4 , 2016)

Features:

- ADD: New Input flag `-logprefix=` which allows you to append a string to the name of the log file generated

## 2.0.3 (May 18 , 2016)

Bug Fixes:

- Fix: Hardcoded number of records that would be processed to 100, removed total is now taken from results from LDAP


## 2.0.2 (May 18 , 2016)

Bug Fixes:

- Fix: Attributes for user Profile were not being populated correctly
- Fix: Potential Race conditions after moving to Goroutines
- FIX: Reduce Potential for any errors and default workers back to 1

## 2.0.1 (May 18 , 2016)

Features:

- ADD: Ability to assign roles to a user on user update

Bug Fixes:

- Fix: conf.json was invalid
- Fix: Potential Race conditions after moving to Goroutines

## 2.0.0 (May 11 , 2016)

Features:

- ADD Support for APIKeys instead of username and password for Authentication to Hornbill
- ADD Concurrent workers, this should improve performance on large imports -workers input flag allows you to override the default of 3
- User Profile Extended Detail Support has been added
- Ability to define what actions should cause Site and Organisations to be associate (On Create | On Update | Both)
- ADD Ability to set a user account status (On Create | On Update | Both)
- ADD Ability to Lookup Manager Name from LDAP Field and map to Hornbill User Id UserManagerMapping

Bug Fixes:

- Fix: Timetaken was not being reported in the ESP Log File
- Fix: Version was not being reported in the ESP Log File
- Fix: HTML Entities were not correct decoded when returned from an LDAP string
- Fix: SiteLookup.Attribute was not expecting LDAP Attributes to be wrapped in [ ] like OrgLookup.Attribute this has been changed for consistency with other LDAP Attributes

Notes:

- The conf.json has been altered so some of the property names like LDAPMapping not correctly reflect there purpose LDAPMApping is now UserMapping the content of this configuration is the same and can be copied from your previous configuration file.

- UserID and Password for you Hornbill Instance is no longer used, you will need to generate an APIKey in the Administration Tool via the User Details - There is now a new tab called API Keys which allow you to generate an API key for a user, this should be done against which ever user account was previously used for the imports.

## 1.7.0 (March 18 , 2016)

Features:

- ADD Organization Lookup when Creating / Updating Users.

## 1.6.2 (March 16 , 2016)

Features:

- ADD Debug LDAP Connection to aid with troubleshooting.

## 1.6.1 (March 2 , 2016)

Bug Fixes:

- Fix SAMAccountName was hardcoded as the userId when checking if a user existed in hornbill

## 1.6.0 (Feburary 25 , 2016)

Features:

- GoLang upgraded to 1.6
- Added support for SSL/TLS connections to LDAP server

## 1.5.3 (Feburary 16 , 2016)

Features:

- Errors are now displayed in red in the console.
- Checks for mandatory values from the configuration file are now done before processing begins.
- Check for latest version will now display in the console and log on startup of the process

Bug Fixes:

- HTTP Client ignores environmental http proxy and https proxy settings
- Better Error handling and XMLMC lib cannot connect to the API Endpoint

## 1.5.2 (Feburary 8, 2016)

Bug Fixes:

- Add additional logging to the login script.

## 1.5.1 (January 8, 2016)

Features:

- CH00138411 - Expose further LDAP Search Configuration options to the Configuration File
- LDAP Search Query is now saved to the Log file
- Errors Creating / Updating Users as well as Assigning Roles are returned to the Hornbill Log File for easier Diagnostics

Bug Fixes:

- Prevent "<invalid Value>" from being used as a value when an LDAP Attribute cannot be found and instead a Error in the log is written.
- Errors while create / Updating users are no loner output to the console breaking the progress bar and instead an error count is output at
the end of the process and will refer the user to the Log File.
Notes:

- The following should be added to the "LDAPConf" section of the configuration file for searching to operate as previous.

"Scope": 1,
"DerefAliases": 1,

## 1.5.0 (December 18, 2015)

Bug Fixes:

- UserID was incorrectly referenced as UserId after some code refactoring
- UpdateUserType was missing from the default configuration

## 1.4.1 (December 14, 2015)

Features:

  - Released to GitHub

## 1.4.0 (December 08, 2015)

Features:

  - Ability to prevent userType from being updated

## 1.3.0 (December 07, 2015)

Features:

  - Updated External Library (ApiLib)
  - Added Debugging for LDAP Returned Data

Bug Fixes:

 - Phone Attribute on User Create/Update was not being set

## 1.2.0 (February 17, 2010)

Features:

  - ESP Logging to Aid Support
  - Ability to Lookup Site

Bug Fixes:

  - Unable to Load Configuration from -file flag

## 1.1.0 (November 9, 2015)

Features:

  - Assigning Roles when a user is created

## 1.0.0 (November 9, 2015)

Features:

  - Initial Release
