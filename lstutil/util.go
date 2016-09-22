package lstutil 	// Ledger Stress Testing utility functions

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"time"
	"obcsdk/chaincode"
	"obcsdk/peernetwork"
)

// A Utility program, contains several utility methods that can be used across
// test programs
const (
	// CHAINCODE_NAME = "example02"
	// CHAINCODE_NAME = "chaincode_example02_addrecs"
	CHAINCODE_NAME = "mycc"
	INIT           = "init"
	INVOKE         = "invoke"
	QUERY          = "query"
	DATA           = "Yh1WWZlw1gGd2qyMNaHqBCt4zuBrnT4cvZ5iMXRRM3YBMXLZmmvyVr0ybWfiX4N3UMliEVA0d1dfTxvKs0EnHAKQe4zcoGVLzMHd8jPQlR5ww3wHeSUGOutios16lxfuQTdnsFcxhXLiGwp83ahyBomdmJ3igAYTyYw2bwXqhBeL9fa6CTK43M2QjgFhQtlcpsh7XMcUWnjJhvMHAyH67Z8Ugke6U8GQMO5aF1Oph0B2HlIQUaHMq2i6wKN8ZXyx7CCPr7lKnIVWk4zn0MLZ16LstNErrmsGeo188Rdx5Yyw04TE2OSPSsaQSDO6KrDlHYnT2DahsrY3rt3WLfBZBrUGhr9orpigPxhKq1zzXdhwKEzZ0mi6tdPqSzMKna7O9STstf2aFdrnsoovOm8SwDoOiyqfT5fc0ifVZSytVNeKE1C1eHn8FztytU2itAl1yDYSfTZQv42tnVgDjWcLe2JR1FpfexVlcB8RUhSiyoThSIFHDBZg8xyULPmp4e6acOfKfW2BXh1IDtGR87nBWqmytTOZrPoXRPq2QXiUjZS2HflHJzB0giDbWEeoZoMeF11364Xzmo0iWsBw0TQ2cHapS4cR49IoEDWkC6AJgRaNb79s6vythxX9CqfMKxIpqYAbm3UAZRS7QU7MiZu2qG3xBIEegpTrkVNneprtlgh3uTSVZ2n2JTWgexMcpPsk0ILh10157SooK2P8F5RcOVrjfFoTGF3QJTC2jhuobG3PIXs5yBHdELe5yXSEUqUm2ioOGznORmVBkkaY4lP025SG1GNPnydEV9GdnMCPbrgg91UebkiZsBMM21TZFbUqP70FDAzMWZKHDkDKCPoO7b8EPXrz3qkyaIWBymSlLt6FNPcT3NkkTfg7wl4DZYDvXA2EYu0riJvaWon12KWt9aOoXig7Jh4wiaE1BgB3j5gsqKmUZTuU9op5IXSk92EIqB2zSM9XRp9W2I0yLX1KWGVkkv2OIsdTlDKIWQS9q1W8OFKuFKxbAEaQwhc7Q5Mm"
)

var TESTNAME string
var logEnabled bool
var logFile *os.File

// Called in teardown methods to messure and display over all execution time
func TimeTracker(start time.Time, info string) {
	elapsed := time.Since(start)
	Logger(fmt.Sprintf("========= %s is %s", info, elapsed))
	CloseLogger()
}

func GetChainHeight(url string) int {
	height := chaincode.Monitor_ChainHeight(url)
	Logger(fmt.Sprintf("=========  Chaincode Height on "+url+" is : %d", height))
	return height
}

// This is a helper function to generate a random string of the requested length
// This is to make each Deploy transaction unique
func RandomString(strlen int) string {
	rand.Seed(time.Now().UTC().UnixNano())
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, strlen)
	for i := 0; i < strlen; i++ {
		result[i] = chars[rand.Intn(len(chars))]
	}
	return string(result)
}

// This is to make a standard Deploy transaction, for re-use
func FixedString(strlen int) string {
	rand.Seed(time.Now().UTC().UnixNano())
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, strlen)
	for i := 0; i < strlen; i++ {
		result[i] = chars[ i % (len(chars)) ]
	}
	return string(result)
}

