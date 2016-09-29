package main

// 
// INSTRUCTIONS:
// 
// 1. Change chco2.CurrentTestName, to set this test name = this filename
// 2. Edit to add your test steps at the bottom.
// 3. go build setupTest.go
// 4. go run setupTest.go  - or better yet, to save all results use script:  gorecord.sh setupTest.go
// 
// 
// SETUP STEPS included already:
// -----------------------------
// Default Setup: 4 peer node network with security CA node using local docker containers.
// (To change network characteristics and tuning parameters, change consts in file ../chco2/chco2.go)
// 
// SETUP 1: Deploy chaincode_example02 with A=1000000, B=1000000 as initial args.
// SETUP 2: Send INVOKES (moving 1 from A to B) once on each peer node.
// SETUP 3: Query all peers to validate values of A, B, and chainheight.
// 


import (
	"os"
	"time"
	"bufio"
	"obcsdk/chco2"
	"obcsdk/chaincode"
	"obcsdk/peernetwork"
	"fmt"
	"strconv"
	"strings"
	// "bufio"
	// "log"
)

var osFile *os.File

func main() {

	//=======================================================================================
	// SET THE TESTNAME:  set the filename/testname here, to display in output results.
	//=======================================================================================

	chco2.CurrentTestName = "CAT_100_Startup_DIQ_API.go"


	//=======================================================================================
	// Getting started: output file, test timing, setup/init, and start & confirm the network
	//=======================================================================================

	if (chco2.Verbose) { fmt.Println("Welcome to test " + chco2.CurrentTestName) }

	chco2.RanToCompletion = false
	startTime := time.Now()
	_, err := os.Stat(chco2.OutputSummaryFileName)	// Stat returns *FileInfo. It will return an error if there is no file.
	if err != nil {
		if os.IsNotExist(err) {
			// File simply does not exist. Create the *File.
			osFile, err = os.Create(chco2.OutputSummaryFileName)
			chco2.Check(err)
		} else {
			chco2.Check(err)  // some other error; panic and exit.
		}
	} else {
		// open the existing file
		osFile, err = os.OpenFile(chco2.OutputSummaryFileName, os.O_RDWR|os.O_APPEND, 0666)
		chco2.Check(err)
	}
	defer osFile.Close()

	chco2.Writer = bufio.NewWriter(osFile)

	// When main() ends, print the test PASS/FAIL line, with elapsed time, to outfile and to stdout
	defer chco2.TimeTrack(startTime, chco2.CurrentTestName)

	// Initialize everything, and start the network: deploy, invoke once on each peer, and query all peers to confirm
	chco2.Setup( chco2.CurrentTestName, startTime )


	//=======================================================================================
	// 
	// OPTIONAL OVERRIDES:
	// 	Tune these booleans to control verbosity and test strictness.
	// 	These booleans are initialized inside chco2.Setup(), as follows.
	// 
	//	Note: Set AllRunningNodesMustMatch to false when need merely enough peers for consensus
	// 	to match results, especially when test involves stopping or pausing peer nodes.
	// 	OR, set it true (default) when all active running peers must match (e.g. at init
	// 	time, or after sending enough invokes after a node outage to guarantee that all
	// 	peers are caught up, in sync, with matching values for chainheight, A & B.
	// 
	//	Simply uncomment any lines here for this testcase to override
	//	the default values, as defined in ../chco2/chco2.go
	// 
	// 	chco2.Verbose = true			// See also: "verbose" in ../chaincode/const.go
	// 	chco2.Stop_on_error = true
	// 	chco2.EnforceQueryTestsPass = false
	//	chco2.EnforceChainHeightTestsPass = false
	//	chco2.AllRunningNodesMustMatch = false 	// Note: chco2 inits to true, but sets this false when restart a peer node
		chco2.CHsMustMatchExpected = true	// not fully implemented and working, so leave this false for most TCs
	//	chco2.QsMustMatchExpected = false 	// Note: until #2148 is solved, you may need to set false here if testcase has complicated multiple stops/restarts
	//	chco2.DefaultInvokesPerPeer = 1		//  1 = default. Uncomment and change this here to override for this testcase.
	//	chco2.TransPerSecRate = 20		// 20 = default. Uncomment and change this here to override for this testcase.


	//=======================================================================================
	// 
	// chco2. API available function calls in ../chco2/chco2.go:
	// 
	//	DeployNew(A int, B int)
	//	Invokes(totalInvokes int)
	//	InvokeOnEachPeer(numInvokesPerPeer int)
	//	InvokeOnThisPeer(totalInvokes int, peerNum int)
	//	QueryAllPeers(stepName string)
	//	StopPeers(peerNums []int)
	//	RestartPeers(peerNums []int)
	//	QueryMatch(currA int, currB int)
	//	SleepTimeSeconds(secs int) time.Duration
	//	SleepTimeMinutes(minutes int) time.Duration
	//	CatchUpAndConfirm()
	// To be implemented soon:
	//	WaitAndConfirm()
	//	PausePeers(peerNums []int)
	//	UnpausePeers(peerNums []int)
	// 
	// Example usages:
	// 
	// chco2.DeployNew( 9000, 1000 )
	// chco2.Invokes( chco2.InvokesRequiredForCatchUp )
	// chco2.InvokeOnEachPeer( chco2.DefaultInvokesPerPeer )
	// InvokeOnThisPeer( 100, 0 )
	// chco2.StopPeers( []int{ 99 } )
	// chco2.QueryAllPeers( "STEP 6, after STOP PEERs " + strconv.Itoa(99) )
	// chco2.RestartPeers( []int{ j, k } )
	// chco2.QueryAllPeers( "STEP 9, after RESTART PEERs " + strconv.Itoa(j) + ", " + strconv.Itoa(k) )
	// if (chco2.Verbose) { fmt.Println("Sleep extra 60 secs") }
	// time.Sleep(chco2.SleepTimeSeconds(60))
	// time.Sleep(chco2.SleepTimeMinutes(1))
	// 
	//=======================================================================================


	//=======================================================================================
	// DEFINE MAIN TESTCASE STEPS HERE
	// 
	// CAT_100_Startup_DIQ_API.go

	// We already performed initial Deploy, Invoke, and Query. Try another (re)Deploy, Query, Invoke, Query

	chco2.DeployNew(1000000,1000000)
	chco2.QueryAllPeers( "STEP 2, Query all peers after Redeploy initial network settings" )
	chco2.InvokeOnEachPeer( 10 )
	chco2.QueryAllPeers( "STEP 4, Query all peers after " + strconv.Itoa(10) + " Invokes on each peer" )

	// API TESTS

	apiTestsPass := true
	myStr := ""

	fmt.Println("\n===== Start userRegisterTest =====")
	user_ip, user_port, user_name, err := peernetwork.PeerOfThisUser(chco2.MyNetwork, "test_user0")
	url := chaincode.GetURL(user_ip, user_port)
	rc := userRegisterTest(url, user_name)
	if err != nil || !rc { apiTestsPass = false }

	fmt.Println("\n===== Start NetworkPeers Test =====")
	response, status := chaincode.NetworkPeers(url)
	if strings.Contains(status, "200") {
		myStr = fmt.Sprintf("PASSED NetworkPeers API TEST")
	} else {
		apiTestsPass = false
		myStr = fmt.Sprintf("FAILED NetworkPeers API TEST !!!!! status=<%s> response=<%s>", status, response)
	}
	fmt.Println(myStr)
	fmt.Fprintln(chco2.Writer, myStr)
	chco2.Writer.Flush()

	fmt.Println("\n===== Get ChainStats Test =====")
	response, status = chaincode.GetChainStats(url)
	if strings.Contains(status, "200") {
		fmt.Println("ChainStats() chain info status: ", status)
		myStr = fmt.Sprintf("PASSED ChainStats API TEST")
	} else {
		apiTestsPass = false
		myStr = fmt.Sprintf("FAILED ChainStats API TEST !!!!! status=<%s> response=<%s>", status, response)
	}
	fmt.Println(myStr)
	fmt.Fprintln(chco2.Writer, myStr)
	chco2.Writer.Flush()


	// send one invoke request, just to get an invokeResponse to use below
	fmt.Println("\nSend one invoke, and retreive the transaction response...")
	peerName := chco2.MyNetwork.Peers[0].PeerDetails["name"]
	invArgs := []string{"a", "b", "1"}
	iAPIArgs := []string{"example02", "invoke", peerName}
	invRes, err := chco2.DoOneInvoke(iAPIArgs, invArgs)
	time.Sleep(2 * time.Second)
	if err != nil {
		apiTestsPass = false
		myStr = fmt.Sprintf("ERROR from invoke: ", err)
		fmt.Println(myStr)
	}

	fmt.Println("\n===== GetBlockStats API Test =====")

	height, err := chaincode.GetChainHeight(peerName)
	if err != nil {
		myStr = fmt.Sprintf("GetChainHeight with peerName=<%s> returned ERROR <%s>", peerName, err)
		fmt.Println(myStr)
	}
	txList, err := chaincode.GetBlockTrxInfoByHost(peerName, height-1)
	if err != nil {
		myStr = fmt.Sprintf("GetBlockTrxInfoByHost with peerName=<%s> and height=%d returned ERROR <%s>", peerName, height-1, err)
		fmt.Println(myStr)
	}

	if err == nil && txList != nil && strings.Contains(txList[0].Txid, invRes) { 	// these should be equal, if the invoke transaction was successful
		myStr = fmt.Sprintf("\nPASSED GetBlocks API TEST: Transaction Successfully stored in Block")
		myStr += fmt.Sprintf("\n CH_Block = %d, txid = %s, InvokeTransactionResult = %s", height-1, txList[0].Txid, invRes)
		fmt.Println(myStr)
		fmt.Fprintf(chco2.Writer, myStr)
		chco2.Writer.Flush()
	} else {
		apiTestsPass = false
                myStr = fmt.Sprintf("\nFAILED GetBlocks API TEST: Transaction NOT stored in CH_Block=%d, InvokeTransactionResult=%s", height-1, invRes)
		if txList != nil { 	// && txList[0] != nil 
			myStr += fmt.Sprintf("\n txid = %s", txList[0].Txid)
		} else {
			myStr += fmt.Sprintf("\n Transaction Result is nil!")
		}
		if err != nil {  myStr += fmt.Sprintf("\nerr = ", err) }
		fmt.Println(myStr)
		fmt.Fprintf(chco2.Writer, myStr)
		chco2.Writer.Flush()
		getBlockTxInfo(0)
	}

	fmt.Println("\n===== Get Transactions API Test =====")
	response, status, trans_err := chaincode.GetChainTransactions(url, invRes)
	if trans_err == nil && strings.Contains(status, "200") {
                myStr = fmt.Sprintf("\nPASSED Get Transactions API TEST: status=<%s> response=<%s>", status, response)
	} else {
		apiTestsPass = false
                myStr = fmt.Sprintf("\nFAILED Get Transactions API TEST: url=<%s> invRes=<%s> status=<%s> response=<%s> err=<%s>", url, invRes, status, response, trans_err)
	}
	fmt.Println(myStr)
	fmt.Fprintf(chco2.Writer, myStr)

	resultStr := "PASS"
	if !apiTestsPass {
		resultStr = "FAIL"
		chco2.AnnexTestPassResult(false) // since our API tests failed, then make sure the final test result is FAILED(false)
	}
	myStr = fmt.Sprintf("\nOVERALL API TEST RESULT = %s", resultStr)
	fmt.Println(myStr)
	fmt.Fprintf(chco2.Writer, myStr)
	chco2.Writer.Flush()

	chco2.RanToCompletion = true	// DO NOT MOVE OR CHANGE THIS. It must remain last.
}





