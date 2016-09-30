package chco2

// Tip: Read https://github.com/hyperledger/fabric/tree/master/examples/events/block-listener
// 	to debug code that tries to validate "expected chainheight"
// And (for fun) read http://www.multichain.com/blog/2016/05/four-genuine-blockchain-use-cases/.

import (
	"errors"
	"bufio"
	"fmt"
	"obcsdk/chaincode"
	"obcsdk/peernetwork"
	"obcsdk/threadutil"
	"strconv"
	"strings"
	"time"
	"log"
	"os"
)


const (

	//==============================================================================================================
	// Internal constants : tunable
	// 

	// Increase this for more traffic

	DefaultInvokesPerPeer = 1


	//==============================================================================================================
	// Internal constant : DO NOT change.
	// Two lines for each test are appended to this file which will contain a handy running summary.

	OutputSummaryFileName = "GO_TESTS_SUMMARY.log"
)


var Writer *bufio.Writer
var url string
var MyNetwork peernetwork.PeerNetwork
var Verbose bool	// Another option: go edit  "verbose" in chaincode/const.go for more info about lower level functions operation
var ExistingNetwork bool
var Stop_on_error bool
var RanToCompletion bool
var CurrentTestName string
var currA, currB int
var initA, initB  string
var currCH  int
var queryTestsPass, chainHeightTestsPass, annexTestsPass bool
var TransPerSecRate int


type peerQData struct {
	resA int
	resB int
}
// Use slices, not an array. NumberOfPeersInNetwork is set after initialization, so leave size open-ended for now.
var qData []peerQData		// latest queried values for A and B for each peer
var qtransPerPeerForCH []int 	// counts of transactions queued per peer, for calculating blockchainheight
var qtrans int			// counts of transactions queued, for calculating expected values of A & B

// bools to control when to stop/abort test (and to print additional error msgs when that happens)

var EnforceQueryTestsPass, EnforceChainHeightTestsPass bool

// bools ...MustMatchExpected... :
// True means strict mode: the values must be in consensus AND match an internal counter based on our testcase logic/expectations.
// False means lenient mode, or "Consensus only": the values must match each other but not an internal counter.
//    Use false when we can't fully understand or complete our own test code logic for the counters.

var QsMustMatchExpected bool 	// Queried values (A & B) must match the internal counters = "expected" values (currA & currB) 
var CHsMustMatchExpected bool 	// ChainHeight values (CH) must match the internal counter = "expected" value (currCH)

// AllRunningNodesMustMatch=true implies for ALL running nodes; false implies only just enough for consensus.
// Typically set false except at very beginning and in CatchUpAndConfirm() - after sending
// many many (how many?) invokes that guarantee every node catches up.

var AllRunningNodesMustMatch bool

var NetworkAlreadyRunning bool

var LoggingLevel string
var IsLocalNetwork bool
var LocalNetworkType string
var Security bool
var ConsensusMode string
var PbftMode string
var NumberOfPeersInNetwork int
var NumberOfPeersOkToFail int
var MaxNumberOfPeersThatCanFailWhileStillHaveConsensus int
var MinNumberOfPeersNeededForConsensus int
var NumberOfPeersNeededForConsensus int
var InvokesRequiredForCatchUp int
var K int
var logmultiplier int
var batchsize int	// Note: default CORE_PBFT_GENERAL_BATCHSIZE=500. If you change this to anything more than 5, then
			//	set InvokesRequiredForCatchUp to 200 for a good effort to ensure catch up - but
			//	be aware that this will be lower than the required number so some testcases that
			//	may fail sometimes if they set CHsMustMatchExpected=true.
var batchtimeout int							// default 2
var batchTimeout string	// CORE_PBFT_GENERAL_TIMEOUT_BATCH=2s		// default 2s
			// CORE_PBFT_GENERAL_TIMEOUT_REQUEST=10s	// default 2
			// CORE_PBFT_GENERAL_TIMEOUT_VIEWCHANGE=2s	// default 2
			// CORE_PBFT_GENERAL_TIMEOUT_NULLREQUEST=1s	// default 0 = disable keep-alive nullrequests

var pauseInsteadOfStop bool	// Set pauseInsteadOfStop to true to run all tests using docker pause/unpause
				// instead of docker stop/restart. This allows tests to be reused, instead of duplicated.





func Setup(testName string, started time.Time) {
	setup_part1(testName, started)
	setup_part2_network()
	setup_part3_verifyNetworkAndDeployCC()
}

func SetupQuick(testName string, started time.Time) {
	setup_part1(testName, started)
	fmt.Println("chco2.SetupQuick(): Skipping creation of a network; assuming one is already running.")
	setup_part3_verifyNetworkAndDeployCC()
}

