package main

//----- Packages -----
import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"encoding/xml"
	"errors"
	"flag"
	"fmt"
	"html"
	"log"
	"os"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/fatih/color" //-- CLI Colour
	"github.com/hornbill/goApiLib"
	"github.com/hornbill/ldap"    //-- Hornbill Clone of "github.com/mavricknz/ldap"
	"github.com/hornbill/pb"      //--Hornbil Clone of "github.com/cheggaaa/pb"
	"github.com/tcnksm/go-latest" //-- For Version checking
)

//----- Main Function -----
func main() {

	//-- Initiate Variables
	initVars()

	//-- Process Flags
	procFlags()

	//-- If configVersion just output version number and die
	if configVersion {
		fmt.Printf("%v \n", version)
		return
	}

	//-- Check for latest
	checkVersion()

	//-- Load Configuration File Into Struct
	ldapImportConf = loadConfig()

	//-- Validation on Configuration File
	err := validateConf()
	if err != nil {
		logger(4, fmt.Sprintf("%v", err), true)
		logger(4, "Please Check your Configuration File: "+fmt.Sprintf("%s", configFileName), true)
		return
	}

	//-- Set Instance ID
	var boolSetInstance = setInstance(configZone, ldapImportConf.InstanceID)
	if boolSetInstance != true {
		return
	}

	//-- Generate Instance XMLMC Endpoint
	ldapImportConf.URL = getInstanceURL()
	logger(1, "Instance Endpoint "+fmt.Sprintf("%v", ldapImportConf.URL), true)

	//-- Once we have loaded the config write to hornbill log file
	espLogger("---- XMLMC LDAP Import Utility V"+fmt.Sprintf("%v", version)+" ----", "debug")

	//-- Query LDAP
	var boolLDAPUsers = queryLdap()

	if boolLDAPUsers {
		processUsersFromWorkers()
	}

	outputEnd()
}

func outputEnd() {
	//-- End output
	if errorCount > 0 {
		logger(4, "Error encountered please check the log file", true)
		logger(4, "Error Count: "+fmt.Sprintf("%d", errorCount), true)
		//logger(4, "Check Log File for Details", true)
	}
	logger(1, "Updated: "+fmt.Sprintf("%d", counters.updated), true)
	logger(1, "Updated Skipped: "+fmt.Sprintf("%d", counters.updatedSkipped), true)

	logger(1, "Created: "+fmt.Sprintf("%d", counters.created), true)
	logger(1, "Created Skipped: "+fmt.Sprintf("%d", counters.createskipped), true)

	logger(1, "Profiles Updated: "+fmt.Sprintf("%d", counters.profileUpdated), true)
	logger(1, "Profiles Skipped: "+fmt.Sprintf("%d", counters.profileSkipped), true)

	//-- Show Time Takens
	endTime = time.Now().Sub(startTime)
	logger(1, "Time Taken: "+fmt.Sprintf("%v", endTime), true)
	//-- complete
	complete()
	logger(1, "---- XMLMC LDAP Import Complete ---- ", true)
}
func procFlags() {
	//-- Grab Flags
	flag.StringVar(&configFileName, "file", "conf.json", "Name of Configuration File To Load")
	flag.StringVar(&configZone, "zone", "eur", "Override the default Zone the instance sits in")
	flag.BoolVar(&configDryRun, "dryrun", false, "Allow the Import to run without Creating or Updating users")
	flag.BoolVar(&configVersion, "version", false, "Output Version")
	flag.IntVar(&configWorkers, "workers", 1, "Number of Worker threads to use")

	//-- Parse Flags
	flag.Parse()

	//-- Output config
	if !configVersion {
		outputFlags()
	}
}
func outputFlags() {
	//-- Output
	logger(1, "---- XMLMC LDAP Import Utility V"+fmt.Sprintf("%v", version)+" ----", true)

	logger(1, "Flag - Config File "+fmt.Sprintf("%s", configFileName), true)
	logger(1, "Flag - Zone "+fmt.Sprintf("%s", configZone), true)
	logger(1, "Flag - Dry Run "+fmt.Sprintf("%v", configDryRun), true)
	logger(1, "Flag - Workers "+fmt.Sprintf("%v", configWorkers), false)
}
func initVars() {
	//-- Start Time for Durration
	startTime = time.Now()
	//-- Start Time for Log File
	timeNow = time.Now().Format(time.RFC3339)
	//-- Remove :
	timeNow = strings.Replace(timeNow, ":", "-", -1)
	//-- Set Counter
	errorCount = 0
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
		logger(3, fmt.Sprintf("%s", version)+" is not latest, you should upgrade to "+fmt.Sprintf("%s", res.Current)+" by downloading the latest package Here https://github.com/hornbill/goLDAPUserImport/releases/tag/v"+fmt.Sprintf("%s", res.Current), true)
	}
}

