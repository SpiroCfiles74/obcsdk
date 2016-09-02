package main

import (
	"fmt"
	"obcsdk/chaincode"
	"obcsdk/peernetwork"
	"os"
	"strconv"
	"strings"
	"time"
	"obcsdk/lstutil"
	"obcsdk/threadutil"
)

var f *os.File
var myNetwork peernetwork.PeerNetwork
var url string
var counter int64

func main() {
        counter = 0

	lstutil.TESTNAME = "BasicFuncExistingNetworkLST"
	lstutil.InitLogger(lstutil.TESTNAME)
	lstutil.Logger("\n\n*********** " + lstutil.TESTNAME + " ***************")

	defer timeTrack(time.Now(), "Total execution time for " + lstutil.TESTNAME)

	setupNetwork()

	//get a URL details to get info n chainstats/transactions/blocks etc.
	aPeer, _ := peernetwork.APeer(myNetwork)
	url = "http://" + aPeer.PeerDetails["ip"] + ":" + aPeer.PeerDetails["port"]

	userRegisterTest(url)

	response, status := chaincode.NetworkPeers(url)
	if strings.Contains(status, "200") {
		lstutil.Logger("NetworkPeers Rest API Test Pass: successful")
		lstutil.Logger(response)
	}

	chaincode.ChainStats(url)

	lstutil.DeployChaincode()

	queryAllHosts("DEPLOY", counter)

	invRes := lstutil.InvokeChaincode()
	time.Sleep(30000 * time.Millisecond)

	queryAllHosts("INVOKE", counter)

	getHeight()

	time.Sleep(30000 * time.Millisecond)

	lstutil.Logger("\nBlockchain: Get Chain  ....")
	chaincode.ChainStats(url)

	lstutil.Logger("\nBlockchain: GET Chain  ....")
	response2 := chaincode.Monitor_ChainHeight(url)

	lstutil.Logger(fmt.Sprintf("\nChain Height", chaincode.Monitor_ChainHeight(url)))

	lstutil.Logger("\nBlock: GET/Chain/Blocks/")
	//chaincode.BlockStats(url, response2)
	nonHashData, _ := chaincode.GetBlockTrxInfoByHost(threadutil.GetPeer(0), response2-1)


	if strings.Contains(nonHashData.TransactionResult[0].Uuid, invRes) {
		lstutil.Logger(fmt.Sprintf("\n\nGetBlocks API Test PASS: Transaction %s Successfully stored in Block ", invRes))
		lstutil.Logger(fmt.Sprintf("\n=============Block:%d UUID: %s \n", response2-1, nonHashData.TransactionResult[0].Uuid))
	} else {
		lstutil.Logger(fmt.Sprintf("GetBlocks API Test FAIL!!!  Transaction %s NOT stored in Block ", invRes))
	}

  //This is for error condition
	//getBlockTxInfo(response2)

	lstutil.Logger("\nTransactions: GET/transactions/" + invRes)
	chaincode.Transaction_Detail(url, invRes)

	lstutil.Logger("\n\n*********** END BasicFuncExistingNetworkLST ***************\n\n")

}

func setupNetwork() {

        lstutil.Logger("========= setupNetwork =========")

	// lstutil.Logger("Setup a new network of peers (after killing old ones) using local_fabric script")
	// peernetwork.SetupLocalNetwork(4, false)

	lstutil.Logger("===== Get existing Network Credentials ===== ")
        peernetwork.GetNC_Local()

	lstutil.Logger("===== Connect to existing network - InitNetwork =====")
        myNetwork = chaincode.InitNetwork()

        lstutil.Logger("===== InitChainCodes =====")
        chaincode.InitChainCodes()

        lstutil.Logger("===== RegisterUsers =====")
        chaincode.RegisterUsers()

        //lstutil.Logger("===== RegisterCustomUsers =====")
        //chaincode.RegisterCustomUsers()


	time.Sleep(10000 * time.Millisecond)
	//peernetwork.PrintNetworkDetails(myNetwork)
	peernetwork.PrintNetworkDetails()
	numPeers := peernetwork.GetNumberOfPeers(myNetwork)

	lstutil.Logger(fmt.Sprintf("Network running successfully with %d peers with pbft and security+privacy enabled\n", numPeers))
}

