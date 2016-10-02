
date; echo "Start new network, and run all LedgerStressTests in the Regression Suite:"
echo "LST_1client1peer20K.go LST_2client1peer20K.go LST_2client2peer20K.go LST_4client1peer20K.go LST_4client4peer20K.go"

##### start anew:

export CORE_PBFT_GENERAL_BATCHSIZE=500
export TEST_EXISTING_NETWORK=FALSE
go run ../CAT/testtemplate.go

### look for some of the output showing parameters used:
### exec.Command:  /opt/gopath/src/github.com/hyperledger/fabric/vendor/obcsdk/automation/local_fabric_gerrit.sh -c 4173edd -n 4 -f 1 -l critical -m pbft -b 500 -s



##### Run LedgerStressTest Regression Tests, using the existing network
##### Ensure no envvars override the parms that each testcase uses

export TEST_EXISTING_NETWORK=TRUE
export TEST_LST_TX_COUNT=""
export TEST_LST_NUM_CLIENTS=""
export TEST_LST_NUM_PEERS=""
export TEST_LST_THROUGHPUT_RATE=""

date; echo "==================== Start of LST Regression Tests ===================="
date; go run LST_1client1peer20K.go
date; go run LST_2client1peer20K.go
date; go run LST_2client2peer20K.go
date; go run LST_4client1peer20K.go
date; go run LST_4client4peer20K.go
date; echo "==================== End of LST Regression Tests, 100,000 Transactions in total ===================="

