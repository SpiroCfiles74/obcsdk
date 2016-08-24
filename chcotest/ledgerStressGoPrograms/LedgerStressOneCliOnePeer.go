package main

import (
	"fmt"
	"strconv"
	"time"
	//"io/ioutil"
	"obcsdk/chaincode"
	"obcsdk/peernetwork"
)

var peerNetworkSetup peernetwork.PeerNetwork
var AVal, BVal, curAVal, curBVal, invokeValue int64
var argA = []string{"a"}
var argB = []string{"counter"}

var data string
var counter int64
const(
	TOTAL_NODES = 4
)
var url string
func setupNetwork() {
	fmt.Println("Creating a local docker network")
  peernetwork.SetupLocalNetwork(4, true)
	//peernetwork.PrintNetworkDetails()
	peerNetworkSetup = chaincode.InitNetwork()
	chaincode.InitChainCodes()
	chaincode.RegisterUsers()
}

//TODO : rather can we have a map for sleep for millis, secs and mins
func sleep(secs int64) {
	time.Sleep(time.Second * time.Duration(secs))
}

func deployChaincode (done chan bool){
	example := "mycc"
	var funcArgs = []string{example, "init"}
	var args = []string{argA[0], data, argB[0], "0"}

	//fmt.Println("\n######## Deploying chaincode ")
	chaincode.Deploy(funcArgs, args)

	sleep(30)
	done <- true
}

func invokeChaincode (i int){// (res1, res2 int64) {
	arg1Construct := []string{"mycc", "invoke"}
	arg2Construct := []string{"a"+strconv.Itoa(i), data, "counter", }

	_, _ = chaincode.Invoke(arg1Construct, arg2Construct)
  //fmt.Println("\n Invoke Transaction ID: ", invRes)
}

func queryChaincode () (res1, res2 string){
	var qargA = []string{"a"+strconv.FormatInt(counter, 10)}
	qAPIArgs0 := []string{"mycc", "query"}
	A, _ := chaincode.Query(qAPIArgs0, qargA)
	B, _ := chaincode.Query(qAPIArgs0, []string{"counter"})
	return A,B
}

func main() {
	//done chan int
	done := make(chan bool, 1)
	data = "!function(a){\"function\"==typeof define&&define.amd?define([\"jquery\"],a):\"object\"==typeof exports?module.exports=a:a(jQuery)}(function(a){function b(b){var g=b||window.event,h=i.call(arguments,1),j=0,l=0,m=0,n=0,o=0,p=0;if(b=a.event.fix(g),b.type=\"mousewheel\",\"detail\"in g&&(m=-1*g.detail),\"wheelDelta\"in g&&(m=g.wheelDelta),\"wheelDeltaY\"in g&&(m=g.wheelDeltaY),\"wheelDeltaX\"in g&&(l=-1*g.wheelDeltaX),\"axis\"in g&&g.axis===g.HORIZONTAL_AXIS&&(l=-1*m,m=0),j=0===m?l:m,\"deltaY\"in g&&(m=-1*g.deltaY,j=m),\"deltaX\"in g&&(l=g.deltaX,0===m&&(j=-1*l)),0!==m||0!==l){if(1===g.deltaMode){var q=a.data(this,\"mousewheel-line-height\");j*=q,m*=q,l*=q}else if(2===g.deltaMode){var r=a.data(this,\"mousewheel-page-height\");j*=r,m*=r,l*=r}if(n=Math.max(Math.abs(m),Math.abs(l)),(!f||f>n)&&(f=n,d(g,n)&&(f/=40)),d(g,n)&&(j/=40,l/=40,m/=40),j=Math[j>=1?\"floor\":\"ceil\"](j/f),l=Math[l>=1?\"floor\":\"ceil\"](l/f),m=Math[m>=1?\"floor\":\"ceil\"](m/f),k.settings.normalizeOffset&&this.getBoundingClientRect){var s=this.getBoundingClientRect();o=b.123456789"
	// Setup the network based on the NetworkCredentials.json provided
	setupNetwork()
  //Deploy chaincode
  deployChaincode (done)

	// time to messure overall execution of the testcase
	defer timeTrack(time.Now(), "execution for LedgerStressOneCliOnePeer.go ")
	/*select {
		case <-done:
    	go schedulerTask()
	}
	sleep(10);*/
  InvokeLoop()

}
//Invokes loop
func InvokeLoop() {
	counter = 0;
	for i := 1;i<=20000;i++{
		counter++;
		fmt.Println ("============== > Iteration#", counter);
		invokeChaincode(i);
	}
}
func displayChainHeight(){
	startValue := 3
	height := 0
	var urlStr string
	for i:=0;i<TOTAL_NODES;i++ {
		urlStr = "http://172.17.0."+strconv.Itoa(startValue+i)+":5000"
		height = chaincode.Monitor_ChainHeight(urlStr)
		fmt.Println("################ Chaincode Height on "+urlStr+" is : ", height)
	}
}
func timeTrack(start time.Time, name string) {
	elapsed := time.Since(start)
  //sleep(20)
	val1,val2 := queryChaincode ()
	displayChainHeight()
	fmt.Printf("\n########### After Query Vals A = %s \nB = %s", val1,val2)
	fmt.Printf("\n\n################# %s took %s \n\n", name, elapsed)
}
