package lstutil 	// Ledger Stress Testing functions

import (
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"
	"strings"
	"obcsdk/chaincode"
	"obcsdk/peernetwork"
)

/* *********************************************************************************************
*   1. Setup 4 node peer network with security enabled, and deploy chaincode
*   2. Caller passes in arguments:
*	number of clients per peer,
*	number of peers, and
*	total number or transactions to be divided among each go client.
*   3. Each client will invoke transactions in parallel
*   4. Confirm the total expected counter value (TX_COUNT) matches with query on "counter"
* 
*   The default test environment is LOCAL. To optionally override,
*   tester may set on command line
*	TEST_NETWORK=Z go run <test.go>
*   or export and reuse for all test executions thereafter:
*	export TEST_NETWORK=Z
*	go run <test.go>
*	go run <test2.go>
* 
*   Tester may also set these env vars to override the test settings:
*	NUM_CLIENTS
*	NUM_PEERS
*	TX_COUNT
*	THROUGHPUT_RATE
* 
*   Ensure the users+passwords are set correctly in credentials file.
* 
********************************************************************************************* */

const (
	// transactions per second = traffic rate that we can expect the network of peers to handle for long durations
	//    v05 LOCAL Docker network: should handle 11 Tx/sec on one client thread
	//    v05 Z or HSBN Network: only 1.5 - 2  on one client thread
	THROUGHPUT_RATE_DEFAULT = 10
	// THROUGHPUT_RATE_MAX = 160 	// normally should be well under 160, but this blast rate might be useful for short tests
	THROUGHPUT_RATE_MAX = 50

	BUNDLE_OF_TRANSACTIONS = 1000 	// in each client, after sending this many transactions, print a status msg and sleep for ntwk to catch up
	MAX_CLIENTS = 20
)

var peerNetworkSetup peernetwork.PeerNetwork
var wg sync.WaitGroup

var TX_COUNT int64
var NUM_CLIENTS int
var NUM_PEERS int
var THROUGHPUT_RATE int
var lstCounter int64

// The seconds required for 1000 transactions to be processed by the peer network: 30 implies a rate of about 33/sec.
// Our network can handle that for example02. And hopefully for this custom network too.
var SLEEP_SECS int

func initNetwork(localNetworkType string) {
	Logger("========= Init Network =========")
	//peernetwork.GetNC_Local()
	peernetwork.GetNC_Local()
	peerNetworkSetup = chaincode.InitNetwork()
	chaincode.InitChainCodes()

	// override if the user set env var to indicate whether or not the network is LOCAL
	if localNetworkType != "" {
		if localNetworkType == "LOCAL" {
			chaincode.SetNetworkLocality(peerNetworkSetup,true) // my local copy
			chaincode.SetNetworkIsLocal(true)
		} else {
			chaincode.SetNetworkLocality(peerNetworkSetup,false) // my local copy
			chaincode.SetNetworkIsLocal(false)
		}
	}

	//Logger("========= Register Custom Users =========")
	//chaincode.RegisterCustomUsers()
}

// this func is not currently used by LST tests, but it works fine for BasicFunc; it just finds and uses first avail user on first avail peer
func InvokeChaincode(mynetwork peernetwork.PeerNetwork, counter *int64) (invokeResponse string) {
        *counter++
        arg1 := []string{CHAINCODE_NAME, INVOKE}
        arg2 := []string{"a" + strconv.FormatInt(*counter, 10), DATA, "counter"}
        invokeResponse, _ = chaincode.InvokeWithNetwork(mynetwork, arg1, arg2)
        return invokeResponse
}

// this func just uses the specified user 
func invokeChaincodeWithUser(user string) {
	lstCounter++
	arg1Construct := []string{CHAINCODE_NAME, INVOKE, user}
	arg2Construct := []string{"a" + strconv.FormatInt(lstCounter, 10), DATA, "counter"}
	_, _ = chaincode.InvokeAsUser(arg1Construct, arg2Construct)
}

// this func finds and uses a username on the specified peer
func invokeChaincodeOnPeer(peer string) {
        lstCounter++
        arg1Construct := []string{CHAINCODE_NAME, INVOKE, peer}
        arg2Construct := []string{"a" + strconv.FormatInt(lstCounter, 10), DATA, "counter"}
        _, _ = chaincode.InvokeOnPeer(arg1Construct, arg2Construct)
}

