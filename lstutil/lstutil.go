package lstutil 	// Ledger Stress Testing functions

import (
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"
	"obcsdk/chaincode"
	"obcsdk/peernetwork"
	"obcsdk/threadutil"
)

/* *********************************************************************************************
*   1. Setup 4 node peer network with security enabled, and deploy chaincode
*   2. Caller passes in arguments:
*	number of clients per peer,
*	number of peers, and
*	total number or transactions to be divided among each go client.
*   3. Each client will invoke transactions in parallel
*   4. Confirm the total expected counter value (TRX_COUNT) matches with query on "counter"
* 
*   The default test environment is LOCAL. To optionally override,
*   tester may set on command line
*	NETWORK=Z go run <test.go>
*   or export and reuse for all test executions thereafter:
*	export NETWORK=Z
*	go run <test.go>
*	go run <test2.go>
* 
*   Tester may also set these env vars to override the test settings:
*	CLIENTS
*	TRX_COUNT
*	THROUGHPUT_RATE
* 
*   Ensure the users+passwords are set correctly in one or both appropriate locations:
* 	for the standard users, refer to:  ../util/NetworkCredentials.json
* 	for extra/custom users, refer to:  ../threadutil/threadutil.go
* 
********************************************************************************************* */

const (
	// transactions per second = traffic rate that we can expect the network of peers to handle for long durations
	//    v05 LOCAL Docker network: should handle 80 Tx/sec
	//    v05 Z or HSBN Network: 
	THROUGHPUT_RATE_DEFAULT = 80
	THROUGHPUT_RATE_MAX = 1000

	BUNDLE_OF_TRANSACTIONS = 1000 	// in each client, after sending this many transactions, print a status msg and sleep for ntwk to catch up
)

var peerNetworkSetup peernetwork.PeerNetwork
var counter int64
var wg sync.WaitGroup

var TRX_COUNT int64
var CLIENTS int
var THROUGHPUT_RATE int

// The seconds required for 1000 transactions to be processed by the peer network: 30 implies a rate of about 33/sec.
// Our network can handle that for example02. And hopefully for this custom network too.
var SLEEP_SECS int

func initNetwork() {
	Logger("========= Init Network =========")
	//peernetwork.GetNC_Local()
	peernetwork.GetNC_Local()
	peerNetworkSetup = chaincode.InitNetwork()
	chaincode.InitChainCodes()
	Logger("========= Register Custom Users =========")
	chaincode.RegisterCustomUsers()
}

// this func is not currently used, but it works; it just finds and uses first avail user on first avail peer
func invokeChaincode() {
        counter++
        arg1 := []string{CHAINCODE_NAME, INVOKE}
        arg2 := []string{"a" + strconv.FormatInt(counter, 10), DATA, "counter"}
        _, _ = chaincode.Invoke(arg1, arg2)
}

func invokeChaincodeWithUser(user string) {
	counter++
	arg1Construct := []string{CHAINCODE_NAME, INVOKE, user}
	arg2Construct := []string{"a" + strconv.FormatInt(counter, 10), DATA, "counter"}
	_, _ = chaincode.InvokeAsUser(arg1Construct, arg2Construct)
}

func invokeChaincodeOnPeer(peer string) {
        counter++
        arg1Construct := []string{CHAINCODE_NAME, INVOKE, peer}
        arg2Construct := []string{"a" + strconv.FormatInt(counter, 10), DATA, "counter"}
        _, _ = chaincode.InvokeOnPeer(arg1Construct, arg2Construct)
}

