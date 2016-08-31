#!/bin/bash


####################################################################################################
# FUNCTIONS
####################################################################################################


################################################################################
################################################################################
################################################################################
# USE THIS _sigs() signal catcher/forwarder to pass signal to the child process.
# Write function _sigs(), or do it with one line:
# trap 'echo I am going down, so killing off my processes..; kill $!; exit' SIGHUP SIGINT SIGQUIT SIGTERM 

_sigs() { 
echo -e "\n=========== go run $testcase ABORTABORTABORTABORTABORTABORT `date` ==========" | tee -a $OUT | tee -a $SUMMARY
echo -e   "=========== SKIPPING any remaining testcases; caught a termination signal!"    | tee -a $OUT | tee -a $SUMMARY
kill -SIGINT "$child" 2>/dev/null
#kill        "$child" 2>/dev/null
exit
}


################################################################################
################################################################################
################################################################################
headerInfo() {
echo -e   "\n=========== GROUP START TIME ${STARTDATE}, using $REPOSITORY_SOURCE COMMIT IMAGE $COMMIT" | tee -a ${SUMMARY}
echo -e     "==========="
echo -e   "\nNote: Output is recorded to file ${OUT}"
echo -e     "Note: Brief test summaries are also written by tests themselves to file ${SUMMARY}"
set ${TESTNAMES}
echo -e   "\nPreparing to run ${#} testcases:\n${*}"
echo -e   "\nUsing test environment variables:\n"

### Set some env vars, which are used by the go tests themselves in func SetupLocalNetwork() in peernetwork/peerNetworkSetup.go.
### Some of these are passed to local_fabric.sh as options; others are hardcoded defaults in that script.
### And also define some here that are not passed to local_fabric.sh but are used directly by the tests (e.g. STOP_OR_PAUSE).
### Echo all of them here for the user to see in the script output file.

echo -e "COMMIT LEVEL IMAGE: $COMMIT"
grep "^PEER_IMAGE="       $LOCAL_FABRIC_SCRIPT
grep "^MEMBERSRVC_IMAGE=" $LOCAL_FABRIC_SCRIPT

echo -e "CORE_PBFT_GENERAL_N, number of peer nodes in network: $CORE_PBFT_GENERAL_N"

echo -e "CORE_LOGGING_LEVEL: $LOGGING_LEVEL"

echo -e "CORE_SECURITY_ENABLED: $CORE_SECURITY_ENABLED"

# consensus mode = pbft , by default
echo -e "CORE_PEER_VALIDATOR_CONSENSUS_PLUGIN: `grep ^CONSENSUS= $LOCAL_FABRIC_SCRIPT | grep -v OPTARG | cut -f2 -d'='`  `grep CONSENSUS_MODE= $LOCAL_FABRIC_SCRIPT | grep -v OPTARG | cut -f2 -d'\"'` "

# pbft mode = batch , by default. Other options: noops 
echo -e "CORE_PBFT_GENERAL_MODE: `grep "^PBFT_MODE=" $LOCAL_FABRIC_SCRIPT | cut -f2 -d'='` "

echo -e "CORE_PBFT_GENERAL_BATCHSIZE, max number of transactions sent in each batch for ordering: $CORE_PBFT_GENERAL_BATCHSIZE"
echo -e "CORE_PBFT_GENERAL_F, max number of nodes that can fail while still have consensus: $CORE_PBFT_GENERAL_F"

# Others: Uncomment here and in other two places in this script when it becomes supported in local_fabric.sh and in go scripts
#echo -e "CORE_PBFT_GENERAL_K: $CORE_PBFT_GENERAL_K"
#echo -e "CORE_PBFT_GENERAL_LOGMULTIPLIER: $CORE_PBFT_GENERAL_LOGMULTIPLIER"
#echo -e "CORE_PBFT_GENERAL_TIMEOUT_BATCH: $CORE_PBFT_GENERAL_TIMEOUT_BATCH"

# These are specifically for the go sdk code scripts. They are not used by
# the hyperledger/fabric nodes or local_fabric.sh

echo -e "DOCKER STOP_OR_PAUSE MODE used by GO tests when disrupting network CA and PEER nodes: $STOP_OR_PAUSE"

echo -e "REPOSITORY_SOURCE used by GO SDK to retrieve the fabric COMMIT image: $REPOSITORY_SOURCE"


# Finally, let's show the commands parameters passed to each docker container
# when we execute "docker run" with the commands "peer node start"
echo -e " "
awk '/^membersrvc_setup/,/^[}]/' $LOCAL_FABRIC_SCRIPT | awk '/name=PEER/,/peer node start/'
echo -e " "

}


################################################################################
################################################################################
################################################################################
trailerInfo() {
echo -e "\n=========== GROUP END, using $REPOSITORY_SOURCE COMMIT IMAGE $COMMIT"
echo -e "${STATS_BEFORE_RUN_GROUP}"
echo -e "${STATS_AFTER_RUN_GROUP}"
}