func setup_part1(testName string, started time.Time) {

	//---------------------------------------------------------------------------------------------------------------
	// configure the booleans, environment variables, and the test parameter constants and slices that depend on them
	//---------------------------------------------------------------------------------------------------------------

	CurrentTestName = testName
	RanToCompletion = false
	Verbose = false			// See also:  "verbose" in chaincode/const.go for lower level functions
	Stop_on_error = false
	ExistingNetwork = false
	queryTestsPass = true
	chainHeightTestsPass = true
	annexTestsPass = true
	EnforceQueryTestsPass = true
	EnforceChainHeightTestsPass = true
	QsMustMatchExpected = true 	// Queried values (A & B) must match the internal counters = "expected" values (currA & currB) 
	CHsMustMatchExpected = false 	// ChainHeight values (CH) must match the internal counter = "expected" value (currCH)
	AllRunningNodesMustMatch = true // values should match on ALL avail nodes - not just enough nodes for consensus;
					// set true after sending enough invokes to ensure all nodes caught up; not sure how many,
					// or why, so most testcases should set this to false after the initial setup and query is done.

	// You MAY change this one. Set TransPerSecRate to be comfortable with enough time to allow the network to
	// process the queued transactions per second - based on your test environment processing rate speeds.
	// Expected network processing time: as of 7/14/16 we can support a little more than the following
	// (for a 10-minute test run on v0.5):
	// 	 11/sec invokes in vagrant/docker environment on PC,
	// 	  2/sec invokes in zACI LPAR environment, or more if using different client threads,
	//	and 3 x those numbers for Queries.
	// A value higher than those above will simply drive it as fast as can go.

	TransPerSecRate = 10

	//---------------------------------------------------------------------------------------------------------------
	// set defaults for ENV variables
	//---------------------------------------------------------------------------------------------------------------
	// Defaults for most of these tuning parameters are in .../fabric/consensus/obcpbft/config.yaml
	// When setting up our network, we need to pass some of them to local_fabric.sh to override the
	// default settings, so that our test code uses the same values as the running peers and we can
	// tune our tests accordingly (e.g. counting transactions).

	IsLocalNetwork = true
	LocalNetworkType = ""
	NumberOfPeersInNetwork = 4	//  CORE_PBFT_GENERAL_N         - number of validating peers in the network
	NumberOfPeersOkToFail = 1	//  CORE_PBFT_GENERAL_F         - max # possible faulty nodes while still can reach consensus
	LoggingLevel = "error"		//  CORE_LOGGING_LEVEL          - [critical|error|warning|notice|info|debug] as defined in peer/core.yaml
	Security = true			//  CORE_SECURITY_ENABLED       - use secure network using MemberSrvc CA [Y|N]
	ConsensusMode = "pbft"		//  CORE_PEER_VALIDATOR_CONSENSUS_PLUGIN - consensus mode [pbft|...]
        PbftMode = "batch"		//  CORE_PBFT_GENERAL_MODE      - pbft mode [batch|noops]
	batchsize = 2			//  CORE_PBFT_GENERAL_BATCHSIZE - max # Tx sent in each batch for ordering; we override the default [500]
	batchTimeout = "2s"		//  CORE_PBFT_GENERAL_TIMEOUT_BATCH=2s
	batchtimeout = 2		//    - default 2 in v0.5 Jun 2016, default 1 in gerrit fabric Aug 2016
	pauseInsteadOfStop = false	//  TEST_STOP_OR_PAUSE          - MODE used by GO tests when disrupting network CA and Peer nodes [STOP|PAUSE]

					// Others that we may use in future:
					//  CORE_PBFT_GENERAL_TIMEOUT_BATCH - batch timeout value, use s for seconds, default=[2s]
	logmultiplier = 4		//  CORE_PBFT_GENERAL_LOGMULTIPLIER - logmultiplier [4]
	K = 10				//  CORE_PBFT_GENERAL_K             - checkpoint period K [10]


	//---------------------------------------------------------------------------------------------------------------
	// read any ENV variables that are set, and override the default values
	//---------------------------------------------------------------------------------------------------------------

	// save Network Type for now, and use it later
	LocalNetworkType = strings.TrimSpace(strings.ToUpper(os.Getenv("TEST_NETWORK")))
	if LocalNetworkType != "" {
		if LocalNetworkType == "LOCAL" {
			IsLocalNetwork = true
		} else {
			IsLocalNetwork = false
			if LocalNetworkType == "Z" { TransPerSecRate = 2 }
		}
	}
	if strings.ToUpper(os.Getenv("TEST_VERBOSE")) == "TRUE" {
		Verbose = true 	// Another option: edit  "verbose" in chaincode/const.go for more info about lower level functions operation
	}
	var envvar string
	envvar = strings.TrimSpace(os.Getenv("CORE_PBFT_GENERAL_N"))
	if envvar != "" { NumberOfPeersInNetwork, _ = strconv.Atoi(envvar) }
	envvar = strings.TrimSpace(os.Getenv("CORE_PBFT_GENERAL_F"))
	if envvar != "" { NumberOfPeersOkToFail, _  = strconv.Atoi(envvar) }
	envvar = strings.TrimSpace(os.Getenv("CORE_LOGGING_LEVEL"))
	if envvar != "" { LoggingLevel = envvar }
	envvar = strings.TrimSpace(os.Getenv("CORE_SECURITY_ENABLED"))
	if strings.ToUpper(envvar) == "N" { Security = false }
	envvar = strings.TrimSpace(os.Getenv("CORE_PEER_VALIDATOR_CONSENSUS_PLUGIN"))
	if envvar != "" { ConsensusMode = envvar }
	envvar = strings.TrimSpace(os.Getenv("CORE_PBFT_GENERAL_MODE"))
	if envvar != "" { PbftMode = strings.ToUpper(envvar) }
	envvar = strings.TrimSpace(os.Getenv("CORE_PBFT_GENERAL_BATCHSIZE"))
	if envvar != "" { batchsize, _ = strconv.Atoi(envvar) }
	// Must read batchTimeout (string) and set batchtimeout (int) after stripping the trailing 's'...
	// envvar = strings.TrimSpace(os.Getenv("CORE_PBFT_GENERAL_TIMEOUT_BATCH"))
	// if envvar != "" { batchtimeout, _  = strconv.Atoi(envvar) }
	envvar = strings.TrimSpace(os.Getenv("TEST_STOP_OR_PAUSE"))
	if strings.ToUpper(envvar) == "PAUSE" { pauseInsteadOfStop = true }
	if strings.ToUpper(os.Getenv("TEST_EXISTING_NETWORK")) == "TRUE" {
		ExistingNetwork = true
		AllRunningNodesMustMatch = false // since we don't know the status of the existing network, the counters may already differ
	}


	//---------------------------------------------------------------------------------------------------------------
	// envvar error checking, and initialize other variables that are dependent on those env vars
	//---------------------------------------------------------------------------------------------------------------


	// validate N and related items

	if Security {
		if NumberOfPeersInNetwork < 4 {
			fmt.Println("WARNING: INVALID VALUE (" + strconv.Itoa(NumberOfPeersInNetwork) + ") provided for N when security is enabled; a secure network must contain a minimum of 4 peer nodes.")
			//fmt.Println("WARNING: INVALID VALUE (" + strconv.Itoa(NumberOfPeersInNetwork) + ") PROVIDED FOR N !!!  When security is enabled, a network must contain a minimum of 4 peer nodes. Resetting N to 4")
			//NumberOfPeersInNetwork = 4 // we could reset it, but then we wouldn't see how the fabric reacts...
		}
	}

	envvar = strings.ToUpper(os.Getenv("TEST_FULL_CATCHUP"))
	if envvar == "TRUE" {
		// InvokesRequiredForCatchUp really should be (K * batchsize * logmultiplier) to ensure recovery.
		// (Although sometimes, depending on timing and state transitions, I wonder if maybe even that is not enough???)
		InvokesRequiredForCatchUp = (K * batchsize * logmultiplier)
	} else {
		// Since it gets so big for larger networks and slows down our tests. (That is one reason why
		// we set a default batchsize of 2 for these consensus tests.) And we often do not need it to be
		// the maximum for tests to pass, so let's just try this, which seems to be enough in most cases:
		InvokesRequiredForCatchUp = (NumberOfPeersInNetwork * 25)
	}

	// validate F and related items

	// ensure we do not set F to a value exceeding (n-1)/3. And set other related vars.

        MaxNumberOfPeersThatCanFailWhileStillHaveConsensus = (NumberOfPeersInNetwork - 1) / 3

	if (strings.ToUpper(ConsensusMode) == "PBFT") && (NumberOfPeersOkToFail > MaxNumberOfPeersThatCanFailWhileStillHaveConsensus) {
		fmt.Println("WARNING: INVALID VALUE (" + strconv.Itoa(NumberOfPeersOkToFail) + ") provided for F !!!  Maximum is (N-1)/3 = " + strconv.Itoa(MaxNumberOfPeersThatCanFailWhileStillHaveConsensus))
		//fmt.Println("WARNING: INVALID VALUE (" + strconv.Itoa(NumberOfPeersOkToFail) + ") PROVIDED FOR F !!!  CHANGING TO (N-1)/3 = " + strconv.Itoa(MaxNumberOfPeersThatCanFailWhileStillHaveConsensus))
		//NumberOfPeersOkToFail = MaxNumberOfPeersThatCanFailWhileStillHaveConsensus // we could reset it, but then we wouldn't see how the fabric reacts...
	}

	// Note: user must ensure NumberOfPeersOkToFail <= MaxNumberOfPeersThatCanFailWhileStillHaveConsensus
	// User may configure the network to require a higher number of agreeing peers than the minimum,
	// by setting F to a smaller value than its maximum (in other words, allowing a fewer number to fail -
	// for example instead of requiring 2/3 to be functional (allowing almost a third to fail), the
	// user may desire 9/10 to be functional.)
	// Another way to look at it: min value for "N" is 3F+1.

	MinNumberOfPeersNeededForConsensus = NumberOfPeersInNetwork - MaxNumberOfPeersThatCanFailWhileStillHaveConsensus
	NumberOfPeersNeededForConsensus = NumberOfPeersInNetwork - NumberOfPeersOkToFail

	// validate batchsize

	if batchsize < 1 {
		fmt.Println("WARNING: INVALID VALUE (" + strconv.Itoa(batchsize) + ") provided for batchsize !!!  CHANGING TO 1")
		batchsize = 1
	}

	if pauseInsteadOfStop {
		// we support PAUSE only for LOCAL docker network, so far...
		if IsLocalNetwork {
			fmt.Println("All STOPS and RESTARTS will be executed with Docker PAUSE and UNPAUSE")
		} else {
			fmt.Println("Unsupported Configuration: TEST_STOP_OR_PAUSE cannot be PAUSE when TEST_NETWORK != LOCAL")
			panic(errors.New("docker-pause is not supported on non-Local configuration"))
		}
	}

	//fmt.Println("INFO: setup_part1(): TransPerSecRate = ", TransPerSecRate)

	//---------------------------------------------------------------------------------------------------------------
	// create and initialize storage slices for queued transactions counters, now that we know size of "N"
	//---------------------------------------------------------------------------------------------------------------

	qData = make([]peerQData, NumberOfPeersInNetwork)
	qtransPerPeerForCH = make([]int, NumberOfPeersInNetwork)
	qtrans = 0
	for i:= 0; i < NumberOfPeersInNetwork; i++ {
		qtransPerPeerForCH[i] = 0
		qData[i].resA = 0
		qData[i].resB = 0
	}


	myStr := fmt.Sprintf("\nBEGIN  %s (Enforce Q=%t CH=%t, MustMatch Q=%t CH=%t AllRunningNodes=%t) [STARTED: %s]", CurrentTestName, EnforceQueryTestsPass, EnforceChainHeightTestsPass, QsMustMatchExpected, CHsMustMatchExpected, AllRunningNodesMustMatch, started)
	fmt.Println(myStr)
	fmt.Fprintln(Writer, myStr)
	Writer.Flush()
}

func setup_part2_network() {
    if ExistingNetwork {
	fmt.Println("chco2.setup_part2_network(): TEST_EXISTING_NETWORK is TRUE, which means:\n (1) we will NOT create a new network, and\n (2) we will IGNORE the COMMIT image and a few other env vars, and\n (3) we will use the existing Network as previously created, and query its values as our starting point.")
    } else {
	fmt.Println("Creating a local docker network with # peers = ", NumberOfPeersInNetwork)
	peernetwork.SetupLocalNetworkWithMoreOptions(
		NumberOfPeersInNetwork,	//  CORE_PBFT_GENERAL_N
		NumberOfPeersOkToFail,	//  CORE_PBFT_GENERAL_F
		LoggingLevel,		//  CORE_LOGGING_LEVEL
		Security,		//  CORE_SECURITY_ENABLED
		ConsensusMode,		//  CORE_PEER_VALIDATOR_CONSENSUS_PLUGIN
        	//PbftMode,		//  CORE_PBFT_GENERAL_MODE
		//batchTimeout,		//  CORE_PBFT_GENERAL_TIMEOUT_BATCH
		batchsize )		//  CORE_PBFT_GENERAL_BATCHSIZE

	if (Verbose) { fmt.Println("Sleep 10 secs extra after setup_part2 created network") }; time.Sleep(10 * time.Second)
    }
}

func setup_part3_verifyNetworkAndDeployCC() {

	peernetwork.PrintNetworkDetails()
	// read in the NetworkCredentials.json file that was used in this (or prior) test to create the network
	MyNetwork = chaincode.InitNetwork()

	// if the user has set env var, then override the PeerNetwork.IsLocal boolean now that we have initialized the network
	if LocalNetworkType != "" { chaincode.SetNetworkIsLocal(IsLocalNetwork) }

	if ExistingNetwork {
		// get the actual IP addresses of all the peers, in case they have changed due to any
		// peer node restarts some time after this network was created in earlier testcase
		chaincode.UpdatePeerIp(&MyNetwork, -1)
	}
	chaincode.InitChainCodes()
	if !chaincode.RegisterUsers() {
		panic(errors.New("setup_part3_verify: ERROR: FAILED TO REGISTER one or more users in NetworkCredentials.json file\n"))
	}

	// get any avail node URL details to get info on chainstats/transactions/blocks etc.
	aPeer, _ := peernetwork.APeer(chaincode.ThisNetwork)
	url := chaincode.GetURL(aPeer.PeerDetails["ip"], aPeer.PeerDetails["port"])

	// Display!
	chaincode.NetworkPeers(url)
	chaincode.ChainStats(url)

	//chaincode.User_Registration_Status(url, "lukas")
	//chaincode.User_Registration_Status(url, "nishi")
	//chaincode.User_Registration_ecertDetail(url, "lukas")

	// Init / Deploy 
	initA = "1000000"		// start with ONE MILLION
	initB = "1000000"		// start with ONE MILLION
	currA, _ = strconv.Atoi(initA)
	currB, _ = strconv.Atoi(initB)
	currCH = 1			// one for genesis block

	// find highest numbered running peer; deploy; and send one invoke request to each peer
        peerNum := NumberOfPeersInNetwork-1
        for ; peerNum >= 0; peerNum-- { if peerIsRunning(peerNum,MyNetwork) { break } }
	if peerNum < 0 {
		fmt.Println("Setup() ERROR: Cannot find any running peer to use for Deploy!!!!!!!!!!")
		panic(errors.New("setup_part3_verify: CANNOT find any running peer node for Deploy!!!!!"))
	} else {
		if Verbose { fmt.Println("setup_part3_verify, before deploy: A/B/chainheight values: ", currA, currB, currCH) }
		DeployInit(peerNum)
		InvokeOnEachPeer(1)
		QueryAllPeers("STEP SETUP, after initial Deployment followed by 1 Invoke on each peer")
	}
}


