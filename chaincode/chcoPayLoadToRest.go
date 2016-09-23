package chaincode

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"obcsdk/peerrest"
)

// These structures describe the response to a  POST /chaincode API call
// Formats defined here:
// https://github.com/hyperledger/fabric/blob/master/docs/API/CoreAPI.md#chaincode
type result_T struct { /* part of a successful restCallResult_T 	*/
	Status  string `json:"status"`
	Message string `json:"message"`
}
type error_T struct { /* part of a rejected restCallResult_T.  Rarely happens. */
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    string `json:"data"`
}
type restCallResult_T struct { /* response to a REST API POST /chaincode  	*/
	Jsonrpc string   `json:"jsonrpc"`
	Result  result_T `json:"result,omitempty"`
	Error   error_T  `json:"error,omitempty"`
	Id      int64    `json:"id"`
}

type Timestamps struct {
	Seconds int `json:"seconds"`
	Nanos int `json:"nanos"`
}

/* **********************************************************************
// v0.5 and prior
type Transactions struct {
	Type int  `json:"type,omitempty"`
	ChaincodeID string  `json:"chaincodeID"`
	Payload string  `json:"payload"`
	Txid string  `json:"txid"`
	Timestamp Timestamps  `json:"timestamp"`
	ConfidentialityLevel int `json:"confidentialityLevel"`
	ConfidentialityProtocolVersion string `json:"confidentialityProtocolVersion"`
	nonce string `json:"nonce"`
	toValidators string `json:"toValidators"`
	cert string `json:"cert"`
	signature string `json:"signature"`
}
type TransactionResults struct {
	Uuid string `json:"uuid,omitempty"`
	Result byte `json:"result,omitempty"`
	ErrorCode int `json:"errorCode,omitempty"`
	Error string `json:"error,omitempty"`
	//chaincodevent ChaincodeEvent `json:"chaincodeEvent,omitempty"`
}
type NonHashData struct {
		LocalLedgerCommitTimestamp Timestamps  `json:"localLedgerCommitTimestamp"`
		TransactionResult []TransactionResults  `json:"transactionResults"`
}
type Block struct {
	TransactionList []Transactions `json:"transactions"`
	StateHash  string `json:"stateHash"`
	PreviousBlockHash string `json:"previousBlockHash"`
	ConsensusMetadata string `json:"consensusMetadata"`
	NonHash NonHashData `json:"nonHashData"`
}
 ********************************************************************* */

// v0.6 and later releases
type Transactions struct {
	Type int  `json:"type,omitempty"`
	ChaincodeID string  `json:"chaincodeID"`
	Payload string  `json:"payload"`
	Txid string  `json:"txid"`
	Timestamp Timestamps  `json:"timestamp"`
	ConfidentialityLevel int `json:"confidentialityLevel"`
	ConfidentialityProtocolVersion string `json:"confidentialityProtocolVersion"`
	nonce string `json:"nonce"`
	toValidators string `json:"toValidators"`
	cert string `json:"cert"`
	signature string `json:"signature"`
}
type ChaincodeEvents struct {
	// 	? 
}
	
type NonHashData struct {
		LocalLedgerCommitTimestamp Timestamps  `json:"localLedgerCommitTimestamp"`
		ChaincodeEvent []ChaincodeEvents  `json:"chaincodeEvents"`
}
type Block struct {
	TransactionList []Transactions `json:"transactions"`
	StateHash  string `json:"stateHash"`
	PreviousBlockHash string `json:"previousBlockHash"`
	ConsensusMetadata string `json:"consensusMetadata"`
	NonHash NonHashData `json:"nonHashData"`
}

/*
  returns height of chain for a network peer.
	url(http//:IP:PORT) is the address of the peerRATN
*/
func Monitor_ChainHeight(url string) int {

	respBody, status := peerrest.GetChainInfo(url + "/chain")
	type ChainMsg struct {
		HT int `json:"height"`
		//curHash string `json:"currentBlockHash"`
		//prevHash string `json:"previousBlockHash"`
	}
	if verbose {
		fmt.Println("Monitor_ChainHeight() chain info status: ", status)
		fmt.Println("Monitor_ChainHeight() chain info respBody: ", respBody)
	}
	resCh := new(ChainMsg)
	err := json.Unmarshal([]byte(respBody), &resCh)
	if err != nil {
		fmt.Println("There was an error in unmarshalling chain info")
	}
	return resCh.HT
}

