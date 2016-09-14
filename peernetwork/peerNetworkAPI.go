package peernetwork

import (
	"fmt"
	"errors"
	"log"
	"strings"
	"time"
	"os/exec"
)

/*
  prints the content of the network: peers, users, and chaincodes.
*/
func PrintNetworkDetails() {

	ThisNetwork := LoadNetwork()
	Peers := ThisNetwork.Peers
	i := 0
	for i < len(Peers) {

		msgStr := fmt.Sprintf("  ip: %s port: %s name: %s ", Peers[i].PeerDetails["ip"], Peers[i].PeerDetails["port"], Peers[i].PeerDetails["name"])
		fmt.Println(msgStr)
		userList := ThisNetwork.Peers[i].UserData
		fmt.Println("  users:")
		for user, secret := range userList {
			msgStr := fmt.Sprintf("    user: %s , secret: %s", user, secret)
			fmt.Println(msgStr)
		}
		i++
	}
	fmt.Println("\nAvailable Chaincodes :")
	libChainCodes := InitializeChainCodes()
	for k, v := range libChainCodes.ChainCodes {
		fmt.Println("Chaincode :", k)
		fmt.Println("Detail :")
		for fieldname, value := range v.Detail {
			msgStr := fmt.Sprintf("  %s  %s", fieldname, value)
			fmt.Println(msgStr)
		}
		//fmt.Println("\n")
	}

}


/*
 Get Number of Peers on network
*/