// This func can be used by chco2 funcs as well as external funcs such as chcotest/BasicFuncExistingNetwork.go
// that have their own network (and which did not call chco2 setup/init functions)

func QueryAllHostsToGetCurrentValues(mynetwork peernetwork.PeerNetwork, a *int, b *int, ch *int) bool {		// using example02

	found_consensus := false
	N := peernetwork.GetNumberOfPeers(mynetwork)	// CORE_PBFT_GENERAL_N ; in chco2, this is same as NumberOfPeersInNetwork
	F := (N - 1) / 3				// Max value for F (if CORE_PBFT_GENERAL_F is actually set lower than F and
							//   if the current number of available peers is somewhere in between, then
							//   we may be OK here but the testcase could conceivably still fail later)
	var queryData []peerQData			// queried values for A and B for each peer
	queryData = make([]peerQData, N)		// queried values for block chain height for each peer
	var ht []int
	ht = make([]int, N)
	runningPeerCounter := 0
        qArgsa := []string{"a"}
        qArgsb := []string{"b"}
	qAPIArgs := []string{"example02", "query", threadutil.GetPeer(0)}

	// loop through and query all hosts to determine the current values
	for n:=0; n < N; n++ {
		queryData[n].resA = 0
		queryData[n].resB = 0
		ht[n] = 0
		if peerIsRunning(n,mynetwork) {
			runningPeerCounter++
			ht[n], _ = chaincode.GetChainHeight(threadutil.GetPeer(n))
			qAPIArgs = []string{ "example02", "query", threadutil.GetPeer(n) }
			chco2_QueryOnHost(qAPIArgs, qArgsa, qArgsb, &queryData[n].resA, &queryData[n].resB)
		}
        	if Verbose { fmt.Println(fmt.Sprintf("QueryAllHosts() found on Peer %d :  A=%d, B=%d, CH=%d", n, queryData[n].resA, queryData[n].resB, ht[n])) }
	}

	// loop through to determine if we have consensus, and obtain the consensus values
	if runningPeerCounter > 2*F {
		// there are enough peers running to have a chance to find consensus
		for i := 0 ; i <= F ; i++ {
			if ht[i] > 0 {
				candidate_cntr := 1
				candidate_A := queryData[i].resA
				candidate_B := queryData[i].resB
				candidate_CH := ht[i]
				for j := i+1 ; j < N ; j++ {
					if queryData[j].resA == candidate_A && queryData[j].resB == candidate_B && ht[j] == candidate_CH {
						candidate_cntr++
						if Verbose { fmt.Println("QueryAllHosts(): values match on peers ", i, j) }
					}
				}
				if candidate_cntr >= 2*F+1 { *a = candidate_A; *b = candidate_B; *ch = candidate_CH; found_consensus = true; break }
			}
		}
	} else { fmt.Println("QueryAllHosts(): NOT ENOUGH RUNNING Peers TO FIND CONSENSUS! #running/#total = ", runningPeerCounter, N) }

	if !found_consensus { fmt.Println("QueryAllHosts(): CANNOT FIND CONSENSUS!") }

	if Verbose { fmt.Println("QueryAllHosts() returning: A, B, CH, found_consensus? : ", *a, *b, *ch, found_consensus) }

	return found_consensus
}

func TestsCurrentlyPass() bool {
	rc := true
	//QueryAllPeers("STEP CHECK Tests Status")
	if (EnforceQueryTestsPass && !queryTestsPass) || (EnforceChainHeightTestsPass && !chainHeightTestsPass) {
		queryTestsPass = true 
		chainHeightTestsPass = true
		//fmt.Println("RecheckCurrentStatus: Tests still not passing.")
		QueryAllPeers("STEP RECHECK Tests Status")
	}
	if (EnforceQueryTestsPass && !queryTestsPass) || (EnforceChainHeightTestsPass && !chainHeightTestsPass) { rc = false }
	return rc
}

func WaitAndConfirm(sleepExtra int) {
	if (EnforceQueryTestsPass && !queryTestsPass) || (EnforceChainHeightTestsPass && !chainHeightTestsPass) {
		queryTestsPass = true 
		chainHeightTestsPass = true
		fmt.Println("WaitAndConfirm: Tests still not passing. Sleep extra (" + strconv.Itoa((int)(sleepExtra)) + " secs) and check again...")
		time.Sleep(SleepTimeSeconds(sleepExtra))
		QueryAllPeers("STEP to WAIT EXTRA TIME and CHECK AGAIN to see if all nodes catch up.")
	}
}

func CatchUpAndConfirm() {
	// Calling this is optional. If you just care about a "current status", to see if
	// everything eventually catches up and synchronizes, then call this method;
	// it will send enough invokes to ensure all active nodes catch up, and then
	// sleep a long time, and finally query all active nodes to confirm.

	if enoughPeersRunningForConsensus() {

	    if (EnforceQueryTestsPass && !queryTestsPass) || (EnforceChainHeightTestsPass && !chainHeightTestsPass) {

		if (Verbose) { fmt.Println("\nCATCH UP AND CONFIRM RESULTS: Something failed along the way; send enough Invokes for peers to catch up, and sleep awhile, and then recheck.") }

		// reset TestsPass booleans: all we care about is the final query, i.e. if all peers have caught up
		// finally and we pass tests here, then we probably can ignore all the preliminary test failures
		// and consider those failures as misunderstood behavior expectations - maybe they were
		// stale values, observable due to delay/timing issues (test expectation errors)

		queryTestsPass = true 
		chainHeightTestsPass = true

		// Do not set the CHsMustMatchExpected boolean to true. If they are not already set (indicating a tested scenario),
		// then we cannot guarantee our code logic and the consensus network will come up with the same answers.
		// // CHsMustMatchExpected = true
		// 
		// QsMustMatchExpected should already be true (unless specific testcases disable it for
		// their own reasons, to workaround known issues).
		// // QsMustMatchExpected = true
		// 
		// Do not set AllRunningNodesMustMatch to true. It won't work; apparently, all nodes may NOT catch up as initially thought.
		// It still can be useful after init, but only until any nodes stopped/paused and then restarted. Once they get out of
		// sync and lag, there is not much that will make them sync up (unless there are exactly 2f+1 nodes running.)
		// AllRunningNodesMustMatch = true
		// 
		// Just check if enough peers for consensus now match each other; somehow 
		// it is not guaranteed that lagging peers will catch up, so don't bother.
		// Consensus is working if a peer that is "running but not caught up"
		// will catch up when needed (when a peer stops, leaving only 2f peers running)
		// after another set of invokes totalling 

		numInvokes := InvokesRequiredForCatchUp		// send enough invokes to ensure queues emptied

		Invokes( numInvokes )

		// sleep again, to allow double the expected processing time, to help ensure all transactions are processed
		if (Verbose) { fmt.Println("Sleep extra time...") }
		time.Sleep(sleepTimeForTrans(numInvokes))

		QueryAllPeers("STEP to CATCH UP AND CONFIRM RESULTS after extra invokes and sleep")

		// if still not passing, wait again extra time, with NO more invokes, and recheck.
		WaitAndConfirm( ((numInvokes * 3) / TransPerSecRate) + batchtimeout )

	    } else { fmt.Println("CatchUpAndConfirm: All enforced test types already passed. Hooray!") }

	} else { if (Verbose) { fmt.Println("CatchUpAndConfirm: CANNOT try, because not enough peers for consensus are running") } }
}

func DeployNew(a int, b int) {
	// find and use the highest numbered node in the network that is running
	peerNum := NumberOfPeersInNetwork-1
        for ; peerNum >=0 ; peerNum-- {
        	if peerIsRunning(peerNum,MyNetwork) { break }
	}
	DeployNewOnPeer(a, b, peerNum)
}

func DeployNewOnPeer(a int, b int, peer int) {
	if peer < 0 || peer >= NumberOfPeersInNetwork {
		panic(errors.New("DeployNew : Invalid value for peer (" + strconv.Itoa(peer) + "). Expecting 0.." + strconv.Itoa(NumberOfPeersInNetwork-1)))
	}
	strA := strconv.Itoa(a)
	strB := strconv.Itoa(b)
	if initA == strA && initB == strB {
		fmt.Println("\nPOST/Chaincode: NEW DEPLOY, on peer " + threadutil.GetPeer(peer) + ", using SAME INIT VALUES (and therefore no new chaincode instance, so this will be ignored), A=" + strA + " B=" + strB)
		// same values for A and B ==>
		// the request will be mapped to same hash ==>
		// there will NOT be a new network deployed with new values ==>
		// afterwards we will continue to use the current (prior) existing expected values with the old network hash
		// Essentially, the Deploy transaction is treated by the peer network like a No-OP. There is no such thing as a "reset".
	} else {
		// A new chaincode instance (and hash) will be created on each peer node, for this new deployed network.
		// Our GO SDK will be using the new values from now on, so set our internal values accordingly.
		// (To access the old one too, refer to usage of deployUsingTagName())
		currA = a
		currB = b
		initA = strA
		initB = strB
	}
	DeployInit(peer)
}