/*
  displays the chain information.
	url (http://IP:PORT) is the address of a network peer
*/
func ChainStats(url string) {
	body, status := peerrest.GetChainInfo(url + "/chain")
	fmt.Println("ChainStats() chain info status: ", status) 
	fmt.Println("ChainStats() chain info body: ", body)
	//return body, status
}

func GetChainStats(url string) (body, status string) {
        body, status = peerrest.GetChainInfo(url + "/chain")
        //fmt.Println("ChainStats() chain info status: ", status)
        //fmt.Println("ChainStats() chain info body: ", body)
        return body, status
}

func ChaincodeBlockHash(url string, block int) string {
	respBody, status := peerrest.GetChainInfo(url + "/chain/blocks/" + strconv.Itoa(block))
	fmt.Println("ChaincodeBlockHash() chain info status: ", status)
	if verbose { fmt.Println("ChaincodeBlockHash() chain info respBody: ", respBody) }
	blockStruct := new(Block)
	err := json.Unmarshal([]byte(respBody), &blockStruct)
	if err != nil {
		fmt.Println("There was an error in unmarshalling chain info", err)
	}
	return blockStruct.StateHash
}

func ChaincodeBlockTrxInfo(url string, block int) []Transactions {
	respBody, status := peerrest.GetChainInfo(url + "/chain/blocks/" + strconv.Itoa(block))
	fmt.Println("ChaincodeBlockTrxInfo() chain info status: ", status)
	if verbose { fmt.Println("ChaincodeBlockTrxInfo() chain info respBody: ", respBody) }
	blockStruct := new(Block)

	err := json.Unmarshal([]byte(respBody), &blockStruct)
	if err!= nil {
		fmt.Println(err)
	}
	//fmt.Println(blockStruct.TransactionList)
	return blockStruct.TransactionList
}

/*
 displays statistics for a specific block.
	url(ip:port) is the address of a peer on the network.
	block is an integer such that 0 < block <= chain height).
*/
func BlockStats(url string, block int) string {
	currBlock := strconv.Itoa(block - 1)
	var body, status string
	var prettyJSON bytes.Buffer
	const JSON_INDENT = "    " // four bytes of indentation
	body, status = peerrest.GetChainInfo(url + "/chain/blocks/" + currBlock)
	fmt.Println("BlockStats() GetChainInfo status: ", status)
	if verbose { fmt.Println("BlockStats() GetChainInfo body: ", body) }

	error := json.Indent(&prettyJSON, []byte(body), "", JSON_INDENT)
	if error != nil {
		fmt.Println("JSON parse error: ", error)
		return "JSON parse error "
	}
	return string(prettyJSON.Bytes())
}

/*
  
*/
func NetworkPeers(url string) (string, string) {
	var body, status string
	var prettyJSON bytes.Buffer
	const JSON_INDENT = "    " // four bytes of indentation
	body, status = peerrest.GetChainInfo(url + "/network/peers")
	fmt.Println("NetworkPeers() GetChainInfo status: ", status)
	if verbose { fmt.Println("NetworkPeers() GetChainInfo body: ", body) }

	error := json.Indent(&prettyJSON, []byte(body), "", JSON_INDENT)
	if error != nil {
		return fmt.Sprintln("JSON parse error: %s", error), status
	}

	if verbose { fmt.Println(string(prettyJSON.Bytes())) }
	return string(prettyJSON.Bytes()), status

}

/*
  displays if the given user has been already registed.
	url  (http://IP:PORT) is the address of network peer
*/
func UserRegister_Status(url string, username string) (responseBody string, status string){
	responseBody, status = peerrest.GetChainInfo(url + "/registrar/" + username)
	fmt.Println("  UserRegister_Status() chain info responseStatus = ", status)
	if verbose { fmt.Println("  UserRegister_Status() chain info responseBody = ", responseBody) }
	return responseBody, status
}