//-- Function to Load Configruation File
func loadConfig() ldapImportConfStruct {
	//-- Check Config File File Exists
	cwd, _ := os.Getwd()
	configurationFilePath := cwd + "/" + configFileName
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
	//-- Process Config File

	return nil
}

//-- Worker Pool Function
func processUsersFromWorkers() {
	bar := pb.StartNew(len(ldapUsers))
	logger(1, "Processing Users", false)
	
	total := len(ldapUsers)
	jobs := make(chan int, total)
	results := make(chan int, total)
	workers := configWorkers

	if total < workers {
		workers = total
	}
	//This starts up 3 workers, initially blocked because there are no jobs yet.
	for w := 1; w <= workers; w++ {
		go processUsers(w, jobs, results, bar)
	}
	//-- Here we send a job for each user we have to process
	for j := 1; j <= total; j++ {
		jobs <- j
	}
	close(jobs)
	//-- Finally we collect all the results of the work.
	for a := 1; a <= total; a++ {
		<-results
	}
	bar.FinishPrint("Processing Complete!")
}

//-- Process Users
func processUsers(id int, jobs <-chan int, results chan<- int, bar *pb.ProgressBar) {

	//-- Range On Jobs for worker
	for j := range jobs {
		//-- Get User Record from Array
		ldapUser := ldapUsers[j-1]
		var buffer bytes.Buffer
		//-- Get User Id based on the mapping
		var userID = strings.ToLower(getFeildValue(ldapUser, "UserID", &buffer))
		if userID == "" {
			buffer.WriteString(loggerGen(1, "Unable to Proceess User - Invalid Id"+userID))
		} else {
			buffer.WriteString(loggerGen(1, "Buffer For Job: "+fmt.Sprintf("%d", j)+" - Worker: "+fmt.Sprintf("%d", id)+" - User: "+userID))

			//-- For Each LDAP Users Check if they already Exist
			boolUpdate, err := checkUserOnInstance(userID)
			if err != nil {
				buffer.WriteString(loggerGen(4, "Unable to Search For User: "+fmt.Sprintf("%v", err)))
			}
			//-- User Exists so Update
			if boolUpdate {
				buffer.WriteString(loggerGen(1, "Update User: "+userID))
				_, errUpdate := updateUser(ldapUser, &buffer)
				if errUpdate != nil {
					buffer.WriteString(loggerGen(4, "Unable to Update User: "+fmt.Sprintf("%v", errUpdate)))
				}
			} else {
				buffer.WriteString(loggerGen(1, "Create User: "+userID))
				//-- User Does not Exist so Create
				if ldapUser != nil {
					_, errorCreate := createUser(ldapUser, &buffer)
					if errorCreate != nil {
						buffer.WriteString(loggerGen(4, "Unable to Create User: "+fmt.Sprintf("%v", errorCreate)))
					}
				}

			}
		}
		//-- Increment
		bar.Increment()
		bufferMutex.Lock()
		loggerWriteBuffer(buffer.String())
		bufferMutex.Unlock()
		buffer.Reset()
		//-- Results
		results <- j * 2
	}

}

//-- Get XMLMC Feild from mapping via User Object
func getFeildValue(u *ldap.Entry, s string, buffer *bytes.Buffer) string {
	//-- Dyniamicly Grab Mapped Value
	r := reflect.ValueOf(ldapImportConf.UserMapping)
	f := reflect.Indirect(r).FieldByName(s)
	//-- Get Mapped Value
	var UserMapping = f.String()
	return processComplexFeild(u, UserMapping, buffer)
}

