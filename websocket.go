package main
import (
    "log"
    "time"
    "github.com/gorilla/websocket"
    "encoding/json"
    "fmt"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 30 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10
)

var cloud_ws_url = "wss://x-space.cloud/ws"

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

   	   _ = ws.WriteMessage(websocket.TextMessage, []byte("{\"method\":\"connect\", \"to\":\"" + getMacAddrs() + "\"}"))
       ch := make(chan bool)
       go SendPingPong(ws, ch)

   	   for{
           _, msg, err2 := ws.ReadMessage()
           if err2 != nil {
              log.Println("read:", err2)
              ch <- true
              goto LOOP
           }

           var cmd Command
           if err3 := json.Unmarshal([]byte(msg), &cmd); err3 != nil {
              log.Println(err3)
           }

           switch cmd.Method {

               case "cmd":
               		    _ = ws.WriteMessage(websocket.TextMessage, []byte("{\"method\":\"test\"}"))
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
