
package threadutil

import (
        "os"
)

// A Utility program, contains several utility methods that can be used across test programs for multithreading

// For now, these are some hardcoded users for custom usage.
//TODO : These values should be configurable for different environments
var NumberOfPeers = 4
var NumberCustomUsersOnLastPeer = 4
var LocalUsersOnLastPeer = []string{"test_user4", "test_user5", "test_user6", "test_user7"}
var LocalUserPasswordsOnLastPeer = []string{"4nXSrfoYGFCP", "yg5DVhm0er1z", "b7pmSxzKNFiw", "YsWZD4qQmYxo"}
var ZUsersOnLastPeer = []string{"dashboarduser_type0_efeeb83216", "dashboarduser_type0_fa08214e3b", "dashboarduser_type0_e00e125cf9", "dashboarduser_type0_e0ee60d5af"}
var ZUserPasswordsOnLastPeer = []string{"", "", "", ""}
var LocalPeers = []string{"PEER0", "PEER1", "PEER2", "PEER3"}
var ZPeers = []string{"vp0", "vp1", "vp2", "vp3"}

//Get the user names based on network environment:  Z | LOCAL [default]
func GetUser(userNumber int) string {
	if os.Getenv("TEST_NETWORK") == "Z" {
		return ZUsersOnLastPeer[userNumber]
	} else {
		return LocalUsersOnLastPeer[userNumber]
	}
}

//Get the peer name based on network environment:  Z | LOCAL [default]
func GetPeer(peerNumber int) string {
	if os.Getenv("TEST_NETWORK") == "Z" {
		return ZPeers[peerNumber]
	} else {
		return LocalPeers[peerNumber]
	}
}

