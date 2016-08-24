# obcsdk
opensource blockchain software development kit, GO code and tests for hyperledger fabric

A test framework for testing Blockchain with GoSDK (written in Go Language)

##Obtain the GoSDK and test programs:
Clone to the src directory where GO is installed (use either $GOROOT/src or $GOPATH/src),
and follow instructions (here) to setup the peer network.

	$ cd $GOPATH/src
	$ git clone https://github.com/scottz64/obcsdk.git -o obcsdk

 
##How to execute the programs:
Change the credentials in NetworkCredentials.json accordingly.
Helpful scripts are located in the automation directory.
Go to the test directories and execute the tests.
PEER LOGFILES are saved in the automation directory.
GO_TEST output files are saved in the current working directory, when running script go_record.sh.
	 
	$ cd obcsdk/chcotest
	$ go run BasicFunc.go
	 
	$ cd obcsdk/ledgerstresstest
	$ NETWORK="LOCAL" go run LedgerStressOneCliOnePeer.go Utils.go
	 
	$ cd obcsdk/chco2test
	$ go run IQ.go
	 
	$ cd obcsdk/CAT
	$ go run testtemplate.go
	$ ../automation/go_record.sh CAT*go
	 

