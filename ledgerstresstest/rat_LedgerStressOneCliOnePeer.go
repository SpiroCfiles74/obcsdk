package main

import (
	"fmt"
	"strconv"
	"time"
	"os"
	"sync"
	"obcsdk/chaincode"
	"obcsdk/peernetwork"
	"obcsdk/util"
)

/********** Test Objective : Ledger Stress with 1 Client and 1 Peer ************
*
*   Setup: 4 node peer network with security enabled
*   1. Deploy chaincode https://goo.gl/TysS79
*   2. Invoke 20K transactions (TODO: make this value configurable ?)
*      After each 10K trxs, sleep for 30 secs, StateTransfer to take place
*      All transactions takes place on single peer with single client
*   3. Check if the counter value(20000) matches with query on "counter"
*
* USAGE: NETWORK="LOCAL" go run LedgerStressOneCliOnePeer.go 
*  This NETWORK env value could be LOCAL or Z
*********************************************************************/
var peerNetworkSetup peernetwork.PeerNetwork
var AVal, BVal, curAVal, curBVal, invokeValue int64
//var argA = []string{"a"}
//var argCounter = []string{"counter"}

var counter int64
var wg sync.WaitGroup

const (
	TRX_COUNT = 5 //20000 	// or 3000000 for long runs
)

func initNetwork() {
	fmt.Println("========= Init Network =========")
	peernetwork.GetNC_Local()
	peerNetworkSetup = chaincode.InitNetwork()
	chaincode.InitChainCodes()
	fmt.Println("========= Register Users =========")
	chaincode.RegisterUsers()
}

// Utility function to invoke on chaincode available @ http://urlmin.com/4r76d
func invokeChaincode() {
	counter++
	arg1 := []string{util.CHAINCODE_NAME, util.INVOKE}
	arg2 := []string{"a" + strconv.FormatInt(counter, 10), util.DATA, "counter"}
	_, _ = chaincode.Invoke(arg1, arg2)
}

//Repeated Invokes based on the thread and Transaction count configuration
func invokeLoop() {

	go func() {
		curTime := time.Now()
		for i := 1; i <= TRX_COUNT; i++ {
			if counter > 1 && counter%1000 == 0 {
				elapsed := time.Since(curTime)
				fmt.Println("=========>>>>>> Iteration#", counter, " Time: ", elapsed)
				util.Sleep(30)
				curTime = time.Now()
			}
			invokeChaincode()
		}
		wg.Done()
	}()
}

//Execution starts from here ...
func main() {
	util.InitLogger("LedgerStressOneCliOnePeer")
	//TODO:Add support similar to GNU getopts, http://goo.gl/Cp6cIg
	if len(os.Args) < 1 {
		fmt.Println("Usage: go run LedgerStressOneCliOnePeer.go ")
		return
	}
	//TODO: Have a regular expression to check if the give argument is correct format
	/*if !strings.Contains(os.Args[1], "http://") {
		fmt.Println("Error: Argument submitted is not right format ex: http://127.0.0.1:5000 ")
		return;
	}*/
	//Get the URL
	//url = os.Args[1]

	// time to messure overall execution of the testcase
	defer util.TimeTracker(time.Now(), "Total execution time for LedgerStressOneCliOnePeer.go ")

	//maintained counter variable to compare with final query value
	counter = 0
	wg.Add(1)
	done := make(chan bool, 1)

	// Setup the network based on the NetworkCredentials.json provided
	initNetwork()

	//Deploy chaincode
	util.DeployChaincode(done)
	fmt.Println("========= Transacations execution stated  =========")
	invokeLoop()
	wg.Wait()
	fmt.Println("========= Transacations execution ended  =========")
	util.TearDown(counter)
}
