/*
* Release V1.1 (2021/08/17)
* Release V1.2 (2021/08/18)
* Release V1.3 (2021/08/25)
* Release V1.4 (2021/08/26)
*/
package main
import (
    "fmt"
    "log"
	"time"
   // "net"
    "os/exec"
    "net/http"
    "net"
    "strconv"
    "encoding/hex"
    "strings"
)

var aesKey = []byte("qzy159pkn333rty2")

func main() {
	fmt.Println("pi server v1.4")
	pi4_mac = getMacAddrs()
	fmt.Printf("mac addrs: %q\n",pi4_mac)
	LoadPassword()
	sys_load_vw_Preset()
	LoadSystemConfig()	
	get_pi4_ipconfig()
	loadEmail()
	loadEvent()
	ast_initial_preset()
	router := NewRouter()
	device_info_load("clear")
	ite_device_info_load("clear")
	history_list_load()
	osd_info_load()
	go ast_node_list()
	go ite_node_list()
	go checkIPFormat()
	go func(){
		log.Fatal(http.ListenAndServe(":8080",router))
		/*s := &http.Server{
		Addr:           ":80",
		Handler:        router,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
		}
		log.Fatal(s.ListenAndServe())*/

	}()

	startUDPServer()

	log.Fatal(http.ListenAndServeTLS(":10443","server.crt", "server.key",router))

	defer func() {
	   device_info_save()
	   history_list_save()
	}()
}

func checkIPFormat() {
	fmt.Printf("checkIPFormat~~~~~~~~~~")
    t := time.NewTicker(10 * time.Second)
    for now := range t.C {
      fmt.Println("tick", now)
       i,_:= net.InterfaceByName("eth0")
       var count = 0

       addrs, _ := i.Addrs()
       for _, a := range addrs {
         fmt.Println(a.String())
         if(strings.Contains(a.String(),"127.0.0.1")){
           fmt.Printf("find 127.0.0.1 ~~~~~~~~~~~~~~~")
           cmd1 := exec.Command("sudo","avahi-autoipd","-k","eth0:0")
           _,err1 := cmd1.Output()
		   if err1 != nil {
		 	fmt.Println("err1 = ",err1, "\n")
		   }
           cmd2 := exec.Command("sudo","avahi-autoipd","-D","--force-bind","eth0:0")
           _,err2 := cmd2.Output()
		    if err2 != nil {
			 	fmt.Println("err2 = ",err2, "\n")
		    }
          }
         count = count + 1;
       }

       if(count < 2){
         fmt.Printf("count < 2 ~~~~~~~~~~~~~~~")
         cmd3 := exec.Command("sudo","avahi-autoipd","-k","eth0:0")
          _,err3 := cmd3.Output()
		    if err3 != nil {
				 	fmt.Println("err3 = ",err3, "\n")
		    }
         cmd4 := exec.Command("sudo","avahi-autoipd","-D","--force-bind","eth0:0")
         _,err4 := cmd4.Output()
		    if err4 != nil {
			 	fmt.Println("err4 = ",err4, "\n")
		    }
       }
    }
}


// Start to listen UDP, ip = 255.255.255.255, port = 5002
func startUDPServer(){

    fmt.Println(" Start UDP Server")
    src := "255.255.255.255:5002"
	listener, err := net.ListenPacket("udp", src)
	if err != nil {
		fmt.Println(err.Error())
	}
	defer listener.Close()

	fmt.Printf("UDP server start and listening on %s.\n", src)

	for {
		buf := make([]byte, 1024)
		n, addr, err := listener.ReadFrom(buf)
		if err != nil {
			continue
		}
		go serve(listener, addr, buf[:n])
	}
}

