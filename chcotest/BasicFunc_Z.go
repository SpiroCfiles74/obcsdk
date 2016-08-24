package main

import (
	"bufio"
	"fmt"
	"obcsdk/chaincode"
	"obcsdk/peernetwork"
	"os"
	"strconv"
	"strings"
	"time"
)

var f *os.File
var writer *bufio.Writer
var myNetwork peernetwork.PeerNetwork
var url string
var testRunStatus bool

func main() {

	var err error
	f, err = os.OpenFile("/tmp/hyperledgerBetaTestrun_Output", os.O_RDWR|os.O_APPEND, 0660)
	if err != nil {
		fmt.Println("Output file does not exist creating one ..")
		f, err = os.Create("/tmp/hyperledgerBetaTestrun_Output")
	}
	//check(err)
	defer f.Close()
	writer = bufio.NewWriter(f)

	myStr := fmt.Sprintf("\n\n*********** BEGIN BASICFUNC.go ***************")
	fmt.Println(myStr)
	fmt.Fprintln(writer, myStr)

	defer timeTrack(time.Now(), "Testcase execution Done")

	setupNetwork()

	//get a URL details to get info n chainstats/transactions/blocks etc.
	//aPeer, _ := peernetwork.APeer(myNetwork)
	//url = "https://" + aPeer.PeerDetails["ip"] + ":" + aPeer.PeerDetails["port"]
        url = "https://0d5a85cf-ed43-48b5-815f-c79bbaad6a8b_vp1-api.zone.blockchain.ibm.com:443"

	userRegisterTest()

	response, status := chaincode.NetworkPeers(url)
	if strings.Contains(status, "200") {
		myStr = fmt.Sprintf("NetworkPeers Rest API Test Pass: successful")
		fmt.Println(myStr)
		fmt.Fprintln(writer, myStr)
		fmt.Println(response)
		fmt.Fprintln(writer, response)
	}

	chaincode.ChainStats(url)

	var inita, initb, curra, currb int
	inita = 111 
	initb = 0  
	curra = inita
	currb = initb
        testRunStatus = false

	deploy()
	time.Sleep(60000 * time.Millisecond)

	query("DEPLOY", curra, currb)

	invRes := invoke()
	curra = curra - 1
	currb = currb + 1

	time.Sleep(60000 * time.Millisecond)

	query("INVOKE", curra, currb)

	getHeight()


	fmt.Println("\nBlockchain: Get Chain  ....")
	chaincode.ChainStats(url)

	fmt.Println("\nBlockchain: GET Chain  ....")
	response2 := chaincode.Monitor_ChainHeight(url)

	fmt.Println("\nChain Height", chaincode.Monitor_ChainHeight(url))

	fmt.Println("\nBlock: GET/Chain/Blocks/")
	//chaincode.BlockStats(url, response2)
	nonHashData, _ := chaincode.GetBlockTrxInfoByHost("vp2", 2)

	if strings.Contains(nonHashData.TransactionResult[0].Uuid, invRes) {
		myStr := fmt.Sprintf("\n\nGetBlocks API Test PASS: Transaction %s Successfully stored in Block ", invRes)
		fmt.Printf(myStr)
		fmt.Fprintf(writer, myStr)

		myStr1 := fmt.Sprintf("\n=============Block:%d UUID: %s \n", response2-1, nonHashData.TransactionResult[0].Uuid)
		fmt.Printf(myStr1)
		fmt.Fprintf(writer, myStr1)
		writer.Flush()
                testRunStatus = true

	} else {
		myStr := fmt.Sprintf("GetBlocks API Test FAIL: Transaction %s NOT stored in Block ", invRes)
		fmt.Printf(myStr)
		fmt.Fprintf(writer, myStr)
                testRunStatus = false
	}

	fmt.Println("\nTransactions: GET/transactions/", url, invRes)
	chaincode.Transaction_Detail(url, invRes)


	writer.Flush()

}