func Init() {
	lstCounter = 0

        var envvar string
        envvar = os.Getenv("TEST_LST_TX_COUNT")
        if envvar != "" {
		TX_COUNT, _ = strconv.ParseInt(envvar, 10, 64)
	}
	if TX_COUNT < 1 { TX_COUNT = 1 }
	if TX_COUNT > 1000000000 { TX_COUNT = 1000000000 }     // 1 billion max

	envvar = os.Getenv("TEST_LST_NUM_CLIENTS")
        if envvar != "" {
		NUM_CLIENTS, _ = strconv.Atoi(envvar)
	}
	if NUM_CLIENTS < 1 { NUM_CLIENTS = 1 }
	if NUM_CLIENTS > 400 {
		NUM_CLIENTS = 400
		Logger(fmt.Sprintf("Too many NUM_CLIENTS requested! Using maximum = %d", NUM_CLIENTS))
	}

	envvar = os.Getenv("TEST_LST_NUM_PEERS")
        if envvar != "" {
		NUM_PEERS, _ = strconv.Atoi(envvar)
	}
	if NUM_PEERS < 1 { NUM_PEERS = 1 }
	if NUM_PEERS > 4 {
		NUM_PEERS = 4
		Logger(fmt.Sprintf("Too many NUM_PEERS requested! Using maximum = %d", NUM_PEERS))
	}

	localNetworkType := strings.TrimSpace(strings.ToUpper(os.Getenv("TEST_NETWORK")))
	THROUGHPUT_RATE = THROUGHPUT_RATE_DEFAULT
	if localNetworkType == "" || localNetworkType == "LOCAL" {
		THROUGHPUT_RATE = NUM_CLIENTS * 11
		//THROUGHPUT_RATE = 11
	} else if localNetworkType == "Z" {
		THROUGHPUT_RATE = NUM_CLIENTS * 2
		//THROUGHPUT_RATE = 2
	}
	envvar = os.Getenv("TEST_LST_THROUGHPUT_RATE")
        if envvar != "" {
		THROUGHPUT_RATE, _ = strconv.Atoi(envvar)
	}
	if THROUGHPUT_RATE < 1 { THROUGHPUT_RATE = 1 }
	if THROUGHPUT_RATE > THROUGHPUT_RATE_MAX { THROUGHPUT_RATE = THROUGHPUT_RATE_MAX }
	SLEEP_SECS = BUNDLE_OF_TRANSACTIONS / THROUGHPUT_RATE

	Logger(fmt.Sprintf("TX_COUNT=%d, NUM_CLIENTS=%d, NUM_PEERS=%d, THROUGHPUT_RATE=%d/sec, SLEEP=%d secs per every %d Tx", TX_COUNT, NUM_CLIENTS, NUM_PEERS, THROUGHPUT_RATE, SLEEP_SECS, BUNDLE_OF_TRANSACTIONS))

	wg.Add(NUM_CLIENTS)

	// Setup the network "peerNetworkSetup" based on the NetworkCredentials.json provided
	initNetwork(localNetworkType)

	//Deploy chaincode
	lstCounter = DeployChaincode(peerNetworkSetup)

	// since may be using an exisiting network that has handled some transactions (and thus has non-zero counter),
	// get the current counter value (which would not be zero since we probably deployed again the same original fixed values for A and Counter)
	queryCounterSuccess := QueryAllHostsToGetCurrentCounter(peerNetworkSetup, TESTNAME, &lstCounter)
	if !queryCounterSuccess {
		Logger(fmt.Sprintf("%s: Init() WARNING: CANNOT find consensus in network for lstCounter value; it may not match expected value later", TESTNAME))
		// panic(errors.New("CANNOT find consensus in existing network"))
	}
	Logger(fmt.Sprintf("%s: AFTER deploy,QueryAllHosts retrieved lstCounter value = %d", TESTNAME, lstCounter))
}