################################################################################
################################################################################
################################################################################
runTests() {
for testcase in ${TESTNAMES}
do
	echo -e "\n\n----------- go run $testcase STARTSTARTSTARTSTARTSTARTSTART `date` -----------" 
	echo -e     "-----------"
	#ls -l $testcase	&
	go run $testcase	&
	# echo go run $testcase ... child PID = $! 
	child="$!"
	wait "$child"
	echo -e     "-----------"
	echo -e     "----------- go run $testcase FINISHFINISHFINISHFINISHFINISH `date` -----------" 
done
}


####################################################################################################
# BEGIN MAIN ROUTINE
####################################################################################################


################################################################################
# Catch signals, process arguments, and define global variables:
#   TESTNAMES, OUT and SUMMARY output file names, STARTDATE, STATS_BEFORE_RUN_GROUP

trap _sigs SIGTERM SIGINT SIGHUP SIGQUIT SIGABRT

## 
#Run test using default network with first committed image level in hyperledger fabric master branch in gerrit:"
#	$0 COMMIT="v05-3e0e80a" testtemplate.go"
## 
#	CORE_PBFT_GENERAL_TIMEOUT_BATCH 	- batch timeout value, use s for seconds, default= [ 2s ]"
#	CORE_PBFT_GENERAL_LOGMULTIPLIER		- logmultiplier [ 4 ]"
#	CORE_PBFT_GENERAL_K			- checkpoint period K [ 10 ]"
## 

USAGE="
Usage:  ${0}					- get help
        ${0} <testname.go> ...			- run testname.go

To override any default script parameters: set and export them from your ENV, or set them on the command line first when executing this script:

	COMMIT 					- hash commit image to use for the ca and peers; use prefix 'master-' for gerrit images
	REPOSITORY_SOURCE 			- location of the fabric COMMIT image [ GERRIT (default, for master) | GITHUB (for v0.5) ]
	CORE_PBFT_GENERAL_N 			- number of validating peers in the network [ 4 ]; note currently users are defined only for 10 nodes
	CORE_PBFT_GENERAL_F 			- max # possible faulty nodes while still can reach consensus [ 1 ] ; do not set to a value exceeding (2n-1)/3
	LOGGING_LEVEL				- [ critical | error | warning | notice | info | debug ] as defined in peer/core.yaml
	CORE_SECURITY_ENABLED 			- use secure network using MemberSrvc CA [ Y | N ]
	CORE_PEER_VALIDATOR_CONSENSUS_PLUGIN 	- consensus mode [ pbft | ... ]
	CORE_PBFT_GENERAL_MODE 			- pbft mode [ batch | noops ]
	CORE_PBFT_GENERAL_BATCHSIZE 		- max # Tx sent in each batch for ordering (Although code dflt=500, this script sets 2 unless overridden)
	STOP_OR_PAUSE 				- MODE used by GO tests when disrupting network CA and PEER nodes [ STOP | PAUSE ]

Examples:

    Run a test in current directory, using default script parameters; these two commands are equivalent:

 	$0 testtemplate.go
 	COMMIT=latest REPOSITORY_SOURCE=GERRIT CORE_PBFT_GENERAL_N=4 CORE_PBFT_GENERAL_F=1 LOGGING_LEVEL=error CORE_SECURITY_ENABLED=Y CORE_PEER_VALIDATOR_CONSENSUS_PLUGIN=pbft CORE_PBFT_GENERAL_MODE=batch CORE_PBFT_GENERAL_BATCHSIZE=2 STOP_OR_PAUSE=STOP $0 testtemplate.go

    This should run a test on the latest hyperledger fabric images using the code default configuration parameters:

 	COMMIT=latest CORE_PBFT_GENERAL_BATCHSIZE=500 $0 testtemplate.go

    Run all GO tests in current directory, using the first gerrit image level committed in hyperledger fabric master branch, with 10 peer nodes:

 	COMMIT=master-5a4bbbc CORE_PBFT_GENERAL_N=10 CORE_PBFT_GENERAL_F=3 $0 *.go

    Run all the CAT tests in current directory with batchsize 10 and collect debug logs, using the v0.5 branch June beta image (from github):

 	COMMIT=3e0e80a REPOSITORY_SOURCE=GITHUB LOGGING_LEVEL=debug CORE_PBFT_GENERAL_BATCHSIZE=10 $0 CAT*.go
"

