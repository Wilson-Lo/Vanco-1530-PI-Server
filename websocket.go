package main
import (
    "log"
    "time"
    "github.com/gorilla/websocket"
    "encoding/json"
    "fmt"
)

var (
    apiConnection  *websocket.Conn
    apiConnected   = false
)
var cloud_ws_url = "ws://139.162.101.171:80/ws"

func initWebSocket(){
   LOOP: for {

        ws, _, err := websocket.DefaultDialer.Dial(cloud_ws_url, nil)
   		if err != nil {
   			fmt.Println("dial err:" + err.Error())
   			log.Printf("wait 5 seconds to redial...")
   			time.Sleep(time.Second * 5)
   			goto LOOP
   		}

   	   defer ws.Close()

   	   _ = ws.WriteMessage(websocket.TextMessage, []byte("{\"method\":\"connect\", \"to\":\"" + getMacAddrs() + "\"}"))
   	   for{
           _, msg, err2 := ws.ReadMessage()
           if err2 != nil {
              log.Println("read:", err2)
              goto LOOP
           }
           var cmd Command
           if err3 := json.Unmarshal([]byte(msg), &cmd); err3 != nil {
              log.Println(err3)
           }
           switch cmd.Method {

               case "cmd":
               		    _ = ws.WriteMessage(websocket.TextMessage, []byte("{\"method\":\"test\"}"))

               default:
               	       log.Printf("~~~~~~~~~~~~~~~~~~~~receive: %s\n", msg)
               	       _ = ws.WriteMessage(websocket.TextMessage, []byte("{\"method\":\"hello\"}"))

               }
       }
   	}
}