//-- Get XMLMC Feild from mapping via profile Object
func getFeildValueProfile(u *ldap.Entry, s string, buffer *bytes.Buffer) string {
	//-- Dyniamicly Grab Mapped Value
	r := reflect.ValueOf(ldapImportConf.UserProfileMapping)

	f := reflect.Indirect(r).FieldByName(s)

	//-- Get Mapped Value
	var UserProfileMapping = f.String()
	return processComplexFeild(u, UserProfileMapping, buffer)
}
func processComplexFeild(u *ldap.Entry, s string, buffer *bytes.Buffer) string {
	//-- Match $variables from String
	re1, err := regexp.Compile(`\[(.*?)\]`)
	if err != nil {
		buffer.WriteString(loggerGen(4, "Regex Error"+fmt.Sprintf("%v", err)))
	}
	//-- Get Array of all Matched max 100
	result := re1.FindAllString(s, 100)

	//-- Loop Matches
	for _, v := range result {
		//-- Grab LDAP Mapping value from result set
		var LDAPAttributeValue = u.GetAttributeValue(v[1 : len(v)-1])
		//-- Check for Invalid Value
		if LDAPAttributeValue == "" {
			buffer.WriteString(loggerGen(4, "Unable to Load LDAP Attribute: "+v[1:len(v)-1]+" For Input Param: "+s))
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
	//-- Curreny WD
	cwd, _ := os.Getwd()
	//-- Log Folder
	logPath := cwd + "/log"
	//-- Log File
	logFileName := logPath + "/LDAP_User_Import_" + timeNow + ".log"
	red := color.New(color.FgRed).PrintfFunc()
	orange := color.New(color.FgCyan).PrintfFunc()
	//-- If Folder Does Not Exist then create it
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		err := os.Mkdir(logPath, 0777)
		if err != nil {
			fmt.Printf("Error Creating Log Folder %q: %s \r", logPath, err)
			os.Exit(101)
		}
	}

	//-- Open Log File
	f, err := os.OpenFile(logFileName, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0777)
	if err != nil {
		fmt.Printf("Error Creating Log File %q: %s \n", logFileName, err)
		os.Exit(100)
	}
	// don't forget to close it
	defer f.Close()
	// assign it to the standard logger
	logFileMutex.Lock()
	log.SetOutput(f)
	logFileMutex.Unlock()
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
	espLogger("Errors: "+fmt.Sprintf("%d", errorCount), "error")
	espLogger("Updated: "+fmt.Sprintf("%d", counters.updated), "debug")
	espLogger("Updated Skipped: "+fmt.Sprintf("%d", counters.updatedSkipped), "debug")
	espLogger("Created: "+fmt.Sprintf("%d", counters.created), "debug")
	espLogger("Created Skipped: "+fmt.Sprintf("%d", counters.createskipped), "debug")
	espLogger("Profiles Updated: "+fmt.Sprintf("%d", counters.profileUpdated), "debug")
	espLogger("Profiles Skipped: "+fmt.Sprintf("%d", counters.profileSkipped), "debug")
	espLogger("Time Taken: "+fmt.Sprintf("%v", endTime), "debug")
	espLogger("---- XMLMC LDAP User Import Complete ---- ", "debug")
}

// Set Instance Id
func setInstance(strZone string, instanceID string) bool {
	//-- Set Zone
	setZone(strZone)
	//-- Check for blank instance
	if instanceID == "" {
		logger(4, "InstanceId Must be Specified in the Configuration File", true)
		return false
	}
	//-- Set Instance
	xmlmcInstanceConfig.instance = instanceID
	return true
}

// Set Instance Zone to Overide Live
func setZone(zone string) {
	xmlmcInstanceConfig.zone = zone

	return
}

//-- Log to ESP
func espLogger(message string, severity string) {
	espXmlmc := apiLib.NewXmlmcInstance(ldapImportConf.URL)
	espXmlmc.SetAPIKey(ldapImportConf.APIKey)
	espXmlmc.SetParam("fileName", "LDAP_User_Import")
	espXmlmc.SetParam("group", "general")
	espXmlmc.SetParam("severity", severity)
	espXmlmc.SetParam("message", message)

	XMLLogger, xmlmcErr := espXmlmc.Invoke("system", "logMessage")
	var xmlRespon xmlmcResponse
	if xmlmcErr != nil {
		logger(4, "Unable to write to log "+fmt.Sprintf("%s", xmlmcErr), true)
	}
	err := xml.Unmarshal([]byte(XMLLogger), &xmlRespon)
	if err != nil {
		logger(4, "Unable to write to log "+fmt.Sprintf("%s", err), true)
	}
	if xmlRespon.MethodResult != constOK {
		logger(4, "Unable to write to log "+xmlRespon.State.ErrorRet, true)
	}
}

func errorCountInc() {
	mutexCounters.Lock()
	errorCount++
	mutexCounters.Unlock()
}
func updateCountInc() {
	mutexCounters.Lock()
	counters.updated++
	mutexCounters.Unlock()
}
func updateSkippedCountInc() {
	mutexCounters.Lock()
	counters.updatedSkipped++
	mutexCounters.Unlock()
}
func createSkippedCountInc() {
	mutexCounters.Lock()
	counters.createskipped++
	mutexCounters.Unlock()
}
func createCountInc() {
	mutexCounters.Lock()
	counters.created++
	mutexCounters.Unlock()
}
func profileCountInc() {
	mutexCounters.Lock()
	counters.profileUpdated++
	mutexCounters.Unlock()
}
func profileSkippedCountInc() {
	mutexCounters.Lock()
	counters.profileSkipped++
	mutexCounters.Unlock()
}

//-- Function Builds XMLMC End Point
func getInstanceURL() string {
	xmlmcInstanceConfig.url = "https://"
	xmlmcInstanceConfig.url += xmlmcInstanceConfig.zone
	xmlmcInstanceConfig.url += "api.hornbill.com/"
	xmlmcInstanceConfig.url += xmlmcInstanceConfig.instance
	xmlmcInstanceConfig.url += "/xmlmc/"

	return xmlmcInstanceConfig.url
}
