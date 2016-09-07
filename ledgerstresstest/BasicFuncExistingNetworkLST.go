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
var subTestsFailures int

func main() {
	subTestsFailures = 0
	lstutil.TESTNAME = "BasicFuncExistingNetworkLST"
	lstutil.InitLogger(lstutil.TESTNAME)
	lstutil.Logger("\n\n*********** " + lstutil.TESTNAME + " ***************")

	defer timeTrack(time.Now(), "Total execution time for " + lstutil.TESTNAME)

	setupNetwork()

	lstutil.Logger("\n===== userRegisterTest =====")
	lstutil.Logger("userRegisterTest: FirstUser=" + peernetwork.FirstUser)
	user_ip, user_port, user_name, err := peernetwork.PeerOfThisUser(myNetwork, peernetwork.FirstUser)
	check(err)
	url = chaincode.GetURL(user_ip, user_port)
	userRegisterTest(url, user_name)

	lstutil.Logger("\n===== NetworkPeers Test =====")
	response, status := chaincode.NetworkPeers(url)
	myStr := "NetworkPeers Rest API TEST "
	if strings.Contains(status, "200") {
		myStr += "PASS. Successful "
	} else {
		subTestsFailures++
		myStr += "FAIL!!! Error "
	}
	myStr += fmt.Sprintf("NetworkPeers response body:\n%s\n", response)
	lstutil.Logger(myStr)

	lstutil.Logger("\n===== Get ChainStats Test =====")
	response, status = chaincode.GetChainStats(url)
        if strings.Contains(status, "200") {
                lstutil.Logger("ChainStats Rest API TEST PASS.")
        } else {
                subTestsFailures++
                lstutil.Logger("ChainStats Rest API TEST FAIL!!!")
        }

	counter = queryAllHostsToGetCurrentCounter(lstutil.TESTNAME)
        height := chaincode.Monitor_ChainHeight(url) // and save the height; it will be needed below for getHeight()

	lstutil.Logger("\n===== Deploy Test =====")
	counter = lstutil.DeployChaincode()  // includes sleep
	height++
	queryAllHosts("DEPLOY", counter)

	lstutil.Logger("\n===== Invoke Test =====")
	invRes := lstutil.InvokeChaincode()  // increments counter inside
	height++
	time.Sleep(10000 * time.Millisecond)
	queryAllHosts("INVOKE", counter)

	lstutil.Logger("\n===== GetChainHeight Test =====")
	getHeight(height)  // this gets height from all peers and validates all match


	lstutil.Logger("\n===== GetBlock Stats API Test =====")
	//chaincode.BlockStats(url, height)
	nonHashData, _ := chaincode.GetBlockTrxInfoByHost(threadutil.GetPeer(0), height-1)
	myStr = "\nGetBlocks API TEST "
	if strings.Contains(nonHashData.TransactionResult[0].Uuid, invRes) {
		myStr += fmt.Sprintf("PASS: Transaction Successfully stored in Block")
	} else {
		subTestsFailures++
		myStr += fmt.Sprintf("FAIL!!! Transaction NOT stored in Block")
	}
	myStr += fmt.Sprintf("\nCH_Block = %d, UUID = %s, InvokeTransactionResult = %s\n", height-1, nonHashData.TransactionResult[0].Uuid, invRes)
	lstutil.Logger(myStr)

		//This is for error condition
		//getBlockTxInfo(height)

	lstutil.Logger("\n===== Get Transaction_Detail Test =====")
	lstutil.Logger("  input url:  " + url)
	lstutil.Logger("  input invRes:  " + invRes)
	lstutil.Logger("  calling Transaction_Detail(url,invRes):  ")
	chaincode.Transaction_Detail(url, invRes)

	if subTestsFailures == 0 {
		myStr = "PASS"
	} else {
        	myStr = fmt.Sprintf("FAIL (failed %d sub-tests)", subTestsFailures)
	}
	lstutil.Logger(fmt.Sprintf("\n\n*********** END BasicFuncExistingNetworkLST OVERALL TEST RESULT = %s ***************\n\n", myStr))
}

