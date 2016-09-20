package peernetwork

import (
	//"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"errors"
	//"github.com/pkg/sftp"
	//"golang.org/x/crypto/ssh"
)

const (
	RUNNING      = 0
	STOPPED      = 1
	STARTED      = 2
	PAUSED       = 3
	UNPAUSED     = 4
	NOTRESPONDIN = 5
)

type Peer struct {
	PeerDetails map[string]string
	UserData    map[string]string
	State       int
}

type PeerNetwork struct {
	Peers []Peer
	Name  string
}

type LibChainCodes struct {
	ChainCodes map[string]ChainCode
}

type ChainCode struct {
	Detail   map[string]string
	Versions map[string]string
}

var peerNetwork PeerNetwork
var FirstUser string

const USER = "ibmadmin"
const PASSWORD = "m0115jan"

//HOST = "urania"
const IP = "9.37.136.147"

//NEW_IP = "9.42.91.158"


/* ********** deprecated Aug 2016
// func SetupLocalNetwork(numPeers int, security bool) {
	//goroot := os.Getenv("GOROOT")
	var cmd *exec.Cmd
	pwd, _ := os.Getwd()
	//fmt.Println("Initially ", pwd)
	os.Chdir(pwd + "/../automation/")
	pwd, _ = os.Getwd()
	//fmt.Println("After change dir ", pwd)
	script := pwd + "/local_fabric.sh"
	arg0 := "-n"
	arg1 := strconv.Itoa(numPeers)
	if security == true {
		arg2 := "-s"
		//cmdStr := script + arg0 + arg1 + arg2
		//fmt.Println("cmd ", cmdStr)
		cmd = exec.Command("sudo", script, arg0, arg1, arg2) // sometimes sudo seems to be needed to run the script
	} else {
		//cmdStr := script + arg0 + arg1
		//fmt.Println("cmd ", cmdStr)
		cmd = exec.Command("sudo", script, arg0, arg1)
	}

	var stdoutBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
	//fmt.Printf("in all caps: \n", stdoutBuf.String())

	GetNC_Local()
	os.Chdir(pwd + "/../chcotest")
	pwd, _ = os.Getwd()
	//fmt.Println("After change back dir ", pwd)
}
********** */

func SetupLocalNetwork(numPeers int, security bool) {
	// wrapper: simply use mostly the default values
	SetupLocalNetworkWithMoreOptions(
        	numPeers,		// N
        	(numPeers-1)/3,		// F
        	"error",		// debug level
        	security,		// secure network T/F
        	"pbft",			// consensusMode
        	// "batch",		// pbftMode
        	// "2s",		// batchTimeout
        	2 )			// batchSize - this is the default for our testing, although the default for the fabric is larger such as 500 or 1000
}

