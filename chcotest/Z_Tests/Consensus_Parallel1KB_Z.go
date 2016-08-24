package main

/******************** Testing Objective consensu:STATE TRANSFER ********
*   Setup: 4 node local docker peer network with security
*   0. Deploy chaincodeexample02 with 100000, 90000 as initial args
*   1. Send Invoke Requests on multiple peers using go routines.
*   2. Verify query results match on vp0 and vp1 after invoke
*********************************************************************/

import (
	"fmt"
	"time"
	"obcsdk/chaincode"
	"sync"
)

var ctrNumReq int

func main() {

	fmt.Println("Using an existing docker network")
	//peernetwork.SetupLocalNetwork(4, true)

	_ = chaincode.InitNetwork()
	chaincode.InitChainCodes()
	chaincode.RegisterUsers()

	time.Sleep(30000 * time.Millisecond)
	//peernetwork.PrintNetworkDetails()

	data := "Yh1WWZlw1gGd2qyMNaHqBCt4zuBrnT4cvZ5iMXRRM3YBMXLZmmvyVr0ybWfiX4N3UMliEVA0d1dfTxvKs0EnHAKQe4zcoGVLzMHd8jPQlR5ww3wHeSUGOutios16lxfuQTdnsFcxhXLiGwp83ahyBomdmJ3igAYTyYw2bwXqhBeL9fa6CTK43M2QjgFhQtlcpsh7XMcUWnjJhvMHAyH67Z8Ugke6U8GQMO5aF1Oph0B2HlIQUaHMq2i6wKN8ZXyx7CCPr7lKnIVWk4zn0MLZ16LstNErrmsGeo188Rdx5Yyw04TE2OSPSsaQSDO6KrDlHYnT2DahsrY3rt3WLfBZBrUGhr9orpigPxhKq1zzXdhwKEzZ0mi6tdPqSzMKna7O9STstf2aFdrnsoovOm8SwDoOiyqfT5fc0ifVZSytVNeKE1C1eHn8FztytU2itAl1yDYSfTZQv42tnVgDjWcLe2JR1FpfexVlcB8RUhSiyoThSIFHDBZg8xyULPmp4e6acOfKfW2BXh1IDtGR87nBWqmytTOZrPoXRPq2QXiUjZS2HflHJzB0giDbWEeoZoMeF11364Xzmo0iWsBw0TQ2cHapS4cR49IoEDWkC6AJgRaNb79s6vythxX9CqfMKxIpqYAbm3UAZRS7QU7MiZu2qG3xBIEegpTrkVNneprtlgh3uTSVZ2n2JTWgexMcpPsk0ILh10157SooK2P8F5RcOVrjfFoTGF3QJTC2jhuobG3PIXs5yBHdELe5yXSEUqUm2ioOGznORmVBkkaY4lP025SG1GNPnydEV9GdnMCPbrgg91UebkiZsBMM21TZFbUqP70FDAzMWZKHDkDKCPoO7b8EPXrz3qkyaIWBymSlLt6FNPcT3NkkTfg7wl4DZYDvXA2EYu0riJvaWon12KWt9aOoXig7Jh4wiaE1BgB3j5gsqKmUZTuU9op5IXSk92EIqB2zSM9XRp9W2I0yLX1KWGVkkv2OIsdTlDKIWQS9q1W8OFKuFKxbAEaQwhc7Q5Mm"


	fmt.Println("\nPOST/Chaincode: Deploying chaincode at the beginning ....")
	dAPIArgs0 := []string{"concurrency", "init", "vp1"}
	depArgs0 := []string{"a", data, "counter", "0"}
	chaincode.DeployOnPeer(dAPIArgs0, depArgs0)

	time.Sleep(240000 * time.Millisecond)
	fmt.Println("\nPOST/Chaincode: Querying a and b after deploy >>>>>>>>>>> ")
	qAPIArgs0 := []string{"concurrency", "query"}
	qArgsa := []string{"a"}
	qArgsb := []string{"counter"}
	A, _ := chaincode.Query(qAPIArgs0, qArgsa)
	B, _ := chaincode.Query(qAPIArgs0, qArgsb)
	myStr := fmt.Sprintf("\nA = %s B= %s", A, B)
	fmt.Println(myStr)

        ctrNumReq = 0
        numReq := 1000
        defer timeTrack(time.Now(), "Testcase execution Done")
        now := time.Now().Unix()
        endTime := now + 10 * 60
        for now < endTime {
            //fmt.Println("Time Now ", time.Now())
            //fmt.Println("end Time ", endTime)
            //fmt.Println("now ", now)
            //schedulerTask()
	    InvokeLoop(numReq, data)
            now = time.Now().Unix()
       }

 
        //defer timeTrack(time.Now(), "Testcase executiion Done")
        //numReq :=1000
	//InvokeLoop(numReq, data)
	//time.Sleep(time.Minute * time.Duration(5))




}

