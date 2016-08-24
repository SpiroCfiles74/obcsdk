package main

/******************** Testing Objective consensu:STATE TRANSFER ********
*   Setup: 4 node local docker peer network with security
*   0. Deploy chaincodeexample02 with 100000, 90000 as initial args
*   1. Send Invoke Requests on multiple peers using go routines.
*   2. Verify query results match on vp0 and vp1 after invoke
*********************************************************************/

import (
	"fmt"
	"strconv"
	"time"

	"obcsdk/chaincode"
	//"obcsdk/peernetwork"
	"sync"
)

var inita, initb, curra, currb int
var ctrNumReq int

func main() {

	fmt.Println("Creating a local docker network")
	//peernetwork.SetupLocalNetwork(4, true)

	time.Sleep(60000 * time.Millisecond)
	_ = chaincode.InitNetwork()
	chaincode.InitChainCodes()
	chaincode.RegisterUsers()


	fmt.Println("\nPOST/Chaincode: Deploying chaincode at the beginning ....")
	dAPIArgs0 := []string{"example02", "init"}
	depArgs0 := []string{"a", "100000", "b", "90000"}
	chaincode.Deploy(dAPIArgs0, depArgs0)

	inita = 100000
	initb = 90000
	curra = inita
	currb = initb

	time.Sleep(60000 * time.Millisecond)
	fmt.Println("\nPOST/Chaincode: Querying a and b after deploy >>>>>>>>>>> ")
	qAPIArgs0 := []string{"example02", "query"}
	qArgsa := []string{"a"}
	qArgsb := []string{"b"}
	A, _ := chaincode.Query(qAPIArgs0, qArgsa)
	B, _ := chaincode.Query(qAPIArgs0, qArgsb)
	myStr := fmt.Sprintf("\nA = %s B= %s", A, B)
	fmt.Println(myStr)

        ctrNumReq = 0

        defer timeTrack(time.Now(), "Testcase execution Done")
        now := time.Now().Unix()
        endTime := now + 2 * 60
        for now < endTime {
            //fmt.Println("Time Now ", time.Now())
            //fmt.Println("end Time ", endTime)
            //fmt.Println("now ", now)
            schedulerTask()
            now = time.Now().Unix()
       } 

}

func schedulerTask() {

	numReq := 1000
	var wg sync.WaitGroup

        var k0, k1, k2, k3 int
	invArgs0 := []string{"a", "b", "1"}

        
        fmt.Println("Beginning of schedulerTask")
	wg.Add(4)
	go func() {

		defer wg.Done()
	        iAPIArgs_vp0 := []string{"example02", "invoke", "vp0"}
		k0 := 1
		for k0 <= numReq {
			chaincode.InvokeOnPeer(iAPIArgs_vp0, invArgs0)
	                //time.Sleep(5000 * time.Millisecond)
			k0++
		}
		fmt.Println("# of Req Invoked on vp0 ", k0)
	}()

	go func() {

		defer wg.Done()
		iAPIArgs_vp1 := []string{"example02", "invoke", "vp1"}
		k1 := 1
		for k1 <= numReq {
			chaincode.InvokeOnPeer(iAPIArgs_vp1, invArgs0)
	                //time.Sleep(5000 * time.Millisecond)
			k1++
		}
		fmt.Println("# of Req Invoked  on vp1", k1)
	}()

	go func() {

		defer wg.Done()
	        iAPIArgs_vp2 := []string{"example02", "invoke", "vp2"}
		k2 := 1
		for k2 <= numReq {
			chaincode.InvokeOnPeer(iAPIArgs_vp2, invArgs0)
	                //time.Sleep(5000 * time.Millisecond)
			k2++
		}
		fmt.Println("# of Req Invoked on vp2 ", k2)
	}()

	go func() {

		defer wg.Done()
	        iAPIArgs_vp3 := []string{"example02", "invoke", "vp3"}
		k3 := 1
		for k3 <= numReq {
			chaincode.InvokeOnPeer(iAPIArgs_vp3, invArgs0)
	                //time.Sleep(5000 * time.Millisecond)
			k3++
		}
		fmt.Println("# of Req Invoked on vp3", k3)
	}()

	wg.Wait()
        ctrNumReq = ctrNumReq + k0 + k1 + k2 + k3
        curra = curra - (k0 + k1 + k2 + k3)
        currb = currb + (k0 + k1 + k2 + k3)
        fmt.Println("Total # of Req Invoked ", ctrNumReq)
        fmt.Println("END of schedulerTask")

}

func QueryMatch(expectedA int, expectedB int) {

        var myStr string

        fmt.Println("BEGINNING of schedulerTask")
	fmt.Println("Sleeping for 1 minutes ")
	time.Sleep(60000 * time.Millisecond)

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

	fmt.Println("Results in a and b PEER0 : ", res0AI, res0BI)
	fmt.Println("Results in a and b PEER1 : ", res1AI, res1BI)
	fmt.Println("Results in a and b PEER2 : ", res2AI, res2BI)
	fmt.Println("Results in a and b PEER3 : ", res3AI, res3BI)

	//if res0AI == expectedA && res1AI == expectedA && res2AI == expectedA && res3AI == expectedA {
	if res0AI == res1AI && res2AI == res3AI && res3AI == res1AI {
		myStr = fmt.Sprintf("Values Verified for A: peer0: %d, peer1: %d, peer2: %d, peer3: %d", res0AI, res1AI, res2AI, res3AI)
		fmt.Println(myStr)
	} else {
		myStr := fmt.Sprintf("\n\nTEST FAIL: Results in A value DO NOT match on all Peers after ")
		fmt.Println(myStr)
	}

	//if res0BI == expectedB && res1BI == expectedB && res2BI == expectedB && res3BI == expectedB {
	if res0BI == res1BI && res2BI == res3BI && res3BI == res1BI {
		myStr = fmt.Sprintf("Values Verified for B: peer0: %d, peer1: %d, peer2: %d, peer3: %d\n\n", res0BI, res1BI, res2BI, res3BI)
		fmt.Println(myStr)
	} else {
		myStr := fmt.Sprintf("\n\n%TEST FAIL: Results in B value DO NOT match on all Peers after ")
		fmt.Println(myStr)
	}


}

func GetHeight() {
	fmt.Println("Sleeping for 2 minutes ")
	time.Sleep(60000 * time.Millisecond)
	ht0, _ := chaincode.GetChainHeight("vp0")
	ht1, _ := chaincode.GetChainHeight("vp1")
	ht2, _ := chaincode.GetChainHeight("vp2")
	ht3, _ := chaincode.GetChainHeight("vp3")
	fmt.Printf("ht0: %d, ht1: %d, ht2: %d, ht3: %d ", ht0, ht1, ht2, ht3)
}

func timeTrack(start time.Time, name string) {
        elapsed := time.Since(start)
	fmt.Printf("\n################# %s took %s \n", name, elapsed)
	time.Sleep(120000 * time.Millisecond)
        GetHeight()
        QueryMatch(curra, currb)
        fmt.Printf("\nValue in totalNumReq sent %d", ctrNumReq)
	fmt.Println("\n################# Execution Completed #################")
}