// use this func when using single thread per peer
func InvokeSingleThreadsOnPeers() {

	// Number of NUM_CLIENTS = Number of peers : this func creates one client thread for each peer

	startTime := time.Now()
	startCount := lstCounter
 /*
	curTime := time.Now()
	nobodySleeping := true

	// All clients are submitting transactions as a whole; they all increment the shared counter.
	// The throughput rate is what can be handled by the network as a whole.

	var sleepSecs int64
	sleepSecs = int64(SLEEP_SECS) 	// do not multiply by NUM_CLIENTS, as long as we monitor the shared "counter"
 */

	// Send invokes to each peer, one client per peer; create NUM_CLIENTS threads: Client0 on Peer0, Client1 on Peer1, etc.

	if NUM_CLIENTS > len(peerNetworkSetup.Peers) { panic("Not enough peers in network to run this test") }

	for t := 0; t < NUM_CLIENTS; t++ {
		go func(clientThread int) {
			peername := peerNetworkSetup.Peers[clientThread].PeerDetails["name"] 	// get a user on the next peer
			if peername == "" { Logger(fmt.Sprintf("Undefined/Blank peername for peer %d", clientThread)) }
			var i int64
			var numTxOnThisClient int64
			numTxOnThisClient = TX_COUNT / int64(NUM_CLIENTS)
			if clientThread == 0 { numTxOnThisClient = numTxOnThisClient + (TX_COUNT % int64(NUM_CLIENTS)) }
			Logger(fmt.Sprintf("========= Started CLIENT %d thread on peer %s, to run %d Tx", clientThread, peername, numTxOnThisClient))
			for i = 0; i < numTxOnThisClient; i++ {
				invokeChaincodeOnPeer(peername)		// this function increments counter too
				if (lstCounter-startCount) % BUNDLE_OF_TRANSACTIONS == 0 {
					Logger(fmt.Sprintf("==== %d Tx accumulated (discovered by client %d), accumulated elapsed time : %s", (lstCounter-startCount), clientThread, time.Since(startTime)))
					/*  *****************************
					 *  **	9/20/16 - We can run only 11 per sec from laptop in local env in one thread, so there is no need
					 *  **  to sleep unless for example we are running 4 peers while the target tps is set to less than 40!

					if nobodySleeping {
						nobodySleeping = false
						elapsed := time.Since(curTime)
						accum := time.Since(startTime)
						curTime = time.Now() 		// yes this will include the sleep time in the next cycle
						Logger(fmt.Sprintf("==== %d Tx accumulated (discovered by client %d). Elapsed Time prev=%s, accum=%s", (lstCounter-startCount), clientThread, elapsed, accum))
						Logger(fmt.Sprintf("Client %d myTx=%d going to sleep %d secs", clientThread, i+1, sleepSecs))
						Sleep( sleepSecs )
						nobodySleeping = true
					} else {
						Logger(fmt.Sprintf("Client %d myTx=%d going to sleep %d secs", clientThread, i+1, sleepSecs))
						Sleep( sleepSecs )
					}
					*********************************  */
				}
			}
			Logger(fmt.Sprintf("========= Finished CLIENT %d thread on peer %s, Tx=%d", clientThread, peername, i))
			wg.Done()
		}(t)
	}
}