if [ ${#} -eq 0 ]
then
	#echo -e "Error:   no testnames provided"
	echo -e "$USAGE"
	exit
fi

TESTNAMES=${*}

if  [ -z "${TESTNAMES}" ]
then
	echo -e "Error:   no  <*.go>  testnames provided"
	echo -e "$USAGE"
	exit
fi

set ${TESTNAMES}
if [ ${#} -eq 1 ]
then
	OUT="GO_TEST__$(echo ${*} | tr -d ' ')__$(date | cut -c 4-80 | tr -d ' ')"
else
	OUT="GO_TEST__MULTI__$(date | cut -c 4-80 | tr -d ' ')"
fi

SUMMARY="GO_TESTS_SUMMARY"
STARTDATE="`date`"
STATS_BEFORE_RUN_GROUP=$(echo -e "=========== STARTTIME ${STARTDATE}")

touch ${SUMMARY}
BEFORE_BEGIN=`grep -c ^BEGIN ${SUMMARY} 2>/dev/null`
BEFORE_PASSED=`grep -c ^PASSED ${SUMMARY} 2>/dev/null`
BEFORE_FAILED=`grep -c ^FAILED ${SUMMARY} 2>/dev/null`
BEFORE_ABORT=`grep -c ABORT ${SUMMARY} 2>/dev/null`


################################################################################
# Set up any more environment variables that the tests require

# the hash commit submission; could set this to use an image from master or v05 or whatever
# : ${COMMIT="master-latest"}
: ${COMMIT="latest"}

# error , debug , critical , info ...
: ${LOGGING_LEVEL="error"}

: ${CORE_SECURITY_ENABLED="Y"}
# Y or N ; capitalized
CORE_SECURITY_ENABLED=$(echo $CORE_SECURITY_ENABLED | tr a-z A-Z)

# number of peers in the network
: ${CORE_PBFT_GENERAL_N="4"}

# number of peers in the network that can fail or go rogue, with the others being correct and can still reach consensus
: ${CORE_PBFT_GENERAL_F="1"}

# Unless the caller has set CORE_PBFT_GENERAL_BATCHSIZE on the command line or in their environment,
# let's change batchsize environment variable here. This will allow scripts to run most quickly.
# Change from the default (500) in config.yaml to 2.
# Any value up to 10 is best for optimal performance, although other values should work too.
# And this will trigger the GO test scripts to also pass option "-b 2" when calling local_fabric.sh
# so the peers use the same value.
#export CORE_PBFT_GENERAL_BATCHSIZE=2

: ${CORE_PBFT_GENERAL_BATCHSIZE="2"}

# These were not initially defined in early version of local_fabric.sh.
# For now, we set them to their defaults, as defined in fabric/consensus/obcpbft/config.yaml.

# : ${CORE_PBFT_GENERAL_K="10"}
# : ${CORE_PBFT_GENERAL_LOGMULTIPLIER="4"}
# : ${CORE_PBFT_GENERAL_TIMEOUT_BATCH="2"}

# Consensus tests use docker stop and docker start when disrupting peer nodes, by default.
# We can override that behavior to use "docker pause" instead, by setting this "PAUSE".

: ${STOP_OR_PAUSE="STOP"}

# This is used by the GO SDK to determine which repository to retrieve the image from.
# Default is GERRIT, representing the official source of the hyperledger fabric master branch.
# Optionally can be set to GITHUB to retrieve images of IBM V0.5 branch.

: ${REPOSITORY_SOURCE="GERRIT"}

# We will grep some of the info directly from the local_fabric bash script.
# Deprecated: LOCAL_FABRIC_SCRIPT=../automation/local_fabric.sh
# There are now two scripts. Determine which one we will be using, based on REPOSITORY_SOURCE.
#   Default is the gerrit script, for the official hyperledger/fabric.
#   GITHUB may be specified when using v0.5 branch commit images for Z and BlueMix.
LOCAL_FABRIC_SCRIPT=../automation/local_fabric_gerrit.sh
if [ "$REPOSITORY_SOURCE" == "GITHUB" ]
then
    LOCAL_FABRIC_SCRIPT=../automation/local_fabric_github.sh
fi

################################################################################
# Print the header info, run the tests, and print the summary trailer info.

headerInfo   | tee -a ${OUT}

runTests     | tee -a ${OUT}

AFTER_BEGIN=`grep -c ^BEGIN ${SUMMARY} 2>/dev/null`
AFTER_PASSED=`grep -c ^PASSED ${SUMMARY} 2>/dev/null`
AFTER_FAILED=`grep -c ^FAILED ${SUMMARY} 2>/dev/null`
AFTER_ABORT=`grep -c ABORT ${SUMMARY} 2>/dev/null`

DIFF_BEGIN=$(($AFTER_BEGIN-$BEFORE_BEGIN))
DIFF_PASSED=$(($AFTER_PASSED-$BEFORE_PASSED))
DIFF_FAILED=$(($AFTER_FAILED-$BEFORE_FAILED))
DIFF_ABORT=$(($AFTER_ABORT-$BEFORE_ABORT))

STATS_AFTER_RUN_GROUP=$(echo -e "=========== STOP TIME `date`, TESTS STARTED=$DIFF_BEGIN PASS=$DIFF_PASSED FAIL=$DIFF_FAILED STOPPED=$DIFF_ABORT")

trailerInfo  | tee -a ${OUT} | tee -a ${SUMMARY}