func DeployInit(peerNum int) {
	peerStr := threadutil.GetPeer(peerNum)
	fmt.Println("\nPOST/Chaincode: DEPLOY chaincode on peer " + peerStr + ", A=" + initA + " B=" + initB)
	dAPIArgs := []string{"example02", "init", peerStr}
	depArgs := []string{"a", initA, "b", initB}
	txId, err := chaincode.DeployOnPeer(dAPIArgs, depArgs)
	Check(err) 	// if we cannot deploy, then panic
	if (Verbose) { fmt.Println("Sleep 60 secs, after deployed, txId=" + txId) }
	time.Sleep(60 * time.Second)
	incrHeightCount(1, peerNum)
	setQueuedTransactionCounter(1)

	// Query the network and Update the counters A, B, and CH.
	if ExistingNetwork {
		// We could be initializing/deploying with the default init values, or could be new values.
		// Even if they are "new" values, they could be some that were used before in prior tests
		// or when rerunning this test using same existing network of peers.

		// STANDARD (RE)DEPLOYMENT OVERRIDE:
		// Useful for rerunning this testcase and others, to use the existing network with existing deployment
		// (by using same counters), rather than creating yet another deployment for the existing peers...
		// but that means the values for A and B will not be reset to 1 million, because they already exist with
		// other values, so we need to go find the actual CURRENT values for A & B

		success := QueryAllHostsToGetCurrentValues(MyNetwork, &currA, &currB, &currCH)
		if !success {
			fmt.Println("setup_part3_verify: CANNOT find consensus in existing network; chainheight/A/B invoke and query tests will likely fail to match any expected values!!!")
			// panic(errors.New("setup_part3_verify: CANNOT find consensus in existing network"))
		}
	}
	if Verbose { fmt.Println("setup_part3_verify, AFTER deploy,QueryAllHosts: A/B/chainheight values: ", currA, currB, currCH) }
}

func Invokes(totalNumInvokes int) {
	// count the num running peers, and determine numInvokes to send to each peer
	numPeersRunning := getNumberOfPeersRunning()
	if (numPeersRunning == 0) {
        	fmt.Println("Invokes: ERROR: CANNOT send INVOKEs : no peers running!")
		return
	}
	if (totalNumInvokes <= 0) {
        	fmt.Println("Invokes: WARNING: no INVOKEs requested")
		return
	}
        fmt.Println("\nPOST/Chaincode: INVOKEs total (" + strconv.Itoa(totalNumInvokes) + ") divided among all " + strconv.Itoa(numPeersRunning) + " running peers")
	numInvokesPerPeer := totalNumInvokes / numPeersRunning
	extras := totalNumInvokes % numPeersRunning
	runningPeerCounter := 0
	firstOne := true
        for peerNum := 0; runningPeerCounter < numPeersRunning && peerNum < NumberOfPeersInNetwork; peerNum++ {
        	if peerIsRunning(peerNum,MyNetwork) {
			runningPeerCounter++
			if firstOne {
				firstOne = false
				doInvoke(numInvokesPerPeer + extras, threadutil.GetPeer(peerNum))
				incrHeightCount(numInvokesPerPeer + extras, peerNum)
				if numInvokesPerPeer == 0 { break }
			} else {
				doInvoke(numInvokesPerPeer, threadutil.GetPeer(peerNum))
				incrHeightCount(numInvokesPerPeer, peerNum)
			}
		}
	}

	if (runningPeerCounter > 0) {
        	setQueuedTransactionCounter(totalNumInvokes)
	} else {
		fmt.Println("Invokes: ERROR: CANNOT send INVOKEs; runningPeerCounter = " + strconv.Itoa(runningPeerCounter))
	}
}

func InvokeOnEachPeer(numInvokesPerPeer int) {
	runningPeerCounter := 0
        fmt.Println("\nPOST/Chaincode: INVOKEs (" + strconv.Itoa(numInvokesPerPeer) + ") being sent to each running peer")
        for peerNum := 0; peerNum < NumberOfPeersInNetwork; peerNum++ {
        	if peerIsRunning(peerNum,MyNetwork) {
			doInvoke(numInvokesPerPeer, threadutil.GetPeer(peerNum))
			incrHeightCount(numInvokesPerPeer, peerNum)
			runningPeerCounter++
		}
	}
	if (runningPeerCounter > 0) {
		setQueuedTransactionCounter(runningPeerCounter * numInvokesPerPeer)
	} else {
		fmt.Println("InvokeOnEachPeer: WARNING: CANNOT send INVOKEs; no peers are running!")
	}
}

func invokeOnAnyPeer(totalNumInvokes int) {
        fmt.Println("\nPOST/Chaincode: INVOKEs (%d) using first available peer", strconv.Itoa(totalNumInvokes))
	sent := false
        for peerNum := 0; peerNum < NumberOfPeersInNetwork; peerNum++ {
        	if peerIsRunning(peerNum,MyNetwork) {
			doInvoke(totalNumInvokes, threadutil.GetPeer(peerNum))
			incrHeightCount(totalNumInvokes, 0)
        		setQueuedTransactionCounter(totalNumInvokes)
			sent = true
			break
		}
	}
	if !sent { fmt.Println("invokeOnAnyPeer: WARNING: CANNOT send INVOKEs; no peers are running!") }
}

func InvokesUniqueOnEveryPeer() {
	powerOf2 := 1
	for i := 0 ; i < NumberOfPeersInNetwork ; i++ {
       		if peerIsRunning(i,MyNetwork) { InvokeOnThisPeer( powerOf2, i ) }
		powerOf2 = powerOf2 * 2
	}
}

func InvokeOnThisPeer(totalNumInvokes int, peerNum int) {
        fmt.Println("\nPOST/Chaincode: INVOKEs (" + strconv.Itoa(totalNumInvokes) + ") using peer " + strconv.Itoa(peerNum))
       	if peerIsRunning(peerNum,MyNetwork) {
		doInvoke(totalNumInvokes, threadutil.GetPeer(peerNum))
		incrHeightCount(totalNumInvokes, 0)
        	setQueuedTransactionCounter(totalNumInvokes)
	} else {
		if Verbose { fmt.Println("InvokeOnThisPeer: ERROR: CANNOT send INVOKEs; peer " + strconv.Itoa(peerNum) + " is not running!") }
	}
}

func incrHeightCount(numInvokesOnThisPeer int, thisPeerNum int) {

	// PREcondition: The associated peer should be RUNNING, otherwise we won't be called (and
	// we shouldn't be queueing up any transactions since the peer isn't there to receive them and queue them).

        if !peerIsRunning(thisPeerNum,MyNetwork) { return }

	// Our current height count actually represents the actual height count.
	// We have NOT already incremented our height counter for these transactions; we will do that now
	// IF we are running WITH consensus working - otherwise queue the transactions and defer the job
	// of computing and incrementing the heightcount until later when we empty the queue when Consensus resumes.
	// Depending on the number of transactions now, versus those in the queue later, they could be
	// bundled/batched differently later than they would be now, so we cannot calulate reliably now.

	if enoughPeersRunningForConsensus() {
		countChainBlocks(numInvokesOnThisPeer)
	} else {
		qtransPerPeerForCH[thisPeerNum] += numInvokesOnThisPeer
	}
}

func countChainBlocks(numInvokesOnThisPeer int) {
        // NOTE: to be called ONLY from incrHeightCount
        // PRECONDITION: THIS peer (which received these invokes) is running, and consensus is working.

        newBlocks := 0
        queuedBlocks := 0

        // Calculate the number of blocks that we will need for all the given "new" transactions
        newBlocks += numInvokesOnThisPeer / batchsize   //  = # full batches
        if (numInvokesOnThisPeer % batchsize > 0) {     //  = Extra transactions are sent in one batch when
                newBlocks += 1                          //    bundle/batchtimeout expires (default every 2 secs)
        }

        // Since consensus is working, there is no need for any queued transactions on any running peer.
        // If this is the handling of the first transactions since Consensus resumed, and/or
        // since a peer node rejoined the network, then we need to count all the queued CH transactions of running nodes.
        // (This may not be precisely correct in all crazy scenarios, but should serve most cases.)

// ### NOTE: this code may be simplified and this LOOP would not be needed if we call this function incrHeightCount for every running peer already...

        peerNum := 0
        for peerNum < NumberOfPeersInNetwork {
                if peerIsRunning(peerNum,MyNetwork) {
                        if (qtransPerPeerForCH[peerNum] > 0) {
                                // Calculate the number of chain blocks needed for all these transactions
                                // on this peer's queue. Since the queue was not empty, the peer probably
                                // recently rejoined the network (or the consensus network itself resumed
                                // operation). This means it would have recently bundled up (or "batched")
                                // all the queued transactions, so let's count them now.

                                queuedBlocks += qtransPerPeerForCH[peerNum] / batchsize
                                if (qtransPerPeerForCH[peerNum] % batchsize) > 0 {
                                        queuedBlocks += 1
                                }
                                qtransPerPeerForCH[peerNum] = 0         // and clear the counter
                        }
                }
                peerNum++
        }
        if (Verbose) { fmt.Println("Increment current ChainHeightBlockCount (" + strconv.Itoa(currCH) + "): + newBlocks(" + strconv.Itoa(newBlocks) + ") + queuedBlocks(" + strconv.Itoa(queuedBlocks) + ")" ) }
        currCH += newBlocks + queuedBlocks
}

func setQueuedTransactionCounter(numTrans int) {
	// Our current A and B counters do not always exactly correspond to the actual A & B chaincode values.
	// Our currA and currB could be higher than the actual ones, because we count those that
	// are queued up whenever Consensus is not working.
	// "qtrans" is the running total NumberOfTotalTransactionsSinceWeHadEnoughPeersForConsensus,
	// and thus represents the difference between currA/currB and the actual chaincode values of A/B.
	
	if enoughPeersRunningForConsensus() {
		if qtrans > 0 {
			if (Verbose) { fmt.Println("Sleep extra to allow processing queued transactions...") }
			time.Sleep(sleepTimeForTrans(qtrans))
		}
        	// Since we have enough nodes running to provide consensus, then reset qtrans to 0 because
		// our transactions will be processed immediately by the peer and network.
		qtrans = 0
	} else {
		// Otherwise increase qtrans by the new number of transactions
		qtrans += numTrans
	}
}

func SleepTimeSeconds(secs int) time.Duration {
	return ( time.Duration(secs) * 1 * time.Second )
}

