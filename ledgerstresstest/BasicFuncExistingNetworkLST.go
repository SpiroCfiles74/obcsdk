package main

import (
	"fmt"
	"obcsdk/chaincode"
	"obcsdk/peernetwork"
	"os"
	//"strconv"
	"strings"
	"time"
	"obcsdk/lstutil"
	"obcsdk/chco2"
)

var f *os.File
//var myNetwork peernetwork.PeerNetwork
var url string
var counter int64
var subTestsFailures int

func main() {
	subTestsFailures = 0
	lstutil.TESTNAME = "BasicFuncExistingNetworkLST"
	lstutil.InitLogger(lstutil.TESTNAME)
	lstutil.Logger("\n*********** " + lstutil.TESTNAME + " ***************")

	defer timeTrack(time.Now(), "Total execution time for " + lstutil.TESTNAME)

	setupNetwork()  // establish chco2.MyNetwork, using networkcredentials

	lstutil.Logger("\n===== userRegisterTest =====")
	lstutil.Logger("userRegisterTest: FirstUser=" + peernetwork.FirstUser)
	user_ip, user_port, user_name, err := peernetwork.PeerOfThisUser(chco2.MyNetwork, peernetwork.FirstUser)
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
	lstutil.Logger(fmt.Sprintf("  ChainStats response status: %s\n  ChainStats response body: %s\n", status, response))

	lstutil.Logger("\n===== Deploy Test =====")
	counter = lstutil.DeployChaincode(chco2.MyNetwork)  // includes sleep 60 secs for Local network or 120 secs for External network
	lstutil.Logger(fmt.Sprintf("-----Deploy Test returned counter: %d", counter))
	// lstutil.Logger("Sleep another 2 min before proceeding..."); time.Sleep(120 * time.Second)

	queryCounterSuccess := lstutil.QueryAllHostsToGetCurrentCounter(chco2.MyNetwork, lstutil.TESTNAME, &counter)
	if !queryCounterSuccess {
		subTestsFailures++
		lstutil.Logger(fmt.Sprintf("BasicFuncExistingNetworkLST: WARNING: CANNOT find consensus in network for actual values; counter value will likely fail to match expected value"))
		// panic(errors.New("BasicFuncExistingNetworkLST: CANNOT find consensus in existing network"))
	}
	lstutil.Logger(fmt.Sprintf("-----BasicFuncExistingNetworkLST, AFTER deploy,QueryAllHosts retrieved counter value (and this is now the expected value): %d", counter))

        height := chaincode.Monitor_ChainHeight(url) // and save the height; it will be needed below for getHeight()

	queryDeploySuccess := lstutil.QueryAllHosts(chco2.MyNetwork, "DEPLOY", counter)
	if !queryDeploySuccess { subTestsFailures++ }

	lstutil.Logger("\n===== Invoke Test =====")
	invRes := lstutil.InvokeChaincode(chco2.MyNetwork, &counter)  // increments counter inside
	height++
	time.Sleep(10 * time.Second)
	queryInvokeSuccess := lstutil.QueryAllHosts(chco2.MyNetwork, "INVOKE", counter)
	if !queryInvokeSuccess { subTestsFailures++ }

	lstutil.Logger("\n===== GetChainHeight Test =====")
	getHeight(chco2.MyNetwork, height)  // this gets height from all peers and validates all match

	lstutil.Logger("\n===== GetBlock Stats API Test =====")
	//chaincode.BlockStats(url, height)

	if len(chco2.MyNetwork.Peers) < 1 { panic("No peers in network; cannot run this test") }
	peername := chco2.MyNetwork.Peers[0].PeerDetails["name"]
	txList, _ := chaincode.GetBlockTrxInfoByHost(peername, height-1)
	myStr = "\nGetBlocks API TEST "
	//if strings.Contains(txList.TransactionResult[0].Uuid, invRes) { 	// v0.5
	if txList != nil && strings.Contains(txList[0].Txid, invRes) {  // these should be equal, if the invoke transaction was successful
		myStr += fmt.Sprintf("PASS: Transaction Successfully stored in Block")
		myStr += fmt.Sprintf("\nCH_Block = %d, txid = %s, InvokeTransactionResult = %s\n", height-1, txList[0].Txid, invRes)
		lstutil.Logger(myStr)
	} else {
		subTestsFailures++
		myStr += fmt.Sprintf("FAIL!!! Transaction NOT stored in Block")
		myStr += fmt.Sprintf("\nCH_Block = %d, txid = %s, InvokeTransactionResult = %s\n", height-1, txList[0].Txid, invRes)
		lstutil.Logger(myStr)
		getBlockTxInfo(chco2.MyNetwork,0)
	}

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
	lstutil.Logger(fmt.Sprintf("\n*********** END BasicFuncExistingNetworkLST OVERALL TEST RESULT = %s ***************\n", myStr))
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
        chco2.MyNetwork = chaincode.InitNetwork()

        lstutil.Logger("----- InitChainCodes -----")
        chaincode.InitChainCodes()
	time.Sleep(50 * time.Second)

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
		

	time.Sleep(10 * time.Second)
	//peernetwork.PrintNetworkDetails(chco2.MyNetwork)
	peernetwork.PrintNetworkDetails()
	numPeers := peernetwork.GetNumberOfPeers(chco2.MyNetwork)

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
	time.Sleep(2 * time.Second)

	lstutil.Logger("\n----- RegisterUser Negative Test -----")
	response, status = chaincode.UserRegister_Status(url, "ghostuserdoesnotexist")
	if (strings.Contains(status, "200")) {
		subTestsFailures++
		lstutil.Logger(fmt.Sprintf("RegisterUser API Negative TEST FAIL!!! User <ghostuserdoesnotexist> was found in Registrar User List but it was never registered!\n status = %s\n response = %s\n", status, response))
	} else {
		if strings.Contains(status, "401") { // Unauthorized
			// Did not find the specified username, and no error occurred while trying;
			// this is a good expected result for our test.
			lstutil.Logger("RegisterUser API Negative TEST PASS: CONFIRMED that user <ghostuserdoesnotexist> is unregistered as expected")
		} else {
			// Did not find the specified username, and we encountered some other error while trying
			lstutil.Logger(fmt.Sprintf("RegisterUser API Negative TEST FAIL!!! ERROR while searching for non-existant user!\n status = %s\n response = %s\n", status, response))
		}
	}
	time.Sleep(2 * time.Second)

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
	time.Sleep(2 * time.Second)
 */

	lstutil.Logger("\n----- UserRegister_ecert Negative Test -----")
	response, status = chaincode.UserRegister_ecertDetail(url, "ghostuserdoesnotexist")
	if ((strings.Contains(status, "200")) == false) {
		lstutil.Logger("UserRegister_ecert API Negative TEST PASS: CONFIRMED that user <ghostuserdoesnotexist> is unregistered as expected")
	} else {
		subTestsFailures++
		lstutil.Logger(fmt.Sprintf("UserRegister_ecert API Negative TEST FAIL!!! User <ghostuserdoesnotexist> was found in Registrar User List but it was never registered!\n status = %s\n response = %s\n", status, response))
	}
	time.Sleep(5 * time.Second)
}