func userRegisterTest(url string) {

	response, status := chaincode.UserRegister_Status(url, "test_user0")

	if (strings.Contains("200", status) && strings.Contains("test_user0", response)) {
		lstutil.Logger(fmt.Sprintf("RegisterUser API Test PASS: User %s Registration is successful", response))
	} else {
		lstutil.Logger(fmt.Sprintf("RegisterUser API Test FAIL: User %s Registration is NOT successful", response))
	}

	response, status = chaincode.UserRegister_Status(url, "nishi")
	if ((strings.Contains("200", status)) == false) {
		lstutil.Logger("RegisterUser API -Ve Test PASS: User Nishi Is Not Registered")
	} else {
		lstutil.Logger("RegisterUser API Test FAIL: User Nishi found in Register user list")
	}

	chaincode.UserRegister_ecertDetail(url, "lukas")
}

/*
func deploy() {							// using example02
	dAPIArgs0 := []string{"example02", "init"}
	depArgs0 := []string{"a", "100", "b", "900"}
	chaincode.Deploy(dAPIArgs0, depArgs0)
	time.Sleep(30000 * time.Millisecond) // minimum delay required, works fine in local environment
}

func invoke() string {						// using example02
	iAPIArgs0 := []string{"example02", "invoke"}
	invArgs0 := []string{"a", "b", "1"}
	invRes, _ := chaincode.Invoke(iAPIArgs0, invArgs0)
	return invRes
}
*/

func queryAllHosts(txName string, expected_count int64) {		// using ratnakar myCC, modified example02
	// loop through and query all hosts
	failedCount := 0
	N := peernetwork.GetNumberOfPeers(myNetwork)
	for peerNumber := 0 ; peerNumber < N ; peerNumber++ {
		result := "SUCCESS"
        	valueStr, counterIdxStr := lstutil.QueryChaincodeOnHost(peerNumber, expected_count)
        	newVal, err := strconv.ParseInt(counterIdxStr, 10, 64)
        	if err != nil { lstutil.Logger(fmt.Sprintf("Failed to convert %s to int64\n Error: %s\n", counterIdxStr, err)) }
        	if err != nil || newVal != expected_count {
			result = "FAILURE"
			failedCount++
		}
        	lstutil.Logger(fmt.Sprintf("QueryOnHost %d %s after %s: expected_count=%d, Actual a%s = %s", peerNumber, result, txName, expected_count, counterIdxStr, valueStr))
	}
	if failedCount > (N-1)/3 {
		lstutil.Logger(fmt.Sprintf("%s TEST FAILED!!!  TOO MANY PEERS (%d) FAILED to obtain the correct count, so network consensus failed!!!\n(If fewer than %d/%d peers are running, then the network does not have enough running nodes to reach consensus.)", txName, failedCount,  ((N-1)/3)*2+1, N ))
	} else {
		lstutil.Logger(fmt.Sprintf("%s TEST PASSED; enough peers (%d/%d) reached consensus on the correct count", txName, N-failedCount, N))
	}
}

