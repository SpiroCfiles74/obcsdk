package peerrest

import (
	"bytes"
	"fmt"
	"io/ioutil"
	//"log"
	"net/http"
	"crypto/tls"
	"os"
	"time"
)

const (
	waitSecs = 30
	waitTimeoutRetries = 1
 )

// Calling GetChainInfo according to http or https api according to the value in env variable "TEST_NETWORK"
// "TEST_NETWORK" == "LOCAL" - would use a network with http protocol
// "TEST_NETWORK" == "Z" - would use https protocol

func GetChainInfo(url string) (respBody string, respStatus string){
	if os.Getenv("TEST_NET_COMM_PROTOCOL") == "HTTPS" || os.Getenv("TEST_NETWORK") == "Z" {
		respBody, respStatus = GetChainInfo_HTTPS(url)
	} else  {
		respBody, respStatus = GetChainInfo_HTTP(url)
	}
	return respBody, respStatus
}

/*
  Issue GET request to BlockChain resource
    url is the GET request.
	respStatus is the HTTP response status code and message
	respBody is the HTTP response body
*/
func GetChainInfo_HTTP(url string) (respBody string, respStatus string) {
	//TODO : define a logger
	//fmt.Println("GetChainInfo_HTTP :", url)

	httpclient := &http.Client{ Timeout: time.Second * waitSecs }
	response, err := httpclient.Get(url)

	if err != nil {
		fmt.Println("Error from httpclient.GET request, url, response: ", url, response)
		fmt.Println("Error from httpclient.GET request, err: ", err)
		return err.Error(), "Error from httpclient.GET request"
	} else {
		defer response.Body.Close()
		contents, err := ioutil.ReadAll(response.Body)
		if err != nil {
			fmt.Println("Error from ioutil.ReadAll during GET request, url, response, contents: ", url, response, contents)
			fmt.Println("Error from ioutil.ReadAll during GET request, err: ", err)
			return err.Error(), "Error from ioutil.ReadAll during GET request"
		}
		return string(contents), response.Status
	}
}

/*
  Issue GET request to BlockChain resource
    url is the GET request.
	respStatus is the HTTPS response status code and message
	respBody is the HTTPS response body
*/
func GetChainInfo_HTTPS(url string) (respBody string, respStatus string) {
	//TODO : define a logger
	//fmt.Println("GetChainInfo_HTTPS :", url)

        tr := &http.Transport{
	         TLSClientConfig:    &tls.Config{RootCAs: nil},
	         DisableCompression: true,
        }
        httpsclient := &http.Client{ Timeout: time.Second * waitSecs, Transport: tr }
        response, err := httpsclient.Get(url)
	if err != nil {
			fmt.Println("ERROR from httpsclient.GET request, url, response: ", url, response)
			fmt.Println("ERROR from httpsclient.GET request, err: ", err)
			return err.Error(), "ERROR from httpsclient.GET request"
	} else {
		defer response.Body.Close()
		contents, err := ioutil.ReadAll(response.Body)
	        if err != nil {
			fmt.Println("ERROR from https ioutil.ReadAll during GET request, url, response, contents: ", url, response, contents)
			fmt.Println("ERROR from https ioutil.ReadAll during GET request, err: ", err)
			return err.Error(), "ERROR from https ioutil.ReadAll during GET request"
		}
		return string(contents), response.Status
	}
}

// Calling GetChainInfo according to http or https api according to the value in env variable "TEST_NETWORK"
// "TEST_NETWORK" == "LOCAL" - would use a network with http protocol
// "TEST_NETWORK" == "Z" || "TEST_NET_COMM_PROTOCOL" = "HTTPS" - we would use https protocol

func PostChainAPI(url string, payLoad []byte) (respBody string, respStatus string){
	if os.Getenv("TEST_NET_COMM_PROTOCOL") == "HTTPS" || os.Getenv("TEST_NETWORK") == "Z" {
		respBody, respStatus = PostChainAPI_HTTPS(url, payLoad)
	} else  {
		respBody, respStatus = PostChainAPI_HTTP(url, payLoad)
	}
	return respBody, respStatus
}

/*
  Issue POST request to BlockChain resource.
	url is the target resource.
	payLoad is the REST API payload
	respStatus is the HTTP response status code and message
	respBody is the HHTP response body
*/
func PostChainAPI_HTTP(url string, payLoad []byte) (respBody string, respStatus string) {

	veryverbose := false 	// for debugging github hyperledger fabric issue #2357

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payLoad))
	//req.Header.Set("X-Custom-Header", "myvalue")
	req.Header.Set("Content-Type", "application/json")

	if veryverbose {
		fmt.Println("PostChainAPI() calling http.Client.Do to url=" + url) 
	}
	httpclient := &http.Client{ Timeout: time.Second * waitSecs }
	resp, err := httpclient.Do(req)
	if veryverbose {
		fmt.Println("PostChainAPI()  AFTER  http.Client.Do(req)")
	}
	if err != nil {
		fmt.Println("PostChainAPI() httpclient.Do Error, url, response: ", url, resp)
		fmt.Println("PostChainAPI() httpclient.Do Error, err: ", err)
		return err.Error(), "httpclient.Do Error"
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if veryverbose {
		fmt.Println("PostChainAPI() >>> response Status:", resp.Status)
		fmt.Println("PostChainAPI() >>> response Body:", body)
	}
	if err != nil {
		fmt.Println("PostChainAPI() Error from ioutil.ReadAll(), url, response: ", url, body)
		fmt.Println("PostChainAPI() Error from ioutil.ReadAll(), err: ", err)
		return err.Error(), "ERROR from ioutil.ReadAll"
	}
	//return string(body), resp.Status
	return string(body), string("")
}

/*
  Issue POST request to BlockChain resource.
	url is the target resource.
	payLoad is the REST API payload
	respStatus is the HTTP response status code and message
	respBody is the HHTP response body
*/
func PostChainAPI_HTTPS(url string, payLoad []byte) (respBody string, respStatus string) {

	veryverbose := false 	// for debugging github hyperledger fabric issue #2357

	if veryverbose {
		fmt.Println("PostChainAPI()_HTTPS url=" + url) 
	}
        tr := &http.Transport{
                 TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	         //TLSClientConfig:    &tls.Config{RootCAs: nil},
	         DisableCompression: true,
        }
        httpclient := &http.Client{ Transport: tr, Timeout: time.Second * waitSecs }
	if veryverbose {
		fmt.Println("PostChainAPI()_HTTPS calling http.Client.Post=" + url) 
	}
	response, err := httpclient.Post(url, "json", bytes.NewBuffer(payLoad))
	if veryverbose {
		fmt.Println("PostChainAPI()  AFTER  http.Client.Post")
	}

	if err != nil {
		fmt.Println("PostChainAPI_HTTPS() httpclient.Post Error, url, response: ", url, response)
		fmt.Println("PostChainAPI_HTTPS() httpclient.Post Error, err: ", err)
		return err.Error(), "httpclient.Post Error"
	}
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Println("PostChainAPI_HTTPS Error from ioutil.ReadAll, url, response: ", url, body)
		fmt.Println("PostChainAPI_HTTPS Error from ioutil.ReadAll, err: ", err)
	}
	if veryverbose {
		fmt.Println("PostChainAPI_HTTPS() secure postchain >>> response Status:", response.Status)
		fmt.Println("PostChainAPI_HTTPS() secure postchain >>> response Body:", body)
	}
	//return string(body), response.Status
	return string(body), string("")
}