// Utility function to deploy chaincode available @ http://urlmin.com/4r76d
func DeployChaincode(mynetwork peernetwork.PeerNetwork) (cntr int64) {		// ledger stress test (LST) chaincode
	var funcArgs = []string{CHAINCODE_NAME, INIT}
	cntr = 0
	// Use FixedString, or hardcoded DATA, to generate the same string every time we run this test
	// so that we simply redeploy / reuse same chaincode instance, without creating multiple deployments on the network!
	myStr := FixedString(1024)
	//myStr := RandomString(1024) 	// OR, we could use this to create a unique string each time this test is run, for a new deployment hash
	var chaincodeDeployArgs = []string{"a0", myStr, "counter", "0"}
	deployID, err := chaincode.DeployWithNetwork(mynetwork, funcArgs, chaincodeDeployArgs)
	if err != nil {
		Logger(fmt.Sprintf("lstutil.DeployChaincode(): Time to PANIC! chaincode.DeployWithNetwork returned (deployID=%s) and (Non-nil error=%s)\n", deployID, err))
		panic(err)
	}
	var sleepTime int64
	sleepTime = 60
	// Wait for deploy to complete; sleep based on network environment:  Z | LOCAL [default]
	// Increase sleep from 60 secs (works in LOCAL network, the defalt) by 60 to sum of 120 secs in external/remote networks
	ntwk := os.Getenv("NETWORK")
	if ntwk != "" && ntwk != "LOCAL" { sleepTime += 60 }
	Logger(fmt.Sprintf("<<<<<< DeployID=%s. cntr=%d. Need to give it some time; sleep for %d secs >>>>>>", deployID, cntr, sleepTime))
	Sleep(sleepTime)
	return cntr
}

// Utility function to query on chaincode available @ http://urlmin.com/4r76d
func QueryChaincode(mynetwork peernetwork.PeerNetwork, counter int64) (aVal, counterIndexStr string) {
	var arg1 = []string{CHAINCODE_NAME, QUERY}
	counterIndexStr, _ = chaincode.QueryWithNetwork(mynetwork, arg1, []string{"counter"})
	var aKey = []string{"a" + strconv.FormatInt(counter, 10)}
	aVal, _ = chaincode.QueryWithNetwork(mynetwork, arg1, aKey)
	return aVal, counterIndexStr
}

func GetChaincodeValuesOnHost(mynetwork peernetwork.PeerNetwork, peerNum int) (aVal, counterIndexStr string) {
        if peerNum >= len(mynetwork.Peers) { panic("peerNum does not exist in network") }
        peername := mynetwork.Peers[peerNum].PeerDetails["name"]
		//Logger(fmt.Sprintf("----------GetChaincodeValuesOnHost peerNum=%d peername=%s", peerNum, peername))
	var arg1 = []string{CHAINCODE_NAME, QUERY, peername}
	counterIndexStr, _ = chaincode.QueryOnHostWithNetwork(mynetwork, arg1, []string{"counter"})
		//Logger(fmt.Sprintf("----------GetChaincodeValuesOnHost retrieved counter=%s", counterIndexStr))
	var aKey = []string{"a" + counterIndexStr}
	aVal, _ = chaincode.QueryOnHostWithNetwork(mynetwork, arg1, aKey)
  		//Logger(fmt.Sprintf("----------GetChaincodeValuesOnHost counter=%s, query (using aKey=%s) result=%s", counterIndexStr, aKey[0], aVal))
	return aVal, counterIndexStr
}