func setupNetwork() {

	fmt.Println("Working with an existing network")
	//peernetwork.SetupLocalNetwork(4, true)
	myNetwork = chaincode.InitNetwork()
	chaincode.InitChainCodes()
	chaincode.RegisterUsers()

	time.Sleep(10000 * time.Millisecond)
	//peernetwork.PrintNetworkDetails(myNetwork)
	peernetwork.PrintNetworkDetails()
	numPeers := peernetwork.GetNumberOfPeers(myNetwork)

	myStr := fmt.Sprintf("Launched Local Docker Network successfully with %d peers with pbft and security+privacy enabled\n", numPeers)
	fmt.Println(myStr)
	fmt.Fprintln(writer, myStr)

}

func userRegisterTest() {

        ip, port, _, _ := peernetwork.AUserFromThisPeer(myNetwork, "vp0")
	url = "https://" + ip + ":" + port
	response, status := chaincode.UserRegister_Status(url, "user_type1_0c5cb888f7")

        //fmt.Println("Response from UR %s", response)
        //fmt.Println("Status from UR %s", status)

	if strings.Contains("200", status) && strings.Contains("user_type1_226a34c9af", response) {
		fmt.Println("RegisterUser API Test PASS: User %s Registration is successful", response)
	} else {
		fmt.Println("RegisterUser API Test FAIL: User %s Registration is NOT successful", response)
	}

	response, status = chaincode.UserRegister_Status(url, "nishi")
	if (strings.Contains("200", status)) == false {
		fmt.Println("RegisterUser API -Ve Test PASS: User Nishi Is Not Registered")
	} else {
		fmt.Println("RegisterUser API Test FAIL: User Nishi found in Register user list")
	}

	chaincode.UserRegister_ecertDetail(url, "lukas")

}

func deploy() {

	fmt.Println("\nPOST/Chaincode: Deploying chaincode at the beginning ....")
	dAPIArgs0 := []string{"example02", "init"}
	depArgs0 := []string{"a", "111", "b", "0"}
	chaincode.Deploy(dAPIArgs0, depArgs0)

}
func invoke() string {

	iAPIArgs0 := []string{"example02", "invoke"}
	invArgs0 := []string{"a", "b", "1"}
	invRes, _ := chaincode.Invoke(iAPIArgs0, invArgs0)
	return invRes

}

func query(txName string, expectedA int, expectedB int) {

	qAPIArgs00 := []string{"example02", "query", "vp0"}
	qAPIArgs01 := []string{"example02", "query", "vp1"}
	qAPIArgs02 := []string{"example02", "query", "vp2"}
	qAPIArgs03 := []string{"example02", "query", "vp3"}
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

	fmt.Println("Results in a and b vp0 : ", res0AI, res0BI)
	fmt.Println("Results in a and b vp1 : ", res1AI, res1BI)
	fmt.Println("Results in a and b vp2 : ", res2AI, res2BI)
	fmt.Println("Results in a and b vp3 : ", res3AI, res3BI)

	if res0AI == expectedA && res1AI == expectedA && res2AI == expectedA && res3AI == expectedA {
		myStr := fmt.Sprintf("\n\n%s TEST PASS : Results in A value match on all Peers after %s", txName, txName)
		fmt.Println(myStr)
		fmt.Fprintln(writer, myStr)
		myStr = fmt.Sprintf("Values Verified : vp0: %d, vp1: %d, vp2: %d, vp3: %d", res0AI, res1AI, res2AI, res3AI)
		fmt.Println(myStr)
		fmt.Fprintln(writer, myStr)
                testRunStatus = true
	} else {
		myStr := fmt.Sprintf("\n\n%s TEST FAIL: Results in A value DO NOT match on all Peers after %s", txName, txName)
		fmt.Println(myStr)
		fmt.Fprintln(writer, myStr)
		myStr = fmt.Sprintf("Values Verified : vp0: %d, vp1: %d, vp2: %d, vp3: %d\n\n", res0AI, res1AI, res2AI, res3AI)
		fmt.Println(myStr)
		fmt.Fprintln(writer, myStr)
                testRunStatus = false
	}

	if res0BI == expectedB && res1BI == expectedB && res2BI == expectedB && res3BI == expectedB {
		myStr := fmt.Sprintf("\n\n%s TEST PASS : Results in B value match on all Peers after %s", txName, txName)
		fmt.Println(myStr)
		fmt.Fprintln(writer, myStr)
		myStr = fmt.Sprintf("Values Verified : vp0: %d, vp1: %d, vp2: %d, vp3: %d", res0BI, res1BI, res2BI, res3BI)
		fmt.Println(myStr)
		fmt.Fprintln(writer, myStr)
                testRunStatus = true
	} else {
		myStr := fmt.Sprintf("\n\n%s TEST FAIL: Results in B value DO NOT match on all Peers after %s", txName, txName)
		fmt.Println(myStr)
		fmt.Fprintln(writer, myStr)
		myStr = fmt.Sprintf("Values Verified : vp0: %d, vp1: %d, vp2: %d, vp3: %d", res0BI, res1BI, res2BI, res3BI)
		fmt.Println(myStr)
		fmt.Fprintln(writer, myStr)
                testRunStatus = false 
	}
	writer.Flush()
}

