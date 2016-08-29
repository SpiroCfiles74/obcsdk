package chaincode

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"obcsdk/peernetwork"
	"os"
	//"obcsdk/util"
	"obcsdk/threadutil"
)

var ThisNetwork peernetwork.PeerNetwork
var Peers = ThisNetwork.Peers
var ChainCodeDetails, Versions map[string]string
var LibCC peernetwork.LibChainCodes

const invokeOnPeerUsage = ("iAPIArgs0 := []string{\"example02\", \"invoke\", \"<PEER_IP_ADDRESS>\" + \"(optional)<tagName>\"}" +
	"invArgs0 := []string{\"a\", \"b\", \"500\"} " +

	"chaincode.Invoke(iAPIArgs0, invArgs0)}")
const invokeAsUserUsage = ("iAPIArgs0 := []string{\"example02\", \"invoke\", \"<Registered_USER_NAME>\" + \"(optional)<tagName>\"}" +
	"invArgs0 := []string{\"a\", \"b\", \"500\"} " +

	"chaincode.Invoke(iAPIArgs0, invArgs0)}")

/**
  initializes users on network using data supplied in NetworkCredentials.json file
*/
func InitNetwork() peernetwork.PeerNetwork {

	ThisNetwork = peernetwork.LoadNetwork()
	return ThisNetwork
}

/**
   initializes chaincodes on network using information supplied in CC_Collections.json file
*/
func InitChainCodes() {
	LibCC = peernetwork.InitializeChainCodes()
}

/*
  initializes network based on files in directory utils
*/
func Init() {
	InitNetwork()
	InitChainCodes()
}

func GetURL(ip, port string) string {
    var url string
    if os.Getenv("NET_COMM_PROTOCOL") == "HTTPS" {
        url = "https://" + ip + ":" + port
    } else {
    	url = "http://" + ip + ":" + port
    }
    return url
}

/*
   Registers each user on the network based on the content of ThisNetwork.Peers.
*/
func RegisterUsers() {
	if verbose { fmt.Println("\nRegisterUsers: register all users in all peers in network") }

	//testuser := peernetwork.AUser(ThisNetwork)
	Peers := ThisNetwork.Peers
	i := 0
	for i < len(Peers) {

		userList := ThisNetwork.Peers[i].UserData
		for user, secret := range userList {
			url := GetURL(Peers[i].PeerDetails["ip"], Peers[i].PeerDetails["port"])
			if verbose {
				msgStr := fmt.Sprintf("\nRegistering %s with password %s on %s using %s", user, secret, Peers[i].PeerDetails["name"], url)
				fmt.Println(msgStr)
			}
			register(url, user, secret)
		}
		fmt.Println("\nRegisterUsers(): Done Registering ", len(userList), "users on ", Peers[i].PeerDetails["name"], "\n")
		i++
	}
}


func RegisterCustomUsers() {
	if verbose { fmt.Println("\nRegisterCustomUsers: register all users in all peers in network, plus custom users") }

	Peers := ThisNetwork.Peers

	for i := 0; i < len(Peers) ; i++ {

		extraUsers := 0
		userList := ThisNetwork.Peers[i].UserData
		for user, secret := range userList {
			url := GetURL(Peers[i].PeerDetails["ip"], Peers[i].PeerDetails["port"])
			msgStr := fmt.Sprintf("\nRegistering %s with password %s on %s using %s", user, secret, Peers[i].PeerDetails["name"], url)
			fmt.Println(msgStr)
			register(url, user, secret)
			if (i == len(Peers)-1) {
				if os.Getenv("NETWORK") == "Z" {
					// custom users in Z network
					for u := 0; u < threadutil.NumberCustomUsersOnLastPeer; u++ {
						user = threadutil.ZUsersOnLastPeer[u]
						secret = threadutil.ZUserPasswordsOnLastPeer[u]
						msgStr = fmt.Sprintf("\nZ NTWK: Registering custom user %s with password %s on %s using %s", user, secret, Peers[i].PeerDetails["name"], url)
						fmt.Println(msgStr)
						register(url, user, secret)
						extraUsers++
					}
				} else {
					// custom users in local network
					for u := 0; u < threadutil.NumberCustomUsersOnLastPeer; u++ {
						user = threadutil.LocalUsersOnLastPeer[u]
						secret = threadutil.LocalUserPasswordsOnLastPeer[u]
						msgStr = fmt.Sprintf("\nLOCAL NTWK: Registering custom user %s with password %s on %s using %s", user, secret, Peers[i].PeerDetails["name"], url)
						fmt.Println(msgStr)
						register(url, user, secret)
						extraUsers++
					}
				}
			}
		}
		fmt.Println("\nRegisterCustomUsers(): Done Registering ", len(userList)+extraUsers, "users on ", Peers[i].PeerDetails["name"], "\n")
	}
}