// arg = a username that was already registered; this func confirms if it was successful
// and confirms user "ghostuserdoesnotexist" is not registered
// and confirms ecert
func userRegisterTest(url string, username string) (passed bool) {

	passed = true
	fmt.Println("\n----- RegisterUser Test -----")
	response, status := chaincode.UserRegister_Status(url, username)
	myStr := ""
	if strings.Contains(status, "200") && strings.Contains(response, username + " is already logged in") {
		myStr += fmt.Sprintf ("PASSED RegisterUser API TEST: %s User Registration was already done successfully", username)
	} else {
		passed = false
		myStr += fmt.Sprintf ("FAILED RegisterUser API TEST: %s User Registration was NOT already done\n status = %s\n response = %s", username, status, response)
	}
	fmt.Println(myStr)
	fmt.Fprintf(chco2.Writer, myStr)
	time.Sleep(1 * time.Second)

	myStr = ""
	fmt.Println("\n----- RegisterUser Negative Test -----")
	response, status = chaincode.UserRegister_Status(url, "ghostuserdoesnotexist")
	if ((strings.Contains(status, "200")) == false) {
		myStr += fmt.Sprintf ("PASSED RegisterUser Negative API TEST: CONFIRMED that user <ghostuserdoesnotexist> is unregistered as expected")
	} else {
		passed = false
		myStr += fmt.Sprintf("FAILED RegisterUser Negative API TEST: User <ghostuserdoesnotexist> was found in Registrar User List but it was never registered!\n status = %s\n response = %s\n", status, response)
	}
	fmt.Println(myStr)
	fmt.Fprintf(chco2.Writer, myStr)
	time.Sleep(1 * time.Second)

 /*
	fmt.Println("\n----- UserRegister_ecert Test -----")
	ecertUser := "lukas"
	response, status = chaincode.UserRegister_ecertDetail(url, ecertUser)
	myStr = ""
	if strings.Contains(status, "200") && strings.Contains(response, ecertUser + " is already logged in") {
		myStr += fmt.Sprintf ("PASSED UserRegister_ecert API TEST: %s ecert User Registration was already done successfully", ecertUser)
	} else {
		passed = false
		myStr += fmt.Sprintf ("FAILED UserRegister_ecert API TEST: %s ecert User Registration was NOT already done\n status = %s\n response = %s\n", username, status, response)
	}
	fmt.Println(myStr)
	fmt.Fprintf(chco2.Writer, myStr)
	time.Sleep(1 * time.Second)
 */

	fmt.Println("\n----- UserRegister_ecert Negative Test -----")
	myStr = ""
	response, status = chaincode.UserRegister_ecertDetail(url, "ecertghostuserdoesnotexist")
	if ((strings.Contains(status, "200")) == false) {
		myStr += fmt.Sprintf("PASSED UserRegister_ecert Negative API TEST: CONFIRMED that user <ecertghostuserdoesnotexist> is unregistered as expected")
	} else {
		passed = false
		myStr += fmt.Sprintf("FAILED UserRegister_ecert Negative API TEST: User <ghostuserdoesnotexist> was found in Registrar User List but it was never registered!\n status = %s\n response = %s\n", status, response)
	}
	fmt.Println(myStr)
	fmt.Fprintf(chco2.Writer, myStr)
	time.Sleep(1 * time.Second)

	chco2.Writer.Flush()
	return passed
}

