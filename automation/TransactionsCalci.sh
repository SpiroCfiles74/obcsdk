#!/bin/bash

# ------------------------------------------------------------------ 
#
# TITLE : Transaction Calculator
#
# AUTHOR: Ratnakar Asara
#
# VERSION: 0.1
#
# DESCRIPTION:
#          The purpose of this script is to calculate number of 
# successful transactions taken place on fabric. Script gets the
# chain height and segragates Deploy, Failed and Total transactions
# executed on blockchain
#
# DEPENDENCY:
#	Download and Install JQ: https://goo.gl/DsDskg
#
# USAGE:
#	TransactionsCalci.sh [OPTIONS]
# OPTIONS:
#	-h/? - Print a usage message
#	-i   - IP and HOST
#       -b   - Block number from where calculations begins
#       -f   - Enable Block info logging to file
#
# SAMPLE :
#	./TransactionsCalci.sh -i http://127.0.0.1:5000 -b 2 -f
#       Transaction infomration is calculates from block 2 and block
# info will be saved to blocks.txt
# ------------------------------------------------------------------

function usage(){
	## Enhance this section
        echo "USAGE : TransactionsCalci.sh -i http://IP:PORT -b <BLOCK_NUMBER_FROM> -f"
	echo "ex: ./TransactionsCalci.sh -i http://127.0.0.1:5000 -b 2 -f"
}

while getopts "\?hefi:b:" opt; do
  case $opt in
     i)   IP_PORT="$OPTARG"
	;;
     f)   ENABLE_LOG="Y"
	;;
     b)   BLOCK_NUM="$OPTARG"
	;;
   \?|h)  usage
          exit 1
        ;;
  esac
done

echo
echo "##################### Letz begin the fun  #######################"
#Wrong way of cheking ?
#TEMP_IP=$IP_PORT

: ${IP_PORT:="http://127.0.0.1:5000"}
: ${BLOCK_NUM:=1}
: ${ENABLE_LOG:="N"}

deployTrxn=0
errTrxn=0
Trxn=0

echo
echo "http[s]://IP:PORT --- $IP_PORT"
IS_SECURE=${IP_PORT:0: 5}
echo
if [ "$IS_SECURE" = "https" ]; then
    TOTAL_TRXNS=$(curl -k $IP_PORT/chain | jq '.height')
else 
    TOTAL_TRXNS=$(curl -s $IP_PORT/chain | jq '.height')
fi

echo "--- Total Blocks to be processed ` expr $TOTAL_TRXNS - 1 ` (Ignore Firt Block Genesis) ---"

if [ "$ENABLE_LOG" == "Y" ] ; then
	echo "############ Begin Writing blocks ############" > blocks.txt
fi

for (( i=$BLOCK_NUM; $i<$TOTAL_TRXNS; i++ ))
do
	#This check is required
	if [ "$IS_SECURE" = "https" ]; then
		curl -k $IP_PORT/chain/blocks/$i | jq '.' > data.json
	else 
		curl -s $IP_PORT/chain/blocks/$i | jq '.' > data.json
	fi

	#Write logs to blocks.txt if block logging enabled
	if [ "$ENABLE_LOG" == "Y" ] ; then
		echo "---------------- Block-$i ----------------" >> blocks.txt
		cat data.json >> blocks.txt
		echo "---------------- Block-$i ----------------"  >> blocks.txt
		echo "" >> blocks.txt
	fi

	#Calculate Deploy transactions
	counter=$(cat data.json | grep "ChaincodeDeploymentSpec" | wc -l)
	deployTrxn=` expr $deployTrxn + $counter `

	#Calculate Error transactions
	counter=$(cat data.json | jq '.["nonHashData"]["transactionResults"]' | grep errorCode | wc -l)
	errTrxn=` expr $errTrxn + $counter `

	#Calculate Error transactions
	counter=$(cat data.json | jq '.["nonHashData"]["transactionResults"]' | grep uuid | wc -l)
	Trxn=` expr $Trxn + $counter `
done

if [ "$ENABLE_LOG" == "Y" ] ; then
	echo "############ End Writing blocks ############" >> blocks.txt
fi

if [ -f ./data.json ]; then
    rm ./data.json
fi

echo
echo "----------- Chaincode Deployment Transactions: $deployTrxn"
echo "----------- Failed Transactions: $errTrxn"
echo "----------- Total Successful Transactions (Exclude Deploy Trxn) : " ` expr $Trxn - $errTrxn`
echo
echo "############ Thatz all I have for now, Letz catchup some other time  ###############"
echo