func RegisterUsers2() {
	if verbose { fmt.Println("\nCalling RegisterUsers2 ") }

	//testuser := peernetwork.AUser(ThisNetwork)
	Peers := ThisNetwork.Peers
	for i:= 0;i < len(Peers)-2;i++ {

		userList := ThisNetwork.Peers[i].UserData
		for user, secret := range userList {
			url := GetURL(Peers[i].PeerDetails["ip"], Peers[i].PeerDetails["port"])
			if verbose {
				msgStr := fmt.Sprintf("\nRegistering %s with password %s on %s using %s", user, secret, Peers[i].PeerDetails["name"], url)
				fmt.Println(msgStr)
			}
			register(url, user, secret)
		}
		fmt.Println("\nRegisterUsers2(): Done Registering ", len(userList), "users on ", Peers[i].PeerDetails["name"], "\n")
	}
}

/*
   deploys a chaincode in the fabric to later execute functions on this deployed chaincode
   Takes two arguments
 	 A. args []string
	   	1.ccName (string)			- name of the chaincode as specified in CC_Collections.json file
		2.funcName (string)			- name of the function to call from chaincode specification
									"init" for chaincodeexample02
		3.tagName(string)(optional)		- tag a deployment to support something like versioning

 	B. depargs []string				- actual arguments passed to initialize chaincode inside the fabric.

		Sample Code:
		dAPIArgs0 := []string{"example02", "init"}
		depArgs0 := []string{"a", "20000", "b", "9000"}

		var depRes string
		var err error
		depRes, err := chaincode.Deploy(dAPIArgs0, depArgs0)
*/
func Deploy(args []string, depargs []string) (id string, err error)  {

	if (len(args) < 2) || (len(args) > 3) {
		return " ", errors.New("FAILURE TO DEPLOY: Incorrect number of arguments. Expecting 2 or 3")
	}
	ccName := args[0]
	funcName := args[1]

	var tagName, txId string
	if len(args) == 2 {
		tagName = ""
	} else if len(args) == 3 {
		tagName = args[2]
	}
	dargs := depargs
	var err1 error
	ChainCodeDetails, Versions, err1 = peernetwork.GetCCDetailByName(ccName, LibCC)
	if err1 != nil {
		fmt.Println("chcoAPI.Deploy() FAILURE TO DEPLOY: Inside chcoAPI.Deploy(), found error: ", err1)
		//log.Fatal("No Chain Code Details, we cannot proceed")
		return " ", errors.New("FAILURE TO DEPLOY: No Chain Code Details; we cannot proceed")
	}
	if strings.Contains(ChainCodeDetails["deployed"], "true") {
		fmt.Println("\nchcoAPI.Deploy()  ** Already deployed ... skipping deploy...")
	} else {
		//msgStr := fmt.Sprintf("** Initializing and deploying chaincode %s on network with args %s", ChainCodeDetails["path"], dargs)
		//fmt.Println(msgStr)
		restCallName := "deploy"
		peer, auser := peernetwork.AUserFromNetwork(ThisNetwork)
	if verbose { fmt.Println("Value in deploying peer.State (0=RUNNING): ", peer.State) }
		url := GetURL(peer.PeerDetails["ip"], peer.PeerDetails["port"])
		if verbose {
			msgStr := fmt.Sprintf("chcoAPI.Deploy() ** Initializing and deploying chaincode %s on network with args %s", ChainCodeDetails["path"], dargs)
			fmt.Println(msgStr)
			fmt.Println("chcoAPI.Deploy() Value in the deploying peer.State (0=RUNNING): ", peer.State, " user=", auser)
			fmt.Println("chcoAPI.Deploy() url=", url)
			fmt.Println("chcoAPI.Deploy() restCallname=", url, " funcName=", funcName)
		}
		txId = changeState(url, ChainCodeDetails["path"], restCallName, dargs, auser, funcName)
        //if verbose { fmt.Println("chcoAPI.Deploy() txID", txId) }
		//storing the value of most recently deployed chaincode inside chaincode details if no tagname or versioning
		ChainCodeDetails["dep_txid"] = txId
		if len(tagName) != 0 {
		  Versions[tagName] = txId
		}
		//fmt.Println("ChainCodeDetails dep_txid = " + ChainCodeDetails["dep_txid"])
	}

	return txId, errors.New("")

}

