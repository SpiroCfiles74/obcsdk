package main

import (
	"fmt"
	"strconv"
	"time"

	"obcsdk/chaincode"
	"obcsdk/peernetwork"
)

func main() {

	var myNetwork peernetwork.PeerNetwork

	fmt.Println("Creating a local docker network")
	//peernetwork.SetupLocalNetwork(4, true)

	time.Sleep(10000 * time.Millisecond)

	myNetwork = chaincode.InitNetwork()
	chaincode.InitChainCodes()
	//chaincode.Init()
	chaincode.RegisterUsers()

	//peernetwork.PrintNetworkDetails(myNetwork)
	peernetwork.PrintNetworkDetails()

	//get a URL details to get info n chainstats/transactions/blocks etc.
	aPeer, _ := peernetwork.APeer(myNetwork)
	url := "http://" + aPeer.PeerDetails["ip"] + ":" + aPeer.PeerDetails["port"]

	chaincode.NetworkPeers(url)

	chaincode.ChainStats(url)

	fmt.Println("\nPOST/Chaincode: Deploying chaincode at the beginning ....")
	dAPIArgs0 := []string{"example02", "init"}
	depArgs0 := []string{"a", "100", "b", "900"}
	chaincode.Deploy(dAPIArgs0, depArgs0)


	var resa, resb string
	var inita, initb, curra, currb, resaI, resbI int
	inita = 100
	initb = 900
	curra = inita
	currb = initb

	time.Sleep(60000 * time.Millisecond);
	fmt.Println("\nPOST/Chaincode: Querying a and b after deploy >>>>>>>>>>> ")
	qAPIArgs0 := []string{"example02", "query"}
	qArgsa := []string{"a"}
	qArgsb := []string{"b"}
	A, _ := chaincode.Query(qAPIArgs0, qArgsa)
	B, _ := chaincode.Query(qAPIArgs0, qArgsb)
	myStr := fmt.Sprintf("\nA = %s B= %s", A,B)
	fmt.Println(myStr)


	fmt.Println("******************************")
	fmt.Println("STOPPING PEER1 and PEER2 .. To Test Consensus")
	fmt.Println("******************************")

	peersToStartStop := []string{"PEER1", "PEER2"}
	peernetwork.StopPeersLocal(myNetwork, peersToStartStop)

/********************************************
	j = 0
	for j < 1{
		iAPIArgs0 := []string{"example02", "invoke"}
		invArgs0 := []string{"a", "b", "1"}
		invRes, _ := chaincode.Invoke(iAPIArgs0, invArgs0)
		fmt.Println("\nFrom Invoke invRes ", invRes)
		curra = curra - 1
		currb = currb + 1

   	j++

	}
	**************************/
	time.Sleep(30000 * time.Millisecond)
	fmt.Println("STARTING PEER1, ... To Test Consensus STATE TRANSFER")
	peernetwork.StartPeerLocal(myNetwork, "PEER1")
	fmt.Println("Sleeping for 2 minutes for PEER1 to sync up - state transfer")


	qAPIArgs00 := []string{"example02", "query", "PEER0"}
	resa, _ = chaincode.QueryOnHost(qAPIArgs00, qArgsa)
	resb, _ = chaincode.QueryOnHost(qAPIArgs00, qArgsb)

	resaI, _ = strconv.Atoi(resa)
	resbI, _ = strconv.Atoi(resb)


	qAPIArgs01 := []string{"example02", "query", "PEER1"}
	resa1, _ := chaincode.QueryOnHost(qAPIArgs01, qArgsa)
	resb1, _ := chaincode.QueryOnHost(qAPIArgs01, qArgsb)

	resa1I, _ := strconv.Atoi(resa1)
	resb1I, _ := strconv.Atoi(resb1)

	if (curra == resaI ) && (currb == resbI ) {
		fmt.Println("Results in a and b after bringing up peers match in PEER0 :")
		valueStr := fmt.Sprintf(" curra : %d, currb : %d, resa : %d , resb : %d", curra, currb, resaI, resbI)
		fmt.Println(valueStr)
	} else {
		fmt.Println("******************************")
		fmt.Println("RESULTS DO NOT MATCH AS EXPECTED ON VP2")
		fmt.Println("******************************")
	}

	if (curra == resa1I ) && (currb == resb1I ) {
		fmt.Println("Results in a and b after bringing up peers match in PEER1 :")
		valueStr := fmt.Sprintf(" curra : %d, currb : %d, resa : %d , resb : %d", curra, currb, resa1I, resb1I)
		fmt.Println(valueStr)
	} else {
		fmt.Println("******************************")
		fmt.Println("RESULTS DO NOT MATCH AS EXPECTED ON VP2")
		fmt.Println("******************************")
	}

}
