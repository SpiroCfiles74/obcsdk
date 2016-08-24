package main

import (
	"fmt"
	"obcsdk/peerrest"
)



func main() {


        enrollId := "user_type1_d4433d2296"
        enrollSecret :=  "e1ebb93796"
        escQuote          := "\""
        openBrace         := "{"
        closeBrace        := "}"
        comma             := ","
        colon             := ":"
        newLine           := "\n"

        RegisterJsonPart1 := newLine + openBrace + newLine + escQuote + "enrollId" + escQuote + colon + escQuote + enrollId
        RegisterJsonPart2 := escQuote + comma + newLine + escQuote + "enrollSecret" + escQuote + colon + escQuote + enrollSecret
        RegisterJsonPart3 := escQuote + newLine + closeBrace + newLine


       fmt.Println("Calling https")
       res, res2 := peerrest.GetChainInfo("https://e86c517d-f670-4f16-95f2-2c20957414e1_vp1-api.zone.blockchain.ibm.com:443/chain")
       fmt.Println("Res: ", res)
       fmt.Println("Res: ", res2)

       fmt.Println("Testing PoST")
       payLoadString := RegisterJsonPart1 + RegisterJsonPart2 + RegisterJsonPart3 
       fmt.Println("payLoadString: ", payLoadString)
       payLoad := []byte(payLoadString)

       res, res2 = peerrest.PostChainAPI("https://e86c517d-f670-4f16-95f2-2c20957414e1_vp0-api.zone.blockchain.ibm.com:443/registrar", payLoad)
       fmt.Println("Res: ", res)
       fmt.Println("Res: ", res2)
}