/*
   deploys a chaincode in the fabric to later execute functions on this deployed chaincode
   Takes two arguments
 	 A. args []string
	   	1.ccName (string)			- name of the chaincode as specified in CC_Collections.json file
		2.funcName (string)			- name of the function to call from chaincode specification
									"init" for chaincodeexample02
		3. host (string)				- hostname or ipaddress to call invoke from
		4. tagName(string)(optional)			- tag the invocation to support something like versioning

 	B. depargs []string				- actual arguments passed to initialize chaincode inside the fabric.

		Sample Code:
		dAPIArgs0 := []string{"example02", "init", "PEER0"}
		p.s. "PEER0" is name of the peer from NetworkCredentials.json file
		depArgs0 := []string{"a", "20000", "b", "9000"}

		var depRes string
		var err error
		depRes, err := chaincode.Deploy(dAPIArgs0, depArgs0)
*/
func DeployOnPeer(args []string, depargs []string) (id string, err error)  {

	if (len(args) < 3) || (len(args) > 4) {
		fmt.Println("DeployOnPeer : Incorrect number of arguments. Expecting 3 or 4 in invokeAPI arguments")
		//fmt.Println(deployOnPeerUsage)
		return "", errors.New("DeployOnPeer : Incorrect number of arguments. Expecting 3 or 4 in function arguments")
	}
	ccName := args[0]
	funcName := args[1]
	host := args[2]

	var tagName, txId string
	if len(args) == 2 {
		tagName = ""
	} else if len(args) == 3 {
		tagName = args[2]
	}
	dargs := depargs
	var err1 error
	ChainCodeDetails, Versions, err1 = peernetwork.GetCCDetailByName(ccName, LibCC)
	if err1 != nil {
		fmt.Println("Inside deploy: ", err1)
		//log.Fatal("No Chain Code Details, we cannot proceed")
		return " ", errors.New("No Chain Code Details we cannot proceed")
	}
	if strings.Contains(ChainCodeDetails["deployed"], "true") {
		fmt.Println("\n\n ** Already deployed ..")
		fmt.Println(" skipping deploy...")
	} else {
		//msgStr := fmt.Sprintf("\n** Initializing and deploying chaincode %s on network with args %s\n", ChainCodeDetails["path"], dargs)
		//fmt.Println(msgStr)
		restCallName := "deploy"
		ip, port, auser, err2 := peernetwork.AUserFromThisPeer(ThisNetwork, host)
		if err2 != nil {
			fmt.Println("Inside invoke3: ", err2)
			return "", err2
		} else {

                      //fmt.Println("Value in State : ", peer.State)
                      //fmt.Println("Value in State : ", peer.PeerDetails["state"])
                      url := GetURL(ip, port)
                      txId = changeState(url, ChainCodeDetails["path"], restCallName, dargs, auser, funcName)
                      //storing the value of most recently deployed chaincode inside chaincode details if no tagname or versioning
                      ChainCodeDetails["dep_txid"] = txId
                      if len(tagName) != 0 {
                        Versions[tagName] = txId
                     }
		}
     }
     return txId, nil    /* nil or errors.New("") */
}

