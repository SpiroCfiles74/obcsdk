package main

// 
// INSTRUCTIONS:
// 
// 1. Change chco2.CurrentTestName, to set this test name = this filename
// 2. Edit to add your test steps at the bottom.
// 3. go build setupTest.go
// 4. go run setupTest.go  - or better yet, to save all results use script:  gorecord.sh setupTest.go
// 
// 
// SETUP STEPS included already:
// -----------------------------
// Default Setup: 4 peer node network with security CA node using local docker containers.
// (To change network characteristics and tuning parameters, change consts in file ../chco2/chco2.go)
// 
// SETUP 1: Deploy chaincode_example02 with A=1000000, B=1000000 as initial args.
// SETUP 2: Send INVOKES (moving 1 from A to B) once on each peer node.
// SETUP 3: Query all peers to validate values of A, B, and chainheight.
// 


import (
	"os"
	"time"
	"bufio"
	"obcsdk/chco2"
	"fmt"
	"strconv"
	// "bufio"
	// "obcsdk/chaincode"
	// "obcsdk/peernetwork"
	// "log"
)

var osFile *os.File

func main() {

	//=======================================================================================
	// SET THE TESTNAME:  set the filename/testname here, to display in output results.
	//=======================================================================================

	chco2.CurrentTestName = "CAT_113_StopAndStartMemberServices.go"


	//=======================================================================================
	// Getting started: output file, test timing, setup/init, and start & confirm the network
	//=======================================================================================

	if (chco2.Verbose) { fmt.Println("Welcome to test " + chco2.CurrentTestName) }

	chco2.RanToCompletion = false
	startTime := time.Now()
	_, err := os.Stat(chco2.OutputSummaryFileName)	// Stat returns *FileInfo. It will return an error if there is no file.
	if err != nil {
		if os.IsNotExist(err) {
			// File simply does not exist. Create the *File.
			osFile, err = os.Create(chco2.OutputSummaryFileName)
			chco2.Check(err)
		} else {
			chco2.Check(err)  // some other error; panic and exit.
		}
	} else {
		// open the existing file
		osFile, err = os.OpenFile(chco2.OutputSummaryFileName, os.O_RDWR|os.O_APPEND, 0666)
		chco2.Check(err)
	}
	defer osFile.Close()
	chco2.Writer = bufio.NewWriter(osFile)

	// When main() ends, print the test PASS/FAIL line, with elapsed time, to outfile and to stdout
	defer chco2.TimeTrack(startTime, chco2.CurrentTestName)

	// Initialize everything, and start the network: deploy, invoke once on each peer, and query all peers to confirm
	chco2.Setup( chco2.CurrentTestName, startTime )


	//=======================================================================================
	// 
	// OPTIONAL OVERRIDES:
	// 	Tune these booleans to control verbosity and test strictness.
	// 	These booleans are initialized inside chco2.Setup(), as follows.
	// 
	//	Note: Set AllRunningNodesMustMatch to false when need merely enough peers for consensus
	// 	to match results, especially when test involves stopping or pausing peer nodes.
	// 	OR, set it true (default) when all active running peers must match (e.g. at init
	// 	time, or after sending enough invokes after a node outage to guarantee that all
	// 	peers are caught up, in sync, with matching values for chainheight, A & B.
	// 
	//	Simply uncomment any lines here for this testcase to override
	//	the default values, as defined in ../chco2/chco2.go
	// 
	// 	chco2.Verbose = true
	// 	chco2.Stop_on_error = true
	// 	chco2.EnforceQueryTestsPass = false
	//	chco2.EnforceChainHeightTestsPass = false
	//	chco2.AllRunningNodesMustMatch = false 	// Note: chco2 inits to true, but sets this false when restart a peer node
	//	chco2.CHsMustMatchExpected = true	// not fully implemented and working, so leave this false
	//	chco2.QsMustMatchExpected = false 	// Note: until #2148 is solved, you may need to set false here if testcase has complicated multiple stops/restarts
	//	chco2.DefaultInvokesPerPeer = 1		// 1 is default. Uncomment and change this here to override for this testcase.
	//	chco2.TransPerSecRate = 2		// 2 is default. Uncomment and change this here to override for this testcase.


	//=======================================================================================
	// 
	// chco2. API available function calls in ../chco2/chco2.go:
	// 
	//	DeployNew(A int, B int)
	//	Invokes(totalInvokes int)
	//	InvokeOnEachPeer(numInvokesPerPeer int)
	//	InvokeOnThisPeer(totalInvokes int, peerNum int)
	//	QueryAllPeers(stepName string)
	//	StopPeers(peerNums []int)
	//	RestartPeers(peerNums []int)
	//	QueryMatch(currA int, currB int)
	//	SleepTimeSeconds(secs int) time.Duration
	//	SleepTimeMinutes(minutes int) time.Duration
	//	CatchUpAndConfirm()
	// To be implemented soon:
	//	WaitAndConfirm()
	//	PausePeers(peerNums []int)
	//	UnpausePeers(peerNums []int)
	// 
	// Example usages:
	// 
	// chco2.DeployNew( 9000, 1000 )
	// chco2.Invokes( chco2.InvokesRequiredForCatchUp )
	// chco2.InvokeOnEachPeer( chco2.DefaultInvokesPerPeer )
	// InvokeOnThisPeer( 100, 0 )
	// chco2.StopPeers( []int{ 99 } )
	// chco2.QueryAllPeers( "STEP 6, after STOP PEERs " + strconv.Itoa(99) )
	// chco2.RestartPeers( []int{ j, k } )
	// chco2.QueryAllPeers( "STEP 9, after RESTART PEERs " + strconv.Itoa(j) + ", " + strconv.Itoa(k) )
	// if (chco2.Verbose) { fmt.Println("Sleep extra 60 secs") }
	// time.Sleep(chco2.SleepTimeSeconds(60))
	// time.Sleep(chco2.SleepTimeMinutes(1))
	// 
	//=======================================================================================


	//=======================================================================================
	// DEFINE MAIN TESTCASE STEPS HERE
	// 

	// CAT_113_StopAndStartMemberServices.go

	chco2.StopMemberServices()

	// Normally: chco2 runs with minimum 2 secs per batch of invokes per peer. We normally run tests with batchsize=2.
	// Since 2 Tx per sec is easily within the limit of the configured processing rate=40 transaction per sec,
	// no extra delays will be inserted by the chco2 layer.

	// All timing computations are for chco2 delays while processing invokes, with assumptionthat the rest of the code, including queries, takes no time.
	// However, we know that if we run out of Tcerts then every invoke and query thereafter would fail after a timeout (couple secs each?) further lengthening the test run!

	success := true

	//loops :=  900 / 8  	// Run for 15 mins = 15 * 60 secs. Each loop takes 8 secs. 112 loops.
				// Note: after 100 loops, we have sent 200 Invokes (100 batches of invokes, plus double that for queries)
				//  - and should have run out of the Tcerts on each peer!

	loops := 6
	i := 1
	for i = 1; i <= loops; i++ {
		if success {
			if !chco2.TestsCurrentlyPass() {
				fmt.Println("\n===FAILURE after stopped MemberServices, secs: " + strconv.Itoa((i-1)*8) )
				success = false
			}
		} else {
			if chco2.TestsCurrentlyPass() {
				fmt.Println("\n===RECOVERY! accumulated forced_delays_secs (does not include delays for timeouts for failed invoke/queries): " + strconv.Itoa((i-1)*8) )
				success = true
			}
		}

		// Send 8 invokes means 1 batch of 2 invokes is sent to all 4 running peers, each with 2 secs delay, so a total of 8 secs.
		// 320 (80 per peer) is the max that still fits within 2 secs each, keeping the loop cycle time at 8 secs

		// chco2.Invokes( 8 )	// 8 (2 per peer) is one per sec total invokes

		// let's hurry along to cause exhaustion of the TCert pool on 3 of the 4 peers, in just a few loops

		// Peers 0 and 1 will run out of TCerts on the 5th loop, since 5x40=200, and we had used 6 during setup.
		// So we will see a handful of these error logs in the GO_TEST output file from each of those two peers:
		//	POST /chaincode returned code =-32002 message=Invocation failure data=Error when invoking chaincode: Failed loading TCerts from TCA
		// And some errors from the associated Queries too:
		//	POST /chaincode returned code =-32003 message=Query failure data=Error when querying chaincode: Failed loading TCerts from TCA

		chco2.InvokeOnThisPeer( 40, 0 )
		chco2.InvokeOnThisPeer( 40, 1 )
		chco2.InvokeOnThisPeer(  1, 2 )  // this peer will not run out of TCerts
		chco2.InvokeOnThisPeer(  1, 3 )  // this peer will not run out of TCerts

		chco2.QueryAllPeers("STEP monitor status after stopped caserver, loop=" + strconv.Itoa(i) + "/" + strconv.Itoa(loops) + "  forced_delays_secs=" + strconv.Itoa(i*8) )
	}

	chco2.RestartMemberServices()

	chco2.QsMustMatchExpected = false
	chco2.QueryAllPeers("STEP first query after restarted caserver")

	// loops = 300 / 8		// 300 = 5 minutes, at 8 secs per loop

	i = 1
	for i = 1; i <= loops; i++ {
		if success {
			if !chco2.TestsCurrentlyPass() {
				fmt.Println("\n===FAILURE!! after RestartMemberServices, secs: " + strconv.Itoa((i-1)*8) )
				success = false
			}
		} else {
			if chco2.TestsCurrentlyPass() {
				fmt.Println("\n===RECOVERY! after RestartMemberServices (does not include timeouts for failed invoke/queries) secs: " + strconv.Itoa((i-1)*8) )
				success = true
			}
		}

		// chco2.Invokes( 8 )	// 8 (2 per peer) is one per sec total invokes, 1 per 4 secs on each peer.
					// Multiply that 8 times 40 (or less) to keep from lengthing test duration.
		chco2.Invokes( 320 )	// 320 (80 per peer) is the max that still fits within 2 secs each, keeping the loop cycle time at 8 secs
		chco2.QueryAllPeers("STEP monitor status after restarted caserver, loop=" + strconv.Itoa(i) + "/" + strconv.Itoa(loops) + "  forced_delays_secs=" + strconv.Itoa(i*8) )
	}

	chco2.CatchUpAndConfirm()			// OPTIONAL. Depends on testcase details and objectives.

	chco2.RanToCompletion = true	// DO NOT MOVE OR CHANGE THIS. It must remain last.
}