func setupNetwork() {

        lstutil.Logger("========= setupNetwork =========")

	// lstutil.Logger("Setup a new network of peers (after killing old ones) using local_fabric script")
	// peernetwork.SetupLocalNetwork(4, false)

	// When running BasicFunc test on local network, the local_fabric shell script creates
	// networkcredentials file. When running this with existing network, create it yourself by
	// putting the service_credentials from the Z network into serv_creds_file and executing
	//	"./update_z.py -b -f <serv_creds_file>"
	// to put the networkcredentials file in automation/ folder.
	// Note: you can skip calling GetNC_Local here if you first ensure the networkcredentials
	// file has already been copied to ../util/NetworkCredentials.json

	lstutil.Logger("----- Get existing Network Credentials ----- ")
        peernetwork.GetNC_Local()  // cp ../automation/networkcredentials ../util/NetworkCredentials.json

	lstutil.Logger("----- Connect to existing network - InitNetwork -----")
        myNetwork = chaincode.InitNetwork()

        lstutil.Logger("----- InitChainCodes -----")
        chaincode.InitChainCodes()
	time.Sleep(50000 * time.Millisecond)

        lstutil.Logger("----- RegisterUsers -----")
        if !chaincode.RegisterUsers() {
		lstutil.Logger("\nERROR: FAILED TO REGISTER one or more users in NetworkCredentials.json file\n")
		subTestsFailures++
        }

        //lstutil.Logger("----- RegisterCustomUsers -----")
        //if !chaincode.RegisterCustomUsers() {
	//	lstutil.Logger("\nERROR: FAILED TO REGISTER one or more CUSTOM users\n")
	//	subTestsFailures++
        //}
		

	time.Sleep(10000 * time.Millisecond)
	//peernetwork.PrintNetworkDetails(myNetwork)
	peernetwork.PrintNetworkDetails()
	numPeers := peernetwork.GetNumberOfPeers(myNetwork)

	if subTestsFailures == 0 {
		lstutil.Logger(fmt.Sprintf("Successfully connected to network with %d peers with pbft and security+privacy enabled\n", numPeers))
	}
}