/*
 changes state of a chaincode by passing arguments to BlockChain REST API invoke.
 Takes two arguments
 	 A. args []string
	    1.ccName (string)			- name of the chaincode as specified in CC_Collections.json file
		2.funcName (string)		- name of the function to call from chaincode specification
								"invoke" for chaincodeexample02
		3.tagName(string)(optional)	- tag a deployment to support something like versioning

	B. invargs []string			- actual arguments passed to change the state of chaincode inside the fabric.

		Sample Code:
		iAPIArgs0 := []string{"example02", "invoke"}
		invArgs0 := []string{"a", "b", "500"}

		var invRes string
		var err error
		invRes,err := chaincode.Invoke(iAPIArgs0, invArgs0)}
*/
func Invoke(args []string, invokeargs []string) (id string, err error) {

	if (len(args) < 2) || (len(args) > 3) {
		fmt.Println("Invoke : Incorrect number of arguments. Expecting 2")
		return "", errors.New("Invoke : Incorrect number of arguments. Expecting 2")
	}
	ccName := args[0]
	funcName := args[1]
	var tagName string
	if len(args) == 2 {
		tagName = ""
	} else if len(args) == 3 {
		tagName = args[2]
	}
	invargs := invokeargs
	//fmt.Println("Inside invoke .....")
	var err1 error
	ChainCodeDetails, Versions, err1 = peernetwork.GetCCDetailByName(ccName, LibCC)
	if err1 != nil {
		fmt.Println("Inside invoke: ", err1)
		log.Fatal("No Chain Code Details we cannot proceed")
		return "", errors.New("No Chain Code Details we cannot proceed")
	}
	restCallName := "invoke"
	aPeer, _ := peernetwork.APeer(ThisNetwork)
	if verbose {
		fmt.Println("Getting AUserFromAPeer at ip,port:", aPeer.PeerDetails["ip"], aPeer.PeerDetails["port"])
	}
	ip, port, auser, _ := peernetwork.AUserFromAPeer(*aPeer)
	url := GetURL(ip, port)
	if verbose {
		msgStr0 := fmt.Sprintf("** Calling %s on chaincode %s with args %s on  %s as %s", funcName, ccName, invargs, url, auser)
		fmt.Println(msgStr0)
	}
	var txId string
	if len(tagName) != 0 {
		txId = changeState(url, Versions[tagName], restCallName, invargs, auser, funcName)
	} else {
		txId = changeState(url, (ChainCodeDetails["dep_txid"]), restCallName, invargs, auser, funcName)
	}
	//fmt.Println("*** END Invoking as  ***", auser, " on a single peer")
	return txId, nil  /* or errors.New("") */
}

/*
 changes state of a chaincode on a specific peer by passing arguments to REST API call
 Takes two arguments
	A. args []string
	   	1. ccName (string)				- name of the chaincode as specified in CC_Collections.json file
		2. funcName(string)				- name of the function to call from chaincode specification
										"invoke" for chaincodeexample02
		3. host (string)				- hostname or ipaddress to call invoke from
		4. tagName(string)(optional)			- tag the invocation to support something like versioning

	B. invargs []string					- actual arguments passed to change the state of chaincode inside the fabric.

		Sample Code:
		iAPIArgs0 := []string{"example02", "invoke", "127.0.0.1"}
		invArgs0 := []string{"a", "b", "500"}

		var invRes string
		var err error
		invRes,err := chaincode.Invoke(iAPIArgs0, invArgs0)}
*/
func InvokeOnPeer(args []string, invokeargs []string) (id string, err error) {

	//fmt.Println("Inside InvokeOnPeer .....")
	if (len(args) < 3) || (len(args) > 4) {
		fmt.Println("InvokeOnPeer : Incorrect number of arguments. Expecting 3 or 4 in invokeAPI arguments")
		fmt.Println(invokeOnPeerUsage)
		return "", errors.New("InvokeOPeer : Incorrect number of arguments. Expecting 3 or 4 in function arguments")
	}
	ccName := args[0]
	funcName := args[1]
	host := args[2]
	var tagName string
	if len(args) == 3 {
		tagName = ""
	} else if len(args) == 4 {
		tagName = args[3]
	}
	invargs := invokeargs
	restCallName := "invoke"
	var err1 error
	var txId string
	ChainCodeDetails, Versions, err1 = peernetwork.GetCCDetailByName(ccName, LibCC)
	if err1 != nil {
		fmt.Println("Inside InvokeOnPeer: ", err1)
		log.Fatal("No Chain Code Details we cannot proceed")
		return "", errors.New("No Chain Code Details we cannot proceed")
	}

	ip, port, auser, err2 := peernetwork.AUserFromThisPeer(ThisNetwork, host)
	if err2 != nil {
		fmt.Println("Inside invoke3: ", err2)
		return "", err2
	} else {
		url := GetURL(ip, port)
		if verbose {
			msgStr0 := fmt.Sprintf("** Calling %s on chaincode %s with args %s on  %s as %s on %s", funcName, ccName, invargs, url, auser, host)
			fmt.Println(msgStr0)
		}
		if (len(tagName) > 0) {
			txId = changeState(url, Versions[tagName], restCallName, invargs, auser, funcName)
		}else {
		        txId = changeState(url, (ChainCodeDetails["dep_txid"]), restCallName, invargs, auser, funcName)
		}
		return txId, nil  /* or errors.New("") */
	}
}

