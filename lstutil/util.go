package lstutil 	// Ledger Stress Testing utility functions

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"time"
	"obcsdk/chaincode"
	"obcsdk/threadutil"
)

// A Utility program, contains several utility methods that can be used across
// test programs
const (
	// CHAINCODE_NAME = "example02"
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

// Utility function to deploy chaincode available @ http://urlmin.com/4r76d
func DeployChaincode() (cntr int64) {
	var funcArgs = []string{CHAINCODE_NAME, INIT}
	cntr = 0
	var chaincodeDeployArgs = []string{"a", RandomString(1024), "counter", strconv.Itoa(int(cntr))}
	//call chaincode deploy function to do actual deployment
	deployID, err := chaincode.Deploy(funcArgs, chaincodeDeployArgs)
	if err != nil {
		Logger(fmt.Sprintf("DeployChaincode() returned (deployID=%s) and (Non-nil error=%s). Time to panic!\n", deployID, err))
		panic(err)
	}
	var sleepTime int64
	sleepTime = 30
	// Wait for deploy to complete; sleep based on network environment:  Z | LOCAL [default]
	// Increase sleep from 30 secs (works in LOCAL network, the defalt) by 90 to sum of 120 secs in "Z" (or anything else)
	ntwk := os.Getenv("NETWORK")
	if ntwk != "" && ntwk != "LOCAL" { sleepTime += 90 }
	Logger(fmt.Sprintf("<<<<<< DeployID=%s. Need to give it some time; sleep for %d secs >>>>>>", deployID, sleepTime))
	Sleep(sleepTime)
	return cntr
}

// Utility function to invoke on chaincode available @ http://urlmin.com/4r76d
/*func InvokeChaincode(counter int64) {
	arg1 := []string{CHAINCODE_NAME, INVOKE}
	arg2 := []string{"a" + strconv.FormatInt(counter, 10), data, "counter"}
	_, _ = chaincode.Invoke(arg1, arg2)
}*/

// Utility function to query on chaincode available @ http://urlmin.com/4r76d
func QueryChaincode(counter int64) (res1, res2 string) {
	var arg1 = []string{CHAINCODE_NAME, QUERY}
	var arg2 = []string{"a" + strconv.FormatInt(counter, 10)}
	val, _ := chaincode.Query(arg1, arg2)
	counterArg, _ := chaincode.Query(arg1, []string{"counter"})
	return val, counterArg
}

func QueryChaincodeOnHost(peerNum int, counter int64) (res1, res2 string) {
	var arg1 = []string{CHAINCODE_NAME, QUERY, threadutil.GetPeer(peerNum)}
	var arg2 = []string{"a" + strconv.FormatInt(counter, 10)}
	val, _ := chaincode.QueryOnHost(arg1, arg2)
	counterArg, _ := chaincode.Query(arg1, []string{"counter"})
	return val, counterArg
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

//Cleanup methods to display useful information
func TearDown(counter int64) {
	Sleep(10)
	val1, val2 := QueryChaincode(counter)
	Logger(fmt.Sprintf("========= After Query values counter=%d, a%s = %s\n", counter, val2, val1))
	newVal, err := strconv.ParseInt(val2, 10, 64)
	if err != nil { Logger(fmt.Sprintf("TearDown() Failed to convert val2 <%s> to int64\n Error: %s\n", val2, err)) }

	//TODO: Block size again depends on the Block configuration in pbft config file
	//Test passes when 2 * block height match with total transactions, else fails

	if err == nil && newVal == counter {
		Logger(fmt.Sprintf("\n######### %s TEST PASSED ######### Inserted %d records\n", TESTNAME, counter))
	} else {
		var sleepSecs = int64(120)	// TODO: calculate sleepSecs correctly, after moving more consts and functions into this file
		Logger(fmt.Sprintf("counter does not match; sleep and recheck after %d secs", sleepSecs))
		Sleep(sleepSecs)

		val1, val2 := QueryChaincode(counter)
		Logger(fmt.Sprintf("========= After Query values counter=%d, a%s = %s\n", counter, val2, val1))
		newVal, err := strconv.ParseInt(val2, 10, 64)
		if err != nil { Logger(fmt.Sprintf("TearDown() Failed to convert %s to int64\n ERROR: %s\n", val2, err)) }
		if err == nil && newVal == counter {
			Logger(fmt.Sprintf("\n######### %s TEST PASSED ######### Inserted %d records\n", TESTNAME, counter))
		} else {
			Logger(fmt.Sprintf("\n######### %s TEST FAILED ######### Inserted %d/%d records #########\n", TESTNAME, newVal, counter))
		}
	}
}