// arg = a username that was already registered; this func confirms if it was successful
// and confirms user "ghostuserdoesnotexist" is not registered
// and confirms 
func userRegisterTest(url string, username string) {

	lstutil.Logger("\n----- RegisterUser Test -----")
	response, status := chaincode.UserRegister_Status(url, username)
	myStr := "RegisterUser API TEST "
	if strings.Contains(status, "200") && strings.Contains(response, username + " is already logged in") {
		myStr += fmt.Sprintf ("PASS: %s User Registration was already done successfully", username)
	} else {
		subTestsFailures++
		myStr += fmt.Sprintf ("FAIL!!! %s User Registration was NOT already done\n status = %s\n response = %s", username, status, response)
	}
	lstutil.Logger(myStr)
	time.Sleep(1000 * time.Millisecond)

	lstutil.Logger("\n----- RegisterUser Negative Test -----")
	response, status = chaincode.UserRegister_Status(url, "ghostuserdoesnotexist")
	if ((strings.Contains(status, "200")) == false) {
		lstutil.Logger("RegisterUser API Negative TEST PASS: CONFIRMED that user <ghostuserdoesnotexist> is unregistered as expected")
	} else {
		subTestsFailures++
		lstutil.Logger(fmt.Sprintf("RegisterUser API Negative TEST FAIL!!! User <ghostuserdoesnotexist> was found in Registrar User List but it was never registered!\n status = %s\n response = %s\n", status, response))
	}
	time.Sleep(1000 * time.Millisecond)

 /*
	lstutil.Logger("\n----- UserRegister_ecert Test -----")
	ecertUser := "lukas"
	response, status = chaincode.UserRegister_ecertDetail(url, ecertUser)
	myEcertStr := "\nUserRegister_ecert TEST "
	if strings.Contains(status, "200") && strings.Contains(response, ecertUser + " is already logged in") {
		myEcertStr += fmt.Sprintf ("PASS: %s ecert User Registration was already done successfully", ecertUser)
	} else {
		subTestsFailures++
		myEcertStr += fmt.Sprintf ("FAIL!!! %s ecert User Registration was NOT already done\n status = %s\n response = %s\n", username, status, response)
	}
	lstutil.Logger(myEcertStr)
	time.Sleep(1000 * time.Millisecond)
 */

	lstutil.Logger("\n----- UserRegister_ecert Negative Test -----")
	response, status = chaincode.UserRegister_ecertDetail(url, "ghostuserdoesnotexist")
	if ((strings.Contains(status, "200")) == false) {
		lstutil.Logger("UserRegister_ecert API Negative TEST PASS: CONFIRMED that user <ghostuserdoesnotexist> is unregistered as expected")
	} else {
		subTestsFailures++
		lstutil.Logger(fmt.Sprintf("UserRegister_ecert API Negative TEST FAIL!!! User <ghostuserdoesnotexist> was found in Registrar User List but it was never registered!\n status = %s\n response = %s\n", status, response))
	}
	time.Sleep(5000 * time.Millisecond)
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

func queryAllHostsToGetCurrentCounter(txName string) (counter int64) {		// using ratnakar myCC, modified example02
	// loop through and query all hosts to find consensus and determine what the current counter value is.
	counter = 0
	failedCount := 0
	N := peernetwork.GetNumberOfPeers(myNetwork)
	F := (N-1)/3
	currValues := make([]int64, N)
	for peerNumber := 0 ; peerNumber < N ; peerNumber++ {
        	_, counterIdxStr := lstutil.QueryChaincodeOnHost(peerNumber, counter)
        	newVal, err := strconv.ParseInt(counterIdxStr, 10, 64)
        	if err != nil {
			lstutil.Logger(fmt.Sprintf("Failed to convert %s to int64\n Error: %s\n", counterIdxStr, err))
			currValues[peerNumber] = 0
			failedCount++
		} else {
			currValues[peerNumber] = newVal
		}
	}
	if failedCount > F {
		subTestsFailures++
		lstutil.Logger(fmt.Sprintf("%s TEST STARTUP FAILURE!!! Failed to query %s peers. RERUN when at least %d/%d peers are running, in order to be able to reach consensus.", txName, failedCount, ((N-1)/3)*2+1, N ))
	} else {
		var consensus_counter int64
		consensus_counter = 0
		found_consensus := false

		for i := 0 ; i <= F ; i++ {
			i_val_cntr := 1
			for j := i+1 ; j < N ; j++ {
				if currValues[j] == currValues[i] { i_val_cntr++ }
			}
			if i_val_cntr >= N-F { consensus_counter = currValues[i]; found_consensus = true; break }
		}
		if found_consensus {
			counter = consensus_counter
			lstutil.Logger(fmt.Sprintf("%s TEST PASS STARTUP: %d/%d peers reached consensus: current count = %d", txName, N-failedCount, N, consensus_counter))
		} else {
		subTestsFailures++
			lstutil.Logger(fmt.Sprintf("%s TEST FAIL upon STARTUP: peers cannot reach consensus on current count", txName))
		}
	}
	return counter
}

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
		subTestsFailures++
		lstutil.Logger(fmt.Sprintf("%s TEST FAIL!!!  TOO MANY PEERS (%d) FAILED to obtain the correct count, so network consensus failed!!!\n(If fewer than %d/%d peers are running, then the network does not have enough running nodes to reach consensus.)", txName, failedCount,  ((N-1)/3)*2+1, N ))
	} else {
		lstutil.Logger(fmt.Sprintf("%s TEST PASS.  %d/%d peers reached consensus on the correct count", txName, N-failedCount, N))
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
		lstutil.Logger(fmt.Sprintf("\n\n%s TEST FAIL!!! Results in A value DO NOT match on all Peers after %s", txName, txName))
	}

	if res0BI == expectedB && res1BI == expectedB && res2BI == expectedB && res3BI == expectedB {
		lstutil.Logger(fmt.Sprintf("\n\n%s TEST PASS : Results in B value match on all Peers after %s", txName, txName))
		lstutil.Logger(fmt.Sprintf("Values Verified : peer0: %d, peer1: %d, peer2: %d, peer3: %d\n\n", res0BI, res1BI, res2BI, res3BI))
	} else {
		lstutil.Logger(fmt.Sprintf("\n\n%s TEST FAIL!!! Results in B value DO NOT match on all Peers after %s", txName, txName))
	}
}
*/

