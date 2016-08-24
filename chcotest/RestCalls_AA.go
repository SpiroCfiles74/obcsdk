package main

import (
	"fmt"
	"obcsdk/peernetwork"
	"obcsdk/chaincode"
	"time"
)

func main() {

	fmt.Println("Peers on network ")
       peernetwork.SetupLocalNetwork(4, false)

	chaincode.Init()
        chaincode.RegisterUsers()

  //get a URL details to get info n chainstats/transactions/blocks etc.
	aPeer, _ := peernetwork.APeer(chaincode.ThisNetwork)
 	url := "http://" + aPeer.PeerDetails["ip"] + ":" + aPeer.PeerDetails["port"]

	//does not work when launching a peer locally
	//fmt.Println("Peers on network ")
	chaincode.NetworkPeers(url)


	fmt.Println("Blockchain: GET Chain  ....")
	chaincode.ChainStats(url)


fmt.Println("**********************************************")
fmt.Println("**********************************************")
fmt.Println("**********************************************")
fmt.Println("**********************************************")
	dAPIArgs11 := []string{"artfun", "init"}
	depArgs10 := []string{"INITIALIZE"}
	chaincode.Deploy(dAPIArgs11, depArgs10)


	iAPIArgs00 := []string{"artfun", "PostUser"}
	iAPIArgs11 := []string{"100", "USER", "Ashley Hart", "PR",  "One Copley Parkway, #216, Morrisville, NC 27560", "9198063535", "admin@itpeople.com", "SUNTRUST", "00017102345", "0234678"}
  iAPIArgs12 := []string{"200", "USER", "Sotheby", "AUCTION HOUSE",  "One Wembley Plaza, #216, London, UK ", "9198063535", "admin@sotheby.com", "Standard Chartered", "00017102345", "0234678"}
	iAPIArgs13 := []string{"300", "USER", "Barry Smith", "COLLECTOR",  "155 Regency Parkway, #111, Cary, 27518 ", "9198063535", "barry@us.ibm.com", "RBC Centura", "00017102345", "0234678"}
	iAPIArgs14 := []string{"400", "USER", "Meghan Kelly", "COLLECTOR",  "155 Sunset Blvd, Beverly Hills, CA, USA ", "9058063535", "barry@us.ibm.com", "RBC Centura", "00017102345", "0234678"}
	iAPIArgs15 := []string{"500", "USER", "Tamara Haskins", "COLLECTOR",  "155 Sunset Blvd, Beverly Hills, CA, USA ", "9058063535", "barry@us.ibm.com", "RBC Centura", "00017102345", "0234678"}

        _, _  = chaincode.Invoke(iAPIArgs00, iAPIArgs11)
	_, _  = chaincode.Invoke(iAPIArgs00, iAPIArgs12)
	_, _  = chaincode.Invoke(iAPIArgs00, iAPIArgs13)
	_, _  = chaincode.Invoke(iAPIArgs00, iAPIArgs14)
	_, _  = chaincode.Invoke(iAPIArgs00, iAPIArgs15)

	time.Sleep(120000 * time.Millisecond);

	qAPIArgs00 := []string{"artfun", "GetUser"}
        qAPIArgs11 := []string {"100", "USER"}
        _, _  = chaincode.Query(qAPIArgs00, qAPIArgs11)

  //post-Item
  iAPIArgs01 := []string{"artfun", "PostItem"}
	iAPIArgs16 :=	[]string{"1000", "ARTINV", "Shadows by Asppen", "Asppen Messer", "10102015", "Original", "Nude", "Canvas", "15 x 15 in", "sample_7.png","$600", "100"}
	_, _  = chaincode.Invoke(iAPIArgs01, iAPIArgs16)

	time.Sleep(80000 * time.Millisecond);

  //post-Auction
	iAPIArgs02 := []string{"artfun", "PostAuctionRequest"}
	iAPIArgs17 :=	[]string{"1111", "AUCREQ", "1000", "200", "100", "04012016", "1200", "INIT", "2016-05-20 11:00:00.3 +0000 UTC","2016-05-23 11:00:00.3 +0000 UTC"}
		  _, _  = chaincode.Invoke(iAPIArgs02, iAPIArgs17)
  time.Sleep(80000 * time.Millisecond);

	fmt.Println("**********************************************")
	fmt.Println("**********************************************")
	fmt.Println("**********************************************")
	fmt.Println("**********************************************")

   //PostBid
	iAPIArgs03 := []string{"artfun", "PostBid"}
	iAPIArgs18 := []string {"1111", "BID", "1", "1000", "300", "1200"}
	iAPIArgs19 := []string{"1111", "BID", "2", "1000", "400", "3000"}
	iAPIArgs20 := []string{"1111", "BID", "3", "1000", "400", "6000"}
	iAPIArgs21 := []string{"1111", "BID", "4", "1000", "500", "7000"}
	iAPIArgs22 := []string{"1111", "BID", "5", "1000", "400", "8000"}
	_, _  = chaincode.Invoke(iAPIArgs03, iAPIArgs18)
	_, _  = chaincode.Invoke(iAPIArgs03, iAPIArgs19)
	_, _  = chaincode.Invoke(iAPIArgs03, iAPIArgs20)
	_, _  = chaincode.Invoke(iAPIArgs03, iAPIArgs21)
	_, _  = chaincode.Invoke(iAPIArgs03, iAPIArgs22)

	time.Sleep(80000 * time.Millisecond);

/****************************
   qAPIArgs01 := []string{"artfun", "GetBid"}
	 qAPIArgs11 := []string {"1111", "1"}
  _, _  = chaincode.Query(qAPIArgs01, qAPIArgs11)
************************/
 qAPIArgs02 := []string{"artfun", "GetListOfBids"}
 qAPIArgs12 := []string {"1111"}
 _, _  = chaincode.Query(qAPIArgs02, qAPIArgs12)

/*************************************
	iAPIArgs11 := []string{"artfun", "PostUserRecord"}
	iArgs11 := []string{"4200", "USER", "Susans Art House", "PERSON",  "One Cary Parkway, #216, Cary, NC 27512", "9198063535", "admin@itpeople.com", "BBT", "00017102345", "0234678"}
	_, _  = chaincode.Invoke(iAPIArgs11, iArgs11)


	time.Sleep(80000 * time.Millisecond);


	qAPIArgs11 := []string{"artfun", "query"}
	qArgs11 := []string{"4200", "USER"}
	_, _  = chaincode.Query(qAPIArgs11, qArgs11)

	iAPIArgs21 := []string{"artfun", "PostUserRecord"}
	iArgs21 := []string{"4300", "USER", "Random Art House", "AUCTION",  "One Cary Parkway, #216, Cary, CA 27512", "9198063535", "admin@itpeople.com", "BBT", "00017102346", "0234677"}
	_, _  = chaincode.Invoke(iAPIArgs21, iArgs21)

	qAPIArgs21 := []string{"artfun", "query"}
	qArgs21 := []string{"4300", "USER"}
	_, _  = chaincode.Query(qAPIArgs21, qArgs21)

	iAPIArgs31 := []string{"artfun", "PostUserRecord"}
	iArgs31 := []string{"4400", "USER", "SIMPLE Art House", "BUSINESS",  "One Cary Parkway, #216, Cary, IL 27512", "9198063535", "admin@itpeople.com", "BBT", "00017102347", "0234679"}
	_, _  = chaincode.Invoke(iAPIArgs31, iArgs31)

	qAPIArgs31 := []string{"artfun", "query"}
	qArgs31 := []string{"4400", "USER"}
	_, _  = chaincode.Query(qAPIArgs31, qArgs31)


	fmt.Println("**********************************************")
	fmt.Println("**********************************************")

	iAPIArgs12 := []string{"artfun", "PostArtRecord"}
	iArgs12 := []string{"2000", "ARTINV", "Modern Art Female Portrait", "Ashley Barber", "10102015", "Original", "Nude", "Canvas", "15 x 15 in", "sample_10.png","$600", "4200"}
	_, _  = chaincode.Invoke(iAPIArgs12, iArgs12)

	time.Sleep(80000 * time.Millisecond);

	qAPIArgs12 := []string{"artfun", "query"}
	qArgs12 := []string{"2000", "ARTINV"}
	_, _  = chaincode.Query(qAPIArgs12, qArgs12)


	fmt.Println("**********************************************")
	fmt.Println("**********************************************")
	iAPIArgs13 := []string{"artfun", "PostAuctionRequest"}
	iArgs13 := []string{"111", "AUCREQ", "2000", "4300", "4200", "04012016", "05312016","$1200"}
	_, _  = chaincode.Invoke(iAPIArgs13, iArgs13)

	time.Sleep(120000 * time.Millisecond);

	qAPIArgs13 := []string{"artfun", "query"}
	qArgs13 := []string{"111", "AUCREQ"}
	_, _  = chaincode.Query(qAPIArgs13, qArgs13)

	fmt.Println("**********************************************")
	fmt.Println("**********************************************")
	iAPIArgs14 := []string{"artfun", "PostTransaction"}
	iArgs14 := []string{"111", "POSTTRAN", "2000", "SALE", "4400", "05312016","$2400","Sold to Private Collecter"}
	//iArgs13 := []string{"111", "AUCREQ", "2000", "4300", "4200", "04012016", "05312016","$1200"}
	_, _  = chaincode.Invoke(iAPIArgs14, iArgs14)

	time.Sleep(120000 * time.Millisecond);

	qAPIArgs14 := []string{"artfun", "query"}
	qArgs14 := []string{"111", "POSTTRAN"}
	_, _  = chaincode.Query(qAPIArgs14, qArgs14)

fmt.Println("**********************************************")
fmt.Println("**********************************************")
fmt.Println("**********************************************")
fmt.Println("**********************************************")
*****************************/

/*********************************************************************
fmt.Println("\nPOST/Chaincode: Deploying chaincode at the beginning ....")
	dAPIArgs0 := []string{"example02", "init"}
	depArgs0 := []string{"a", "20000", "b", "9000"}
  	chaincode.Deploy(dAPIArgs0, depArgs0)
	//fmt.Println("From Deploy error ", err)


	time.Sleep(20000 * time.Millisecond);

  	fmt.Println("\nPOST/Chaincode : Querying a and b after a deploy  ")
	qAPIArgs0 := []string{"example02", "query"}
	qArgsa := []string{"a"}
	_, _  = chaincode.Query(qAPIArgs0, qArgsa)
	qArgsb := []string{"b"}
	_, _  = chaincode.Query(qAPIArgs0, qArgsb)


  	fmt.Println("\nPOST/Chaincode : Invoke on a and b after a deploy >>>>>>>>>>> ")
	iAPIArgs0 := []string{"example02", "invoke"}
	invArgs0 := []string{"a", "b", "500"}
	invRes, _ := chaincode.Invoke(iAPIArgs0, invArgs0)
	fmt.Println("\nFrom Invoke invRes ", invRes)

	fmt.Println("Sleeping 5secs for invoke to complete on ledger")

	time.Sleep(5000 * time.Millisecond);

	fmt.Println("\nBlockchain: Get Chain  ....")
  	chaincode.ChainStats(url)

	fmt.Println("\nPOST/Chaincode: Querying a and b after invoke >>>>>>>>>>> ")
	_, _  = chaincode.Query(qAPIArgs0, qArgsa)
	_, _  = chaincode.Query(qAPIArgs0, qArgsb)


	fmt.Println("\nBlockchain: GET Chain  ....")
  	response2 := chaincode.Monitor_ChainHeight(url)

	fmt.Println("\nChain Height", chaincode.Monitor_ChainHeight(url))

	fmt.Println("\nBlock: GET/Chain/Blocks/")
  	chaincode.BlockStats(url, response2)


  	fmt.Println("\nPOST/Chaincode : Invoke on a and b after a deploy >>>>>>>>>>> ")
	iAPIArgs0 = []string{"example02", "invoke"}
	invArgs0 = []string{"a", "b", "500"}
	invRes, _ = chaincode.Invoke(iAPIArgs0, invArgs0)
	fmt.Println("\nFrom Invoke invRes ", invRes)

	fmt.Println("\nBlockchain: Getting Transaction detail for   ....", invRes)

  	time.Sleep(50000 * time.Millisecond);

	fmt.Println("\nBlockchain: GET Chain  ....")
  	response2 = chaincode.Monitor_ChainHeight(url)

	fmt.Println("\nChain Height", chaincode.Monitor_ChainHeight(url))

	fmt.Println("\nBlock: GET/Chain/Blocks/")
  	chaincode.BlockStats(url, response2)

	//fmt.Println("\nTransactions: GET/transactions/" + invRes)
	chaincode.Transaction_Detail(url, invRes)

	fmt.Println("\nBlockchain: GET Chain .... ")
	time.Sleep(10000 * time.Millisecond);
	chaincode.ChainStats(url)

*************************/
	/*** let's call deploy with tagName */
	//deployUsingTagName()


}