// PRE-CONDITIONS: all parameters must be non-null; otherwise the shell script will have problems setting up the network.
func SetupLocalNetworkWithMoreOptions(
        numPeers int,
        f int,
        logging string,
        security bool,
        consensusMode string,
        // pbftMode string,
        // batchTimeout string,
        batchsize int ) 	{

	var cmd *exec.Cmd

	// Depending on the repository source, we need to run a different local_fabric script.
	// User must ensure the -c commit_image is in the specified -r repository
	// - official hyperledger/fabric images are accessible via gerrit
	// - ibm images such as those on branch v05 are still available via github/rameshthoomu

	repository := strings.ToUpper(os.Getenv("REPOSITORY_SOURCE"))	// [ GERRIT | GITHUB ]
	script_name := "/local_fabric_gerrit.sh"	// default: pull hyperledger/fabric via gerrit
	if repository == "GITHUB" { script_name = "/local_fabric_github.sh" }	// pull from github/rameshthoomu

	// Now get the commit image name string. use "latest" as the default, if the tester does not override.
	commitImage := "latest"
        commit_envvar := strings.TrimSpace(os.Getenv("COMMIT"))
        if commit_envvar != "" { commitImage = commit_envvar }

	// run script located in the ../automation directory
	pwd, _ := os.Getwd()
	if strings.Contains(pwd, "obcsdk/") {
		// good: it looks like we are probably in a testing subdirectory of obcsdk; hopefully it is an immediate subdirectory.
		os.Chdir(pwd + "/../automation")
	} else {
		if strings.Contains(pwd, "obcsdk") {
			// ok, we are IN ../obcsdk so lets work with that...
			os.Chdir(pwd + "/automation")
		} else {
			fmt.Println("peernetwork/peerNetworkSetup.go SetupLocalNetworkWithMoreOptions(): ERROR: you must be in a subdirectory of obcsdk/ when running go tests:\ncurrent pwd: ", pwd)
			panic(errors.New("ERROR: first cd to a test directory underneath obcsdk/ to run go tests"))
		}
	}
	pwd_automation, _ := os.Getwd()
	script_cmd := pwd_automation + script_name

 // ================
 //   arg1,2     -c   - specific commit image [latest]
 //   arg3,4     -n   - N = Number of peers to launch
 //   arg5,6     -f   - F = Number of peers that can fail in a secure network, when using pbft for consensus, maximum (N-1)/3
 //   arg7,8     -l   - logging detail level
 //   arg13      -s   - enable Security and Privacy, optional
 //   arg9,10    -m   - consensus mode
 //   arg11,12   -b   - set batch size, useful when using consensus pbft mode of batch
 //                   - Others are not yet unsupported, such as pbftMode ["batch"], batchTimeout ["2s"], K [10], and logmultiplier [4]
 // ================

	arg1 := "-c"
	arg2 := commitImage
	arg3 := "-n"
	arg4 := strconv.Itoa(numPeers)
	arg5 := "-f"
	arg6 := strconv.Itoa(f)
	arg7 := "-l"
	arg8 := logging
	arg9 := "-m"
	arg10 := consensusMode
	arg11 := "-b"
	arg12 := strconv.Itoa(batchsize)
	arg13 := ""
	if security == true { arg13 = "-s" }

	fmt.Println("exec.Command: ", script_cmd, arg1, arg2, arg3, arg4, arg5, arg6, arg7, arg8, arg9, arg10, arg11, arg12, arg13 )
	cmd =        exec.Command(    script_cmd, arg1, arg2, arg3, arg4, arg5, arg6, arg7, arg8, arg9, arg10, arg11, arg12, arg13 )
	fmt.Println("exec.Command done")

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	// cmd.Run()
/*
	var stdoutBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
*/
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}

	GetNC_Local()
	os.Chdir(pwd)
	pwd, _ = os.Getwd()
	//fmt.Println("Returned to starting directory pwd: ", pwd)
	//errStr := "SetupLocalNetworkWithMoreOptions done"
	//fmt.Println(errStr)
	//log.Fatal(errors.New(errStr))
}

func GetNC_Local() {
	inFileName := "../automation/networkcredentials"

	inputfile, err := os.Open(inFileName)
	if err != nil {
		log.Fatal(err)
	}
	outFileName := "../util/NetworkCredentials.json"
	outfile, err := os.Create(outFileName)
	if err != nil {
		fmt.Println("Error in creating NetworkCredentials file ", err)
	}

	_, err = io.Copy(outfile, inputfile)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("peerNetworkSetup GetNC_Local() Copied contents from " + inFileName + " to " + outFileName)
	//log.Println(inputfile)

	outfile.Close()
}