func QueryChaincodeOnHost(mynetwork peernetwork.PeerNetwork, peerNum int, counter int64) (aVal, counterIndexStr string) {
        if peerNum >= len(mynetwork.Peers) { panic("peer does not exist in network") }
        peername := mynetwork.Peers[peerNum].PeerDetails["name"]
	var arg1 = []string{CHAINCODE_NAME, QUERY, peername}

	counterIndexStr, _ = chaincode.QueryOnHostWithNetwork(mynetwork, arg1, []string{"counter"})
	var aKey = []string{"a" + strconv.FormatInt(counter, 10)}

	aVal, _ = chaincode.QueryOnHostWithNetwork(mynetwork, arg1, aKey)
	// Logger(fmt.Sprintf("---QueryChaincodeOnHost counter=%s, query (using aKey=%s) result aVal=%s", counterIndexStr, aKey[0], aVal))
	return aVal, counterIndexStr
}

func Sleep(secs int64) {
	time.Sleep(time.Second * time.Duration(secs))
}

func InitLogger(fileName string) {
	layout := "Jan__2_2006"
	// Format Now with the layout const.
	t := time.Now()
	res := t.Format(layout)
	var err error
	logFile, err = os.OpenFile(res+"-"+fileName+".txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		panic(fmt.Sprintf("error opening file: %s", err))
	}

	logEnabled = true
	log.SetOutput(logFile)
	//log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.SetFlags(log.LstdFlags)
}

func Logger(printStmt string) {
	fmt.Println(printStmt)
	if !logEnabled {
		return
	}
	//TODO: Should we disable logging ?
	log.Println(printStmt)
}

func CloseLogger() {
	if logEnabled && logFile != nil {
		logFile.Close()
	}
}

//Cleanup methods to display useful information for LST tests
func TearDown(mynetwork peernetwork.PeerNetwork) {
	Sleep(10)
	testPassed := false
	_, cntrStr := QueryChaincode(mynetwork, lstCounter)
	Logger(fmt.Sprintf("========= After Query values lstCounter=%d, ledger cntr = <%s>\n", lstCounter, cntrStr))
	cntrVal, err := strconv.ParseInt(cntrStr, 10, 64)
	if err != nil { Logger(fmt.Sprintf("TearDown() Failed to convert cntr <%s> to int64\n Error: %s\n", cntrStr, err)) }

	//TODO: Block size again depends on the Block configuration in pbft config file
	//Test passes when 2 * block height match with total transactions, else fails

	if err == nil && cntrVal == lstCounter {
		testPassed = true
	}

	// keep rechecking, as long as there are no errors, and cntr keeps advancing (catching up processing backlog of invokes)

	var sleepSecs = int64(10) 	// sleep and query again every 10 secs
	prevCntr := int64(0)
	for  err == nil && cntrVal < lstCounter && prevCntr != cntrVal {
		Logger(fmt.Sprintf("current ledger counter %d DOES NOT MATCH expected lstCounter %d ; wait %d secs to try to catch up and recheck", cntrVal, lstCounter, sleepSecs))
		Sleep(sleepSecs)
		prevCntr = cntrVal
		_, cntrStr = QueryChaincode(mynetwork, lstCounter)
		//Logger(fmt.Sprintf("========= After Query values lstCounter=%d, ledger cntrIndex = <%s>\n", lstCounter, cntrStr))
		cntrVal, err = strconv.ParseInt(cntrStr, 10, 64)
		if err != nil {
			Logger(fmt.Sprintf("TearDown() Failed to convert cntrStr <%s> to int64\n ERROR: %s\n", cntrStr, err))
		} else if cntrVal == lstCounter {
			testPassed = true
		} else {
			// failure, but no error, so retry (and let loop guards handle things)
		}
	}
	if testPassed {
		Logger(fmt.Sprintf("\n######### %s TEST PASSED ######### ledger counter = %d\n", TESTNAME, lstCounter))
	} else {
		Logger(fmt.Sprintf("\n######### %s TEST FAILED ######### ledger counter = %d, expected = %d #########\n", TESTNAME, cntrVal, lstCounter))
	}
}

