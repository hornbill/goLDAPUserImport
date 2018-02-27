package main

import (
	"fmt"

	"github.com/hornbill/goApiLib"
	"github.com/hornbill/pb"
)

var (
	hIF *apiLib.XmlmcInstStruct
)

//-- Second Connection here for finalising data not always used
func initXMLMCFinalise() {
	hIF = apiLib.NewXmlmcInstance(Flags.configInstanceId)
	hIF.SetAPIKey(Flags.configApiKey)
	hIF.SetTimeout(5)
	hIF.SetJSONResponse(true)
}
func finaliseData() {
	logger(1, "Finalising User Data", true)
	initXMLMCFinalise()
	bar := pb.StartNew(len(HornbillCache.UsersWorking))

	// This we could do in threads
	for index := range HornbillCache.UsersWorking {
		currentUser := HornbillCache.UsersWorking[index]

		if currentUser.Jobs.create {
			createUser(currentUser)
		}

		if currentUser.Jobs.update || currentUser.Jobs.updateType || currentUser.Jobs.updateSite {
			updateUser(currentUser)
		}
		if currentUser.Jobs.updateProfile {
			updateUserProfile(currentUser)
		}
		if currentUser.Jobs.updateImage {
			updateUserImage(currentUser)
		}
		if len(currentUser.Groups) > 0 {
			updateUserGroups(currentUser)
		}
		if len(currentUser.Roles) > 0 {
			updateUserRoles(currentUser)
		}

		bar.Increment()
	}

	bar.FinishPrint("Finalising User Data Complete")
}

func createUser(currentUser *userWorkingDataStruct) {
	b, err := userCreate(currentUser)
	if b {
		CounterInc(1)
	} else {
		CounterInc(7)
		logger(4, "Unable to Create User: "+currentUser.Account.UserID+" Error: "+fmt.Sprintf("%s", err), false)
	}

}

func updateUser(currentUser *userWorkingDataStruct) {
	b, err := userUpdate(currentUser)
	if b {
		CounterInc(2)
	} else {
		CounterInc(7)
		logger(4, "Unable to Update User: "+currentUser.Account.UserID+" Error: "+fmt.Sprintf("%s", err), false)
	}
}

func updateUserProfile(currentUser *userWorkingDataStruct) {
	b, err := userProfileUpdate(currentUser)
	if b {
		CounterInc(3)
	} else {
		CounterInc(7)
		logger(4, "Unable to Update User Profile: "+currentUser.Account.UserID+" Error: "+fmt.Sprintf("%s", err), false)
	}
}

func updateUserImage(currentUser *userWorkingDataStruct) {
	b, err := userImageUpdate(currentUser)
	if b {
		CounterInc(4)
	} else {
		CounterInc(7)
		logger(4, "Unable to Update User Image: "+currentUser.Account.UserID+" Error: "+fmt.Sprintf("%s", err), false)
	}
}

func updateUserGroups(currentUser *userWorkingDataStruct) {
	b, err := userGroupsUpdate(currentUser)
	if b {
		CounterInc(5)
	} else {
		CounterInc(7)
		logger(4, "Unable to Update User Groups: "+currentUser.Account.UserID+" Error: "+fmt.Sprintf("%s", err), false)
	}
}

func updateUserRoles(currentUser *userWorkingDataStruct) {
	b, err := userRolesUpdate(currentUser)
	if b {
		CounterInc(6)
	} else {
		CounterInc(7)
		logger(4, "Unable to Update User Roles: "+currentUser.Account.UserID+" Error: "+fmt.Sprintf("%s", err), false)
	}
}
