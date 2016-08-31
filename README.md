# obcsdk
opensource blockchain software development kit, GO code and tests for hyperledger fabric

A test framework for testing Blockchain with GoSDK (written in Go Language)

##Obtain the GoSDK and test programs:
Clone to the src directory where GO is installed (use either $GOROOT/src or $GOPATH/src).

	$ cd $GOPATH/src
	$ git clone https://github.com/scottz64/obcsdk.git -o obcsdk

Enclosed local_fabric bash scripts will create docker containers to run the peers.
For more information, 
[read these instructions] (https://github.com/rameshthoomu/fabric1/blob/tools/localpeersetup/local_Readme.md)
to setup a peer network.
 
##How to execute the programs:
- If you wish to connect to an existing network, change the credentials in NetworkCredentials.json as needed.
- Helpful shell scripts are located in the obcsdk/automation directory:
```
	../automation/go_build_all.sh           - execute this from any of the test directories, to build all the *.go tests there
	../automation/go_record.sh <tests.go>   - execute this from any of the test directories, to run go tests and record stdout logs in GO_TEST* files in the current working directory.
```
- LOGFILES for all Peers are saved in the automation directory. Run go_record.sh (or local_fabric.sh) without parameters to get help with the options.
- Go to the test directories and execute the tests. Good luck!
```
	$ cd obcsdk/chcotest
	$ go run BasicFunc.go
	 
	Run Consensus Acceptance Tests and two long-run tests:
	$ cd obcsdk/CAT
	$ go run testtemplate.go
	$ ../automation/go_record.sh CAT*go
	$ ../automation/go_record.sh CRT_501_StopAndRestartRandom_10Hrs.go
	$ ../automation/go_record.sh CRT_502_StopAndRestart1or2_10Hrs.go
	 
	Run other tests being developed in chco2test directory:
	$ cd obcsdk/chco2test
	$ go run IQ.go
	 
	Run various tests being developed in chcotest directory:
	$ cd obcsdk/chcotest
	$ go run ConsensusBasic.go
	 
	Run ledger stress tests: first start up a network, and connect your tests to it by
	configuring obcsdk/util/NetworkCredentials.json
	$ cd obcsdk/CAT
	$ go run testtemplate.go
	$ cd obcsdk/ledgerstresstest
	$ NETWORK=LOCAL go run LST_2Client2Peer.go
```
- Running tests on Z network requires some tweaking to make things run.
```
	Such as:
	  - define its own usernames/passwords (may need to edit threadutil/threadutil.go)
	  - put the IP address info of the peers into the util/NetworkCredentials.json
	  - set environment variable NET_COMM_PROTOCOL to use HTTPS instead of default HTTP
	    (this is required for ledger stress tests on Z network, but also may be set
	    when running other tests in other networks that could use HTTPS), 
	  - (ledgerstresstests only): set correct CHAINCODE_NAME name in lstutil/util.go
	    to change from mycc (special version of example02) to whatever is being run
	    on the actual Z network to which you are connecting; and rebuild the tests
	    before running them:    "cd lstutil; ../automation/go_build_all.sh"

	For example:

	$ cd CAT; NET_COMM_PROTOCOL=HTTPS go run CAT_101*.go

	$ cd ledgerstresstest; NETWORK=Z NET_COMM_PROTOCOL=HTTPS go run LST_1Client1Peer.go
```
