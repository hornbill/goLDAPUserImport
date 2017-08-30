package main

import (
	"bytes"
	"crypto/tls"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/hornbill/goApiLib"
	"github.com/hornbill/ldap"
)

func userAddImage(p *ldap.Entry, buffer *bytes.Buffer, espXmlmc *apiLib.XmlmcInstStruct) {

	UserID := getFeildValue(p, "UserID", buffer)

	//-- Work out the value of URI which may contain [] for LDAP attribute references or just a string
	value := processComplexFeild(p, ldapImportConf.ImageLink.URI, buffer)
	buffer.WriteString(loggerGen(1, "Image Lookup URI: "+fmt.Sprintf("%s", value)))

	strContentType := "image/jpeg"
	if ldapImportConf.ImageLink.ImageType != "jpg" {
		strContentType = "image/png"
	}

	if strings.ToUpper(ldapImportConf.ImageLink.UploadType) != "URL" {
		// get binary to upload via WEBDAV and then set value to relative "session" URI
		var imageB []byte
		var Berr error
		switch strings.ToUpper(ldapImportConf.ImageLink.UploadType) {
		//-- Get Local URL
		case "URI":
			//-- Add Support for local HTTPS URLS with invalid cert
			tr := &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: ldapImportConf.ImageLink.InsecureSkipVerify},
			}
			client := &http.Client{Transport: tr}
			resp, err := client.Get(value)
			if err != nil {
				buffer.WriteString(loggerGen(4, "Unable to get image URI: "+value+" ("+fmt.Sprintf("%v", http.StatusInternalServerError)+") ["+fmt.Sprintf("%v", err)+"]"))
				return
			}
			defer resp.Body.Close()
			if resp.StatusCode == 201 || resp.StatusCode == 200 {
				imageB, _ = ioutil.ReadAll(resp.Body)
			} else {
				buffer.WriteString(loggerGen(4, "Unsuccesful download: "+fmt.Sprintf("%v", resp.StatusCode)))
				return
			}
		case "AD":
			imageB = []byte(value)
		default:
			imageB, Berr = hex.DecodeString(value[2:]) //stripping leading 0x
			if Berr != nil {
				buffer.WriteString(loggerGen(4, "Unsuccesful Decoding: "+fmt.Sprintf("%v", Berr)))
				return
			}
		}
		//WebDAV upload
		relLink := "session/" + UserID + "." + ldapImportConf.ImageLink.ImageType
		strDAVurl := espXmlmc.DavEndpoint + relLink
		buffer.WriteString(loggerGen(1, "DAV Upload URL: "+fmt.Sprintf("%s", strDAVurl)))
		if len(imageB) > 0 {
			putbody := bytes.NewReader(imageB)
			req, Perr := http.NewRequest("PUT", strDAVurl, putbody)
			if Perr != nil {
				buffer.WriteString(loggerGen(4, "Unsuccesful Request PUT Creation: "+fmt.Sprintf("%v", Perr)))
				return
			}
			req.Header.Set("Content-Type", strContentType)
			req.Header.Add("Authorization", "ESP-APIKEY "+ldapImportConf.APIKey)
			req.Header.Set("User-Agent", "Go-http-client/1.1")
			response, Perr := client.Do(req)
			if Perr != nil {
				buffer.WriteString(loggerGen(4, "PUT connection Issue: ("+fmt.Sprintf("%v", http.StatusInternalServerError)+") ["+fmt.Sprintf("%v", Perr)+"]"))
				return
			}
			defer response.Body.Close()
			_, _ = io.Copy(ioutil.Discard, response.Body)
			if response.StatusCode == 201 || response.StatusCode == 200 {
				buffer.WriteString(loggerGen(1, "Uploaded"))
				value = "/" + relLink
			} else {
				buffer.WriteString(loggerGen(4, "Unsuccesful Upload: "+fmt.Sprintf("%v", response.StatusCode)))
				return
			}
		} else {
			buffer.WriteString(loggerGen(4, "No Image to upload"))
			return
		}
	}
	buffer.WriteString(loggerGen(1, "Profile Set Image URL: "+fmt.Sprintf("%s", value)))
	espXmlmc.SetParam("objectRef", "urn:sys:user:"+UserID)
	espXmlmc.SetParam("sourceImage", value)

	XMLSiteSearch, xmlmcErr := espXmlmc.Invoke("activity", "profileImageSet")
	var xmlRespon xmlmcprofileSetImageResponse
	if xmlmcErr != nil {
		log.Fatal(xmlmcErr)
		buffer.WriteString(loggerGen(4, "Unable to associate Image to User Profile: "+fmt.Sprintf("%v", xmlmcErr)))
	}
	err := xml.Unmarshal([]byte(XMLSiteSearch), &xmlRespon)
	if err != nil {
		buffer.WriteString(loggerGen(4, "Unable to Associate Image to User Profile: "+fmt.Sprintf("%v", err)))
	} else {
		if xmlRespon.MethodResult != constOK {
			buffer.WriteString(loggerGen(4, "Unable to Associate Image to User Profile: "+xmlRespon.State.ErrorRet))
		} else {
			buffer.WriteString(loggerGen(1, "Image added to User: "+UserID))
		}
	}
}