func SleepTimeMinutes(mins int) time.Duration {
	return ( time.Duration(mins) * SleepTimeSeconds(60) )
}

func sleepTimeForTrans(nTrans int) time.Duration {
	// Given the number of transactions to be processed, determine the sleep for an amount of time (seconds)
	// based on a predetermined processing rate, e.g.:
	//     if expected rate is 2 transactions per second, and we receive 20 transactions, then we will sleep 10 secs

	numSecs := nTrans / TransPerSecRate

	// If there are transactions queued, then add sleep time for processing them too
	if qtrans > 0 {
		numSecs += qtrans / TransPerSecRate
	}

	// To enable some deterministic testing in low-volume testcases ... let's
	// ensure all are batched and processed.
	// The timer is short (default 2 secs), so this shouldn't have a big impact on test duration.

	if numSecs < batchtimeout {
		numSecs = batchtimeout
		//numSecs = 1 	//reduce sleep to only 1 sec, just for some of our "catchup" tests
	}

	// convert to msecs and time.Duration format, and return

	return ( SleepTimeSeconds(numSecs) )
}

// This func can be used by chco2 funcs as well as external funcs such as chcotest/BasicFuncExistingNetwork.go
// that have their own network (and which did not call chco2 setup/init functions)

func peerIsRunning(peerNum int, mynetwork peernetwork.PeerNetwork) bool {
	if peerNum < len(mynetwork.Peers) {
		if (mynetwork.Peers[peerNum].State == peernetwork.RUNNING) {
			return true
		}
	}
	return false
}

func getNumberOfPeersRunning() int {
	numPeersRunning := 0
	for i:=0; i < NumberOfPeersInNetwork; i++ {		//  NumberOfPeersInNetwork is len(MyNetwork.Peers)
		if peerIsRunning(i,MyNetwork) { numPeersRunning++ }
	}
	return numPeersRunning
}

func enoughPeersRunningForConsensus() bool {
	if (getNumberOfPeersRunning() >= NumberOfPeersNeededForConsensus) { 		// or MinNumberOfPeersNeededForConsensus ???
		return true
	}
	return false
}

func QueryAllPeers(stepName string) {

	// SIDE NOTE: After starting a peer node, if EnforceQueryTestsPass is enabled/true, then
	// hopefully we sent enough invoke transactions to ensure all are in sync before querying.
	//     K * batchsize * logmultiplier
	// The relevant parms are found fabric/consensus/obcpbft/config.yaml (and others are in peer/core.yaml)
	//     = 20000, when using: K=10, logmultiplier=4, batchsize=500
	//     =    80, when using: K=10, logmultiplier=4, batchsize=2
	// CORE_PBFT_GENERAL_MODE=batch
	// CORE_PBFT_GENERAL_K=10
	// CORE_PBFT_GENERAL_LOGMULTIPLIER=4
	// CORE_PBFT_GENERAL_BATCHSIZE=2

	fmt.Println("\nPOST/Chaincode: QUERY all running peers for a and b, and chainheight\n" + stepName)
        qArgsa := []string{"a"}
        qArgsb := []string{"b"}
	qAPIArgs := []string{"example02", "query", threadutil.GetPeer(0) }
	n := 0
	for n=0; n < NumberOfPeersInNetwork; n++ {
		qData[n].resA = 0
		qData[n].resB = 0
		if peerIsRunning(n,MyNetwork) {
			qAPIArgs = []string{ "example02", "query", threadutil.GetPeer(n) }
			chco2_QueryOnHost(qAPIArgs, qArgsa, qArgsb, &qData[n].resA, &qData[n].resB)
		}
	}

// TODO: here we may need to add code -
// but DO WE NEED TO KNOW IF AN INVOKE HAS OCCURRED TOO (SINCE ANY NODE WAS RESTARTED)?
// if enoughPeersRunningForConsensus() then call directly a new function to process the
// qtrans queue (and maybe? CH queues), to update our curr/expected values. (And remove that code out of the other places.)
// That way, qtrans will be correctly cleared (or else remain set) before comparisons are made below.
// (The Invoke is not needed to clear the backlog of queued transactions;
// an Invoke is needed only to help get a newly started/joined node up to speed.)

	// Validate all the query results obtained from all the peers; are they what is needed for success?

	if QsMustMatchExpected {
		// the queried result values (A & B) must match the internal counters - also known as the
		// "expected" values (currA & currB), plus-or-minus the queued transactions counters
		passedCount := 0
		for n=0; n < NumberOfPeersInNetwork; n++ {
			if peerIsRunning(n,MyNetwork) {
				if validPeerQueryResults(currA+qtrans, currB-qtrans, qData[n].resA, qData[n].resB, threadutil.GetPeer(n)) {passedCount++}
			}
		}
		printQtrans()

		if enoughPeersRunningForConsensus(){
			if ((passedCount < NumberOfPeersNeededForConsensus) || (AllRunningNodesMustMatch && (passedCount < getNumberOfPeersRunning()))) {
				// FAILURE
               			myStr := fmt.Sprintf("FAILED QUERY TEST: the required peers do NOT match!!!!!!!!!!\nEXPECTED A/B: %9d %9d.\nACTUALs:", currA, currB)
        			for n = 0; n < NumberOfPeersInNetwork; n++ {
        				if peerIsRunning(n,MyNetwork) { myStr += fmt.Sprintf("\nPeer%2d        %9d %9d", n, qData[n].resA, qData[n].resB) }
				}
               			fmt.Println(myStr)

				handleQueryFailure(stepName)
			} else {
				// PASS, Match Expected
				myStr := fmt.Sprintf("PASSED QUERY TEST: Expected A/B (%d/%d) MATCHED on enough/appropriate Peers. ACTUALs (node:A/B): ", currA, currB)
        			for n = 0; n < NumberOfPeersInNetwork; n++ {
        				if peerIsRunning(n,MyNetwork) { myStr += fmt.Sprintf("%d:%d/%d ", n, qData[n].resA, qData[n].resB) }
				}
               			fmt.Println(myStr)
			}
		} else {
				myStr := fmt.Sprintf("SKIPPED QUERY VALIDATION: not enough peer nodes running for consensus. Expected A/B (%d/%d). ACTUALs (node:A/B): ", currA, currB)
        			for n = 0; n < NumberOfPeersInNetwork; n++ {
        				if peerIsRunning(n,MyNetwork) { myStr += fmt.Sprintf("%d:%d/%d ", n, qData[n].resA, qData[n].resB) }
				}
               			fmt.Println(myStr)
		}

	} else {
		// validate that enoughPeersRunningForConsensus() contain the same values (which are now stored in qData[]) -
		// which would mean that they are in sync - but it does not have to equal the internal "expected" values currA and currB

		if enoughPeersRunningForConsensus() {
			// we probably could restructure this section, putting all logic into validPeerQueryResults or other function.

			foundEnoughInConsensus := false
			consensusValueA := 0
			consensusValueB := 0
			consensusValueCount := 0
			for n=0; n < NumberOfPeersInNetwork && !foundEnoughInConsensus; n++ {
				currentPeerValueOfA := qData[n].resA
				currentPeerValueOfB := qData[n].resB
				if currentPeerValueOfA != 0 || currentPeerValueOfB != 0 {
					currentCount := 1
					for p := n+1; p < NumberOfPeersInNetwork; p++ {
						if qData[p].resA == currentPeerValueOfA && qData[p].resB == currentPeerValueOfB { currentCount++ } 
					}
					if currentCount >= NumberOfPeersNeededForConsensus  {
						consensusValueCount = currentCount
						consensusValueA = currentPeerValueOfA
						consensusValueB = currentPeerValueOfB
						foundEnoughInConsensus = true
					}
				}
			}
			if foundEnoughInConsensus {
				// PASS, Consensus
				myStr := fmt.Sprintf("PASSED QUERY TEST: Enough (%d) peers agree for Consensus (required=%d) with values A/B %d/%d. It is not required to match expected values A/B %d/%d. ACTUALs (node:A/B): ", consensusValueCount, NumberOfPeersNeededForConsensus, consensusValueA, consensusValueB, currA, currB)
        			for n = 0; n < NumberOfPeersInNetwork; n++ {
        				if peerIsRunning(n,MyNetwork) { myStr += fmt.Sprintf("%d:%d/%d ", n, qData[n].resA, qData[n].resB) }
				}
               			fmt.Println(myStr)
			} else {
				// FAILURE
               			myStr := fmt.Sprintf("FAILED QUERY TEST: peers do not agree!!!!!!!!!! (even though it is NOT required to match Expected A/B %d/%d.\nACTUALs:", currA, currB)
        			for n = 0; n < NumberOfPeersInNetwork; n++ {
        				if peerIsRunning(n,MyNetwork) { myStr += fmt.Sprintf("\nPeer%2d        %9d %9d", n, qData[n].resA, qData[n].resB) }
				}
               			fmt.Println(myStr)

				handleQueryFailure(stepName)
			}

		} else {
				myStr := fmt.Sprintf("SKIPPED QUERY VALIDATION: not enough peer nodes running for consensus. It is not required to match expected values A/B %d/%d. ACTUALs (node:A/B): ", currA, currB)
        			for n = 0; n < NumberOfPeersInNetwork; n++ {
        				if peerIsRunning(n,MyNetwork) { myStr += fmt.Sprintf("%d:%d/%d ", n, qData[n].resA, qData[n].resB) }
				}
               			fmt.Println(myStr)
		}
	}

	// here we do the same checks for chainheight, but wrapped in a function...
	// if (!validChainHeights()) {
	validated, _ := checkAllChainHeights(true)
	if !validated {
		 handleChainHeightFailure(stepName)
	}
}

func printQtrans() {
        if (Verbose) {
                myOutStr := fmt.Sprintf(" qtrans (total) = %5d", qtrans)
                for n:=0; n < NumberOfPeersInNetwork; n++ {
                        myOutStr += fmt.Sprintf("\n qtransPerPeerForCH[%2d] = %5d", n, qtransPerPeerForCH[n])
                }
                fmt.Println(myOutStr)
        }
}