func getBlockTxInfo(blockNumber int) {
	errTransactions := 0
	peerName := chco2.MyNetwork.Peers[0].PeerDetails["name"]
	height, _ := chaincode.GetChainHeight(peerName)
	myStr := fmt.Sprintf("++++++++++ getBlockTxInfo() Total Blocks # %d\n", height)
	fmt.Printf(myStr)
	fmt.Fprintf(chco2.Writer, myStr)

	for i := 1; i <= height; i++ {
		fmt.Printf("+++++ Current BLOCK %d +++++\n", i)
		//nonHashData, _ := chaincode.GetBlockTrxInfoByHost(peerName, i)
		txList, _ := chaincode.GetBlockTrxInfoByHost(peerName, i)
		length := len(txList)
		for j := 0; j < length; j++ {
				myStr1 := fmt.Sprintln("Block[%d] TX [%d] Txid [%d]", i, j, txList[j].Txid)
				fmt.Println(myStr1)
				fmt.Fprintln(chco2.Writer, myStr1)
		//	// Print Error info only when transaction failed
		//	if nonHashData.TransactionResult[j].ErrorCode > 0 {
		//		myStr1 := fmt.Sprintln("\nBlock[%d] UUID [%d] ErrorCode [%d] Error: %s\n", i, nonHashData.Transactions[j].Uuid, nonHashData.TransactionResult[j].ErrorCode, nonHashData.TransactionResult[j].Error)
		//		fmt.Println(myStr1)
		//		fmt.Fprintln(chco2.Writer, myStr1)
		//		errTransactions++
		//	}
		}
	}
	if errTransactions > 0 {
		myStr = fmt.Sprintf("\nTotal Blocks ERRORS # %d\n", errTransactions)
		fmt.Println(myStr)
		fmt.Fprintln(chco2.Writer, myStr)
	}
	chco2.Writer.Flush()
}
