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
	"obcsdk/threadutil"
)

/********** Test Objective : Ledger Stress with 2 Clients and 1 Peer ************
*
*   Setup: 4 node peer network with security enabled
*   1. Deploy chaincode https://goo.gl/TysS79
*   2. Invoke 10K transactions from each client, simultaneously on a single peer
*   3. Check if the counter value(20000) matches with query on "counter"
*
* USAGE: NETWORK="LOCAL" go run LedgerStressTwoCliOnePeer.go 
*  This NETWORK env value could be LOCAL or Z
*********************************************************************/
var peerNetworkSetup peernetwork.PeerNetwork
var AVal, BVal, curAVal, curBVal, invokeValue int64
var argA = []string{"a"}
var argB = []string{"counter"}
var counter int64
var wg sync.WaitGroup

const (
	TRX_COUNT = 5 	// 20000
	CLIENTS   = 2

	// The seconds required for 1000 transactions to be processed by the peer network: 30 implies a rate of about 33/sec.
	// Our network can handle that for example02. And hopefully for this custom network too.

	THROUGHPUT_RATE = 30 		// transactions per second = traffic rate that we can expect the peers to handle for long durations

	BUNDLE_OF_TRANSACTIONS = 1000 	// in each client, after sending this many transactions, print a status msg and sleep for ntwk to catch up

	SLEEP_SECS = BUNDLE_OF_TRANSACTIONS / THROUGHPUT_RATE
)

func initNetwork() {
	util.Logger("========= Init Network =========")
	//peernetwork.GetNC_Local()
	peernetwork.GetNC_Local()
	peerNetworkSetup = chaincode.InitNetwork()
	chaincode.InitChainCodes()
	util.Logger("========= Register Custom Users =========")
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

	// Note: we have defined 4 custom users in package threadutil that we can use for each thread.

	// create multiple threads to send invokes to the last/highest peer, from different users in each thread
	for t := 0; t < CLIENTS && t < threadutil.NumberCustomUsersOnLastPeer; t++ {
		util.Logger(fmt.Sprintf("========= Starting thread for CLIENT %d", t+1))
		go func(clientThread int) {
			username := threadutil.GetUser(clientThread)
			for i := 1; i <= TRX_COUNT/CLIENTS; i++ {
				if counter % BUNDLE_OF_TRANSACTIONS == 0 {
					// Here in thread code, multiply sleep_secs by the number of concurrent peers
					// (which are all submitting transactions/blocks at the same rate) to get the sleeptime
					elapsed := time.Since(curTime)
					util.Logger(fmt.Sprintf("=========>>>>>> Iteration# %d Time: %s CLIENT %d", counter, elapsed, clientThread+1))
					if counter > 1 { util.Sleep( int64(SLEEP_SECS * CLIENTS) ) }
					curTime = time.Now()
				}
				invokeChaincode(username)
			}
			wg.Done()
		}(t)
	}

/*
	go func() {
		username := threadutil.GetUser(0)
		for i := 1; i <= TRX_COUNT/CLIENTS; i++ {
			if counter%1000 == 0 {
				elapsed := time.Since(curTime)
				util.Logger(fmt.Sprintf("=========>>>>>> Iteration# %d Time: %s CLIENT-1", counter, elapsed))
				if counter > 1 { util.Sleep(30) }
				curTime = time.Now()
			}
			invokeChaincode(username)
		}
		wg.Done()
	}()
	go func() {
		username := threadutil.GetUser(1)
		for i := 1; i <= TRX_COUNT/CLIENTS; i++ {
			if counter%1000 == 0 {
				elapsed := time.Since(curTime)
				util.Logger(fmt.Sprintf("=========>>>>>> Iteration# %d Time: %s CLIENT-2", counter, elapsed))
				if counter > 1 { util.Sleep(30) }
				curTime = time.Now()
			}
			invokeChaincode(username)
		}
		wg.Done()
 	}()
 */
}

//Execution starts here ...
func main() {
	util.InitLogger("LedgerStressTwoCliOnePeer")
	//TODO:Add support similar to GNU getopts, http://goo.gl/Cp6cIg
	if len(os.Args) < 1 {
		util.Logger("Usage: go run LedgerStressTwoCliOnePeer.go ")
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
	defer util.TimeTracker(time.Now(), "Total execution time for LedgerStressTwoCliOnePeer.go ")

	Init()
	util.Logger("========= Transacations execution stated  =========")
	InvokeMultiThreads()
	wg.Wait()
	util.Logger("========= Transacations execution ended  =========")
	util.TearDown(counter)
}