func handleQueryFailure(stepName string) {
	queryTestsPass = false
	if ( Stop_on_error && EnforceQueryTestsPass ) {
		myOutStr := CurrentTestName + " FAILURE during QUERY : " + stepName
		fmt.Fprintln(Writer, myOutStr)		// write to the output results file
		Writer.Flush()
		log.Fatal (myOutStr)			// write to stdout, and stop the test
	}
}

func handleChainHeightFailure(stepName string) {
	chainHeightTestsPass = false
	if ( Stop_on_error && EnforceChainHeightTestsPass ) {
		myOutStr := CurrentTestName + " FAILURE with CHAINHEIGHT : " + stepName
		fmt.Fprintln(Writer, myOutStr)		// write to the output results file
		Writer.Flush()
		log.Fatal (myOutStr)			// write to stdout, and stop test
	}
}

func StopPeers(peerNumsToStopStart []int) {
	rootPeer := false
	if (len(peerNumsToStopStart) == 0) {
		if pauseInsteadOfStop { fmt.Println("\nPAUSE Peers:  [none requested]")
		} else {                fmt.Println("\nSTOP Peers:  [none requested]") }
	} else {
		myOutStr := fmt.Sprintf("\n")
		if pauseInsteadOfStop { myOutStr += fmt.Sprintf("PAUSE Peers():")
		} else {                myOutStr += fmt.Sprintf("STOP Peers():") }

		var peersToStopStart []string
		peersToStopStart = make([]string, NumberOfPeersInNetwork)

		//  if !buildPeersList(peerNumsToStopStart, &peersToStopStart, &myOutStr) { return }
		i:= 0
		for i < len(peerNumsToStopStart) {
			peerNum := peerNumsToStopStart[i]
			peerName := fmt.Sprintf(threadutil.GetPeer(peerNum))
			myOutStr += "  " + peerName
			if peerNum >= len(MyNetwork.Peers) { 	// if peerName is not in (MyNetwork.Peers)
				myOutStr += fmt.Sprintf(" --> Peer NOT FOUND! Returning without touching any peer nodes!")
				fmt.Println(myOutStr)
				return 
			} else {
				if (MyNetwork.Peers[peerNum].State != peernetwork.RUNNING) {
					myOutStr += fmt.Sprintf("(alreadyNotRUNNING)")
				} else {
					if (peerNum == 0) || ((peerNum==1) && !peerIsRunning(0,MyNetwork) && !rootPeer) {
						// TODO: enhance for larger networks when more nodes could be down.

						rootPeer = true	// we are impacting the primary/root peer

						// incr is needed here for CAT_07 and others - but NOT in all similar scenarios where rootPeer is stop/restarted,
						// so THIS COULD USE more analysis, if we are going to try to validate CH in testcases
						//   incrHeightCount(1, peerNum)
					}
				}
			}
			peersToStopStart[i] = peerName
			i++
		}
		fmt.Println(myOutStr)

		//peernetwork.StopPeersLocal(MyNetwork, peersToStopStart)

		for j:=0; j < i; j++ {
			if pauseInsteadOfStop {
				peernetwork.PausePeerLocal(MyNetwork, peersToStopStart[j]) 	// includes sleeping 5 secs after each Pause
			} else {
				peernetwork.StopPeerLocal(MyNetwork, peersToStopStart[j]) 	// includes sleeping 5 secs after each Stop
			}
		}
		if (rootPeer) {
			// sleep extra when stopping/starting primary/root peer0
			fmt.Println("Sleep 10 secs more after stop, + extra 20 secs because stopping primary") 
			time.Sleep(30 * time.Second) 
		} else {
			fmt.Println("Sleep 10 secs more after stop")
			time.Sleep(10 * time.Second) 
		}
	}
}

func RestartPeers(peerNumsToStopStart []int) {
	rootPeer := false
	if (len(peerNumsToStopStart) == 0) {
		if pauseInsteadOfStop { fmt.Println("\nUNPAUSE Peers:  [none requested]")
		} else {                fmt.Println("\nRESTART Peers:  [none requested]") }
	} else {
		myOutStr := fmt.Sprintf("\n")
		if pauseInsteadOfStop { myOutStr += fmt.Sprintf("UNPAUSE Peers():")
		} else {                myOutStr += fmt.Sprintf("RESTART Peers():") }

		var peersToStopStart []string
		peersToStopStart = make([]string, NumberOfPeersInNetwork)

		//  if !buildPeersList(peerNumsToStopStart, &peersToStopStart, &myOutStr) { return }
		i:= 0
		for i < len(peerNumsToStopStart) {
			peerNum := peerNumsToStopStart[i]
			peerName := fmt.Sprintf(threadutil.GetPeer(peerNum))
			myOutStr += "  " + peerName
			if peerNum >= len(MyNetwork.Peers) { 	// if peerName is not in (MyNetwork.Peers)
				myOutStr += fmt.Sprintf(" --> Peer NOT FOUND! Returning without touching any peer nodes!")
				fmt.Println(myOutStr)
				return 
			}
			if (MyNetwork.Peers[peerNum].State == peernetwork.RUNNING) {
				myOutStr += fmt.Sprintf("(alreadyRUNNING)")
			}
			if (peerNum == 0) || ((peerNum==1) && !peerIsRunning(0,MyNetwork)) {
					// TODO: enhance "if" check for larger networks when more nodes could be down.

					rootPeer = true		// we are impacting the primary/root peer0

					// TODO incrHeightCount(1, peerNum) 	// is incr needed here? Observed no consistency in all similar
										// scenarios where rootPeer is stop/restarted, so THIS COULD USE IMPROVEMENT
			}
			peersToStopStart[i] = peerName
			i++
		}
		fmt.Println(myOutStr)
		//peernetwork.StartPeersLocal(MyNetwork, peersToStopStart)
		for j:= 0; j < i; j++ {
			// Once we stop and restart at least one peer node, (assuming StartPeerLocal() was successful),
			// CH/A/B may be short/wrong in any extra nodes beyond the number required for consensus.
			// Set false because the testcases shouldn't fail as long as we maintain consensus -
			// but only when the node we are restarting is extra (more than the minimum required for consensus).

			if enoughPeersRunningForConsensus() {
				// We already have enough peer nodes running for consensus, so
				// this one will be extra and therefore does not have to sync up exactly.

				AllRunningNodesMustMatch = false
			}

			//if pauseInsteadOfStop || (MyNetwork.Peers[peerNumsToStopStart[j]].State == peernetwork.PAUSED) {
			if pauseInsteadOfStop {
				peernetwork.UnpausePeerLocal(MyNetwork, peersToStopStart[j]) 	// includes sleeping 5 secs after each Unpause
			} else {
				peernetwork.StartPeerLocal(MyNetwork, peersToStopStart[j]) 	// includes sleeping 5 secs after each Restart
			}
		}
		if (rootPeer) {
			// sleep extra when stopping/starting primary/root peer
			fmt.Println("Sleep 60 secs more after a restart = 30 plus extra 30 secs because restarting potential primary") 
			time.Sleep(60 * time.Second) 
		} else {
			fmt.Println("Sleep 30 secs more after a restart")
			time.Sleep(30 * time.Second) 
		}
		// now that the nodes have had time to recover, retrieve the actual IP addresses (which possibly changed)
		for k:= 0; k < len(peerNumsToStopStart); k++ {
			chaincode.UpdatePeerIp(&MyNetwork, peerNumsToStopStart[k])
		}
	}
}

// currently unused:
// to call:  if !buildPeersList(peerNumsToStopStart, &peersToStopStart, &myOutStr) { return }
func buildPeersList(peerNumsToStopStart []int, peersToStopStart *[]string, myOutStr *string) bool {
		i := 0
		for i < len(peerNumsToStopStart) {
			peerNum := peerNumsToStopStart[i]
			peerName := fmt.Sprintf(threadutil.GetPeer(peerNum))
			*myOutStr = *myOutStr + "  " + peerName
			(*peersToStopStart)[i] = peerName
			if peerNum >= len(MyNetwork.Peers) { 	// if peerName is not in (MyNetwork.Peers)
				*myOutStr = *myOutStr + fmt.Sprintf(" --> Peer NOT FOUND! Returning without touching any peer nodes!")
				fmt.Println(myOutStr)
				return false
			}
			i++
		}
		fmt.Println(myOutStr)
		return true
}

func StopMemberServices() {
	fmt.Println("\n\n\n\nSTOP MemberServices (caserver)!\n\n\n")
	//peernetwork.StopMemberServices(MyNetwork)
	peernetwork.StopPeerLocal(MyNetwork, "caserver")
}

func RestartMemberServices() {
	fmt.Println("\n\n\n\nRESTART MemberServices (caserver)!\n\n\n")
	peernetwork.StartPeerLocal(MyNetwork, "caserver")

	// Is it possible that the IP could have changed? If so, what do we do???
}

func TimeTrack(start time.Time, name string) {
	//fmt.Println("+++ENTERED_TIMETRACK+++")
        elapsed := time.Since(start)
        preStr := ""
        postStr := ""
        myOutStr := fmt.Sprintf(" %s (Q_Pass=%t CH_Pass=%t, Enforce Q=%t CH=%t, MustMatch Q=%t CH=%t AllVP=%t) [%s]  ",
			 		name, queryTestsPass, chainHeightTestsPass, EnforceQueryTestsPass, EnforceChainHeightTestsPass,
					QsMustMatchExpected, CHsMustMatchExpected, AllRunningNodesMustMatch, elapsed)
	if !RanToCompletion {
			// NOTE: If the user types ^C to abort the script and stop running the test,
			// execution should still get here and report a result.
			// A good indicator of an interrupted test is a run time much shorter than usual.
        		preStr += fmt.Sprintf("ABORTED")
        		postStr = fmt.Sprintf("--------------------")
	} else {
		if ( (!queryTestsPass && EnforceQueryTestsPass) || (!chainHeightTestsPass && EnforceChainHeightTestsPass) || !annexTestsPass ) {
        		preStr += fmt.Sprintf("FAILED")
        		postStr = fmt.Sprintf("!!!!!!!!!!!!!!!!!!!!")
        		if !annexTestsPass { postStr += fmt.Sprintf(" (annex tests failed)") }
		} else {
        		preStr += fmt.Sprintf("PASSED")
		}
	}
	fmt.Println("\n" + preStr + myOutStr + postStr + "\n")
	fmt.Fprintln(Writer, preStr + myOutStr + postStr)
	Writer.Flush()

	restore_all()
}