func Init() {
	done := make(chan bool, 1)
	counter = 0

        var envvar string
        envvar = os.Getenv("TRX_COUNT")
        if envvar != "" {
		TRX_COUNT, _ = strconv.ParseInt(envvar, 10, 64)
	}
	if TRX_COUNT < 1 { TRX_COUNT = 1 }
	if TRX_COUNT > 1000000000 { TRX_COUNT = 1000000000 }     // 1 billion max

	envvar = os.Getenv("CLIENTS")
        if envvar != "" {
		CLIENTS, _ = strconv.Atoi(envvar)
	}
	if CLIENTS < 1 { CLIENTS = 1 }
	if CLIENTS > threadutil.NumberCustomUsersOnLastPeer { CLIENTS = threadutil.NumberCustomUsersOnLastPeer }

	THROUGHPUT_RATE = THROUGHPUT_RATE_DEFAULT
	envvar = os.Getenv("THROUGHPUT_RATE")
        if envvar != "" {
		THROUGHPUT_RATE, _ = strconv.Atoi(envvar)
	}
	if THROUGHPUT_RATE < 2 { THROUGHPUT_RATE = 2 }
	if THROUGHPUT_RATE > THROUGHPUT_RATE_MAX { THROUGHPUT_RATE = THROUGHPUT_RATE_MAX }
	SLEEP_SECS = BUNDLE_OF_TRANSACTIONS / THROUGHPUT_RATE

	Logger(fmt.Sprintf("TRX_COUNT=%d, CLIENTS=%d, THROUGHPUT_RATE=%d/sec, SLEEP=%d secs per every %d Tx", TRX_COUNT, CLIENTS, THROUGHPUT_RATE, SLEEP_SECS, BUNDLE_OF_TRANSACTIONS))

	wg.Add(CLIENTS)

	// Setup the network based on the NetworkCredentials.json provided
	initNetwork()

	//Deploy chaincode
	DeployChaincode(done)
}

// use this func when using single thread per peer
func InvokeSingleThreadsOnPeers() {

	// Number of CLIENTS = Number of peers : this func creates one client thread for each peer

	startTime := time.Now()
	curTime := time.Now()
	nobodySleeping := true

	// All clients are submitting transactions as a whole; they all increment the shared counter.
	// The throughput rate is what can be handled by the network as a whole.

	var sleepSecs int64
	sleepSecs = int64(SLEEP_SECS) 	// do not multiply by CLIENTS, as long as we monitor the shared "counter"

	// Send invokes to each peer, one client per peer; create CLIENTS threads: Client0 on Peer0, Client1 on Peer1, etc.

	for t := 0; t < CLIENTS; t++ {		// && t < numPeersInNtwk
		go func(clientThread int) {
			peername := threadutil.GetPeer(clientThread) 	// get a user on the next peer
			var i int64
			var numTxOnThisClient int64
			numTxOnThisClient = TRX_COUNT / int64(CLIENTS)
			if clientThread == 0 { numTxOnThisClient = numTxOnThisClient + (TRX_COUNT % int64(CLIENTS)) }
			Logger(fmt.Sprintf("========= Started CLIENT-%d thread on peer %s, to run %d Tx", clientThread, threadutil.GetPeer(clientThread), numTxOnThisClient))
			for i = 0; i < numTxOnThisClient; i++ {
				invokeChaincodeOnPeer(peername)		// this function increments counter too
				if counter % BUNDLE_OF_TRANSACTIONS == 0 {
					if nobodySleeping {
						nobodySleeping = false
						elapsed := time.Since(curTime)
						accum := time.Since(startTime)
						curTime = time.Now() 		// yes this will include the sleep time in the next cycle
						Logger(fmt.Sprintf("==== %d Tx accumulated (discovered by client %d). Elapsed Time prev=%s, accum=%s", counter, clientThread, elapsed, accum))
						Logger(fmt.Sprintf("Client %d myTx=%d going to sleep %d secs", clientThread, i+1, sleepSecs))
						Sleep( sleepSecs )
						nobodySleeping = true
					} else {
						Logger(fmt.Sprintf("Client %d myTx=%d going to sleep %d secs", clientThread, i+1, sleepSecs))
						Sleep( sleepSecs )
					}
				}
			}
			Logger(fmt.Sprintf("========= Finished CLIENT-%d thread on peer %s, Tx=%d", clientThread, threadutil.GetPeer(clientThread), i))
			wg.Done()
		}(t)
	}
}

