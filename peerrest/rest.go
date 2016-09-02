package peerrest

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"crypto/tls"
	"os"
	"time"
)


// Calling GetChainInfo according to http or https api according to the value in env variable "NETWORK"
// "NETWORK" = "LOCAL" - would use a network with http protocol
// "NETWORK" = "Z" - would use https protocol
func GetChainInfo(url string) (respBody string, respStatus string){
	if os.Getenv("NET_COMM_PROTOCOL") == "HTTPS" { /* http */
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

	httpclient := &http.Client{ Timeout: time.Second * 10 }
	response, err := httpclient.Get(url)

	if err != nil {
		fmt.Printf("%s", err)
		return err.Error(), "Error from GET request"
	} else {
		defer response.Body.Close()
		contents, err := ioutil.ReadAll(response.Body)
		if err != nil {
			fmt.Printf("%s", err)
			return err.Error(), "Error from GET request"
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
        httpclient := &http.Client{ Timeout: time.Second * 10, Transport: tr }
        response, err := httpclient.Get(url)
	if err != nil {
			fmt.Printf("%s", err)
			return err.Error(), "Error from GET request"
	} else {
		defer response.Body.Close()
		contents, err := ioutil.ReadAll(response.Body)
	        if err != nil {
			fmt.Printf("%s", err)
			return err.Error(), "Error from GET request"
		}
		return string(contents), response.Status
	}
}

// Calling GetChainInfo according to http or https api according to the value in env variable "NETWORK"
// "NETWORK" = "LOCAL" - would use a network with http protocol
// "NETWORK" = "Z" - would use https protocol

func PostChainAPI(url string, payLoad []byte) (respBody string, respStatus string){
	if os.Getenv("NET_COMM_PROTOCOL") == "HTTPS" { /* http */
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
	httpclient := &http.Client{ Timeout: time.Second * 10 }
	resp, err := httpclient.Do(req)
	if veryverbose {
		fmt.Println("PostChainAPI()  AFTER  http.Client.Do(req)")
	}
	if err != nil {
		log.Println("httpclient.Do Error", url, err)
		return err.Error(), "httpclient.Do Error"
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if veryverbose {
		fmt.Println("PostChainAPI() >>> response Status:", resp.Status)
		fmt.Println("PostChainAPI() >>> response Body:", body)
	}
	if err != nil {
		//fmt.Println("PostChainAPI() Error from ioutil.ReadAll()")
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
        httpclient := &http.Client{ Transport: tr, Timeout: time.Second * 10 }
	if veryverbose {
		fmt.Println("PostChainAPI()_HTTPS calling http.Client.Post=" + url) 
	}
	response, err := httpclient.Post(url, "json", bytes.NewBuffer(payLoad))
	if veryverbose {
		fmt.Println("PostChainAPI()  AFTER  http.Client.Post")
	}

	if err != nil {
		log.Println("httpclient.Post Error", url, err)
		return err.Error(), "httpclient.Post Error"
	}
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Println("Error from ioutil.ReadAll")
	}
	if veryverbose {
		fmt.Println("PostChainAPI() secure postchain >>> response Status:", response.Status)
		fmt.Println("PostChainAPI() secure postchain >>> response Body:", body)
	}
	//return string(body), response.Status
	return string(body), string("")
}