func restore_all() {

//	// This is what we really want to do:    docker ps -aq -f status=paused | xargs docker unpause  1>/dev/null 2>&1
//	// because docker cannot stop or kill or rm containers that are paused, for some reason.

//	cmd := "docker ps -aq -f status=paused | xargs docker unpause  1>/dev/null 2>&1"
//	out, err := exec.Command("/bin/sh", "-c", cmd).Output()
//	if (err != nil) {
//		fmt.Println("restore_all: Could not unpause all peers: ")
//		fmt.Println(out)
//		// log.Fatal(err)
//	}

	cntr := 0
	for i :=0 ; i < NumberOfPeersInNetwork ; i++ {
		// DO NOT leave any nodes paused, or stopped
		if (MyNetwork.Peers[i].State == peernetwork.PAUSED) {
			if Verbose { fmt.Println("restore_all(): unpause peer " + threadutil.GetPeer(i)) }
			peernetwork.UnpausePeerLocal(MyNetwork, threadutil.GetPeer(i))
			chaincode.UpdatePeerIp(&MyNetwork, i)
			cntr++
		} else 
		//if (MyNetwork.Peers[i].State == peernetwork.STOPPED) {
		if (MyNetwork.Peers[i].State != peernetwork.RUNNING) {
			if Verbose { fmt.Println("restore_all(): restart peer " + threadutil.GetPeer(i)) }
			peernetwork.StartPeerLocal(MyNetwork, threadutil.GetPeer(i))
			cntr++
		}
	}
	// if cntr > 0 { fmt.Println("Sleep 20 secs extra after restarting/unpausing peers") }; time.Sleep(20 * time.Second)
}

func clean_up() {
}

func chco2_QueryOnHost(apiArgs00 []string, argsA []string, argsB []string, resAI *int, resBI *int)  {
	resA, _ := chaincode.QueryOnHost(apiArgs00, argsA)
	resB, _ := chaincode.QueryOnHost(apiArgs00, argsB)
	*resAI, _ = strconv.Atoi(resA)
	*resBI, _ = strconv.Atoi(resB)
}

func DoOneInvoke(apiArgs []string, invokeArgs []string) (id string, err error) {
	id, err = chaincode.InvokeOnPeer(apiArgs, invokeArgs)
	currA--
	currB++
	return id, err
	// Hardcoded: we only subtract one from A and add it to B. Transaction size is 1.
	// We could actually improve this by the following, but it would be slower:
	// 	size := strconv.Itoa(invokeArgs[2])
	// 	currA -= size
	// 	currB += size
}

// PREcondition: peer node must be running
func doInvoke(num_invokes int, nodename string)  {

        if Verbose { fmt.Println("doInvoke() calling chaincode.InvokeOnPeer " + strconv.Itoa(num_invokes) + " times on peer " + nodename) }

	// We sleep now only if we have consensus and the invokes can be processed now;
	// otherwise we will sleep when we empty the queue later when consensus is resumed...
	// Get the sleep time based on number of transactions.

	mustSleep := enoughPeersRunningForConsensus()

	startTime := time.Now()
	invArgs := []string{"a", "b", "1"}
	iAPIArgs := []string{"example02", "invoke", nodename}
	for j:=1; j <= num_invokes; j++ {
		_, _ = DoOneInvoke(iAPIArgs, invArgs)

		// if Verbose {
			// Show some progress...
			// Print . for 10 invokes; Print + for 100 invokes; Print newline after 1000 invokes on this peer.
			if j % 10 == 0 {
				if j % 100 == 0 {
					fmt.Printf("+")
					if j % 1000 == 0 { fmt.Println(" ") }
				} else {
					fmt.Printf(".")
				}
			}
		// }
	}

	if mustSleep {
		if num_invokes/TransPerSecRate > batchtimeout {
			if Verbose && num_invokes/TransPerSecRate > 10 {
				fmt.Println("\nSleep up to " + strconv.Itoa(num_invokes/TransPerSecRate) + " secs after sending " + strconv.Itoa(num_invokes) + " invokes")
			}
			// Sleep an amount of time allowed (based on the specified TransPerSec rate) to process the
			// number of invokes we have done, after subtracting the time elapsed since we started.
			time.Sleep( sleepTimeForTrans(num_invokes) - time.Since(startTime) )
		} else { time.Sleep(time.Duration(batchtimeout)*time.Second) }	// sleep at least 2 secs, to give time for the transactions to be batched and
	}									// sent through (so any queries following immediately would be more likely to work)
}

func validPeerQueryResults(a int, b int, resA int, resB int, nodename string) bool {
	var passfail bool
	passfail = true
	valueStr := ""
	if !enoughPeersRunningForConsensus() {
		if ( EnforceQueryTestsPass ) {	// if we care, print status
        		if (Verbose) {
				valueStr = fmt.Sprintf("SKIPPED QUERY VALIDATION on %s: not enough peer nodes running for consensus. EXPECTED/ACTUAL: A=%d/%d, B=%d/%d.", nodename, a, resA, b, resB)
				fmt.Println(valueStr)	// print to stdout only
			}
		}
	} else if (a == resA) && (b == resB) {
		if ( EnforceQueryTestsPass ) {	// if we care, print status
        		if (Verbose) {
				valueStr = fmt.Sprintf("PASS on %s: QUERY RESULTS MATCH expected values: A=%d, B=%d.", nodename, a, b)
				fmt.Println(valueStr)	// print to stdout only
			}
		}
	} else {
		passfail = false
		if ( EnforceQueryTestsPass ) {	// if we care, print status
			valueStr = fmt.Sprintf("FAIL on %s: QUERY RESULTS: EXPECTED/ACTUAL: A=%d/%d, B=%d/%d. *****FAIL*****", nodename, a, resA, b, resB)
        		if (Verbose) {
				fmt.Println(valueStr)
			}
			if ( Stop_on_error ) {
                		fmt.Fprintln(Writer, valueStr)	// print to results file too, since this is the reason we will be stopping shortly
        			Writer.Flush()
			}
		}
	}
	return passfail
}

func QueryMatch(currA int, currB int) { 	// legacy previous API
        qArgsa := []string{"a"}
        qArgsb := []string{"b"}
        fmt.Println("\nPOST/Chaincode: QUERY a and b >>>>>>>>>>> ")
        qAPIArgs00 := []string{"example02", "query", threadutil.GetPeer(0)}
        res0A, _ := chaincode.QueryOnHost(qAPIArgs00, qArgsa)
        res0B, _ := chaincode.QueryOnHost(qAPIArgs00, qArgsb)
        res0AI, _ := strconv.Atoi(res0A)
        res0BI, _ := strconv.Atoi(res0B)

        qAPIArgs01 := []string{"example02", "query", threadutil.GetPeer(1)}
        res1A, _ := chaincode.QueryOnHost(qAPIArgs01, qArgsa)
        res1B, _ := chaincode.QueryOnHost(qAPIArgs01, qArgsb)
        res1AI, _ := strconv.Atoi(res1A)
        res1BI, _ := strconv.Atoi(res1B)

        qAPIArgs02 := []string{"example02", "query", threadutil.GetPeer(2)}
        res2A, _ := chaincode.QueryOnHost(qAPIArgs02, qArgsa)
        res2B, _ := chaincode.QueryOnHost(qAPIArgs02, qArgsb)
        res2AI, _ := strconv.Atoi(res2A)
        res2BI, _ := strconv.Atoi(res2B)

        qAPIArgs03 := []string{"example02", "query", threadutil.GetPeer(3)}
        res3A, _ := chaincode.QueryOnHost(qAPIArgs03, qArgsa)
        res3B, _ := chaincode.QueryOnHost(qAPIArgs03, qArgsb)
        res3AI, _ := strconv.Atoi(res3A)
        res3BI, _ := strconv.Atoi(res3B)

        if (currA == res0AI) && (currB == res0BI) {
                fmt.Println("Results in a and b match with Invoke values on peer 0: PASS")
                valueStr := fmt.Sprintf(" currA : %d, currB : %d, resA : %d , resB : %d", currA, currB, res0AI, res0BI)
                fmt.Println(valueStr)
        } else {
                fmt.Println("******************************")
                fmt.Println("RESULTS DO NOT MATCH on peer 0 : FAIL")
                valueStr := fmt.Sprintf(" currA : %d, currB : %d, resA : %d , resB : %d", currA, currB, res0AI, res0BI)
                fmt.Println(valueStr)

                fmt.Println("******************************")
        }

        if (currA == res1AI) && (currB == res1BI) {
                fmt.Println("Results in a and b match with Invoke values on peer 1: PASS")
                valueStr := fmt.Sprintf(" currA : %d, currB : %d, resA : %d , resB : %d", currA, currB, res1AI, res1BI)
                fmt.Println(valueStr)
        } else {
                fmt.Println("******************************")
                fmt.Println("RESULTS DO NOT MATCH on peer 1 : FAIL")
                valueStr := fmt.Sprintf(" currA : %d, currB : %d, resA : %d , resB : %d", currA, currB, res1AI, res1BI)
                fmt.Println(valueStr)
                fmt.Println("******************************")
        }
        if (currA == res2AI) && (currB == res2BI) {
                fmt.Println("Results in a and b match with Invoke values on peer 2: PASS")
                valueStr := fmt.Sprintf(" currA : %d, currB : %d, resA : %d , resB : %d", currA, currB, res2AI, res2BI)
                fmt.Println(valueStr)
        } else {
                fmt.Println("******************************")
                fmt.Println("RESULTS DO NOT MATCH on peer 2 : FAIL")
                valueStr := fmt.Sprintf(" currA : %d, currB : %d, resA : %d , resB : %d", currA, currB, res2AI, res2BI)
                fmt.Println(valueStr)

                fmt.Println("******************************")
        }

        if (currA == res3AI) && (currB == res3BI) {
                fmt.Println("Results in a and b match with Invoke values on peer 3: PASS")
                valueStr := fmt.Sprintf(" currA : %d, currB : %d, resA : %d , resB : %d", currA, currB, res3AI, res3BI)
                fmt.Println(valueStr)
        } else {
                fmt.Println("******************************")
                fmt.Println("RESULTS DO NOT MATCH on peer 3 : FAIL")
                valueStr := fmt.Sprintf(" currA : %d, currB : %d, resA : %d , resB : %d", currA, currB, res3AI, res3BI)
                fmt.Println(valueStr)

                fmt.Println("******************************")
        }
}