func getHeight_deprecated() {
	ht0, _ := chaincode.GetChainHeight(threadutil.GetPeer(0))
	ht1, _ := chaincode.GetChainHeight(threadutil.GetPeer(1))
	ht2, _ := chaincode.GetChainHeight(threadutil.GetPeer(2))
	ht3, _ := chaincode.GetChainHeight(threadutil.GetPeer(3))

	if (ht0 == 3) && (ht1 == 3) && (ht2 == 3) && (ht3 == 3) {
		lstutil.Logger(fmt.Sprintf("CHAIN HEIGHT TEST PASS : Results in A value match on all Peers after deploy and single invoke:"))
		lstutil.Logger(fmt.Sprintf("  Height Verified: ht0=%d, ht1=%d, ht2=%d, ht3=%d", ht0, ht1, ht2, ht3))
	} else {
		subTestsFailures++
		lstutil.Logger(fmt.Sprintf("CHAIN HEIGHT TEST FAIL!!! value in chain height DOES NOT MATCH ON ALL PEERS after deploy and single invoke:"))
		lstutil.Logger(fmt.Sprintf("  All heights DO NOT MATCH expected value: ht0=%d, ht1=%d, ht2=%d, ht3=%d", ht0, ht1, ht2, ht3))
	}
}


func getHeight(expectedToMatch int) {

        ht0, _ := chaincode.GetChainHeight(threadutil.GetPeer(0))
        ht1, _ := chaincode.GetChainHeight(threadutil.GetPeer(1))
        ht2, _ := chaincode.GetChainHeight(threadutil.GetPeer(2))
        ht3, _ := chaincode.GetChainHeight(threadutil.GetPeer(3))

        numPeers := peernetwork.GetNumberOfPeers(myNetwork)
        if numPeers != 4 { fmt.Println(fmt.Sprintf("TEST FAILURE: TODO: Must fix code %d peers, rather than default=4 peers in network!!!", numPeers)) }
        // before declaring failure, we will first check if we at least have consensus, with enough nodes with the correct height
        agree := 1
        if (ht0 == ht1) { agree++ }
        if (ht0 == ht2) { agree++ }
        if (ht0 == ht3) { agree++ }
        if agree < numPeers - ((numPeers-1) / 3) {
                agree = 1
                if (ht1 == ht2) { agree++ }
                if (ht1 == ht3) { agree++ }
        }

        if (ht0 == expectedToMatch) && (ht1 == expectedToMatch) && (ht2 == expectedToMatch) && (ht3 == expectedToMatch) {
                myStr := fmt.Sprintf("CHAIN HEIGHT TEST PASS : value match on all Peers, after deploy and single invoke:\n")
                myStr += fmt.Sprintf("  Height Verified: ht0=%d, ht1=%d, ht2=%d, ht3=%d", ht0, ht1, ht2, ht3)
                lstutil.Logger(myStr)
        } else if agree >= numPeers - ((numPeers-1) / 3) {
                myStr := fmt.Sprintf("CHAIN HEIGHT TEST PASS : value match on enough Peers for consensus, after deploy and single invoke:\n")
                myStr += fmt.Sprintf("  Height Verified: ht0=%d, ht1=%d, ht2=%d, ht3=%d", ht0, ht1, ht2, ht3)
                lstutil.Logger(myStr)
        } else {
                subTestsFailures++
                myStr := fmt.Sprintf("CHAIN HEIGHT TEST FAIL : value in chain height DOES NOT MATCH expected value %d ON ALL PEERS after deploy and single invoke:\n", expectedToMatch)
                myStr += fmt.Sprintf("  All heights DO NOT MATCH expected value: ht0=%d, ht1=%d, ht2=%d, ht3=%d", ht0, ht1, ht2, ht3)
                lstutil.Logger(myStr)
        }
}


func getBlockTxInfo(blockNumber int) {
	errTransactions := 0
	height, _ := chaincode.GetChainHeight(threadutil.GetPeer(0))
	lstutil.Logger(fmt.Sprintf("\n############### Total Blocks # %d", height))

	for i := 1; i < height; i++ {
		//fmt.Printf("\n============================== Current BLOCKS %d ==========================\n", i)
		nonHashData, _ := chaincode.GetBlockTrxInfoByHost(threadutil.GetPeer(0), i)
		length := len(nonHashData.TransactionResult)
		for j := 0; j < length; j++ {
			// Print Error info only when transaction failed
			if nonHashData.TransactionResult[j].ErrorCode > 0 {
				lstutil.Logger(fmt.Sprintln("\nBlock[%d] UUID [%d] ErrorCode [%d] Error: %s", i, nonHashData.TransactionResult[j].Uuid, nonHashData.TransactionResult[j].ErrorCode, nonHashData.TransactionResult[j].Error))
				errTransactions++
			}
		}
	}
	if errTransactions > 0 {
		lstutil.Logger(fmt.Sprintf("\nTotal Blocks ERRORS # %d", errTransactions))
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
