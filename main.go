package main

//----- Packages -----
import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color" //-- CLI Colour
	"github.com/hornbill/goApiLib"
	//-- Hornbill Clone of "github.com/mavricknz/ldap"
	//--Hornbil Clone of "github.com/cheggaaa/pb"
	"github.com/tcnksm/go-latest" //-- For Version checking
)

var (
	once        sync.Once
	onceLog     sync.Once
	mutexLogger = &sync.Mutex{}
	loggerAPI   *apiLib.XmlmcInstStruct
	mutexLog    = &sync.Mutex{}
	f           *os.File
)

// Main
func main() {
	//-- Start Time for Durration
	Time.startTime = time.Now()
	//-- Start Time for Log File
	Time.timeNow = time.Now().Format(time.RFC3339)
	//-- Remove :
	Time.timeNow = strings.Replace(Time.timeNow, ":", "-", -1)

	//-- Process Flags
	procFlags()

	//-- Used for Building
	if Flags.configVersion {
		fmt.Printf("%v \n", version)
		return
	}

	//-- Check for latest
	checkVersion()

	//-- Load Configuration File Into Struct
	ldapImportConf = loadConfig()

	//-- Validation on Configuration File
	configError := validateConf()

	//-- Check for Error
	if configError != nil {
		logger(4, fmt.Sprintf("%v", configError), true)
		logger(4, "Please Check your Configuration File: "+Flags.configFileName, true)
		return
	}

	//-- Once we have loaded the config write to hornbill log file
	logged := espLogger("---- XMLMC LDAP Import Utility V"+fmt.Sprintf("%v", version)+" ----", "debug")

	//-- Check for Connections
	if !logged {
		logger(4, "Unable to Connect to Instance", true)
		return
	}
	//-- Query LDAP
	queryLdap()

	//-- Process LDAP User Data First
	//-- So we only store data about users we have
	processLDAPUsers()

	//-- Fetch Users from Hornbill
	loadUsers()

	//-- Load User Roles
	loadUsersRoles()

	//-- Fetch Sites
	loadSites()

	//-- Fetch Groups
	loadGroups()

	//-- Fetch User Groups
	loadUserGroups()

	//-- Create List of Actions that need to happen
	//-- (Create,Update,profileUpdate,Assign Role, Assign Group, Assign Site)
	processData()

	//-- End Ouput
	outputEnd()
}

//-- Process Input Flags
func procFlags() {
	//-- Grab Flags
	flag.StringVar(&Flags.configFileName, "file", "conf.json", "Name of Configuration File To Load")
	flag.StringVar(&Flags.configLogPrefix, "logprefix", "", "Add prefix to the logfile")
	flag.BoolVar(&Flags.configDryRun, "dryrun", false, "Allow the Import to run without Creating or Updating users")
	flag.BoolVar(&Flags.configVersion, "version", false, "Output Version")
	flag.IntVar(&Flags.configWorkers, "workers", 1, "Number of Worker threads to use")

	//-- Parse Flags
	flag.Parse()

	//-- Output config
	if !Flags.configVersion {
		logger(1, "---- XMLMC LDAP Import Utility V"+fmt.Sprintf("%v", version)+" ----", true)
		logger(1, "Flag - Config File "+Flags.configFileName, true)
		logger(1, "Flag - Log Prefix "+Flags.configLogPrefix, true)
		logger(1, "Flag - Dry Run "+fmt.Sprintf("%v", Flags.configDryRun), true)
		logger(1, "Flag - Workers "+fmt.Sprintf("%v", Flags.configWorkers)+"\n", true)
	}
}

//-- Generate Output
func outputEnd() {
	logger(1, "Import Complete", true)
	//-- End output
	if counters.errors > 0 {
		logger(4, "Error encountered please check the log file", true)
		logger(4, "Error Count: "+fmt.Sprintf("%d", counters.errors), true)
		//logger(4, "Check Log File for Details", true)
	}
	logger(1, "Updated: "+fmt.Sprintf("%d", counters.updated), true)
	logger(1, "Updated Skipped: "+fmt.Sprintf("%d", counters.updatedSkipped), true)

	logger(1, "Created: "+fmt.Sprintf("%d", counters.created), true)
	logger(1, "Created Skipped: "+fmt.Sprintf("%d", counters.createskipped), true)

	logger(1, "Profiles Updated: "+fmt.Sprintf("%d", counters.profileUpdated), true)
	logger(1, "Profiles Skipped: "+fmt.Sprintf("%d", counters.profileSkipped), true)

	//-- Show Time Takens
	Time.endTime = time.Since(Time.startTime).Round(time.Second)
	logger(1, "Time Taken: "+fmt.Sprintf("%s", Time.endTime), true)
	//-- complete
	complete()
	logger(1, "---- XMLMC LDAP Import Complete ---- ", true)
}