func Check(e error) {
        if e != nil {
                panic(e)
        }
}

func validChainHeights() bool {

	// The expected block chain height is always the same as our currHeightCount.
	// Note: We increment our currCH only when consensus and transactions are
	// processed and bundled/batched into the network. Otherwise, we queue transactions
	// and defer calculation/incrementation of our currCH until later.
	// So our currHeightCount should always match whatever we get/query, even if
	// some more invokes had been sent to a good peer while the network lacked enough
	// nodes for Consensus and therefore could not process them.

	testStatus  := true
	matchedCount := 0
	var ht []int
	ht = make([]int, NumberOfPeersInNetwork)

	runningPeerCounter := 0
        for peerNum := 0; peerNum < NumberOfPeersInNetwork; peerNum++ {
        	if peerIsRunning(peerNum,MyNetwork) {
			ht[peerNum], _ = chaincode.GetChainHeight(threadutil.GetPeer(peerNum))
			if (ht[peerNum] == currCH)  { matchedCount++ }
			runningPeerCounter++
		} else { ht[peerNum] = 0 }
	}

	if (runningPeerCounter >= NumberOfPeersNeededForConsensus) && ((matchedCount < NumberOfPeersNeededForConsensus) || (AllRunningNodesMustMatch && (matchedCount < getNumberOfPeersRunning()))) {
		//handle failure
		testStatus = false
               	myStr := fmt.Sprintf("FAILED CHAIN HEIGHT TEST: required peers do NOT match expected ChainHeight (%d).  Actual CH: ", currCH)
        	for peerNum := 0; peerNum < NumberOfPeersInNetwork; peerNum++ {
        		if peerIsRunning(peerNum,MyNetwork) { myStr += fmt.Sprintf("%s(%d) ", threadutil.GetPeer(peerNum), ht[peerNum]) }
		}
               	myStr += fmt.Sprintf("!!!!!!!!!!")
               	fmt.Println(myStr)					// always print to stdout
		if (Stop_on_error && EnforceChainHeightTestsPass) {	// if we care, print to results file too
                	fmt.Fprintln(Writer, myStr)
			Writer.Flush()
		}
	} else {
		testStatus = true
		myStr := fmt.Sprintf("PASSED CHAIN HEIGHT TEST: Expected height (%d) matched on enough/appropriate Peers. Actual CH: ", currCH)
        	for peerNum := 0; peerNum < NumberOfPeersInNetwork; peerNum++ {
        		if peerIsRunning(peerNum,MyNetwork) { myStr += fmt.Sprintf("%s(%d) ", threadutil.GetPeer(peerNum), ht[peerNum]) }
		}
                fmt.Println(myStr)					// always print to stdout
		if (Stop_on_error && EnforceChainHeightTestsPass) {	// if we care, print status in results file too
                	fmt.Fprintln(Writer, myStr)
			Writer.Flush()
		}
        }
	return testStatus
}

func checkAllChainHeights(printResults bool) (testStatus bool, consensusCH int) {
	// CALL BY: 
	// 	validated, _ := checkAllChainHeights(true)
	// 	if !validated { ...handle error... }
	// OR:
	// 	validated, groupCH := checkAllChainHeights(false)
	// 	if groupCH > 0 { ...we found consensus myNewChainHeight = consensusCH }

	consensusCH = 0
	testStatus 		= true
	enoughMatchExpectedCH 	:= true
	allMatchExpectedCH 	:= true
	allMatchEachOther 	:= true
	consensusPossible 	:= true
	consensusFound 		:= false

	//====================================================================================================================
	// first get the chainheight from each peer node

	var ht []int
	ht = make([]int, NumberOfPeersInNetwork)
	countMatchingExpectedValue := 0
	runningPeerCounter := 0
        for peerNum := 0; peerNum < NumberOfPeersInNetwork; peerNum++ {
        	if peerIsRunning(peerNum,MyNetwork) {
			ht[peerNum], _ = chaincode.GetChainHeight(threadutil.GetPeer(peerNum))
			if (ht[peerNum] == currCH)  { countMatchingExpectedValue++ } 
			runningPeerCounter++
		} else { ht[peerNum] = 0 }
	}

	// Do the chainheights of all the running peers match the EXPECTED value? (STRICT mode, AllRunningNodesMustMatch)
	if countMatchingExpectedValue < runningPeerCounter { allMatchExpectedCH = false }

	// Do chainheights match expected value on the number of peers needed for consensus ?
	if countMatchingExpectedValue < NumberOfPeersNeededForConsensus { enoughMatchExpectedCH = false }

	//====================================================================================================================
	// Determine whether "consensusFound" or "allMatchEachOther"
	// 
	// "consensusFound" = there are enough running peer nodes with matching chainheights for a consensus (LENIENT MODE)
	//	(but not necessarily match the expected value)
	// 
	// "allMatchEachOther" = the chainheights of all the running peers match each other
	//	(and may but not necessarily match the expected value)
	//	(and there may be more or less than enough running nodes to reach consensus - although they ALL match)

	numPeersRunning := getNumberOfPeersRunning() 
	if (numPeersRunning < NumberOfPeersNeededForConsensus) {
		consensusPossible = false
	} else {
		matchCounter := 0
		matchStartPoints := numPeersRunning - NumberOfPeersNeededForConsensus + 1
		for n := 0 ; (n <= NumberOfPeersInNetwork - NumberOfPeersNeededForConsensus) && (matchStartPoints > 0) && !consensusFound; n++ {
        		if peerIsRunning(n,MyNetwork) && ht[n] > 0 {
				matchCounter = 1
				for i := n+1 ; (i < NumberOfPeersInNetwork) ; i++ {
        				if peerIsRunning(i,MyNetwork) {
						if (ht[n] == ht[i]) { matchCounter++ } else { allMatchEachOther = false }
					}
				}
				if (matchCounter >= NumberOfPeersNeededForConsensus) {
					consensusFound = true
					consensusCH = ht[n]
				}
				matchStartPoints--
			}
		}
	}

	//====================================================================================================================

	if printResults {
		myStr := fmt.Sprintf("")
		if (!consensusPossible) {
			myStr += fmt.Sprintf("SKIPPED CHAINHEIGHT VALIDATION: Only %d peer nodes running, but %d are required for consensus in this network of %d. Expected CH (%d). Actual CHs: ", numPeersRunning, NumberOfPeersNeededForConsensus, NumberOfPeersInNetwork, currCH)
        		for peerNum := 0; peerNum < NumberOfPeersInNetwork; peerNum++ {
        			if peerIsRunning(peerNum,MyNetwork) { myStr += fmt.Sprintf("%s(%d) ", threadutil.GetPeer(peerNum), ht[peerNum]) }
			}
                	fmt.Println(myStr)					// always print to stdout
		} else
			// We have enough nodes for consensus. Use cases are the following.
			//   1. Either they all match each other, or,
			//   2. they don't all match - but that is OK because it is not required that they ALL match -
			//      AND we still have enough matching each other for consensus, or,
			//   3. we don't even have enough matching in agreement for consensus.
			// 
			// If (1 or 2), then that is good - but pass only if we meet an additional condition (A or B or C):
			//   A. They do not need to match the expected CH value, or, 
			//   B. They do need to match expected CH value AND all do match, or,
			//   C. They do need to match expected CH value AND enough for consensus match expected value (which is all that is required), or,
			//   D. They do need to match expected CH value, but their value doesn't match the expected value.

		if (allMatchEachOther || (!AllRunningNodesMustMatch && consensusFound)) && (!CHsMustMatchExpected || (allMatchExpectedCH || (enoughMatchExpectedCH && !AllRunningNodesMustMatch))) {
			// SUCCESS
			myStr += fmt.Sprintf("PASSED CHAIN HEIGHT TEST: matches on enough/appropriate Peers. Expected CH (%d). Actual CHs: ", currCH)
        		for peerNum := 0; peerNum < NumberOfPeersInNetwork; peerNum++ {
        			if peerIsRunning(peerNum,MyNetwork) { myStr += fmt.Sprintf("%s(%d) ", threadutil.GetPeer(peerNum), ht[peerNum]) }
			}
                	fmt.Println(myStr)					// always print to stdout
		} else {
			// FAILURE
			testStatus = false
               		myStr += fmt.Sprintf("FAILED CHAIN HEIGHT TEST: enough required peers do NOT match. Expected ChainHeight (%d). Actual CHs: ", currCH)
        		for peerNum := 0; peerNum < NumberOfPeersInNetwork; peerNum++ {
        			if peerIsRunning(peerNum,MyNetwork) { myStr += fmt.Sprintf("%s(%d) ", threadutil.GetPeer(peerNum), ht[peerNum]) }
			}
               		myStr += fmt.Sprintf("!!!!!!!!!!")
               		fmt.Println(myStr)					// always print to stdout
		}
		if (Stop_on_error && EnforceChainHeightTestsPass) {	// if we care, print status in results file too
			fmt.Fprintln(Writer, myStr)
			Writer.Flush()
		}
	}
	return testStatus, consensusCH
}

func AnnexTestPassResult(testStatus bool) {
	annexTestsPass = testStatus
}