/*
func query(txName string, expectedA int, expectedB int) {	// using example02
	qAPIArgs00 := []string{CHAINCODE_NAME, QUERY, threadutil.GetPeer(0)}
	qAPIArgs01 := []string{CHAINCODE_NAME, QUERY, threadutil.GetPeer(1)}
	qAPIArgs02 := []string{CHAINCODE_NAME, QUERY, threadutil.GetPeer(2)}
	qAPIArgs03 := []string{CHAINCODE_NAME, QUERY, threadutil.GetPeer(3)}
	qArgsa := []string{"a"}
	qArgsb := []string{"b"}

	res0A, _ := chaincode.QueryOnHost(qAPIArgs00, qArgsa)
	res0B, _ := chaincode.QueryOnHost(qAPIArgs00, qArgsb)
	res0AI, _ := strconv.Atoi(res0A)
	res0BI, _ := strconv.Atoi(res0B)

	res1A, _ := chaincode.QueryOnHost(qAPIArgs01, qArgsa)
	res1B, _ := chaincode.QueryOnHost(qAPIArgs01, qArgsb)
	res1AI, _ := strconv.Atoi(res1A)
	res1BI, _ := strconv.Atoi(res1B)

	res2A, _ := chaincode.QueryOnHost(qAPIArgs02, qArgsa)
	res2B, _ := chaincode.QueryOnHost(qAPIArgs02, qArgsb)
	res2AI, _ := strconv.Atoi(res2A)
	res2BI, _ := strconv.Atoi(res2B)

	res3A, _ := chaincode.QueryOnHost(qAPIArgs03, qArgsa)
	res3B, _ := chaincode.QueryOnHost(qAPIArgs03, qArgsb)
	res3AI, _ := strconv.Atoi(res3A)
	res3BI, _ := strconv.Atoi(res3B)

	lstutil.Logger("Results in a and b vp0 : ", res0AI, res0BI)
	lstutil.Logger("Results in a and b vp1 : ", res1AI, res1BI)
	lstutil.Logger("Results in a and b vp2 : ", res2AI, res2BI)
	lstutil.Logger("Results in a and b vp3 : ", res3AI, res3BI)

	if res0AI == expectedA && res1AI == expectedA && res2AI == expectedA && res3AI == expectedA {
		lstutil.Logger(fmt.Sprintf("\n\n%s TEST PASS : Results in A value match on all Peers after %s", txName, txName))
		lstutil.Logger(fmt.Sprintf("Values Verified : peer0: %d, peer1: %d, peer2: %d, peer3: %d", res0AI, res1AI, res2AI, res3AI))
	} else {
		lstutil.Logger(fmt.Sprintf("\n\n%s TEST FAIL: Results in A value DO NOT match on all Peers after %s", txName, txName))
	}

	if res0BI == expectedB && res1BI == expectedB && res2BI == expectedB && res3BI == expectedB {
		lstutil.Logger(fmt.Sprintf("\n\n%s TEST PASS : Results in B value match on all Peers after %s", txName, txName))
		lstutil.Logger(fmt.Sprintf("Values Verified : peer0: %d, peer1: %d, peer2: %d, peer3: %d\n\n", res0BI, res1BI, res2BI, res3BI))
	} else {
		lstutil.Logger(fmt.Sprintf("\n\n%s TEST FAIL: Results in B value DO NOT match on all Peers after %s", txName, txName))
	}
}
*/

func getHeight() {

	ht0, _ := chaincode.GetChainHeight(threadutil.GetPeer(0))
	ht1, _ := chaincode.GetChainHeight(threadutil.GetPeer(1))
	ht2, _ := chaincode.GetChainHeight(threadutil.GetPeer(2))
	ht3, _ := chaincode.GetChainHeight(threadutil.GetPeer(3))

	if (ht0 == 3) && (ht1 == 3) && (ht2 == 3) && (ht3 == 3) {
		lstutil.Logger(fmt.Sprintf("CHAIN HEIGHT TEST PASSED : Results in A value match on all Peers after "))
		lstutil.Logger(fmt.Sprintf("  Height Verified: ht0: %d, ht1: %d, ht2: %d, ht3: %d ", ht0, ht1, ht2, ht3))
	} else {
		lstutil.Logger(fmt.Sprintf("All heights do NOT match : ht0: %d, ht1: %d, ht2: %d, ht3: %d ", ht0, ht1, ht2, ht3))
		lstutil.Logger(fmt.Sprintf("CHAIN HEIGHT TEST FAILED : value in chain height do not match on all Peers after deploy and single invoke"))
	}

}

func getBlockTxInfo(blockNumber int) {
	errTransactions := 0
	height, _ := chaincode.GetChainHeight(threadutil.GetPeer(0))

	lstutil.Logger(fmt.Sprintf("############### Total Blocks %d #", height))
	lstutil.Logger(fmt.Sprintf("\n\nTotal Blocks # %d\n", height))

	for i := 1; i < height; i++ {
		//fmt.Printf("\n============================== Current BLOCKS %d ==========================\n", i)
		nonHashData, _ := chaincode.GetBlockTrxInfoByHost(threadutil.GetPeer(0), i)
		length := len(nonHashData.TransactionResult)
		for j := 0; j < length; j++ {
			// Print Error info only when transatcion failed
			if nonHashData.TransactionResult[j].ErrorCode > 0 {
				myStr1 := fmt.Sprintln("\n=============Block[%d] UUID [%d] ErrorCode [%d] Error: %s\n", i, nonHashData.TransactionResult[j].Uuid, nonHashData.TransactionResult[j].ErrorCode, nonHashData.TransactionResult[j].Error)
				fmt.Println(myStr1)
				errTransactions++
			}
		}
	}

}

func timeTrack(start time.Time, name string) {
	elapsed := time.Since(start)
	lstutil.Logger(fmt.Sprintf("\n################# %s took %s \n", name, elapsed))
	lstutil.Logger("################# Execution Completed #################")
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}
