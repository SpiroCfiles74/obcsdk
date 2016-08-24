
package main

import (
	"fmt"
	//"workingbackup/obcsdk/peernetwor"
	//"obc-test/obcsdk/chaincode"
	"time"
	"obcsdk/chaincode"
)

func deployUsingTagName() {

/** Another invoke ***/


	fmt.Println("\nPOST/Chaincode: Deploying chaincode instance2 with tagName....")
	dAPIArgs1 := []string{"example02", "init", "sdktest1"}
	depArgs1:= []string{"a", "100", "b", "200"}
	chaincode.Deploy(dAPIArgs1, depArgs1)
	//fmt.Println("From Deploy error ", err)


	time.Sleep(20000 * time.Millisecond);

	fmt.Println("\nPOST/Chaincode : Querying a and b after a deploy  ")
	qAPIArgs1 := []string{"example02", "query", "sdktest1"}
	qArgsa := []string{"a"}
	_, _  = chaincode.Query(qAPIArgs1, qArgsa)
	qArgsb := []string{"b"}
	_, _  = chaincode.Query(qAPIArgs1, qArgsb)


	fmt.Println("\n<<<<<<<<<<<<<  POST/Chaincode : Invoke on a and b after a deploy INSTANCE with incorrect args >>>>>>>>>>> ")
	iAPIArgs1 := []string{"example02", "invoke"}
	invArgs1 := []string{"a", "b", "5"}
	chaincode.InvokeAsUser(iAPIArgs1, invArgs1)
	//fmt.Println("\nFrom Invoke invRes ", invRes)

	fmt.Println("\n<<<<<<<<<<<<<  Now Invoke on a and b after a deploy INSTANCE with correct args >>>>>>>>>>> ")
	iAPIArgs2 := []string{"example02", "invoke", "jim", "sdktest1"}
	invRes, _ := chaincode.InvokeAsUser(iAPIArgs2, invArgs1)
	fmt.Println("\nFrom Invoke invRes ", invRes)

	time.Sleep(20000 * time.Millisecond);
	fmt.Println("\n<<<<<<<<<<<<<<<<<<<<<<<<   POST/Chaincode : Querying a and b after invokeAsUser  ")
	qAPIArgs2 := []string{"example02", "query", "sdktest1"}
	qArgsa = []string{"a"}
	_, _  = chaincode.Query(qAPIArgs2, qArgsa)
	qArgsb = []string{"b"}
	_, _  = chaincode.Query(qAPIArgs1, qArgsb)

}

func main() {
	fmt.Println("\nHello World\n")
}