/*
 changes state of a chaincode using a specific user credential
  Takes two arguments
 	A. args []string
	   	1. ccName (string)				- name of the chaincode as specified in CC_Collections.json file
		2. funcName(string)				- name of the function to call from chaincode specification
										"invoke" for chaincodeexample02
		3. user (string)				- login name of a registered user
		4. tagName(string)(optional)			- tag the invocation to support something like versioning

	B. invargs []string					- actual arguments passed to change the state of chaincode inside the fabric.

		Sample Code:
		iAPIArgs0 := []string{"example02", "invoke", "jim"}
		invArgs0 := []string{"a", "b", "500"}

		var invRes string
		var err error
		invRes,err := chaincode.Invoke(iAPIArgs0, invArgs0)}
*/
func InvokeAsUser(args []string, invokeargs []string) (id string, err error) {
	if (len(args) < 3) || (len(args) > 4) {
		fmt.Println("InvokeAsUser : Incorrect number of arguments. Expecting 3 or 4 in invokeAPI arguments")
		fmt.Println(invokeAsUserUsage)
		return "", errors.New("InvokeAsUser : Incorrect number of arguments. Expecting 3 or 4 number in invokeAPI arguments")
	}
	ccName := args[0]
	funcName := args[1]
	userName := args[2]
	var tagName string
	if len(args) == 3 {
		tagName = ""
	} else if len(args) == 4 {
		tagName = args[3]
	}
	invargs := invokeargs
	var err1 error
	ChainCodeDetails, Versions, err1 = peernetwork.GetCCDetailByName(ccName, LibCC)
	if err1 != nil {
		fmt.Println("Inside InvokeAsUser err1: ", err1)
		log.Fatal("No Chain Code Details we cannot proceed")
		return "", errors.New("No Chain Code Details we cannot proceed")
	}
	restCallName := "invoke"
	ip, port, auser, err2 := peernetwork.PeerOfThisUser(ThisNetwork, userName)
	if err2 != nil {
		fmt.Println("inside InvokeAsUser err2: ", err2)
		//return "", err2
		log.Fatal("Cannot cannot find PeerOfThisUser " + userName + ", ccName=" + ccName + ", funcName=" + funcName)
		return "", errors.New("Cannot cannot find PeerOfThisUser " + userName + ", ccName=" + ccName + ", funcName=" + funcName)
	} else {
		url := GetURL(ip, port)
		if verbose {
			msgStr0 := fmt.Sprintf("InvokeAsUser: ** Calling %s on chaincode %s with args %s on  %s as %s", funcName, ccName, invargs, url, auser)
			fmt.Println(msgStr0)
		}
		txId := changeState(url, Versions[tagName], restCallName, invargs, auser, funcName)
		return txId, errors.New("")
	}
}