/*
func SetupRemoteNetwork(numPeers int) {

	config := &ssh.ClientConfig{
		User: USER,
		Auth: []ssh.AuthMethod{
			ssh.Password(PASSWORD),
		},
	}

	fmt.Sprintf("Connecting to remote network: %s as %s ", IP, USER)

	//var i int
	//fmt.Println("Please enter an integer for number of peers to be created on remote network")
	//_, err := fmt.Scanf("%d", &i)

	//if err != nil {
	//	fmt.Println("No input value entered")
	//}
	//client, err := Dial("tcp", "yourserver.com:22", config)
	//if i == 0 {
	//	fmt.Println("Initializing network with 4 peers")
	//	i = 4
	//}

	cmd := "./setupObcPeers.sh -O \"pbft batch\" " + " -C -n " + strconv.Itoa(numPeers) + " -p vp -N obcvm"
	//+ " -o http://rtpgsa.ibm.com/gsa/rtpgsa/projects/b/blockchainfvt/openchain.pbft.yaml"

	fmt.Println("cmd : ", cmd)
	executeCommand(cmd, IP, config)
	//fmt.Println(result)

}

func TearDownRemoteNetwork() {

	//USER := "ibmadmin"
	//PASSWORD := "m0115jan"
	//HOST := "urania"
	//IP := "9.37.136.147"
	//NEW_IP := "9.42.91.158"

	config := &ssh.ClientConfig{
		User: USER,
		Auth: []ssh.AuthMethod{
			ssh.Password(PASSWORD),
		},
	}

	var i int
	fmt.Println("Please enter an integer for number of peers that were created on remote network")

	_, err := fmt.Scanf("%d", &i)

	if err != nil {
		fmt.Println("No input value entered")
	}
	//client, err := Dial("tcp", "yourserver.com:22", config)
	if i == 0 {
		fmt.Println("Tearing down network of 4 peers")
		i = 4
	}
	cmd := "./setupObcPeers.sh -D -p vp -N obcvm"
	fmt.Println("cmd : ", cmd)
	executeCommand(cmd, IP, config)
	//fmt.Println(result)
}*/

func stream(stdoutPipe io.Reader) {

	buffer := make([]byte, 100, 1000)
	for {
		n, err := stdoutPipe.Read(buffer)
		if err == io.EOF {
			//stdoutPipe.Close()
			break
		}
		buffer = buffer[0:n]
		os.Stdout.Write(buffer)
	}

}

/*func executeCommand(cmd, IP string, config *ssh.ClientConfig) {

	addr := fmt.Sprintf("%s:%d", IP, 22)
	fmt.Println("Dialing Connection to ", addr)
	conn, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		log.Fatal("unable to connect to [%s]: %v", addr, err)
	}
	defer conn.Close()
	session, err := conn.NewSession()
	defer session.Close()
	if err != nil {
		panic("Failed to create session: " + err.Error())
	}
	fmt.Println("Got connection from client")
	var stdoutBuf bytes.Buffer
	session.Stdout = &stdoutBuf

	//stdout, _ := session.StdoutPipe()
	//stderr, _ := session.StderrPipe()

	//in := bufio.NewReaderSize(io.MultiReader(stdout, stderr), 100)
	//cmd.Start()

	session.Run(cmd)
	//for {
	//  l, _ := in.ReadString('\n')
	// log.Printf(string(l))
	//}

	//stream(stdout)

	//time.Sleep(20000 * time.Millisecond)

        fmt.Println(stdoutBuf.String())

	//make a sftp connection only when creating the network for the first time
	if strings.Contains(cmd, "-C") {
		callSFTP(conn)
	}
	//return stdoutBuf.String()
}

func callSFTP(connection *ssh.Client) {

	sftp, err := sftp.NewClient(connection)
	if err != nil {
		log.Fatal(err)
	}
	inFileName := "/home/ibmadmin/PEERENVARS"
	inputfile, err := sftp.OpenFile(inFileName, os.O_RDONLY)
	if err != nil {
		log.Fatal(err)
	}
	pwd, _ := os.Getwd()
	outFileName := pwd+"/../util/NetworkCredentials.json"
	outfile, err := os.Create(outFileName)
	if err != nil {
		fmt.Println("Error in creating NetworkCredentials file ", err)
	}

	_, err = io.Copy(outfile, inputfile)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("peerNetworkSetup callSFTP() Copied contents from " + inFileName + " to " + outFileName)
	//log.Println(inputfile)

	outfile.Close()
	sftp.Close()

}*/

