## 1.5.3 (Feburary X , 2016)

Features:

- Errors are now displayed in red in the console.
- Checks for mandatory values from the configuration file are now done before processing begins.
- Check for latest version will now display in the console and log on startup of the process


## 1.5.2 (Feburary 8, 2016)

Bugfixes:

- Add additional logging to the login script.

## 1.5.1 (January 8, 2016)

Features:

- CH00138411 - Expose further LDAP Search Configuration options to the Configuration File
- LDAP Search Query is now saved to the Log file
- Errors Creating / Updating Users as well as Assigning Roles are returned to the Hornbill Log File for easier Diagnostics

Bugfixes:

- Prevent "<invalid Value>" from being used as a value when an LDAP Attribute cannot be found and instead a Error in the log is written.
- Errors while create / Updating users are no loner output to the console breaking the progress bar and instead an error count is output at
the end of the process and will refer the user to the Log File.
Notes:

- The following should be added to the "LDAPConf" section of the configuration file for searching to operate as previous.

"Scope": 1,
"DerefAliases": 1,

## 1.5.0 (December 18, 2015)

Bugfixes:

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

Bugfixes:

 - Phone Attribute on User Create/Update was not being set

## 1.2.0 (February 17, 2010)

Features:

  - ESP Logging to Aid Support
  - Ability to Lookup Site

Bugfixes:

  - Unable to Load Configuration from -file flag

## 1.1.0 (November 9, 2015)

Features:

  - Assigning Roles when a user is created

## 1.0.0 (November 9, 2015)

Features:

  - Initial Release