func InvokeLoop(numReq int, data string) {

	var wg sync.WaitGroup

	iAPIArgs := []string{"a", data, "counter"}

	wg.Add(4)
	go func() {

		defer wg.Done()
		k := 1
	        invArgs0 := []string{"concurrency", "invoke", "vp0"}
		for k <= numReq {
		   go chaincode.InvokeOnPeer(invArgs0, iAPIArgs)
	           time.Sleep(5000 * time.Millisecond)
		   k++
		}
		fmt.Println("# of Req Invoked on vp0 ", k)
	}()
	go func() {

		defer wg.Done()
		k := 1
	        invArgs1 := []string{"concurrency", "invoke", "vp1"}
		for k <= numReq {
		   go chaincode.InvokeOnPeer(invArgs1, iAPIArgs)
	           time.Sleep(5000 * time.Millisecond)
		   k++
		}
		fmt.Println("# of Req Invoked on vp1 ", k)
	}()


	go func() {

		defer wg.Done()
		k := 1
	        invArgs2 := []string{"concurrency", "invoke", "vp2"}
		for k <= numReq {
		   go chaincode.InvokeOnPeer(invArgs2, iAPIArgs)
	           time.Sleep(5000 * time.Millisecond)
		   k++
		}
		fmt.Println("# of Req Invoked on vp2 ", k)
	}()

	go func() {

		defer wg.Done()
		invArgs3 := []string{"concurrency", "invoke", "vp3"}
		k := 1
		for k <= numReq {
		    go chaincode.InvokeOnPeer(invArgs3, iAPIArgs)
	            time.Sleep(5000 * time.Millisecond)
		    k++
		}
		fmt.Println("# of Req Invoked  on vp3", k)
	}()

	wg.Wait()
}

func QueryHeight() {

	fmt.Println("\nPOST/Chaincode: Querying a and b after invoke >>>>>>>>>>> ")
	qAPIArgs00 := []string{"concurrency", "query", "vp0"}
	qAPIArgs01 := []string{"concurrency", "query", "vp1"}
	qAPIArgs02 := []string{"concurrency", "query", "vp2"}
	qAPIArgs03 := []string{"concurrency", "query", "vp3"}

	qArgsa := []string{"a"}
	qArgsb := []string{"counter"}

	res0A, _ := chaincode.QueryOnHost(qAPIArgs00, qArgsa)
	res0B, _ := chaincode.QueryOnHost(qAPIArgs00, qArgsb)

	res1A, _ := chaincode.QueryOnHost(qAPIArgs01, qArgsa)
	res1B, _ := chaincode.QueryOnHost(qAPIArgs01, qArgsb)

	res2A, _ := chaincode.QueryOnHost(qAPIArgs02, qArgsa)
	res2B, _ := chaincode.QueryOnHost(qAPIArgs02, qArgsb)

	res3A, _ := chaincode.QueryOnHost(qAPIArgs03, qArgsa)
	res3B, _ := chaincode.QueryOnHost(qAPIArgs03, qArgsb)

	fmt.Println("Results in a and b vp0 : ", res0A, res0B)
	fmt.Println("Results in a and b vp1 : ", res1A, res1B)
	fmt.Println("Results in a and b vp2 : ", res2A, res2B)
	fmt.Println("Results in a and b vp3 : ", res3A, res3B)

	ht0, _ := chaincode.GetChainHeight("vp0")
	ht1, _ := chaincode.GetChainHeight("vp1")
	ht2, _ := chaincode.GetChainHeight("vp2")
	ht3, _ := chaincode.GetChainHeight("vp3")

	fmt.Printf("ht0: %d, ht1: %d, ht2: %d, ht3: %d ", ht0, ht1, ht2, ht3)
}


func timeTrack(start time.Time, name string) {
	elapsed := time.Since(start)
        QueryHeight()
	fmt.Printf("\n################# %s took %s \n", name, elapsed)
	fmt.Println("################# Execution Completed #################")
}