/*
  creates network as defined in NetworkCredentials.json, distributing users evenly among the peers of the network.
*/
func LoadNetwork() PeerNetwork {

	p, n := initializePeers()

	peerNetwork := PeerNetwork{Peers: p, Name: n}
	return peerNetwork
}

/*
  reads CC_Collection.json and returns a library of chain codes.
*/
func InitializeChainCodes() LibChainCodes {
	pwd, _ := os.Getwd()
	file, err := os.Open(pwd + "/../util/CC_Collection.json")
	if err != nil {
		log.Fatal("Error in opening CC_Collection.json file ")
	}

	poolChainCode, err := unmarshalChainCodes(file)
	if err != nil {
		log.Fatal("Error in unmarshalling")
	}

	//make a map to hold each chaincode detail
	ChCos := make(map[string]ChainCode)
	for i := 0; i < len(poolChainCode); i++ {
		//construct a map for each chaincode detail
		detail := make(map[string]string)
		detail["type"] = poolChainCode[i].TYPE
		detail["path"] = poolChainCode[i].PATH
		//detail["dep_txid"] = poolChainCode[i].DEP_TXID
		//detail["deployed"] = poolChainCode[i].DEPLOYED

		versions := make(map[string]string)
		CC := ChainCode{Detail: detail, Versions: versions}
		//add the structure to map of chaincodes
		ChCos[poolChainCode[i].NAME] = CC
	}
	//finally add this map - collection of chaincode detail to library of chaincodes
	libChainCodes := LibChainCodes{ChainCodes: ChCos}
	return libChainCodes
}

func initializePeers() (peers []Peer, name string) {

	fmt.Println("Getting and Initializing Peer details from network")
	peerDetails, userDetails, Name := initNetworkCredentials()
	numOfPeersOnNetwork := len(peerDetails)
	numOfUsersOnNetwork := len(userDetails)
	fmt.Println("After reading NetworkCredentials:", numOfPeersOnNetwork)
	fmt.Println("Num of Peers:", numOfPeersOnNetwork)
	fmt.Println("Num of Users:", numOfUsersOnNetwork)
	fmt.Println("Network Name:", Name)
	if numOfPeersOnNetwork == 0 || numOfUsersOnNetwork == 0 {
		fmt.Println("WARNING: UNUSABLE NETWORK (no peers or no users)!!!\nExamine files networkcredentials and NetworkCredentials.json for errors.\nIf local network, try running the exec.Command directly on command line to possibly see more startup errors.\nAnd check the COMMIT level image name and confirm it is in the identified REPOSITORY_SOURCE.")
	}
	FirstUser = userDetails[0].USER
	factor := numOfUsersOnNetwork / numOfPeersOnNetwork
	remainder := numOfUsersOnNetwork % numOfPeersOnNetwork
	allPeers := make([]Peer, numOfPeersOnNetwork)
	i := 0
	k := 0
	//for each peerDetail we construct a new peer evenly distributing the list of users
	for i < numOfPeersOnNetwork {
		aPeer := new(Peer)
		aPeerDetail := make(map[string]string)
		//name := "vp" + strconv.Itoa(i)
		aPeerDetail["ip"] = peerDetails[i].IP
		aPeerDetail["port"] = peerDetails[i].PORT
		//aPeerDetail["name"] = name
		aPeerDetail["name"] = peerDetails[i].NAME

		//fmt.Println(aPeerDetail["ip"], aPeerDetail["port"], aPeerDetail["name"])

		//fmt.Println("Getting and Initializing User details from network")

		j := 0
		userInfo := make(map[string]string)
		for j < factor {
			for k < numOfUsersOnNetwork {
				//fmt.Println(" **********value in inside i", i , "k ", k, "factor", factor, " j ", j)
				userInfo[userDetails[k].USER] = userDetails[k].SECRET
				j++
				k++
				if j == factor {
					break
				}
			}
		}

		aPeer.PeerDetails = aPeerDetail
		aPeer.UserData = userInfo
		aPeer.State = RUNNING
		allPeers[i] = *aPeer
		i++
	}
	//do we have any left over users details
	if remainder > 0 {
		for m := 0; m < remainder; m++ {
			allPeers[m].UserData[userDetails[k].USER] = userDetails[k].SECRET
			k++
		}
	}
	return allPeers, Name
}

