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

       var cmd CommandObject

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

               case  "switch_channel"://switch channel
                       var bodyData = strings.ReplaceAll(cmd.Body, SaltFirst, "")
                        bodyData = strings.ReplaceAll(bodyData, SaltAfter, "")
                        byteData, err := b64.StdEncoding.DecodeString(bodyData)
                        if(err != nil){
                            cmd.Method = "cmd"
                            cmd.Body = EncryptionData("{ \"result\": \"failed\"}")
                            cmd.Sign = GetSign(cmd)
                            SendBackToHttps(ws, cmd)
                            return
                        }
                        var switchChannelObject SwitchChannelObject
                        if err := json.Unmarshal([]byte(byteData), &switchChannelObject); err != nil {
                           log.Println(err)
                        }
                        log.Println("ip = ", switchChannelObject.IP, " Channel = ", switchChannelObject.Channel, " type = ", switchChannelObject.Type)
                        var feedback = astswitch(switchChannelObject.IP ,switchChannelObject.Channel ,switchChannelObject.Type)
                        log.Println("switch feedback = ", feedback)
                        if(strings.Contains(feedback, "OK")){
                           cmd.Method = "cmd"
                           cmd.Body = EncryptionData("{ \"result\": \"ok\"}")
                           cmd.Sign = GetSign(cmd)
                           SendBackToHttps(ws, cmd)
                        }else{
                           cmd.Method = "cmd"
                           cmd.Body = EncryptionData("{ \"result\": \"failed\", \"message\": \"Switch channel failed !\"}")
                           cmd.Sign = GetSign(cmd)
                           SendBackToHttps(ws, cmd)
                        }
                        break

               case "get_info"://get control box info
                        var systemConfig = GetSystemConfig()
                        cmd.Method = "cmd"
                        cmd.Body = EncryptionData("{ \"result\": \"ok\", \"fw\": \"" + GetSystemFWVersion() + "\",  \"ip\": \"" + systemConfig.IP + "\", \"mask\": \"" + systemConfig.MASK + "\", \"gateway\": \"" + systemConfig.GATEWAY + "\"}")
                        cmd.Sign = GetSign(cmd)
                        SendBackToHttps(ws, cmd)
                        break

               case "node_list": //get device list
               		    cmd.Method = "cmd"
                        cmd.Body = EncryptionData("{ \"result\": \"ok\", \"device_list\":" + ws_node_list() + "}")
                        cmd.Sign = GetSign(cmd)
                        SendBackToHttps(ws, cmd)
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

func SendBackToHttps(conn *websocket.Conn, cmd CommandObject){
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