/*
  Query fetches the value of the arguments supplied to query function from the fabric.
  Takes two arguments
 	A. args []string
	   	1. ccName (string)				- name of the chaincode as specified in CC_Collections.json file
		2. funcName(string)				- name of the function to call from chaincode specification
										"query" for chaincodeexample02
		3. tagName(string)(optional)	- tag the invocation to support something like versioning

	B. qargs []string					- actual arguments passed to get the values as stored inside fabric.

		Sample Code:
		qAPIArgs0 := []string{"example02", "query"}
		qArgsa := []string{"a"}

		var queryRes string
		var err error
		queryRes,err := chaincode.Query(qAPIArgs0, qArgsa)
*/
func Query(args []string, queryArgs []string) (id string, err error) {

	if (len(args) < 2) || (len(args) > 3) {
		return "", errors.New("Incorrect number of arguments. Expecting 2")
	}
	ccName := args[0]
	funcName := args[1]
	var tagName string
	if len(args) == 2 {
		tagName = ""
	} else if len(args) == 3 {
		tagName = args[2]
	}
	qargs := queryArgs
	var err1 error

	ChainCodeDetails, Versions, err1 = peernetwork.GetCCDetailByName(ccName, LibCC)
	if err1 != nil {
		fmt.Println("Inside Query: ", err1)
		fmt.Println("No Chain Code Details we cannot proceed")
		return "", errors.New("No Chain Code Details we cannot proceed")
	}
	restCallName := "query"
	peer, auser := peernetwork.AUserFromNetwork(ThisNetwork)
	url := GetURL(peer.PeerDetails["ip"], peer.PeerDetails["port"])

	var txId string
	if verbose {
		msgStr0 := fmt.Sprintf("** Calling %s on chaincode %s with args %s on  %s as %s", funcName, ccName, queryArgs, url, auser)
		fmt.Println(msgStr0)
	}

	if len(tagName) != 0 {
		txId = readState(url, Versions[tagName], restCallName, qargs, auser, funcName)
	} else {
		txId = readState(url, (ChainCodeDetails["dep_txid"]), restCallName, qargs, auser, funcName)
	}

	return txId, nil  /* errors.New("") */
}


/*
/*
  Query fetches the value of the arguments supplied to query function from the fabric.
  Takes two arguments
 	A. args []string
	  1. ccName (string)				- name of the chaincode as specified in CC_Collections.json file
		2. funcName(string)				- name of the function to call from chaincode specification
		3. host (string)				- hostname or ipaddress to call query
		4. tagName(string)(optional)	- tag the invocation to support something like versioning

	B. qargs []string					- actual arguments passed to get the values as stored inside fabric.

		Sample Code:
		qAPIArgs0 := []string{"example02", "query", "vp2"}
		qArgsa := []string{"a"}

		var queryRes string
		var err error
		queryRes,err := chaincode.Query(qAPIArgs0, qArgsa)
*/



func QueryOnHost(args []string, queryargs []string) (id string, err error) {
	if (len(args) < 3) || (len(args) > 4) {
		fmt.Println("QueryOnHost : Incorrect number of arguments. Expecting 3 or 4 in invokeAPI arguments")
		fmt.Println(invokeOnPeerUsage)
		return "", errors.New("QueryOnHost : Incorrect number of arguments. Expecting 3 or 4 in function arguments")
	}
	ccName := args[0]
	funcName := args[1]
	host := args[2]
	var tagName string
	if len(args) == 3 {
		tagName = ""
	} else if len(args) == 4 {
		tagName = args[3]
	}
	qryargs := queryargs
	var err1 error
	var txId string
	ChainCodeDetails, Versions, err1 = peernetwork.GetCCDetailByName(ccName, LibCC)
	if err1 != nil {
		fmt.Println("Inside QueryOnHost: ", err1)
		log.Fatal("No Chain Code Details we cannot proceed")
		return "", errors.New("No Chain Code Details we cannot proceed")
	}
	restCallName := "query"
	ip, port, auser, err2 := peernetwork.AUserFromThisPeer(ThisNetwork, host)
	if err2 != nil {
		fmt.Println("Inside Query: ", err2)
		return "", err2
	} else {
		url := GetURL(ip, port)
		if verbose {
			msgStr0 := fmt.Sprintf("** Calling %s on chaincode %s with args %s on  %s as %s on %s", funcName, ccName, qryargs, url, auser, host)
			fmt.Println(msgStr0)
		}
		if (len(tagName) > 0) {
// why are we not using readState here???
			txId = changeState(url, Versions[tagName], restCallName, qryargs, auser, funcName)
		}else {
			txId = changeState(url, (ChainCodeDetails["dep_txid"]), restCallName, qryargs, auser, funcName)
		}
		return txId, nil  /*  errors.New("") */
	}

}