func initNetworkCredentials() ([]peerHTTP, []userData, string) {
	pwd, _ := os.Getwd()
	fmt.Println("PWD :", pwd)
	file, err := os.Open(pwd + "/../util/NetworkCredentials.json")

	if err != nil {
		fmt.Println("Error in opening NetworkCredentials file ", err)
		log.Fatal("Error in opening Network Credential json File")
	}
	networkCredentials, err := unmarshalNetworkCredentials(file)
	if err != nil {
		log.Fatal("Error in unmarshalling")
	}
	//peerdata := make(map[string]string)
	peerData := networkCredentials.PEERHTTP
	userData := networkCredentials.USERDATA
	name := networkCredentials.NAME
        //fmt.Println("peerData", peerData)
        //fmt.Println("userData", userData)
        //fmt.Println("name", name)
	return peerData, userData, name
}

/*
 Gets ChainCode detail for a given chaincode name
  Takes two arguments
	1. name (string)			- name of the chaincode as specified in CC_Collections.json file
	2. lcc (LibChainCodes)		- LibChainCodes struct having current collection of all chaincodes loaded in the network.
  Returns:
 	1. ccDetail map[string]string  	- chaincode details of the chaincode requested as a map of key/value pairs.
	2. Versions map[string]string   - versioning or tagging details on the chaincode requested as a map of key/value pairs
*/
/*************************
func GetCCDetailByName(name string, lcc LibChainCodes) (ccDetail map[string]string, versions map[string]string, err error) {
	var errStr string
	var err1 error
	for k, v := range lcc.ChainCodes {
		if strings.Contains(k, name) {
			return v.Detail, v.Versions, err1
		}
	}
	//no more chaincodes construct error string and empty maps
	errStr = fmt.Sprintf("chaincode %s does not exist on the network", name)
	//need to check for this
	j := make(map[string]string)
	return j, j, errors.New(errStr)
}

*****************/

/** utility functions to aid users in getting to a valid URL on network
 ** to post chaincode rest API
 **/

/*
  gets any one peer from 'thisNetwork' as set by chaincode.Init()
*/
/***********************
func APeer(thisNetwork PeerNetwork) *Peer {
	//thisNetwork := LoadNetwork()
	Peers := thisNetwork.Peers
	var aPeer *Peer
	//get any peer that has at a minimum one userData and one peerDetails
	for peerIter := range Peers {
		//if (len(Peers[peerIter].UserData) > 0) && (len(Peers[peerIter].PeerDetails) > 0) {
		if (len(Peers[peerIter].UserData) > 0) && (len(Peers[peerIter].PeerDetails) > 0) && (Peers[peerIter].State == RUNNING) {
			aPeer = &Peers[peerIter]
		}
	}
	//fmt.Println(" * aPeer ", *aPeer)
	//fmt.Println(" ip ", aPeer.PeerDetails["ip"])
	return (aPeer)
}
**************/

