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

func loadImageFromValue(importDate *userWorkingDataStruct) []byte {

	//-- Work out the value of URI which may contain [] for LDAP attribute references or just a string
	value := processComplexFeild(importDate.LDAP, ldapImportConf.User.Image.URI)
	logger(1, "Image Lookup URI: "+fmt.Sprintf("%s", value), false)

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
			resp, err := client.Get(value)
			if err != nil {
				logger(4, "Unable to get image URI: "+value+" ("+fmt.Sprintf("%v", http.StatusInternalServerError)+") ["+fmt.Sprintf("%v", err)+"]", false)
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
			imageB = []byte(value)
		default:
			imageB, Berr = hex.DecodeString(value[2:]) //stripping leading 0x
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
	//-- Try and Load from Cache
	_, found := HornbillCache.Images[ldapImportConf.User.Image.URI]
	if found {
		image = HornbillCache.Images[ldapImportConf.User.Image.URI]
	} else {
		//- Load Image if we have one into bytes
		imageBytes = loadImageFromValue(importData)

		//-- Validate Sha1 hex string against what we currently have
		imageCheckSumHex := fmt.Sprintf("%x", sha1.Sum(imageBytes))

		//-- Store in cache
		image.imageBytes = imageBytes
		image.imageCheckSum = imageCheckSumHex
		HornbillCache.Images[ldapImportConf.User.Image.URI] = image
	}

	return image
}

// Write DN and User ID to Cache
func writeImageToCache(URI string, image imageStruct) {
	_, found := HornbillCache.Images[URI]
	if !found {
		HornbillCache.Images[URI] = image
	}
}