func getHeight(mynetwork peernetwork.PeerNetwork, expectedToMatch int) {
        ht0, _ := chaincode.GetChainHeight(mynetwork.Peers[0].PeerDetails["name"])
        ht1, _ := chaincode.GetChainHeight(mynetwork.Peers[1].PeerDetails["name"])
        ht2, _ := chaincode.GetChainHeight(mynetwork.Peers[2].PeerDetails["name"])
        ht3, _ := chaincode.GetChainHeight(mynetwork.Peers[3].PeerDetails["name"])
        numPeers := peernetwork.GetNumberOfPeers(chco2.MyNetwork)
        if numPeers != 4 {
		fmt.Println(fmt.Sprintf("TEST FAILURE: TODO: Must fix code %d peers, rather than default=4 peers in network!!!", numPeers))
		panic("Not enough peers in network to run this test")
	}


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

// pass in 0 to get ALL blocks
func getBlockTxInfo(mynetwork peernetwork.PeerNetwork, blockNumber int) {
	errTransactions := 0
	if len(mynetwork.Peers) < 1 { panic("No peers in network to run this test") }
	peername := mynetwork.Peers[0].PeerDetails["name"]
	height, _ := chaincode.GetChainHeight(peername)
	lstutil.Logger(fmt.Sprintf("++++++++++ getBlockTxInfo(%d) Total Blocks # %d", blockNumber, height))

	for i := 1; i < height; i++ {
	    if blockNumber == 0 || blockNumber == i {
		lstutil.Logger(fmt.Sprintf("+++++ Current BLOCK %d +++++", i))
		txList, _ := chaincode.GetBlockTrxInfoByHost(peername, i)
		length := len(txList)
		for j := 0; j < length; j++ {
			lstutil.Logger(fmt.Sprintln("Block[%d] TX[%d] Txid [%d]", i, j, txList[j].Txid))
		}
	    }
	}
	if errTransactions > 0 {
		lstutil.Logger(fmt.Sprintf("\nTotal Blocks ERRORS # %d", errTransactions))
	}
}

func timeTrack(start time.Time, name string) {
	elapsed := time.Since(start)
	lstutil.Logger(fmt.Sprintf("\n*********** %s , elapsed %s\n", name, elapsed))
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}
