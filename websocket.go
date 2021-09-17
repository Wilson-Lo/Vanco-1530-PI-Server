package main
import (
    "log"
    "time"
    "github.com/gorilla/websocket"
    "encoding/json"
   // "fmt"
    // "reflect"
)

var (
    apiConnection  *websocket.Conn
    apiConnected   = false
)
var cloud_ws_url = "ws://139.162.101.171:80/ws"

func initWebSocket(){

    apiConnection = connectApi()
   /* for {
        ws, _, err := websocket.DefaultDialer.Dial(cloud_ws_url, nil)
       fmt.Sprintf("%T", c)
		if err != nil {
			fmt.Println("dial err:" + err.Error())
			log.Printf("wait 5 seconds to redial...")
			time.Sleep(time.Second * 5)
			continue
		}
		c = ws
		defer c.Close()
		break
	}*/

   /* c, _, err := websocket.DefaultDialer.Dial(cloud_ws_url, nil)
	if err != nil {
	    log.Println("dial = ", err)
	//	log.Fatal("dial:", err)
	}*/
	//defer c.Close()

    _ = apiConnection.WriteMessage(websocket.TextMessage, []byte("{\"method\":\"connect\", \"to\":\"" + getMacAddrs() + "\"}"))
   /* if err1 != nil {
        log.Println(err1)
        return
    }*/

	//go func() {
        /*for true {
            log.Printf("send~~~~~~ \n")
            err = c.WriteMessage(websocket.TextMessage, []byte("ping"))
            if err != nil {
           	   log.Println(err)
           	   return
            }
            time.Sleep(5 * time.Second)
        }*/
 //   }()


    for{
        _, msg, err2 := apiConnection.ReadMessage()
	    if err2 != nil {
		    log.Println("read:", err2)

		    return
	    }
        var cmd Command
	    if err3 := json.Unmarshal([]byte(msg), &cmd); err3 != nil {
           log.Println(err3)
        }
        switch cmd.Method {

	      case "cmd":
		         _ = apiConnection.WriteMessage(websocket.TextMessage, []byte("{\"method\":\"test\"}"))

	     default:
	          _ = apiConnection.WriteMessage(websocket.TextMessage, []byte("{\"method\":\"hello\"}"))
	          log.Printf("~~~~~~~~~~~~~~~~~~~~receive: %s\n", msg)

	}


    }
}

func connectApi() *websocket.Conn {

    for {
        log.Println("reconnecting")
        var err error
        apiConnection, _, err = websocket.DefaultDialer.Dial(cloud_ws_url, nil)
        if err != nil {
            apiConnected = false
            log.Println("dial: ", err)
            time.Sleep(4 * time.Second)
            continue
        }

        log.Println("API connected")
        break
    }

    apiConnected = true
    return apiConnection
}
