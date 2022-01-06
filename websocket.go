package main
import (
    "log"
    "bytes"
    "time"
    "github.com/gorilla/websocket"
    "encoding/json"
    "fmt"
    "strings"
     b64 "encoding/base64"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 30 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10
)

var cloud_ws_url = "wss://x-space.cloud/api/ws"

func initWebSocket(){
  // Calling NewTicker method

   LOOP: for {

        ws, _, err := websocket.DefaultDialer.Dial(cloud_ws_url, nil)
   		if err != nil {
   			fmt.Println("dial err:" + err.Error())
   			log.Printf("wait 5 seconds to redial...")
   			time.Sleep(time.Second * 5)
   			goto LOOP
   		}

   	   defer ws.Close()

       //connect to cloud server
   	   _ = ws.WriteMessage(websocket.TextMessage, []byte("{\"method\":\"connect\", \"to\":\"" + getMacAddrs() + "\"}"))
       ch := make(chan bool)
       go SendPingPong(ws, ch)

       var cmd Command

   	   for{
           _, msg, err2 := ws.ReadMessage()
           if err2 != nil {
              log.Println("read:", err2)
              ch <- true
              goto LOOP
           }
           log.Println("receive msg = " , msg)
           if err3 := json.Unmarshal([]byte(msg), &cmd); err3 != nil {
              log.Println(err3)
           }

           switch cmd.Method {

               case "get_info"://get device info
                        var bodyData = strings.ReplaceAll(cmd.Body, SaltFirst, "")
                        bodyData = strings.ReplaceAll(bodyData, SaltAfter, "")
                        byteData, err := b64.StdEncoding.DecodeString(bodyData)
                        if(err != nil){
                          cmd.Etag = cmd.Etag
                          cmd.Method = "cmd"
                          cmd.Body = EncryptionData("{ \"result\": \"failed\"}")
                          cmd.Sign = GetSign(cmd)
                          SendBackToHttps(ws, cmd)
                          return
                        }

                        json.Unmarshal(byteData, &cmd)

                        var systemConfig = GetSystemConfig()

                        cmd.Etag = cmd.Etag
                        cmd.Method = "cmd"
                        cmd.Body = EncryptionData("{ \"result\": \"ok\", \"fw\": \"" + GetSystemFWVersion() + "\",  \"ip\": \"" + systemConfig.IP + "\", \"mask\": \"" + systemConfig.MASK + "\", \"gateway\": \"" + systemConfig.GATEWAY + "\"}")
                        cmd.Sign = GetSign(cmd)
                        SendBackToHttps(ws, cmd)
                        break

               //get device list
               case "node_list":
                        var data =  ws_node_list()
               		    _ = ws.WriteMessage(websocket.TextMessage, []byte("{\"body\":" + data + "}"))
                        break

               case  "pong":
                       log.Println("receive pong")
                       break

               default:
               	       log.Printf("~~~~~~~~~~~~~~~~~~~~receive: %s\n", msg)
               	       _ = ws.WriteMessage(websocket.TextMessage, []byte("{\"method\":\"hello\"}"))

               }
       }
   	}
}

func SendBackToHttps(conn *websocket.Conn, cmd Command){
     reqBodyBytes := new(bytes.Buffer)
     json.NewEncoder(reqBodyBytes).Encode(cmd)
     _ = conn.WriteMessage(websocket.TextMessage, []byte(reqBodyBytes.Bytes()))
}


/**
* Send Ping to cloud server & receive pong from cloud server
*/
func SendPingPong(conn *websocket.Conn, c chan bool){
     ticker := time.NewTicker(pingPeriod)

	 defer func() {
	    log.Printf("ticker.Stop")
		ticker.Stop()
	 }()

     for {
       select{

          case <-c:
               log.Printf("stop timer")
        	   return

		  case <-ticker.C:
			 log.Printf("send ping pong")
             _ = conn.WriteMessage(websocket.TextMessage, []byte("ping"))
		     break
		 }
	 }
}