// use this func when using multiple threads to send requests to one peer
func InvokeMultiThreads() {

	//  Number of CLIENTS = Number of threads : this func creates multiple client threads sending to one peer (the last vp, PEER3)
	//  assuming the usernames are all registered on that same peer (which they are)

	startTime := time.Now()
	curTime := time.Now()
	nobodySleeping := true

	// All clients are submitting transactions as a whole; they all increment the shared counter.
	// The throughput rate is what can be handled by the network as a whole.

	var sleepSecs int64
	sleepSecs = int64(SLEEP_SECS) 	// do not multiply by CLIENTS, as long as we monitor the shared "counter"

	// Note: we have defined 4 additional custom users on the last peer (3) in package threadutil that we can use for each thread.
	// Create multiple parallel client threads to send invokes to the last/highest peer

	for t := 0; t < CLIENTS && t < threadutil.NumberCustomUsersOnLastPeer; t++ {
		go func(clientThread int) {
			username := threadutil.GetUser(clientThread) 	// get another user on the same (last) peer
			var i int64
			var numTxOnThisClient int64
			numTxOnThisClient = TRX_COUNT / int64(CLIENTS)
			if clientThread == 0 { numTxOnThisClient = numTxOnThisClient + (TRX_COUNT % int64(CLIENTS)) }
			Logger(fmt.Sprintf("========= Started CLIENT-%d thread on peer %s, to run %d Tx", clientThread, threadutil.GetPeer(threadutil.NumberOfPeers-1), numTxOnThisClient))
			for i = 0; i < numTxOnThisClient; i++ {
				invokeChaincodeWithUser(username)	// this function increments counter too, and sends invoke to the peer that contains the given user
				if counter % BUNDLE_OF_TRANSACTIONS == 0 {
					// Here in thread code, multiply sleep_secs by the number of concurrent peers
					// (which are all submitting transactions/blocks at the same rate) to get the sleeptime.
					// Each peer will hit this; the throughput rate is what can be handled by the network as a whole, not each peer.
					if nobodySleeping {
						nobodySleeping = false
						elapsed := time.Since(curTime)
						accum := time.Since(startTime)
						curTime = time.Now() 	// doing this here means we include the sleep time within the next cycle elapsed time
						Logger(fmt.Sprintf("==== %d Tx accumulated (discovered by client %d). Elapsed Time prev=%s, accum=%s", counter, clientThread, elapsed, accum))
						Logger(fmt.Sprintf("Client %d myTx=%d going to sleep %d secs", clientThread, i+1, sleepSecs))
						Sleep( sleepSecs )
						nobodySleeping = true
					} else {
						Logger(fmt.Sprintf("Client %d myTx=%d going to sleep %d secs", clientThread, i+1, sleepSecs))
						Sleep( sleepSecs )
					}
				}
			}
			Logger(fmt.Sprintf("========= Finished CLIENT-%d thread on peer %s, Tx=%d", clientThread, threadutil.GetPeer(threadutil.NumberOfPeers-1), i))
			wg.Done()
		}(t)
	}
}

//Execution starts here ...
func RunLedgerStressTest(testname string, numClients int, numPeers int, numTx int64) {
	TESTNAME = testname
	InitLogger(TESTNAME)
	CLIENTS = numClients
	TRX_COUNT = numTx

	//TODO:Add support similar to GNU getopts, http://goo.gl/Cp6cIg
	//if len(os.Args) < 1 {
	//	Logger("Usage: go run " + TESTNAME + ".go")
	//	return
	//}
	//TODO: Have a regular expression to check if the give argument is correct format
	//if !strings.Contains(os.Args[1], "http://") {
	//	Logger("Error: Argument submitted is not right format ex: http://127.0.0.1:5000 ")
	//	return;
	//}
	//Get the URL
	//url := os.Args[1]

	// time to messure overall execution of the testcase
	defer TimeTracker(time.Now(), "Total execution time for " + TESTNAME)

	Init()
	Logger("========= Transactions execution started  =========")

	if numPeers > 1 || CLIENTS == 1 {
		InvokeSingleThreadsOnPeers()	// use this for tests using single thread per peer, e.g. FourClientsFourPeers
	} else {
		InvokeMultiThreads()		// use this for multiple threads sending requests to one peer, such as TwoClientsOnePeer
	}

	wg.Wait()
	Logger("========= Transactions execution ended  =========")
	TearDown(counter)
}
