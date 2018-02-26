package main

import (
	"crypto/sha1"
	"crypto/tls"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

func loadImageFromValue(imageURI string) []byte {

	logger(1, "Image Lookup URI: "+fmt.Sprintf("%s", imageURI), false)

	if strings.ToUpper(ldapImportConf.User.Image.UploadType) != "URL" {
		// get binary to upload via WEBDAV and then set value to relative "session" URI
		var imageB []byte
		var Berr error
		switch strings.ToUpper(ldapImportConf.User.Image.UploadType) {
		//-- Get Local URL
		case "URI":
			//-- Add Support for local HTTPS URLS with invalid cert
			tr := &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: ldapImportConf.User.Image.InsecureSkipVerify},
			}
			client := &http.Client{Transport: tr}
			resp, err := client.Get(imageURI)
			if err != nil {
				logger(4, "Unable to get image URI: "+imageURI+" ("+fmt.Sprintf("%v", http.StatusInternalServerError)+") ["+fmt.Sprintf("%v", err)+"]", false)
				return nil
			}
			defer resp.Body.Close()
			if resp.StatusCode == 201 || resp.StatusCode == 200 {
				imageB, _ = ioutil.ReadAll(resp.Body)
			} else {
				logger(4, "Unsuccesful download: "+fmt.Sprintf("%v", resp.StatusCode), false)
				return nil
			}
		case "AD":
			imageB = []byte(imageURI)
		default:
			imageB, Berr = hex.DecodeString(imageURI[2:]) //stripping leading 0x
			if Berr != nil {
				logger(4, "Unsuccesful Decoding: "+fmt.Sprintf("%v", Berr), false)
				return nil
			}
		}
		return imageB
	}
	//-- Must be a URL
	response, err := http.Get(ldapImportConf.User.Image.URI)
	if err != nil {
		logger(4, "Unsuccesful Download: "+fmt.Sprintf("%v", err), false)
		return nil
	}
	defer response.Body.Close()
	htmlData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		logger(4, "Unsuccesful Image Download: "+fmt.Sprintf("%v", err), false)
		return nil
	}
	return htmlData

}

func getImage(importData *userWorkingDataStruct) imageStruct {
	var image imageStruct
	var imageBytes []byte

	//-- Work out the value of URI which may contain [] for LDAP attribute references or just a string
	importData.ImageURI = processComplexFeild(importData.LDAP, ldapImportConf.User.Image.URI)
	//-- Try and Load from Cache
	_, found := HornbillCache.Images[importData.ImageURI]
	if found {
		image = HornbillCache.Images[importData.ImageURI]
	} else {
		//- Load Image if we have one into bytes
		imageBytes = loadImageFromValue(importData.ImageURI)

		//-- Validate Sha1 hex string against what we currently have
		imageCheckSumHex := fmt.Sprintf("%x", sha1.Sum(imageBytes))

		//-- Store in cache
		image.imageBytes = imageBytes
		image.imageCheckSum = imageCheckSumHex
		HornbillCache.Images[importData.ImageURI] = image
	}
	return image
}