/*
  Gets the ecert for the given registered user
	url  (http://IP:PORT) is the address of network peer
*/
func UserRegister_ecertDetail(url string, username string) (response string, status string) {
	var body string
	body,status = peerrest.GetChainInfo(url + "/registrar/" + username + "/ecert")
	fmt.Println("  UserRegister_ecertDetail() chain info responseStatus = ", status)
	if verbose { fmt.Println("  UserRegister_ecertDetail() chain info responseBody = ", body) }
	return body, status
}

/*
  displays information about a given transaction.
	url  (http://IP:PORT) is the address of network peer
	txId is the transaction ID that is returned from Invoke and Deploy calls
*/
func Transaction_Detail(url string, txid string) {
	//currTxId := strconv.Atoi(txid)
	var body, status string
	var prettyJSON bytes.Buffer
	const JSON_INDENT = "    " // four bytes of indentation
	body, status = peerrest.GetChainInfo(url + "/transactions/" + txid)
	fmt.Println("Transaction_Detail() chain info status: ", status)
	if verbose { fmt.Println("Transaction_Detail() chain info body: ", body) }
	error := json.Indent(&prettyJSON, []byte(body), "", JSON_INDENT)
	if error != nil {
		fmt.Println("JSON parse error: ", error)
		return
	}
	fmt.Println(string(prettyJSON.Bytes()))
}

//
// Use POST /chaincode endpoint to deploy, invoke, and
// query a target chaincode.
func changeState(url string, path string, restCallName string,
	args []string, user string, funcName string) string {

	//  Build a payload for the REST API call
	depPL := make(chan []byte)
	go genPayLoadForChaincode(depPL, path, funcName, args, user, restCallName)
	depPayLoad := <-depPL

	//	Build a URL for the REST API call using the caller's url
	restUrl := url + "/chaincode/"
	if verbose {
		msgStr := fmt.Sprintf("**changeState() Sending Rest Request to url: %s", restUrl)
		fmt.Println(msgStr)
	}

	//  issue REST call
	respBody, respStatus := peerrest.PostChainAPI(restUrl, depPayLoad)

	// Parse the response
	res := new(restCallResult_T)
	err := json.Unmarshal([]byte(respBody), &res)
	if err != nil {
		errMsg := fmt.Sprintf("\nchangeState() ERROR in json.Unmarshal !!!!!\n")
		errMsg += fmt.Sprintf("  path:                      %s\n", path)
		errMsg += fmt.Sprintf("  restCallName:              %s\n", restCallName)
		errMsg += fmt.Sprintf("  args:                      %s\n", args)
		errMsg += fmt.Sprintf("  user:                      %s\n", user)
		errMsg += fmt.Sprintf("  funcName:                  %s\n", funcName)
		errMsg += fmt.Sprintf("  respStatus:                %s\n", respStatus)
		errMsg += fmt.Sprintf("  respBody:                  %s\n", respBody)
		errMsg += fmt.Sprintf("  Sent Rest Request to url:  %s\n", restUrl)
		errMsg += fmt.Sprintf("  json.Unmarshal result:     %s\n", res)
		errMsg += fmt.Sprintf("  json.Unmarshal error:      %s\n", err)
		fmt.Println(errMsg)
		DisplayNetworkDebugInfo()
		log.Fatal("changeState: ERROR in Unmarshalling! ", err)
	}
	if verbose { 
		if res.Result.Message != "" {
			fmt.Println("POST /chaincode returned Result message, extracted from json: ", res.Result.Message)
		}
	}
	if res.Error.Message != "" {
		if verbose { fmt.Println("POST /chaincode returned Error message, extracted from json: ", res.Error.Message) }
		fmt.Printf("POST /chaincode returned code =%v message=%v data=%v\n",
			res.Error.Code, res.Error.Message, res.Error.Data)
		return ""
	}
	if res.Result.Message == "" { /* neither error nor result was returned */
		printJSON(respBody)
		panic("POST /chaincode returned unexpected output: No Result Message AND No Error Message!")
	}
	return res.Result.Message
}

