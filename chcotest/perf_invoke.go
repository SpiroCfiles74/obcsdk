package main

import (
	"bufio"
	"fmt"
        "strconv"
	"obcsdk/chaincode"
	"obcsdk/peernetwork"
	"os"
	//"strings"
	"time"
)

var f *os.File
var writer *bufio.Writer
var myNetwork peernetwork.PeerNetwork
var url string
var testRunStatus bool

func getNowMillis() int64 {
	nanos := time.Now().UnixNano()
	return nanos / 1000000
}

func main() {


	myStr := fmt.Sprintf("\n\n*********** BEGIN BASICFUNC.go ***************")
	fmt.Println(myStr)

        setupNetwork()


	deploy()
	fmt.Printf("DEPLOYED .... \n")
	time.Sleep(60000 * time.Millisecond)


	fmt.Printf("QUERYING once .... \n")
        qAPIArgs00 := []string{"example02", "query", "vp1"}
        qArgsa := []string{"a"}
        res0A, qerr := chaincode.QueryOnHost(qAPIArgs00, qArgsa)
	if qerr != nil {
		fmt.Printf("Could not query %s\n", qerr)
		return
	}
        res0AI, _ := strconv.Atoi(res0A)
	fmt.Printf("QUERY result %d .... \n", res0AI)
	
        invoke(1000, "vp1")
        //curra := res0AI - 1

        //time.Sleep(10000 * time.Millisecond)

        //query(100, "INVOKE", curra, "vp1")

	fmt.Printf("DONE\n")
}

func setupNetwork() {

        fmt.Println("Working with an existing network")
        myNetwork = chaincode.InitNetwork()
        chaincode.InitChainCodes()
        chaincode.RegisterUsers()

        time.Sleep(10000 * time.Millisecond)


}

func deploy() {

	fmt.Println("\nPOST/Chaincode: Deploying chaincode at the beginning ....")
	dAPIArgs0 := []string{"example02", "init"}
	depArgs0 := []string{"a", "10000", "b", "0"}
	chaincode.Deploy(dAPIArgs0, depArgs0)

}
func invoke(iter int, peerName string)  {

	start := getNowMillis()
	fmt.Printf("Starting: %d\n", start)
	for i := 0; i < iter; i++ {
		iAPIArgs0 := []string{"example02", "invoke", peerName}
		invArgs0 := []string{"a", "b", "1"}
		_, invErr := chaincode.InvokeOnPeer(iAPIArgs0, invArgs0)
		if invErr != nil {
			fmt.Printf("Could not invoke %s\n", invErr)
			return
		}
        }
	end := getNowMillis()
	fmt.Printf("Ending: %d\n", end)
	fmt.Printf("Elapsed : %d millis for %d invokes\n", end-start, iter)

}

func getHeight() {

	ht0, _ := chaincode.GetChainHeight("vp0")
	ht1, _ := chaincode.GetChainHeight("vp1")
	ht2, _ := chaincode.GetChainHeight("vp2")
	ht3, _ := chaincode.GetChainHeight("vp3")

	if (ht0 == 3) && (ht1 == 3) && (ht2 == 3) && (ht3 == 3) {
		myStr := fmt.Sprintf("\n\nGET CHAIN HEIGHT TEST PASS : Results in A value match on all Peers after ")
		fmt.Println(myStr)
		myStr = fmt.Sprintf("Height Verified: ht0: %d, ht1: %d, ht2: %d, ht3: %d ", ht0, ht1, ht2, ht3)
		fmt.Println(myStr)
	} else {
		fmt.Printf(" All heights do NOT match : ht0: %d, ht1: %d, ht2: %d, ht3: %d ", ht0, ht1, ht2, ht3)
		myStr := fmt.Sprintf("\n\nGET CHAIN HEIGHT TEST FAIL : value in chain height match on all Peers after deploy and single invoke")
		fmt.Println(myStr)
		myStr = fmt.Sprintf("Height Verified: ht0: %d, ht1: %d, ht2: %d, ht3: %d ", ht0, ht1, ht2, ht3)
		fmt.Println(myStr)
	}

}

func query(iter int, txName string, expectedA int, peerName string) {

        qAPIArgs00 := []string{"example02", "query", peerName}
        qArgsa := []string{"a"}

	start := getNowMillis()
	fmt.Printf("Starting: %d\n", start)
	for i := 0; i < iter; i++ {
            res0A, _ := chaincode.QueryOnHost(qAPIArgs00, qArgsa)
            res0AI, _ := strconv.Atoi(res0A)


            if res0AI != expectedA {
                myStr := fmt.Sprintf("\n\n%s TEST FAIL: Results in A value DO NOT match on %s", txName, peerName)
                fmt.Println(myStr)
		return
            }
	}
	end := getNowMillis()
	fmt.Printf("Ending: %d\n", end)
	fmt.Printf("Elapsed : %d millis for %d queries\n", end-start, iter)
}



func timeTrack(start time.Time, name string) {
	elapsed := time.Since(start)
        if (testRunStatus == true ) {
          myStr := fmt.Sprintln("\n*************** TEST BASICFUNC.go PASSED *********************** \n")
          fmt.Fprintln(writer, myStr)
        }
        if (testRunStatus == false ) {
          myStr := fmt.Sprintln("\n*************** TEST BASICFUNC.go FAILED *********************** \n")
          fmt.Fprintln(writer, myStr)
        }

	myStr := fmt.Sprintf("\n################# %s took %s \n", name, elapsed)
	fmt.Fprintln(writer, myStr)
	fmt.Println(myStr)
	myStr = fmt.Sprintf("\n\n*********** END BASICFUNC.go ***************\n\n")
	fmt.Println(myStr)
	fmt.Fprintln(writer, myStr)
	writer.Flush()
}

