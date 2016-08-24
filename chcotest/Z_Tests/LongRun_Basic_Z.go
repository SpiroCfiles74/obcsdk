package main

/******************** Testing Objective consensu: Sending Invoke Requests in Parallel ********
*   Setup: 8 node local docker peer network with security
*   0. Deploy chaincodeexample02 with 100000000, 1 as initial args
*   1. Send Invoke Requests in parallel on multiple peers using go routines.
    2. Logic is set so that one complete set of invoke requests on 8 peers would send 144 requests.
*   3. After each such one loop, verify query results match on PEER0 and PEER1
*********************************************************************/

import (
	"fmt"
	"strconv"
	"time"
        "os"
        "bufio"
        
	"obcsdk/chaincode"
	"obcsdk/peernetwork"
)

var f *os.File
var writer *bufio.Writer


func main() {

	var myNetwork peernetwork.PeerNetwork


        var err error
        f, err = os.OpenFile("/tmp/hyperledgerBetaTestrun_Output", os.O_RDWR|os.O_APPEND, 0660)
        if ( err != nil) {
          fmt.Println("Output file does not exist creating one ..")
          f, err = os.Create("/tmp/hyperledgerBetaTestrun_Output")
        }
        //check(err)
        defer f.Close()
        writer = bufio.NewWriter(f)

	fmt.Println("Creating a local docker network")
	peernetwork.SetupLocalNetwork(4, true)

	//with the InvokeLoop designed with this 8 peer 4 req setup we would have 144 requests in one round
	numPeers := 4
	numReq := 4

	time.Sleep(60000 * time.Millisecond)
	myNetwork = chaincode.InitNetwork()
	chaincode.InitChainCodes()
	chaincode.RegisterUsers()

	//get a URL details to get info n chainstats/transactions/blocks etc.
	aPeer, _ := peernetwork.APeer(myNetwork)
	url := "https://" + aPeer.PeerDetails["ip"] + ":" + aPeer.PeerDetails["port"]

	fmt.Println("Peers on network ")
	chaincode.NetworkPeers(url)

	fmt.Println("\nPOST/Chaincode: Deploying chaincode at the beginning ....")

	dAPIArgs0 := []string{"example02", "init"}
	depArgs0 := []string{"a", "100000000", "b", "0"}
	chaincode.Deploy(dAPIArgs0, depArgs0)

	time.Sleep(60000 * time.Millisecond)
	fmt.Println("\nPOST/Chaincode: Querying a and b after deploy >>>>>>>>>>> ")

	qAPIArgs0 := []string{"example02", "query"}
	qArgsa := []string{"a"}
	qArgsb := []string{"b"}
	A, _ := chaincode.Query(qAPIArgs0, qArgsa)
	B, _ := chaincode.Query(qAPIArgs0, qArgsb)
	myStr := fmt.Sprintf("\nA = %s B= %s", A, B)
	fmt.Println(myStr)

	defer timeTrack(time.Now(), "Testcase executiion Done")
	InvokeLoop(numPeers, numReq)
	time.Sleep(time.Minute * time.Duration(5))

}

func InvokeLoop(numPeers int, numReq int) {

	var inita, initb, curra, currb int
	inita = 100000000
	initb = 0
	curra = inita
	currb = initb

	invArgs0 := []string{"a", "b", "1"}
	//i := 0
	for {
		j := 0
		k := 1
		for j < numPeers {
			k = 1
			currPeer := "vp" + strconv.Itoa(j)
			iAPIArgsCurrPeer := []string{"example02", "invoke", currPeer}
			for k <= numReq {
				go chaincode.InvokeOnPeer(iAPIArgsCurrPeer, invArgs0)
				k++
			}
			m := j - 1
			for m >= 0 {
				loopPeer := "vp" + strconv.Itoa(m)
				iAPIArgsLoopPeer := []string{"example02", "invoke", loopPeer}
				k = 1
				for k <= numReq {
					go chaincode.InvokeOnPeer(iAPIArgsLoopPeer, invArgs0)
					k++
				}
				m = m - 1
			}
			j++
		}
		curra = curra - 60 
		currb = currb + 60 
		QueryMatch(curra, currb)
		//i++
	}
}

func QueryMatch(curra int, currb int) {

	fmt.Println("Inside Query match ********************************* &&&&&&&&&&&&&& %%%%%%%%%%%%%%%%")
	fmt.Println("Sleeping for 2 minutes ")
	time.Sleep(120000 * time.Millisecond)

	fmt.Println("\nPOST/Chaincode: Querying a and b after invoke >>>>>>>>>>> ")
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

	if (curra == res0AI) && (currb == res0BI) {
		fmt.Println("Results in a and b match with Invoke values on vp0: PASS")
		valueStr := fmt.Sprintf(" curra : %d, currb : %d, resa : %d , resb : %d", curra, currb, res0AI, res0BI)
		fmt.Println(valueStr)
	} else {
		fmt.Println("******************************")
		fmt.Println("RESULTS DO NOT MATCH on vp0 : FAIL")

		fmt.Println("******************************")
	}

	if (curra == res1AI) && (currb == res1BI) {
		fmt.Println("Results in a and b match with Invoke values on vp1: PASS")
		valueStr := fmt.Sprintf(" curra : %d, currb : %d, resa : %d , resb : %d", curra, currb, res1AI, res1BI)
		fmt.Println(valueStr)
	} else {
		fmt.Println("******************************")
		fmt.Println("RESULTS DO NOT MATCH on vp1 : FAIL")
		fmt.Println("******************************")
	}


	if (curra == res2AI) && (currb == res2BI) {
                fmt.Println("Results in a and b match with Invoke values on vp2: PASS")
                valueStr := fmt.Sprintf(" curra : %d, currb : %d, resa : %d , resb : %d", curra, currb, res2AI, res2BI)
                fmt.Println(valueStr)
        } else {
                fmt.Println("******************************")
                fmt.Println("RESULTS DO NOT MATCH on vp2 : FAIL")
                fmt.Println("******************************")
        }


	if (curra == res3AI) && (currb == res3BI) {
                fmt.Println("Results in a and b match with Invoke values on vp3: PASS")
                valueStr := fmt.Sprintf(" curra : %d, currb : %d, resa : %d , resb : %d", curra, currb, res3AI, res3BI)
                fmt.Println(valueStr)
        } else {
                fmt.Println("******************************")
                fmt.Println("RESULTS DO NOT MATCH on vp3 : FAIL")
                fmt.Println("******************************")
        }

}

func timeTrack(start time.Time, name string) {

	elapsed := time.Since(start)
	myStr := fmt.Sprintf("\n################# %s took %s \n", name, elapsed)
	fmt.Println(myStr)
	fmt.Fprintln(writer, myStr)
	myStr = fmt.Sprintf("################# Execution Completed #################")
	fmt.Fprintln(writer, myStr)
	fmt.Println(myStr)
	writer.Flush()
}