//-- Check Latest
func checkVersion() {
	githubTag := &latest.GithubTag{
		Owner:      "hornbill",
		Repository: "goLDAPUserImport",
	}

	res, err := latest.Check(githubTag, version)
	if err != nil {
		logger(4, fmt.Sprintf("%s", err), true)
		return
	}
	if res.Outdated {
		logger(3, version+" is not latest, you should upgrade to "+res.Current+" by downloading the latest package Here https://github.com/hornbill/goLDAPUserImport/releases/tag/v"+res.Current, true)
	}
}

//-- Function to Load Configruation File
func loadConfig() ldapImportConfStruct {
	//-- Check Config File File Exists
	cwd, _ := os.Getwd()
	configurationFilePath := cwd + "/" + Flags.configFileName
	logger(1, "Loading Config File: "+configurationFilePath, false)
	if _, fileCheckErr := os.Stat(configurationFilePath); os.IsNotExist(fileCheckErr) {
		logger(4, "No Configuration File", true)
		os.Exit(102)
	}

	//-- Load Config File
	file, fileError := os.Open(configurationFilePath)
	//-- Check For Error Reading File
	if fileError != nil {
		logger(4, "Error Opening Configuration File: "+fmt.Sprintf("%v", fileError), true)
	}
	//-- New Decoder
	decoder := json.NewDecoder(file)
	//-- New Var based on ldapImportConf
	eldapConf := ldapImportConfStruct{}

	//-- Decode JSON
	err := decoder.Decode(&eldapConf)
	//-- Error Checking
	if err != nil {
		logger(4, "Error Decoding Configuration File: "+fmt.Sprintf("%v", err), true)
	}

	//-- Return New Congfig
	return eldapConf
}

func validateConf() error {

	//-- Check for API Key
	if ldapImportConf.APIKey == "" {
		err := errors.New("API Key is not set")
		return err
	}
	//-- Check for Instance ID
	if ldapImportConf.InstanceID == "" {
		err := errors.New("InstanceID is not set")
		return err
	}
	//-- Check LDAP Sever Connection type
	if ldapImportConf.LDAP.Server.ConnectionType != "" && ldapImportConf.LDAP.Server.ConnectionType != "SSL" && ldapImportConf.LDAP.Server.ConnectionType != "TLS" {
		err := errors.New("Invalid ConnectionType: '" + ldapImportConf.LDAP.Server.ConnectionType + "' Should be either '' or 'TLS' or 'SSL'")
		return err
	}
	//-- Process Config File

	return nil
}

func loggerGen(t int, s string) string {

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
	currentTime := time.Now().UTC()
	time := currentTime.Format("2006/01/02 15:04:05")
	return time + " " + errorLogPrefix + s + "\n"
}
func loggerWriteBuffer(s string) {
	logger(0, s, false)
}

//-- Loggin function
func logger(t int, s string, outputtoCLI bool) {

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
		errorLogPrefix = ""
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

//-- complete
func complete() {
	//-- End output
	espLogger("Errors: "+fmt.Sprintf("%d", counters.errors), "error")
	espLogger("Updated: "+fmt.Sprintf("%d", counters.updated), "debug")
	espLogger("Updated Skipped: "+fmt.Sprintf("%d", counters.updatedSkipped), "debug")
	espLogger("Created: "+fmt.Sprintf("%d", counters.created), "debug")
	espLogger("Created Skipped: "+fmt.Sprintf("%d", counters.createskipped), "debug")
	espLogger("Profiles Updated: "+fmt.Sprintf("%d", counters.profileUpdated), "debug")
	espLogger("Profiles Skipped: "+fmt.Sprintf("%d", counters.profileSkipped), "debug")
	espLogger("Time Taken: "+fmt.Sprintf("%v", Time.endTime), "debug")
	espLogger("---- XMLMC LDAP User Import Complete ---- ", "debug")
}

//-- Log to ESP
func espLogger(message string, severity string) bool {

	// We lock the whole function so we dont reuse the same connection for multiple logging attempts
	mutexLogger.Lock()
	defer mutexLogger.Unlock()

	// We initilaise the connection pool the first time the function is called and reuse it
	// This is reuse the connections rather than creating a pool each invocation
	once.Do(func() {

		loggerAPI = apiLib.NewXmlmcInstance(ldapImportConf.InstanceID)
		loggerAPI.SetAPIKey(ldapImportConf.APIKey)
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

// CounterInc Generic Counter Increment
func CounterInc(counter uint32) {
	Mutex.Counters.Lock()
	counter++
	Mutex.Counters.Unlock()
}