func GetNumberOfPeers(thisNetwork PeerNetwork) int{
	Peers := thisNetwork.Peers
        return len(Peers)
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
func GetCCDetailByName(name string, lcc LibChainCodes) (ccDetail map[string]string, versions map[string]string, err error) {
	var errStr string
	for k, v := range lcc.ChainCodes {
		if strings.Contains(k, name) {
			return v.Detail, v.Versions, nil
		}
	}
	//no more chaincodes construct error string and empty maps
	errStr = fmt.Sprintf("chaincode name <%s> does not exist on the network", name)
	//need to check for this
	j := make(map[string]string)
	return j, j, errors.New(errStr)
}

/******************************
func AddCCToLibrary(path string, name string) {
	myCCDetail := make(map[string]string)
	myCCDetail["type"] = "GOLANG"
	myCCDetail["path"] = ccPath

	versions := make(map[string]string)
  myCC := ChainCode{Detail: myCCDetail, Versions: versions}
  append(myCC[name], LibCC)
}
******************************/


/** utility functions to aid users in getting to a valid URL on network
 ** to post chaincode rest API
 **/

/*
  gets any one running peer from 'thisNetwork' as set by chaincode.Init()
*/
func APeer(thisNetwork PeerNetwork) (thisPeer *Peer, err error) {
	//thisNetwork := LoadNetwork()
	Peers := thisNetwork.Peers
	var aPeer *Peer
	var errStr string
	//get any running peer that has at a minimum one userData and one peerDetails
	for peerIter := range Peers {
		if (len(Peers[peerIter].UserData) > 0) && (len(Peers[peerIter].PeerDetails) > 0) {
			if Peers[peerIter].State == RUNNING {
			//if Peers[peerIter].State == 0 || Peers[peerIter].State == 2 || Peers[peerIter].State == 4 {
				aPeer = &Peers[peerIter]
			}
		}
	}
	if aPeer != nil {
		return (aPeer), nil
	} else {
		errStr = fmt.Sprintf("Not found valid running peer on network")
		return aPeer, errors.New(errStr)
	}

}

/*
  gets IP address of a Peer given it's name on the entire network.
*/
func IPPeer(thisNetwork PeerNetwork, peername string) (IP string, err error) {

	//fmt.Println("Values inside IPPeer ", ip, port, user)
	Peers := thisNetwork.Peers
	var aPeer *Peer
	var errStr string
	peerFullName, _ := GetFullPeerName(thisNetwork, peername)
	//get any peer that has at a minimum one userData and one peerDetails
	for peerIter := range Peers {
		if (len(Peers[peerIter].UserData) > 0) && (len(Peers[peerIter].PeerDetails) > 0) {
			if Peers[peerIter].PeerDetails["name"] == peerFullName {
				aPeer = &Peers[peerIter]
			}
		}
	}
	if aPeer != nil {
		return (aPeer.PeerDetails["IP"]), nil
	} else {
		errStr = fmt.Sprintf("Not found %s peer on network", peername)
		return aPeer.PeerDetails["IP"], errors.New(errStr)
	}
}

/*
  gets any one user from any Peer on the entire network.
*/
func AUserFromNetwork(thisNetwork PeerNetwork) (thisPeer *Peer, user string) {

	//fmt.Println("Values inside AUserFromNetwork ", ip, port, user)

	// get any user from any running peer

	var u string
	aPeer, _ := APeer(thisNetwork)
	users := aPeer.UserData

	for u, _ = range users {
		break
	}
	return aPeer, u
}

/*
  finds any one user associated with the given peer
*/
func AUserFromAPeer(thisPeer Peer) (ip string, port string, user string, err error) {

	//var aPeer *Peer
	aPeer := thisPeer
	var curUser string
	var err1 error
	userList := aPeer.UserData
	for curUser, _ = range userList {
		break
	}
	if curUser == " " {
		errStr := fmt.Sprintf("%s, Not found on network", curUser)
		return "", "", "", errors.New(errStr)
	} else {
		return aPeer.PeerDetails["ip"], aPeer.PeerDetails["port"], curUser, err1
	}
}

/*
 gets a user from a Peer with the given IP or host name on the PeerNetwork
*/
func AUserFromThisPeer(thisNetwork PeerNetwork, host string) (ip string, port string, user string, err error) {

	//var aPeer *Peer
	Peers := thisNetwork.Peers
	var aPeer *Peer
	var u string
	var errStr string
	var err1 error

	//get a random peer that has at a minimum one userData and one peerDetails
	//for p := range Peers {
	for p := 0; p < len(Peers); p++  {
		//fmt.Println("AUserFromThisPeer: peer %d state %d",p,Peers[p].State)
		//if Peers[p].State == 0 || Peers[p].State == 2 || Peers[p].State == 4 {
		if Peers[p].State == RUNNING || Peers[p].State == STARTED || Peers[p].State == UNPAUSED {
				if (strings.Contains(host, ":")) {
					//host: ip address
					if strings.Contains(Peers[p].PeerDetails["ip"], host) {
						aPeer = &Peers[p]
					}
				}else { //host: "vp1"
					if strings.Contains(Peers[p].PeerDetails["name"], host) {
						//fmt.Println("Inside name IP resolution")
						aPeer = &Peers[p]
					}
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
	}else {
			errStr = fmt.Sprintf("Peer (host=%s) Not Found running on thisNetwork:%s,%s,%s,%s", host, Peers[0].PeerDetails["name"], Peers[1].PeerDetails["name"],Peers[2].PeerDetails["name"],Peers[3].PeerDetails["name"])
			return "", "", "", errors.New(errStr)
	}
}


/*
  finds the peer address corresponding to a given user
    thisNetwork as set by chaincode.init
	ip, port are the address of the peer.
	user is the user details: name, credential.
	err	is an error message, or nil if no error occurred.
*/
func PeerOfThisUser_OLD(thisNetwork PeerNetwork, username string) (ip string, port string, user string, err error) {

	//var aPeer *Peer
	Peers := thisNetwork.Peers
	var aPeer *Peer
	var errStr string
	var err1 error
	//fmt.Println("Inside function")
	//get a random peer that has at a minimum one userData and one peerDetails
	for peerIter := range Peers {
		if (len(Peers[peerIter].UserData) > 0) && (len(Peers[peerIter].PeerDetails) > 0) {
			if Peers[peerIter].State == RUNNING {
			//if Peers[peerIter].State == 0 || Peers[peerIter].State == 2 || Peers[peerIter].State == 4 {
				if _, ok := Peers[peerIter].UserData[username]; ok {
					fmt.Println("Found %s in network", username)
					aPeer = &Peers[peerIter]
				}
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


func PeerOfThisUser(thisNetwork PeerNetwork, username string) (ip string, port string, user string, err error) {

	//var aPeer *Peer
	Peers := thisNetwork.Peers
	var aPeer *Peer
	var errStr string
	var err1 error
	//fmt.Println("Inside function")
	//get a random peer that has at a minimum one userData and one peerDetails
	for peerIter := range Peers {
		if len(Peers[peerIter].UserData) > 0 && len(Peers[peerIter].PeerDetails) > 0 && (Peers[peerIter].State == RUNNING ||  Peers[peerIter].State == STARTED){
				if _, ok := Peers[peerIter].UserData[username]; ok {
					//fmt.Printf("Found %s in network on peer %d\n", username, peerIter)
					fmt.Printf("Found %s in network on peer %d\n", username, peerIter)
					aPeer = &Peers[peerIter]
				}
		}
	}
	if aPeer == nil {
		//TODO: we hardcoded some users on peer3. however must change this to a permanent solution. (Change these details on Z as well, below.)
		if (username == "test_user4" || username == "test_user5" || username == "test_user6" || username == "test_user7") {
			aPeer = &Peers[3]
			return aPeer.PeerDetails["ip"], aPeer.PeerDetails["port"], username, err1
		}
		//TODO: Right now testing on Z, need a permanent solution

		//if (username == "test_user4" || username == "test_user5" || username == "test_user6" || username == "test_user7") {
		if (username == "user_type1_5ab5186957" || username == "user_type1_dcc045d54f" || username == "user_type1_5998a3ce42" || username == "user_type1_b13badcce7") {
			aPeer = &Peers[3]
			return aPeer.PeerDetails["ip"], aPeer.PeerDetails["port"], username, err1
		}
		errStr = fmt.Sprintf("PeerOfThisUser   %s, Not found on network", username)
		return "", "", "", errors.New(errStr)
	} else {
		return aPeer.PeerDetails["ip"], aPeer.PeerDetails["port"], username, nil
	}
}


/*Gets the peer details corresponding to a given peer-name
state if running/stopped/unresponsive/paused:0/1/2/3
err	is an error message, or nil if no error occurred.
*/
func GetPeerState(thisNetwork PeerNetwork, peername string) (currPeer *Peer, err error) {

	//var aPeer *Peer
	Peers := thisNetwork.Peers
	var aPeer *Peer
	var errStr string
	fullName, _ := GetFullPeerName(thisNetwork, peername)
	for peerIter := range Peers {
		if (len(Peers[peerIter].UserData) > 0) && (len(Peers[peerIter].PeerDetails) > 0) {
			if strings.Contains(Peers[peerIter].PeerDetails["name"], fullName) {
				aPeer = &Peers[peerIter]
			}
		}
	}

	if aPeer == nil {
		errStr = fmt.Sprintf("%s, Not found on network", peername)
		emptyPD := new(Peer)
		return emptyPD, errors.New(errStr)
	} else {
		return aPeer, nil
	}
}

/*
  sets the peer details corresponding to a given peer-name
  state if running/stopped/unresponsive/paused:0/1/2/3
	err	is an error message, or nil if no error occurred.
*/
func SetPeerState(thisNetwork PeerNetwork, peername string, curstate int) (peerDetails map[string]string, err error) {

	//var aPeer *Peer
	Peers := thisNetwork.Peers
	var aPeer *Peer
	var errStr string
	//fmt.Println("Inside function")
	//get a peerDetails from peername
	fullName, _ := GetFullPeerName(thisNetwork, peername)
	for peerIter := range Peers {
		if (len(Peers[peerIter].UserData) > 0) && (len(Peers[peerIter].PeerDetails) > 0) {
			if strings.Contains(Peers[peerIter].PeerDetails["name"], fullName) {
				aPeer = &Peers[peerIter]
			}
		}
	}

	if aPeer == nil {
		errStr = fmt.Sprintf("%s, Not found on network", peername)
		emptyPD := make(map[string]string)
		return emptyPD, errors.New(errStr)
	} else {
		aPeer.State = curstate
		// fmt.Println("SetPeerState=", curstate)   // RUNNING=0 STOPPED=1 NOTRESPONDIN=2 PAUSED=3
		return aPeer.PeerDetails, nil
	}
}





func PausePeersLocal(thisNetwork PeerNetwork, peers []string) {

	for i:=0 ; i < len(peers); i++ {
		cmd := "docker pause " + peers[i]
		out, err := exec.Command("/bin/sh", "-c", cmd).Output()
                if (err != nil) {
					fmt.Println("PausePeersLocal: Could not Pause peer ", peers[i])
					fmt.Println(out)
					log.Fatal(err)
                }
		//fmt.Println("Paused peer " + peers[i])
		SetPeerState(thisNetwork, peers[i], PAUSED)
	}
	fmt.Println("After pause peers, sleep 5 secs")
	time.Sleep(5000 * time.Millisecond)
}


func PausePeerLocal(thisNetwork PeerNetwork, peer string) {

	cmd := "docker pause " + peer
        out, err := exec.Command("/bin/sh", "-c", cmd).Output()
        if (err != nil) {
			fmt.Println("PausePeerLocal: Could not Pause peer " + peer)
			fmt.Println(out)
			log.Fatal(err)
		} else {
			//fmt.Println("Paused peer " + peer)
			fmt.Println("After pause peer, sleep 5 secs")
			time.Sleep(5000 * time.Millisecond)
			SetPeerState(thisNetwork, peer, PAUSED)
	}

}



func UnpausePeersLocal(thisNetwork PeerNetwork, peers []string) {

	for i:=0; i < len(peers); i++ {
		cmd := "docker unpause " + peers[i]
                out, err := exec.Command("/bin/sh", "-c", cmd).Output()
                if (err != nil) {
					fmt.Println("UnpausePeersLocal: Could not Unpause peer " + peers[i])
					fmt.Println(out)
					log.Fatal(err)
                }
		//exec.Command(cmd)
		//fmt.Println("Unpaused peer " + peers[i])
		SetPeerState(thisNetwork, peers[i], RUNNING)
	}
	fmt.Println("After unpause peers, sleep 5 secs")
	time.Sleep(5000 * time.Millisecond)
}



func UnpausePeerLocal(thisNetwork PeerNetwork, peer string) {

        fmt.Println("UnpausePeerLocal(): peer=" + peer)
	cmd := "docker unpause " + peer
        out, err := exec.Command("/bin/sh", "-c", cmd).Output()
        if (err != nil) {
			fmt.Println("UnpausePeerLocal: Could not Unpause peer " + peer)
			fmt.Println(out)
			log.Fatal(err)
        } else {
			fmt.Println("After unpause peer, sleep 5 secs")
			time.Sleep(5000 * time.Millisecond)
			SetPeerState(thisNetwork, peer, RUNNING)
	}
}


func StopPeersLocal(thisNetwork PeerNetwork, peers []string) {

	for i:=0; i < len(peers); i++ {
		cmd := "docker stop " + peers[i]
                out, err := exec.Command("/bin/sh", "-c", cmd).Output()
                if (err != nil) {
                   fmt.Println("StopPeersLocal: Could not exec docker stop " + peers[i])
					fmt.Println(out)
                   log.Fatal(err)
                }
		//exec.Command(cmd)
		SetPeerState(thisNetwork, peers[i], STOPPED)
	}
	fmt.Println("After stop peers, sleep 5 secs")
	time.Sleep(5000 * time.Millisecond)
}

func StartPeersLocal(thisNetwork PeerNetwork, peers []string) {

	for i:=0; i < len(peers); i++ {
		cmd := "docker start " + peers[i]
		out, err := exec.Command("/bin/sh", "-c", cmd).Output()
		if (err != nil) {
			fmt.Println("StartPeersLocal: Could not exec docker start " + peers[i])
			fmt.Println(out)
			log.Fatal(err)
		} else {
			//exec.Command(cmd)
			SetPeerState(thisNetwork, peers[i], RUNNING)
		}
		fmt.Println("After start peers, sleep 5 secs")
		time.Sleep(5000 * time.Millisecond)
	}
}
func StartPeerLocal(thisNetwork PeerNetwork, peer string) {

	cmd := "docker start " + peer
	out, err := exec.Command("/bin/sh", "-c", cmd).Output()
	if (err != nil) {
		fmt.Println("ERROR: Could not exec docker start " + peer)
		fmt.Println(out)
		log.Fatal(err)
	} else {
		if peer != "caserver" {
			fmt.Println("After start peer, sleep 5 secs")
			time.Sleep(5000 * time.Millisecond)
			SetPeerState(thisNetwork, peer, RUNNING)
		}
	}
}

func StopPeerLocal(thisNetwork PeerNetwork, peer string) {

	cmd := "docker stop " + peer
        out, err := exec.Command("/bin/sh", "-c", cmd).Output()
        if (err != nil) {
           fmt.Println("ERROR: Could not exec docker stop " + peer)
	   fmt.Println(out)
           log.Fatal(err)
        } else {
		if peer != "caserver" {
			fmt.Println("After stop peer, sleep 5 secs")
			time.Sleep(5000 * time.Millisecond)
			SetPeerState(thisNetwork, peer, STOPPED)
		}
	}
}

func GetFullPeerName(thisNetwork PeerNetwork, shortname string) (name string, err error) {
	Peers := thisNetwork.Peers
	var aPeer *Peer
	var errStr string
	//fmt.Println("Inside function")
	//get a peerDetails from peername
	for peerIter := range Peers {
		if (len(Peers[peerIter].UserData) > 0) && (len(Peers[peerIter].PeerDetails) > 0) {
			if strings.Contains(Peers[peerIter].PeerDetails["name"], shortname) {
				aPeer = &Peers[peerIter]
			}
		}
	}

	if aPeer == nil {
		errStr = fmt.Sprintf("%s, Not found on network", shortname)
		return "", errors.New(errStr)
	} else {
		return aPeer.PeerDetails["name"], nil

	}
}

func AddAPeerNetwork() {

}

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