func serve(listener net.PacketConn, addr net.Addr, buf []byte) {

     netAddrArray := strings.Split(addr.String(), ":")
     //receiveCmd := string(buf)
     // fmt.Printf("%s\t: %s\n", netAddrArray[0], receiveCmd)

     if((len(buf)%16) == 0){
      fmt.Printf("phone ip = %s  receive size = %d\n", netAddrArray[0],len(buf))
          decodeData :=AesDecrypt(buf, aesKey)
          fmt.Printf("after aes decode = %s\n", decodeData)

          if(strings.Contains(decodeData, "ETH_REQ")){
             fmt.Printf("GoMax Device request")
          }else{
             fmt.Printf("Others Device request")
          }

         //get IP address
         addrs, err := net.InterfaceAddrs()

         if err != nil {
           fmt.Println(err)
         }

         var tmp_ip string
         var temp_mask [4] string
         var mask net.IPMask
         ipCnt := 0

         for _, addr := range addrs {

         	if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
             var ip net.IP
             switch v := addr.(type) {
                 case *net.IPNet:
                     ip = v.IP
                     mask = v.Mask
                 case *net.IPAddr:
                     ip = v.IP
                     mask = ip.DefaultMask()
                 }
                 if ip == nil {
                     continue
                 }
                 ip = ip.To4()
                 if ip == nil {
                     continue
                 }

         		ipCnt = ipCnt+1

         		if ipCnt>2 {
         		continue
         		}
         	  tmp_ip = ip.String()
              fmt.Println("")
              fmt.Println("ip = " + tmp_ip)
              fmt.Println("mask = ", mask[0], mask[1], mask[2], mask[3])

              temp_mask[0] = strconv.Itoa(int(mask[0]))
              temp_mask[1] = strconv.Itoa(int(mask[1]))
              temp_mask[2] = strconv.Itoa(int(mask[2]))
              temp_mask[3] = strconv.Itoa(int(mask[3]))
         	}

         }
        var feedBackArray [38]byte

        //Device name (Control-Box)
        feedBackArray[5] = 0x43
        feedBackArray[6] = 0x6f
        feedBackArray[7] = 0x6e
        feedBackArray[8] = 0x74
        feedBackArray[9] = 0x72
        feedBackArray[10] = 0x6f
        feedBackArray[11] = 0x6c
        feedBackArray[12] = 0x2d
        feedBackArray[13] = 0x42
        feedBackArray[14] = 0x6f
        feedBackArray[15] = 0x78

        //IP
        ipArray := strings.Split(tmp_ip, ".")
        var ipIndex = 27
        for counter := 0; counter < len(ipArray); counter++ {
            if(ipIndex > 30){
               break
            }
            ip, err := strconv.Atoi(ipArray[counter])
            if err != nil {
            }
            feedBackArray[ipIndex] = byte(ip)
            ipIndex++
        }

        //mac address
        data, err := hex.DecodeString(getMacAddrs())
        if(err != nil){
           fmt.Println("mac address error : " , err)
        }

        var macIndex = 21
        for counter := 0; counter < len(data); counter++ {
            if(macIndex > 26){
               break
            }
            feedBackArray[macIndex] = data[counter]
            macIndex++
        }

       //mask
       var maskIndex = 31
       for counter := 0; counter < 4; counter++ {
           if(maskIndex > 34){
              break
           }
           fmt.Println("temp_mask : " , temp_mask[counter])
           mask, err := strconv.Atoi(temp_mask[counter])
           if err != nil {
           }

           feedBackArray[maskIndex] = byte(mask)
           maskIndex++
       }

        //send UDP feedback
        encodeData := AesEncrypt(feedBackArray[:], aesKey)
        sendUDP("255.255.255.255:65088", encodeData)

     }else{
       fmt.Println("non aes data")
     }
}

//addr
func sendUDP(addr string, msg []byte) {

    fmt.Println("sendUDP data size: %d" , len(msg))
	conn, _ := net.Dial("udp", addr)

    _, err := conn.Write(msg)
	if err != nil{
	   fmt.Println("send UDP err ", err)
	}
	// listen for reply
/*	bs := make([]byte, 1024)
	conn.SetDeadline(time.Now().Add(3 * time.Second))
	len, err := conn.Read(bs)
	if err != nil {
		return "", err
	} else {
		return string(bs[:len]), err
	} */
}