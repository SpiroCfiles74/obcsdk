package main

import (
	"obcsdk/lstutil"
)

/*************** Test Objective : Ledger Stress with 2 Clients and 1 Peers *********************
* 
*   1. Connect to a 4 node peer network with security enabled, and deploy a modified version of
*	chaincode_example02 that stores an additional block of data with every transaction
*	Refer to lstutil.go for more details, including parameters and further configuration.
*   2. Invoke transactions in parallel, divided among each go client thread on the peer
*   3. Check if the expected total counter value matches with query values
* 
*   To use this test script:			go run <testname.go>
*   or (to save output files):			../automation/go_record.sh <testname.go>
* 
***********************************************************************************************/

func main() {
	// args:  testname, # client threads, # peers, total # transactions
	lstutil.RunLedgerStressTest( "LST_2client1peer20K", 2, 1, 20000 )
}