// use this func when using multiple threads to send requests to one peer
func InvokeMultiThreads() {

	peerNum := 1

	//  Number of NUM_CLIENTS = Number of threads : this func creates multiple client threads sending to one peer (the last vp, PEER3)
	//  assuming the usernames are all registered on that same peer (which they are)

	startTime := time.Now()
	startCount := lstCounter
 /*
	curTime := time.Now()
	nobodySleeping := true

	// All clients are submitting transactions as a whole; they all increment the shared counter.
	// The throughput rate is what can be handled by the network as a whole.

	var sleepSecs int64
	sleepSecs = int64(SLEEP_SECS) 	// do not multiply by NUM_CLIENTS, as long as we monitor the shared "counter"
 */

	for t := 0; t < NUM_CLIENTS ; t++ {
		go func(clientThread int) {
			username := peernetwork.GetAUserFromPeer(peerNetworkSetup, peerNum) 	// get a user on selected peer; probably same user for all clients
			if username == "" { panic(fmt.Sprintf("Cannot find a user on peer %d", peerNum)) }
			var i int64
			var numTxOnThisClient int64
			numTxOnThisClient = TX_COUNT / int64(NUM_CLIENTS)
			if clientThread == 0 { numTxOnThisClient = numTxOnThisClient + (TX_COUNT % int64(NUM_CLIENTS)) }
			Logger(fmt.Sprintf("========= Started CLIENT %d thread on peer %d, to run %d Tx", clientThread, peerNum, numTxOnThisClient))
			for i = 0; i < numTxOnThisClient; i++ {
				invokeChaincodeWithUser(username)	// this function increments counter too, and sends invoke to the peer that contains the given user
				if (lstCounter-startCount) % BUNDLE_OF_TRANSACTIONS == 0 {
					Logger(fmt.Sprintf("==== %d Tx accumulated (discovered by client %d), accumulated elapsed time : %s", (lstCounter-startCount), clientThread, time.Since(startTime)))
					/*  *****************************
					 *  **	9/20/16 - We can run only 11 per sec from laptop in local env in one thread, so there is no need
					 *  **  to sleep unless for example we are running 4 peers while the target tps is set to less than 40!
					 *  **	Using multiple threads will increase the rate (maybe up to 180/sec), so we should check the
					 *  **	elapsed time and sleep only as much as needed.
					 *  **	e.g., if tps = 40, then we should allow the peers network 25 seconds to process 1000 transactions;
					 *  **	e.g., if elapsed time since the last sleep was 22 seconds, then sleep for 3 secs.

					// Here in thread code, multiply sleep_secs by the number of concurrent peers
					// (which are all submitting transactions/blocks at the same rate) to get the sleeptime.
					// Each peer will hit this; the throughput rate is what can be handled by the network as a whole, not each peer.
					if nobodySleeping {
						nobodySleeping = false
						elapsed := time.Since(curTime)
						accum := time.Since(startTime)
						curTime = time.Now() 	// doing this here means we include the sleep time within the next cycle elapsed time
						Logger(fmt.Sprintf("==== %d Tx accumulated (discovered by client %d). Elapsed Time prev=%s, accum=%s", (lstCounter-startCount), clientThread, elapsed, accum))
						Logger(fmt.Sprintf("Client %d myTx=%d going to sleep %d secs", clientThread, i+1, sleepSecs))
						Sleep( sleepSecs )
						nobodySleeping = true
					} else {
						Logger(fmt.Sprintf("Client %d myTx=%d going to sleep %d secs", clientThread, i+1, sleepSecs))
						Sleep( sleepSecs )
					}
					*********************************  */
				}
			}
			Logger(fmt.Sprintf("========= Finished CLIENT %d thread on peer %s, Tx=%d", clientThread, peerNum, i))
			wg.Done()
		}(t)
	}
}