func GetChainHeight(host string) (ht int, err error) {

			//fmt.Println("Inside GetChainHeight chcoAPI.....")
			ip, port, _, err2 := peernetwork.AUserFromThisPeer(ThisNetwork, host)
			if err2 != nil {
				fmt.Println("Inside GetChainHeight: ", err2)
				return -1, err2
			} else {
				url := GetURL(ip, port)
				ht := Monitor_ChainHeight(url)
				return ht, errors.New("")
			}

}

func GetBlockTrxInfoByHost(host string, block int) (bsNonHash NonHashData, err error) {
	//respBody, status := peerrest.GetChainInfo(url + "/chain/blocks/" + strconv.Itoa(block))
	ip, port, _, err2 := peernetwork.AUserFromThisPeer(ThisNetwork, host)
	if err2 != nil {
		fmt.Println("Inside GetBlockTrxInfoByHost: ", err2)
		var emptyNonHashData NonHashData
		return emptyNonHashData, err2
	} else {
		url := GetURL(ip, port)
		bsNonHashData := ChaincodeBlockTrxInfo(url, block)
		fmt.Println(bsNonHashData)
		return bsNonHashData, errors.New("")
	}
}

/* *** WORKING UNDER CONSTRUCTIOn ***
   deploys a chaincode in the fabric to later execute functions on this deployed chaincode
   Takes two arguments
 	 A. args []string
	   	1.ccName (string)			- name of the chaincode as specified in CC_Collections.json file
		2.funcName (string)			- name of the function to call from chaincode specification
									"init" for chaincodeexample02
		3.ccPath		- deployment path

 	B. depargs []string				- actual arguments passed to initialize chaincode inside the fabric.

		Sample Code:
		dAPIArgs0 := []string{"example02", "init", "github.com/hyperledger/fabric/chaincode/go/chaincodeexample02"}
		depArgs0 := []string{"a", "20000", "b", "9000"}

		var depRes string
		var err error
		depRes, err := chaincode.Deploy(dAPIArgs0, depArgs0)
*/

/***************************************************************************
func DeployWithCCPATH(args []string, depargs []string) error {

	if (len(args) < 3) || (len(args) > 4) {
		return errors.New("Deploy : Incorrect number of arguments. Expecting 3")
	}
	ccName := args[0]
	funcName := args[1]
	ccPath := args[2]
	//var tagName string
	//if len(args) == 2 {
		//tagName = ""
	//} else if len(args) == 3 {
		//tagName = args[2]
	//}
	dargs := depargs
	var err error
	ChainCodeDetails, Versions, err = peernetwork.GetCCDetailByName(ccName, LibCC)
	if err != nil {
		fmt.Println("Inside deploy: ", err)

		if (ccPath == nil) {
			//log.Fatal("No Chain Code Details, we cannot proceed")
			 return "", errors.New("No Chain Code Details we cannot proceed")
		}

	}
	if strings.Contains(ChainCodeDetails["deployed"], "true") {
		fmt.Println("\n\n ** Already deployed ..")
		fmt.Println(" skipping deploy...")
	} else {
		msgStr := fmt.Sprintf("\n** Initializing and deploying chaincode %s on network with args %s\n", ccPath, dargs)
		fmt.Println(msgStr)
		restCallName := "deploy"
		peer, auser := peernetwork.AUserFromNetwork(ThisNetwork)
		url := "https://" + peer.PeerDetails["ip"] + ":" + peer.PeerDetails["port"]
		txId := changeState(url, ccPath, restCallName, dargs, auser, funcName)
		//storing the value of most recently deployed chaincode inside chaincode details if no tagname or versioning
		//add chaincode to library of chaincodes if does not exist
		AddCCToLibrary(ccPath, ccName)

    ChainCodeDetails, _, err = peernetwork.GetCCDetailByName(ccName, LibCC)
		if err != nil {
		ChainCodeDetails["dep_txid"] = txId
		return txId, errors.New("")
	}

}
**************************************************/
