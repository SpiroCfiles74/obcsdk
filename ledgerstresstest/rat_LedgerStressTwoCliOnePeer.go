package main

import (
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"

	"obcsdk/chaincode"
	"obcsdk/peernetwork"
	"obcsdk/util"
)

/********** Test Objective : Ledger Stress with 2 Clients and 1 Peer ************
*
*   Setup: 4 node peer network with security enabled
*   1. Deploy chaincode https://goo.gl/TysS79
*   2. Invoke 10K transactions from each client, simultaneously on a single peer
*   3. Check if the counter value(20000) matches with query on "counter"
*
* USAGE: NETWORK="LOCAL" go run LedgerStressOneCliOnePeer.go Utils.go
*  This NETWORK env value could be LOCAL or Z
*********************************************************************/
var peerNetworkSetup peernetwork.PeerNetwork
var AVal, BVal, curAVal, curBVal, invokeValue int64
var argA = []string{"a"}
var argB = []string{"counter"}
var counter int64
var wg sync.WaitGroup

const (
	TRX_COUNT = 20000
	CLIENTS   = 2
)

func initNetwork() {
	util.Logger("========= Init Network =========")
	//peernetwork.GetNC_Local()
	peerNetworkSetup = chaincode.InitNetwork()
	chaincode.InitChainCodes()
	util.Logger("========= Register Users =========")
	chaincode.RegisterCustomUsers()
}

func invokeChaincode(user string) {
	counter++
	arg1Construct := []string{util.CHAINCODE_NAME, util.INVOKE, user}
	arg2Construct := []string{"a" + strconv.FormatInt(counter, 10), util.DATA, "counter"}

	_, _ = chaincode.InvokeAsUser(arg1Construct, arg2Construct)
}

func Init() {
	//initialize
	done := make(chan bool, 1)
	counter = 0
	wg.Add(CLIENTS)
	// Setup the network based on the NetworkCredentials.json provided
	initNetwork()

	//Deploy chaincode
	util.DeployChaincode(done)
}

func InvokeMultiThreads() {
	curTime := time.Now()
	go func() {
		for i := 1; i <= TRX_COUNT/CLIENTS; i++ {
			if counter%1000 == 0 {
				elapsed := time.Since(curTime)
				util.Logger(fmt.Sprintf("=========>>>>>> Iteration# %d Time: %s CLIENT-1", counter, elapsed))
				util.Sleep(60)
				curTime = time.Now()
			}
			invokeChaincode(util.GetUser(0))
			//invokeChaincode("user_type1_1c1b79bb45") //For Z Machine
			//invokeChaincode("test_user3")//For local Machine
		}
		wg.Done()
	}()
	go func() {
		for i := 1; i <= TRX_COUNT/CLIENTS; i++ {
			if counter%1000 == 0 {
				elapsed := time.Since(curTime)
				util.Logger(fmt.Sprintf("=========>>>>>> Iteration# %d Time: %s CLIENT-2", counter, elapsed))
				util.Sleep(60)
				curTime = time.Now()
			}
			invokeChaincode(util.GetUser(1))
			//invokeChaincode("dashboarduser_type0_efeeb83216") //For Z -Machine
			//invokeChaincode("test_user4")
		}
		wg.Done()
	}()
}

//Execution starts here ...
func main() {
	util.InitLogger("LedgerStressTwoCliOnePeer")
	//TODO:Add support similar to GNU getopts, http://goo.gl/Cp6cIg
	if len(os.Args) < 1 {
		util.Logger("Usage: go run LedgerStressTwoCliOnePeer.go Utils.go")
		return
	}
	//TODO: Have a regular expression to check if the give argument is correct format
	/*if !strings.Contains(os.Args[1], "http://") {
		util.Logger("Error: Argument submitted is not right format ex: http://127.0.0.1:5000 ")
		return;
	}*/
	//Get the URL
	//url := os.Args[1]
	// time to messure overall execution of the testcase
	defer util.TimeTracker(time.Now(), "Total execution time for LedgerStressOneCliOnePeer.go ")

	Init()
	util.Logger("========= Transacations execution stated  =========")
	InvokeMultiThreads()
	wg.Wait()
	util.Logger("========= Transacations execution ended  =========")
	util.TearDown(counter)
}
