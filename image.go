package main

import (
	"bytes"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/hornbill/goApiLib"
	"github.com/hornbill/ldap"
)

func userAddImage(p *ldap.Entry, buffer *bytes.Buffer, espXmlmc *apiLib.XmlmcInstStruct) {

	UserID := getFeildValue(p, "UserID", buffer)

	//	value := p.GetAttributeValue(ldapImportConf.ImageLink.URI)
	value := getFeildValue(p, ldapImportConf.ImageLink.URI, buffer)
	fmt.Println(value)

	strContentType := "image/jpeg"
	if ldapImportConf.ImageLink.ImageType != "jpg" {
		strContentType = "image/png"
	}

	if strings.ToUpper(ldapImportConf.ImageLink.UploadType) != "URI" {
		// get binary to upload via WEBDAV and then set value to relative "session" URI
		client := http.Client{
			Transport: &http.Transport{
				Proxy: http.ProxyFromEnvironment,
			},
			Timeout: time.Duration(10 * time.Second),
		}

		rel_link := "session/" + UserID
		strDAVurl := ldapImportConf.DAVURL + rel_link

		var imageB []byte
		var Berr error
		switch strings.ToUpper(ldapImportConf.ImageLink.UploadType) {

		case "URL":
			resp, err := http.Get(value)
			if err != nil {
				buffer.WriteString(loggerGen(4, "Unable to find "+value+" ["+fmt.Sprintf("%v", http.StatusInternalServerError)+"]"))
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
			value = p.GetAttributeValue(ldapImportConf.ImageLink.URI)
			imageB = []byte(value)

		default:
			imageB, Berr = hex.DecodeString(value[2:]) //stripping leading 0x
			if Berr != nil {
				buffer.WriteString(loggerGen(4, "Unsuccesful Decoding "+fmt.Sprintf("%v", Berr)))
				return
			}

		}
		//WebDAV upload
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
				buffer.WriteString(loggerGen(4, "PUT connection issue: "+fmt.Sprintf("%v", http.StatusInternalServerError)))
				return
			}
			defer response.Body.Close()
			_, _ = io.Copy(ioutil.Discard, response.Body)
			if response.StatusCode == 201 || response.StatusCode == 200 {
				buffer.WriteString(loggerGen(1, "Uploaded"))
				value = "/" + rel_link
			} else {
				buffer.WriteString(loggerGen(4, "Unsuccesful Upload: "+fmt.Sprintf("%v", response.StatusCode)))
				return
			}
		} else {
			buffer.WriteString(loggerGen(4, "No Image to upload"))
			return
		}
	}

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