func getHeight() {

	ht0, _ := chaincode.GetChainHeight("vp0")
	ht1, _ := chaincode.GetChainHeight("vp1")
	ht2, _ := chaincode.GetChainHeight("vp2")
	ht3, _ := chaincode.GetChainHeight("vp3")

	if (ht0 == 3) && (ht1 == 3) && (ht2 == 3) && (ht3 == 3) {
		myStr := fmt.Sprintf("\n\nGET CHAIN HEIGHT TEST PASS : Results in A value match on all Peers after ")
		fmt.Println(myStr)
		fmt.Fprintln(writer, myStr)
		myStr = fmt.Sprintf("Height Verified: ht0: %d, ht1: %d, ht2: %d, ht3: %d ", ht0, ht1, ht2, ht3)
		fmt.Println(myStr)
		fmt.Fprintln(writer, myStr)
                testRunStatus = true
	} else {
		fmt.Printf(" All heights do NOT match : ht0: %d, ht1: %d, ht2: %d, ht3: %d ", ht0, ht1, ht2, ht3)
		myStr := fmt.Sprintf("\n\nGET CHAIN HEIGHT TEST FAIL : value in chain height match on all Peers after deploy and single invoke")
		fmt.Println(myStr)
		fmt.Fprintln(writer, myStr)
		myStr = fmt.Sprintf("Height Verified: ht0: %d, ht1: %d, ht2: %d, ht3: %d ", ht0, ht1, ht2, ht3)
		fmt.Println(myStr)
		fmt.Fprintln(writer, myStr)
                testRunStatus = false
	}
	writer.Flush()

}

func timeTrack(start time.Time, name string) {
	elapsed := time.Since(start)
        if (testRunStatus == true ) {
          myStr := fmt.Sprintln("\n*************** TEST BASICFUNC.go PASSED *********************** \n")
          fmt.Fprintln(writer, myStr)
        }
        if (testRunStatus == false ) {
          myStr := fmt.Sprintln("\n*************** TEST BASICFUNC.go FAILED *********************** \n")
          fmt.Fprintln(writer, myStr)
        }

	myStr := fmt.Sprintf("\n################# %s took %s \n", name, elapsed)
	fmt.Fprintln(writer, myStr)
	fmt.Println(myStr)
	myStr = fmt.Sprintf("\n\n*********** END BASICFUNC.go ***************\n\n")
	fmt.Println(myStr)
	fmt.Fprintln(writer, myStr)
	writer.Flush()
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}