func readState(url string, path string, restCallName string, args []string, user string, funcName string) string {
	msgStr := fmt.Sprintf("entering readState: path=%s, restCallName=%s, user=%s, funcName=%s, args=%v\n", path, restCallName, user, funcName, args)
	depPL := make(chan []byte)
	go genPayLoadForChaincode(depPL, path, funcName, args, user, restCallName)
	depPayLoad := <-depPL

	restUrl := url + "/chaincode/"
	msgStr += fmt.Sprintf("**readState() Sending Rest Request to url: %s", restUrl)
	if verbose { fmt.Println(msgStr) }

	respBody, respStatus := peerrest.PostChainAPI(restUrl, depPayLoad)

	//	commented for less output messages
	//	fmt.Println("Response from readState() REST call peerrest.PostChainAPI: >>>")
	//	printJSON(respBody)

	res := new(restCallResult_T)
	err := json.Unmarshal([]byte(respBody), &res)
	if err != nil {
		errMsg := fmt.Sprintf("\nreadState() ERROR in json.Unmarshal !!!!!\n")
		errMsg += fmt.Sprintf("  path:                      %s\n", path)
		errMsg += fmt.Sprintf("  restCallName:              %s\n", restCallName)
		errMsg += fmt.Sprintf("  args:                      %s\n", args)
		errMsg += fmt.Sprintf("  user:                      %s\n", user)
		errMsg += fmt.Sprintf("  funcName:                  %s\n", funcName)
		errMsg += fmt.Sprintf("  respStatus:                %s\n", respStatus)
		errMsg += fmt.Sprintf("  respBody:                  %s\n", respBody)
		errMsg += fmt.Sprintf("  Sent Rest Request to url:  %s\n", restUrl)
		errMsg += fmt.Sprintf("  json.Unmarshal result:     %s\n", res)
		errMsg += fmt.Sprintf("  json.Unmarshal error:      %s\n", err)
		fmt.Println(errMsg)
		log.Fatal("readState: ERROR in Unmarshalling! ", err)
	}
	if verbose { fmt.Println("res = ", *res) }
	if res.Error.Message != "" {
		if verbose { fmt.Println("Error extracted from json: res.Error.Message", res.Error.Message) }
		fmt.Printf("POST /chaincode returned code =%v message=%v data=%v\n",
			res.Error.Code, res.Error.Message, res.Error.Data)
		return ""
	}
	if res.Result.Message == "" { /* neither error nor result was returned */
		printJSON(respBody)
		panic("POST /chaincode returned unexpected output")
	}
	return res.Result.Message
}

func genPayLoad(PL chan []byte, pathName string, funcName string, args []string, user string, restCallName string) {

	//formatting args to fit needs of payload
	var argsReady string
	var payLoadString string
	buffer := bytes.NewBufferString("")
	for i := 0; i < len(args); i++ {
		myArgs := args[i]
		buffer.WriteString("\"")
		buffer.WriteString(myArgs)
		buffer.WriteString("\"")
		//omit , for the last arg
		if i != (len(args) - 1) {
			buffer.WriteString(",")
		}
	}

	argsReady = buffer.String()

	switch restCallName {
	case "deploy":
		payLoadString = S1 + "\"path\":\"" + pathName + S2 + funcName + S3 + argsReady + S4 + user + S5
		//payLoadString = S1 + "\"path\":\"" + pathName + S2 + "init" + S3 + argsReady + S4NOSEC
		//fmt.Println("\ndeploy PayLoad \n", payLoadString)
	case "invoke":
		payLoadString = IQSTART + S1 + "\"name\":\"" + pathName + S2 + funcName + S3 + argsReady + S4 + user + S5 + IQEND
		//payLoadString = IQSTART + S1 + "\"name\":\"" + pathName + S2 + funcName + S3 + argsReady + S4NOSEC
		//fmt.Println("\nInvoke PayLoad \n", payLoadString)
	case "query":
		payLoadString = IQSTART + S1 + "\"name\":\"" + pathName + S2 + funcName + S3 + argsReady + S4 + user + S5 + IQEND
		//payLoadString = IQSTART + S1 + "\"name\":\"" + pathName + S2 + funcName + S3 + argsReady + S4NOSEC
		//fmt.Println("\nQuery PayLoad \n", payLoadString)
	}
	payLoadInBytes := []byte(payLoadString)
	PL <- payLoadInBytes
} /* genPayLoad() */

