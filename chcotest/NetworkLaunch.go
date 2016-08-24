package main

import (
	"bufio"
	"fmt"
	"obcsdk/chaincode"
	"obcsdk/peernetwork"
	"os"
	"time"
)

var f *os.File
var writer *bufio.Writer

func main() {

	var err error
	f, err = os.Create("/tmp/hyperledgerBetaTestrun_Output")
	check(err)
	defer f.Close()
	writer = bufio.NewWriter(f)

	myStr := fmt.Sprintf("\n\n*********** BEGIN NetworkLaunch.go ***************")
	fmt.Println(myStr)
	fmt.Fprintln(writer, myStr)



  i := 4
	for ( i < 16) {

	 t := time.Now()

		peernetwork.SetupLocalNetwork(i, false)

		myNetwork := chaincode.InitNetwork()
		chaincode.InitChainCodes()
		//peernetwork.PrintNetworkDetails(myNetwork)
		peernetwork.PrintNetworkDetails()

		numPeers := peernetwork.GetNumberOfPeers(myNetwork)

		if numPeers == i {

			fmt.Println("Test PASS")
			myStr = fmt.Sprintf("Launched Local Docker Network successfully with %d peers with pbft and security+privacy enabled\n", numPeers)
			fmt.Println(myStr)
			fmt.Fprintln(writer, myStr)
			myStr := fmt.Sprintf("NetworkLaunch Test PASSED")
			fmt.Fprintln(writer, myStr)

			} else {
				myStr = fmt.Sprintf("Failed to Launch Local Docker Network  with %d peers with pbft and security+privacy enabled\n", numPeers)
				fmt.Println(myStr)
				fmt.Fprintln(writer, myStr)

				myStr := fmt.Sprintf("NetworkLaunch Test FAILED")
				fmt.Fprintln(writer, myStr)
			}
      elapsed := time.Since(t)
			myStr := fmt.Sprintf("\nTime to launch %d peers: %s \n", i, elapsed)
			fmt.Println(myStr)
			fmt.Fprintln(writer, myStr)

			myStr = fmt.Sprintf("################# Execution Completed #################")
			fmt.Fprintln(writer, myStr)
			fmt.Println(myStr)
			writer.Flush()

			i++
		}
		myStr = fmt.Sprintf("*********** END NetworkLaunch.go ***************\n\n")
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

func check(e error) {
	if e != nil {
		panic(e)
	}
}
