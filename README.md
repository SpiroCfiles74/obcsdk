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
- If you wish to connect to an existing network, change the credentials in NetworkCredentials.json as needed, 
(and you may also need to modify some of the GO code too).
- Helpful shell scripts are located in the obcsdk/automation directory.
- LOGFILES for all Peers are saved in the automation directory.
- GO_TEST output files are saved in the current working directory, when running script go_record.sh.
- Go to the test directories and execute the tests.
```
	$ cd obcsdk/chcotest
	$ go run BasicFunc.go
	 
	$ cd obcsdk/ledgerstresstest
	$ NETWORK="LOCAL" go run LedgerStressOneCliOnePeer.go Utils.go
	 
	$ cd obcsdk/chco2test
	$ go run IQ.go
	 
	$ cd obcsdk/CAT
	$ go run testtemplate.go
	$ ../automation/go_record.sh CAT*go
```