// Build the payload for the POST /chaincode APIs deploy, invoke, and query
// Payload formats defined here:
// https://github.com/hyperledger/fabric/blob/master/docs/API/CoreAPI.md#chaincode
func genPayLoadForChaincode(PL chan []byte, pathName string, funcName string,
	dargs []string, user string, restCallName string) {

	// Structure Chaincode_T for chaincodeID member of payload
	// "chaincodeID" : {"path":"<pathname>"}
	//  	or
	//	"chaincodeID" : {"name": "<chaincode name>"}
	type ChaincodeID_T struct {
		Path string `json:"path,omitempty"`
		Name string `json:"name,omitempty"`
	}

	type CTORMSG_T struct {
		Function string   `json:"function"`
		Args     []string `json:"args"`
	}

	type Parameters_T struct {
		Itype         int           `json:"type"`
		ChaincodeID   ChaincodeID_T `json:"chaincodeID"`
		Ctormsg       CTORMSG_T     `json:"ctorMsg"`
		SecureContext string        `json:"secureContext"`
	}

	type payLoad_T struct {
		Jsonrpc string       `json:"jsonrpc"` //constant
		Method  string       `json:"method"`  //pass  variable
		Params  Parameters_T `json:"params"`
		ID      int64        `json:"id"` // correlation ID
	}

	var PN ChaincodeID_T // Allocate PN to build chaincodeID part

	//
	// chaincodeID member content of our payload differs according to the
	// rest call, so populate it correctly
	//

	//	restCallName = "bogus"
	if strings.Contains(restCallName, "deploy") {
		PN = ChaincodeID_T{Path: pathName}
	} else {
		if strings.Contains(restCallName, "invoke") ||
			strings.Contains(restCallName, "query") {
			PN = ChaincodeID_T{Name: pathName}
		} else {
			logMsg := fmt.Sprintf("Rest call=%s is not supported", restCallName)
			log.Fatal(logMsg)
		}
	}
	// build a unique ID for this REST API call
	PostChaincodeCount++ // number of POST /chaincode calls

	// Allocate 'payLoad' structure and populate it with content we want
	// in our payload json
	payLoadInstance := &payLoad_T{
		Jsonrpc: "2.0",
		Method:  restCallName,
		Params: Parameters_T{
			Itype:       1,
			ChaincodeID: PN,
			Ctormsg: CTORMSG_T{
				Function: funcName,
				Args:     dargs,
			},
			SecureContext: user,
		},
		ID: PostChaincodeCount,
	} /* payLoad */

	payLoadInBytes, err := json.Marshal(payLoadInstance)
	if err != nil {
		log.Fatal("genPayloadForChaincode: error marshalling JSON ", err)
	}

	//printJSON(string(payLoadInBytes))

	//	payLoadInBytes := []byte(payLoadString)
	PL <- payLoadInBytes
} /* genPayLoadforChaincode() */

func register(url string, user string, secret string) string {
	payLoad := make(chan []byte)
	// fmt.Println("register() url,user,secret:", url, user, secret)
	go genRegPayLoad(payLoad, user, secret)
	regPayLoad := <-payLoad
	regUrl := url + "/registrar"
	msgStr := fmt.Sprintf("**register() Sending Rest Request to url %s user=%s secret=%s", regUrl, user, secret)
	fmt.Println(msgStr)
	respBody, status := peerrest.PostChainAPI(regUrl, regPayLoad)
	fmt.Println(respBody)
	return status
}

func genRegPayLoad(payLoad chan []byte, user string, secret string) {
	//fmt.Println("\nRegistering user : ", user + " with secret :", secret)
	registerJsonPayLoad := RegisterJsonPart1 + user + RegisterJsonPart2 + secret + RegisterJsonPart3
	regPayLoadInBytes := []byte(registerJsonPayLoad)
	if verbose { fmt.Println("Register PayLoad \n", registerJsonPayLoad) }
	payLoad <- regPayLoadInBytes
}

func printJSON(inbuf string) {
	var formattedJSON bytes.Buffer // Allocate buffer for formatted JSON
	const JSON_INDENT = "    "     // four bytes of indentation

	error := json.Indent(&formattedJSON, []byte(inbuf), "", JSON_INDENT)
	if error != nil {
		fmt.Println("printJSON: inbuf = \n", inbuf)
		fmt.Println("printJSON: parse error: ", error)
		return
	}

	fmt.Println(string(formattedJSON.Bytes()))
} /* printJSON() */