func QueryAllHostsToGetCurrentCounter(mynetwork peernetwork.PeerNetwork, txName string, counter *int64) (querySuccess bool) {	// using ratnakar myCC - a modified example02
	// loop through and query all hosts to find consensus and determine what the current counter value is.
	querySuccess = true
	*counter = 0
	failedCount := 0
	N := peernetwork.GetNumberOfPeers(mynetwork)
	F := (N-1)/3
	currPeerCounterValue := make([]int64, N)
	for peerNumber := 0 ; peerNumber < N ; peerNumber++ {
        	_, counterIdxStr := GetChaincodeValuesOnHost(mynetwork, peerNumber)
        	Logger(fmt.Sprintf("-----QueryAllHostsToGetCurrentCounter: on peer %d, counter=%s", peerNumber, counterIdxStr))
        	newCounterValue, err := strconv.ParseInt(counterIdxStr, 10, 64)
        	if err != nil {
			Logger(fmt.Sprintf("-----QueryAllHostsToGetCurrentCounter() Failed to convert counterIdxStr <%s> to int64\n Error: %s\n", counterIdxStr, err))
			currPeerCounterValue[peerNumber] = 0
			failedCount++
		} else {
			currPeerCounterValue[peerNumber] = newCounterValue
		}
	}
	if failedCount > F {
		querySuccess = false
		Logger(fmt.Sprintf("%s TEST STARTUP FAILURE!!! Failed to query %d peers. RERUN when at least %d/%d peers are running, in order to be able to reach consensus.", txName, failedCount, ((N-1)/3)*2+1, N ))
	} else {
		var consensus_counter int64 = 0
		found_consensus := false
		for i := 0 ; i <= F ; i++ {
			i_val_cntr := 1
			for j := i+1 ; j < N ; j++ {
				if currPeerCounterValue[j] == currPeerCounterValue[i] { i_val_cntr++ }
			}
			if i_val_cntr >= N-F { consensus_counter = currPeerCounterValue[i]; found_consensus = true; break }
		}
		if found_consensus {
			*counter = consensus_counter
			Logger(fmt.Sprintf("%s TEST PASS STARTUP: %d/%d peers reached consensus: current count = %d", txName, N-failedCount, N, consensus_counter))
		} else {
			querySuccess = false
			Logger(fmt.Sprintf("%s TEST FAIL upon STARTUP: peers cannot reach consensus on current counter value!!!", txName))
		}
	}
	return querySuccess
}

func QueryAllHosts(mynetwork peernetwork.PeerNetwork, txName string, expected_count int64) (querySuccess bool) {	// using ratnakar myCC - a modified example02

	// loop through and query all hosts
	querySuccess = true
	failedCount := 0
	N := peernetwork.GetNumberOfPeers(mynetwork)
	for peerNumber := 0 ; peerNumber < N ; peerNumber++ {
		result := "SUCCESS"
        	valueStr, counterIdxStr := GetChaincodeValuesOnHost(mynetwork, peerNumber)
        	Logger(fmt.Sprintf("-----QueryAllHosts: on peer %d, counter=%s", peerNumber, counterIdxStr))
        	newVal, err := strconv.ParseInt(counterIdxStr, 10, 64)
        	if err != nil { Logger(fmt.Sprintf("QueryAllHosts: Failed to convert counterIdxStr <%s> recvd from GetChaincodeValuesOnHost to int64\n Error: %s\n", counterIdxStr, err)) }
        	if err != nil || newVal != expected_count {
			result = "FAILURE"
			failedCount++
		}
        	Logger(fmt.Sprintf("QueryAllHosts host=%d %s after %s: expected_count=%d, Actual a<%s> = <%s>", peerNumber, result, txName, expected_count, counterIdxStr, valueStr))
	}
	if failedCount > (N-1)/3 {
		querySuccess = false
		Logger(fmt.Sprintf("%s TEST FAIL!!!  TOO MANY PEERS (%d) FAILED to obtain the correct expected count (%d), so network consensus failed!!!\n(If fewer than %d/%d peers are running, then the network does not have enough running nodes to reach consensus.)", txName, failedCount, expected_count, ((N-1)/3)*2+1, N ))
	} else {
		Logger(fmt.Sprintf("%s TEST PASS.  %d/%d peers reached consensus on the correct count", txName, N-failedCount, N))
	}
	return querySuccess
}