// I think we can probably use this for all now, and get rid of the two functions above...
func InvokeAllThreadsOnAllPeers() {

	//  Number of NUM_CLIENTS = Number of threads : this func creates multiple client threads on each peer

	startCount := lstCounter
	startTime := time.Now()
	curTime := time.Now()
	nobodySleeping := true

	// All clients are submitting transactions as a whole; they all increment the shared counter.
	// The throughput rate is what can be handled by the network as a whole.
	// If we assume our tests take zero time to run and there are no transmission delays,
	// then "secsPerTxGroup" is the time we should force each client to sleep/delay at each juncture,
	// to keep pace so all accumlated thread invokes occur at the THROUGHPUT_RATE in the network.

	var secsPerTxGroup int
	secsPerTxGroup = SLEEP_SECS 	// for monitoring the shared "lstCounter"

	for t := 0; t < NUM_CLIENTS ; t++ {
		go func(clientThread int) {

			// Determine the peer to run this client
			// e.g. to use 3 of the 4 peers, and 10 clients: clients numbered 0..9 are started on peers: 0 1 2 0 1 2 0 1 2 0
			peerNum := clientThread % NUM_PEERS

		 	// Get a user on selected peer; probably we will be using the same user on each peer for all its clients
			username := peernetwork.GetAUserFromPeer(peerNetworkSetup, peerNum)
			if username == "" { panic(fmt.Sprintf("Cannot find a user on peer %d", peerNum)) }

			var i int64
			var numTxOnThisClient int64
			numTxOnThisClient = TX_COUNT / int64(NUM_CLIENTS)
			if clientThread == 0 { numTxOnThisClient = numTxOnThisClient + (TX_COUNT % int64(NUM_CLIENTS)) }
			Logger(fmt.Sprintf("========= Started CLIENT %d thread on peer %d, to run %d Tx", clientThread, peerNum, numTxOnThisClient))
			for i = 0; i < numTxOnThisClient; i++ {
				invokeChaincodeWithUser(username)	// this function increments counter too, and sends invoke to the peer that contains the given user
				if (lstCounter-startCount) % BUNDLE_OF_TRANSACTIONS == 0 {
					// Logger(fmt.Sprintf("==== %d Tx accumulated (discovered by client %d), total elapsed: %s", (lstCounter-startCount), clientThread, time.Since(startTime)))
					//  *****************************
					//  **	9/20/16 - We can run only 11 per sec from laptop in local env in one thread, so there is no need
					//  **  to sleep unless for example we are running 4 peers while the target tps is set to less than 40!
					//  **	Using multiple threads will increase the rate (maybe up to 180/sec), so we should check the
					//  **	elapsed time and sleep only as much as needed.
					//  **	e.g., if tps = 40, then we should allow the peers network 25 seconds to process 1000 transactions;
					//  **	e.g., if elapsed time since the last sleep was 22 seconds, then sleep for 3 secs.

					accum := time.Since(startTime)
					elapsed := time.Since(curTime)
					elapsedSecs := elapsed.Seconds()
					sleepSecs := int64(0)
					if elapsedSecs < float64(secsPerTxGroup) { sleepSecs = int64(float64(secsPerTxGroup) - elapsedSecs) }

					// Here in thread code, multiply sleep_secs by the number of concurrent peers
					// (which are all submitting transactions/blocks at the same rate) to get the sleeptime.
					// Each peer will hit this; the throughput rate is what can be handled by the network as a whole, not each peer.

					if nobodySleeping {

						nobodySleeping = false
						//  Logger(fmt.Sprintf("==== %d Tx accumulated (discovered by client %d). Elapsed Time prev=%s, accum=%s", (lstCounter-startCount), clientThread, elapsed, accum))
						//  Logger(fmt.Sprintf("Client %d myTx=%d, sleepSecs = %d", clientThread, i+1, sleepSecs))

						if (lstCounter-startCount) % (BUNDLE_OF_TRANSACTIONS * 10) == 0 {
							Logger(fmt.Sprintf("%d=Tx prev=%s accum=%s (client=%d myTx=%d, sleep=%d)", (lstCounter-startCount), elapsed, accum, clientThread, i+1, sleepSecs))
						}
						if sleepSecs > 0 { Sleep( sleepSecs ) }

						// setting curTime here after sleeping means we include the sleep time within the current/prior cycle of elapsed time, which means the other clients will also use this as their "previous started time" and compute their own sleep time similar to this first client
						curTime = time.Now()

						nobodySleeping = true

					} else {
						//  Logger(fmt.Sprintf("Client %d myTx=%d, sleepSecs = %d", clientThread, i+1, sleepSecs))
						if sleepSecs > 0 { Sleep( sleepSecs ) }
					}
				}
			}
			Logger(fmt.Sprintf("========= Finished CLIENT %d thread on peer %d, Tx=%d, Elapsed Time prev=%s accum=%s", clientThread, peerNum, i, time.Since(curTime), time.Since(startTime)))
			wg.Done()
		}(t)
	}
}

//Execution starts here ...
func RunLedgerStressTest(testname string, numClients int, numPeers int, numTx int64) {
	TESTNAME = testname
	InitLogger(TESTNAME)
	NUM_CLIENTS = numClients
	NUM_PEERS = numPeers
	TX_COUNT = numTx

	if NUM_CLIENTS > MAX_CLIENTS { Logger(fmt.Sprintf("Test not supported yet for more than %d clients", MAX_CLIENTS)); return }

	// time to measure overall execution of the testcase
	defer TimeTracker(time.Now(), "Total execution time for " + TESTNAME)

	Init()
	Logger("========= Transactions execution started =========")

	InvokeAllThreadsOnAllPeers()	// use this for multiple threads on multiple peers, such as TwoClientsPerFourPeers (total 8 threads)

    /*  ***********
	if NUM_CLIENTS > 4 {
		InvokeAllThreadsOnAllPeers()	// use this for multiple threads on multiple peers, such as TwoClientsPerFourPeers (total 8 threads)
	} else {
		if numPeers > 1 || NUM_CLIENTS == 1 {
			InvokeSingleThreadsOnPeers()	// use this for tests using single thread per peer, e.g. FourClientsFourPeers
		} else {
			InvokeMultiThreads()		// use this for multiple threads sending requests to one peer, such as TwoClientsOnePeer
		}
	}
    ************** */

	wg.Wait()
	Logger("========= Transactions execution ended =========")
	TearDown(peerNetworkSetup)
}