/*
  gets any one user from any Peer on the entire network.
*/
/*****************************
func AUserFromNetwork(thisNetwork PeerNetwork) (ip string, port string, user string) {

	//fmt.Println("Values inside AUserFromNetwork ", ip, port, user)
	var u string
	aPeer := APeer(thisNetwork)
	users := aPeer.UserData

	//fmt.Println(" ip ", aPeer.PeerDetails["ip"])
	for u, _ = range users {
		break
	}
	//fmt.Println(" ip ", aPeer.UserData["ip"])
	//fmt.Println(" ip ", user)
	return aPeer.PeerDetails["ip"], aPeer.PeerDetails["port"], u
}
******************/
/*
  finds any one user associated with the given peer
*/
/************************
func AUserFromAPeer(thisPeer Peer) (ip string, port string, user string) {

	//var aPeer *Peer
	aPeer := thisPeer
	var curUser string
	userList := aPeer.UserData
	for curUser, _ = range userList {
		break
	}
	//fmt.Println(" ip ", aPeer.UserData["ip"])
	//fmt.Println(" ip ", user)
	return aPeer.PeerDetails["ip"], aPeer.PeerDetails["port"], curUser
}
*********************/
/*
 gets a particular user from a given Peer on the PeerNetwork
*/
/*************************
 func AUserFromThisPeer(thisNetwork PeerNetwork, host string) (ip string, port string, user string, err error) {
	//var aPeer *Peer
	Peers := thisNetwork.Peers
	var aPeer *Peer
	var u string
	var errStr string
	var err1 error
	//get a random peer that has at a minimum one userData and one peerDetails
	for peerIter := range Peers {
		if (len(Peers[peerIter].UserData) > 0) && (len(Peers[peerIter].PeerDetails) > 0) {
			if strings.Contains(Peers[peerIter].PeerDetails["ip"], host) {
				aPeer = &Peers[peerIter]
			}
		}
	}
	//fmt.Println(" * aPeer ", *aPeer)
	if aPeer != nil {
		users := aPeer.UserData
		for u, _ = range users {
			break
		}
		return aPeer.PeerDetails["ip"], aPeer.PeerDetails["port"], u, err1
	} else {
		errStr= fmt.Sprintf("%s, Not found on network", host)
		return "", "", "", errors.New(errStr)
	}
}
*********************/
/*
  finds the peer address corresponding to a given user
    thisNetwork as set by chaincode.init
	ip, port are the address of the peer.
	user is the user details: name, credential.
	err	is an error message, or nil if no error occurred.
*/
/***************************
func PeerOfThisUser(thisNetwork PeerNetwork, username string) (ip string, port string, user string, err error) {

	//var aPeer *Peer
	Peers := thisNetwork.Peers
	var aPeer *Peer
	var errStr string
	var err1 error
	//fmt.Println("Inside function")
	//get a random peer that has at a minimum one userData and one peerDetails
	for peerIter := range Peers {
		if (len(Peers[peerIter].UserData) > 0) && (len(Peers[peerIter].PeerDetails) > 0) {
			if _, ok := Peers[peerIter].UserData[username]; ok {
				fmt.Println("Found %s in network", username)
				aPeer = &Peers[peerIter]
			}
		}
	}

	if aPeer == nil {
		errStr = fmt.Sprintf("%s, Not found on network", username)
		return "", "", "", errors.New(errStr)
	} else {
		return aPeer.PeerDetails["ip"], aPeer.PeerDetails["port"], username, err1
	}
}
*****************/
/********************
type PeerNetworks struct {
	PNetworks      []PeerNetwork
}


func AddAPeerNetwork() {

}

func DeleteAPeerNetwork() {

}

func AddUserOnAPeer(){

}

func RemoveUserOnAPeer(){

}


func LoadNetworkByName(name string) PeerNetwork {

  networks := LoadPeerNetworks()
	pnetworks := networks.PNetworks
	for peerIter := range pnetworks {
		//fmt.Println(pnetworks[peerIter].Name)
		if strings.Contains(pnetworks[peerIter].Name, name) {
			return pnetworks[peerIter]
		}
	}
	//return *new(PeerNetwork)
}
*********************/
