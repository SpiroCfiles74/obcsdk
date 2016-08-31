package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"obcsdk/chaincode"
	"obcsdk/peernetwork"
	"obcsdk/util"
)

var peerNetworkSetup peernetwork.PeerNetwork
var AVal, BVal, curAVal, curBVal, invokeValue int64
var argA = []string{"a"}
var argB = []string{"counter"}
var counter int64

const(
	TRX_COUNT = 20000
)

func initNetwork() {
	fmt.Println("========= Init Network =========")
	peernetwork.GetNC_Local()
	peerNetworkSetup = chaincode.InitNetwork()
	chaincode.InitChainCodes()
	fmt.Println("========= Register Users =========")
	chaincode.RegisterUsers()
}

func invokeChaincode(peer string ) {
	counter++
	arg1Construct := []string{"concurrency", util.INVOKE, peer}
	arg2Construct := []string{"a" + strconv.FormatInt(counter, 10), util.DATA, "counter"}

	_,_ = chaincode.InvokeOnPeer(arg1Construct, arg2Construct)
}

func Init() {
	//initialize
	done := make(chan bool, 1)
	counter = 0
	// Setup the network based on the NetworkCredentials.json provided
	initNetwork()

	//Deploy chaincode
	util.DeployChaincode(done)
}

func InvokeLoop() {
	  curTime := time.Now()
		for j := 1; j <= TRX_COUNT/2; j++ {
			if counter > 1 && counter%1000 == 0 {
				elapsed := time.Since(curTime)
				fmt.Println("=========>>>>>> Iteration#", counter, " Time: ", elapsed)
				curTime = time.Now()
			}
			//TODO: Change this to InvokeAsUser ?
			invokeChaincode("PEER0")
			invokeChaincode("PEER1")
		}
}

//Cleanup methods to display useful information
func tearDown(url string) {
	var errTransactions int64
	errTransactions = 0
	fmt.Println("....... State transfer is happening, Lets take a nap for 2 mins ......")
	util.Sleep(120)
	val1, val2 := util.QueryChaincode(counter)
	fmt.Println("========= After Query Vals A = ",val1," \n B = ",  val2,"\n")

/*	height := getChainHeight(url) //Remove hardcoding ??
	fmt.Println("========= Total Blocks #", height)
	for i := 1; i < height; i++ {
		//TODO: Don't hard code IP , can we take this as argument ?
		nonHashData := chaincode.ChaincodeBlockTrxInfo(url, i)
		length := len(nonHashData.TransactionResult)
		for j := 0; j < length; j++ {
			if nonHashData.TransactionResult[j].ErrorCode > 0 {
				fmt.Printf("\n========= Block[%d] Trx#[%s] UUID [%d] ErrorCode [%d] Error: %s\n", i, counter, nonHashData.TransactionResult[j].Uuid, nonHashData.TransactionResult[j].ErrorCode, nonHashData.TransactionResult[j].Error)
				errTransactions++
			}
		}
	}
	*/
	if errTransactions > 0 {
		fmt.Println("========= Failed transactions  #", errTransactions)
	}
	fmt.Println("========= Successful transactions  #", counter-errTransactions)

	newVal,err := strconv.ParseInt(val2, 10, 64);

	if  err != nil {
			fmt.Println("Failed to convert ",val2," to int64\n Error: ", err)
	}

	//TODO: Block size again depends on the Block configuration in pbft config file
	//Test passes when 2 * block height match with total transactions, else fails
	if (newVal == counter) {
		fmt.Println("######### TEST PASSED #########")
	} else {
		fmt.Println("######### TEST FAILED #########")
	}

}

//Execution starts here ...
func main() {
	//TODO:Add support similar to GNU getopts, http://goo.gl/Cp6cIg
	if len(os.Args) <  2{
		fmt.Println("Usage: go run LedgerStressTwoCliOnePeer.go Utils.go <http://IP:PORT>")
		return;
	}
	//TODO: Have a regular expression to check if the give argument is correct format
	if !strings.Contains(os.Args[1], "http://") {
		fmt.Println("Error: Argument submitted is not right format ex: http://127.0.0.1:5000 ")
		return;
	}
	//Get the URL
	url := os.Args[1]

	// time to messure overall execution of the testcase
	defer util.TimeTracker(time.Now(), "Total execution time for LedgerStressTwoCliTwoPeer.go ")

	Init()
	InvokeLoop()
	tearDown(url);
}
