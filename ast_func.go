package main
import (
	"fmt"
	"net/http"
    "log"
    "net"
    "bytes"
    "unsafe"
    "reflect"
    "time"    
    "strings"
    "strconv"
    "encoding/json"
	"encoding/base64"
    "io/ioutil"
    "github.com/gorilla/mux"
	"image/png"
	"image/jpeg"
	"os"
)

type reply struct {
	device_type uint32	
	device_function uint32
	device_status[32] byte 
	device_name[256] byte
}

type Ast_node struct {
	IP string `json:"ip"`
	Is_host string `json:"is_host"`
	Status string  `json:"status"`
	Host_name string  `json:"host_name"`
}

type Ast_info struct {	
	IP string `json:"ip"`
	Host_type string `json:"type"`
	Host_name string  `json:"host_name"`
	Device_mac string  `json:"mac"`
	Device_pin string  `json:"pin"`
	Device_id string  `json:"id"`
	Uptime string  `json:"uptime"`
	Alive string  `json:"alive"`
	aliveCnt int  `json:"aliveCount"`
	videoStatus string  `json:"video"`
	FW_ver string  `json:"fw"`	
	OSD string `json:"osd_index"`
	RHost_name string  `json:"rhost_name"`//user define hostname
}

type Ast_preset struct {
	INDEX string `json:"index"`
    SETTING []Ast_info `json:"setting"` 
}

type Ast_preset_array struct {	
    PRESET []Ast_preset `json:"preset"` 
}

type Ast_setup_info struct {	
	FW string `json:"fw"`
	RX_TV string `json:"rx_tv"`
	RX_SCALE string `json:"rx_scale"`
	RX_Rotate string `json:"rx_rotate"`
	TX_SRC string `json:"tx_src"`
	TX_NET_PROFILE string `json:"tx_net"`
	RS232 string `json:"rs232"`
}

type Ast_vw_info struct {	
	VWH string `json:"vw_h"`
	VWV string `json:"vw_v"`
	VWP string `json:"vw_pos"`
	VWL string `json:"vw_l"`
	VWR string `json:"vw_r"`
	VWU string `json:"vw_u"`
	VWB string `json:"vw_b"`
}

type ast_osd_info struct {	
	//gerenal
	OSD_name string `json:"name"`
	OSD_type bool `json:"type"`
	OSD_alpha string `json:"alpha"`
	OSD_mask string `json:"mask"`
	OSD_offtime string `json:"offtime"`	
	//font	
	OSD_string string `json:"str"`	
	OSD_fontSize string `json:"fsize"`
	OSD_fontColor string `json:"fcolor"`
	//Position
	OSD_position string `json:"pos"`
	OSD_px string `json:"x"` 
	OSD_py string `json:"y"`	
	//image	
	OSD_imgName string `json:"pic"`
		
}

var debugFlag_stayTime bool = false
var ast_node_array[64] Ast_node
var ast_node_num int
var ast_info_map =make(map[string]*Ast_info)  
var ast_preset_map[8] Ast_preset
var ast_osd_map[8] ast_osd_info
var ast_busy = false 


func ast_initial_preset(){

	ast_loadPreset("1")
	ast_loadPreset("2")
	ast_loadPreset("3")
	ast_loadPreset("4")
	ast_loadPreset("5")
	ast_loadPreset("6")
	ast_loadPreset("7")
	ast_loadPreset("8")
}


func ast_del_node_preset(mac string){
	

	var dev_type string
	dev_id := "0000"
	var tmp []Ast_info
	var flag bool
	var name string
	
	dev_type = "n"
	for _, map_data := range ast_info_map {	
		if map_data.Device_mac == mac{
		dev_type = map_data.Host_type
		if dev_type =="t"{
		dev_id = map_data.Device_id
		name = map_data.Host_name
		break
		}		
		}
	}
	add_event("Setup","To delete device "+name)
	
	for i := 0; i < 8; i++ {
	tmp = nil
	flag = false	
	if dev_type =="n"{
	continue
	}else if dev_type =="t"{
		//fmt.Println("delete the tx")
		for _, map_data := range ast_preset_map[i].SETTING {
					if map_data.Device_id == dev_id	{
						fmt.Println("found " + dev_id)
						flag = true
						continue
					}else{
					tmp = append(tmp, map_data)		
					}
			}
			if flag==true{
			ast_preset_map[i].SETTING = tmp
			}
	}else{
	//fmt.Println("delete the rx")
	for _, map_data := range ast_preset_map[i].SETTING {
				if map_data.Device_mac == mac	{
					fmt.Println("found " + mac)
					flag = true
					break
				}else{
				tmp = append(tmp, map_data)		
				}
		}
		if flag==true{
		ast_preset_map[i].SETTING = tmp
		}
	}
	//save preset 
	//var modifyNode_slice []Ast_info = tmp
		if flag==true{
			index := strconv.Itoa(i+1)   
			path := "./preset/preset"+ index + "_info.json"
			json_node_modify ,_:= json.Marshal(tmp)
			_ = ioutil.WriteFile(path,json_node_modify, 0644)
			fmt.Println("preset " +index+" changed")
		}
	}
	
}

func ast_modifyPreset(index string, data []Ast_info)(string){  

	k, _ := strconv.Atoi(index)
	if k>8{
	return "DAIL_ERR"
	}	
	var modifyNode_slice []Ast_info = data
	path := "./preset/preset"+ index + "_info.json"
	json_node_modify ,_:= json.Marshal(modifyNode_slice)
	_ = ioutil.WriteFile(path,json_node_modify, 0644)	
	
	ast_loadPreset(index)	
	return "OK"
}


func ast_savePreset(index string)(string){  

	k, _ := strconv.Atoi(index)
	if k>8{
	return "DAIL_ERR"
	}
	node_slice := make([]Ast_info, 0, len(ast_info_map))
	for _, tx := range ast_info_map {
      node_slice = append(node_slice, *tx)
	} 
	path := "./preset/preset"+ index + "_info.json"
	json_node_list ,_:= json.Marshal(node_slice)
	_ = ioutil.WriteFile(path,json_node_list, 0644)
	event := "Set current configuration to preset "+ index
	add_event("Mapping",event)
	ast_loadPreset(index)	
	return "OK"
}

func ast_loadPreset(index string){

	k, _ := strconv.Atoi(index)
	if k>8{
	return
	}
	//clear array
	ast_preset_map[k-1].SETTING = nil
	
	path := "./preset/preset"+ index + "_info.json"
	file, err := ioutil.ReadFile(path)
	ast_preset_map[k-1].INDEX = index
	if err!=nil {	
	return;
	}
	var node_slice []Ast_info
	_ = json.Unmarshal([]byte(file), &node_slice)
   
	for i := 0; i < len(node_slice); i++ {
			ast_preset_map[k-1].SETTING = append(ast_preset_map[k-1].SETTING, node_slice[i])	
		}
}

func api_debug(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")//delete this line when FW is publishing **************************

	r_updateTime := r.FormValue("updateTime")
	if r_updateTime != ""{
		if r_updateTime == "true"{
		debugFlag_stayTime = true
		fmt.Println("start device time log")
		}else{
		debugFlag_stayTime = false
		fmt.Println("stop device time log")
		}
	}else{
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w,"{\"result\":\"%v\"}","OK")
	w.(http.Flusher).Flush()
}


func insertNth(s string, n int) string {
	var buffer bytes.Buffer
	var n_1 = n - 1
	var l_1 = len(s) - 1
	for i, rune := range s {
		buffer.WriteRune(rune)
		if i%n == n_1 && i != l_1 {
			buffer.WriteRune(':')
		}
	}
	return buffer.String()
}

func astLogin(con net.Conn,bs []byte)  (error){
	//var s_dbg1,s_dbg2,s_dbg3 string //debug
	fmt.Println("astLogin start")
	 // set SetReadDeadline
    err := con.SetReadDeadline(time.Now().Add(3 * time.Second))
    if err != nil {
       fmt.Println("SetReadDeadline failed:", err)
    }
	len, err := con.Read(bs)
	s_tmp := string(bs[:len])
	//s_dbg1 = s_tmp  ////debug	
	if !strings.Contains(s_tmp,"login:")	{
		len, _ = con.Read(bs)
		s_tmp = string(bs[:len])
		//s_dbg2 = s_tmp  ////debug
		if !strings.Contains(s_tmp,"login:")	{
			len, _ = con.Read(bs)
		//	s_tmp = string(bs[:len])
		//	s_dbg3 = s_tmp  ////debug		
		}
	}

	fmt.Println("account root")
	con.Write([]byte("root\r\n"))
	fmt.Println("read 1")
	len, err = con.Read(bs)
	if err != nil {
       if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
           fmt.Println("read timeout:", err)
           // time out
           return err
       } else {
           fmt.Println("read error:", err)
           // some error else, do something else, for example create new conn
           return err
       }
    }
	s_tmp = string(bs[:len])
	if(strings.Contains(s_tmp, "/ #")){
	   fmt.Println("login success")
	   return nil
	}else{
	    fmt.Println("wait login again")
	    len, err = con.Read(bs)
	    if err != nil {
           if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
               fmt.Println("read timeout:", err)
                   // time out
               return err
           } else {
               fmt.Println("read error:", err)
                   // some error else, do something else, for example create new conn
               return err
           }
        }
    	if string(bs[:len]) != "/ # " {
    	  fmt.Println("astLogin(con,bs) != nil")
    	  //fmt.Println("dbg1="+s_dbg1)	////debug
    	  //fmt.Println("dbg2="+s_dbg2)	////debug
    	  //fmt.Println("dbg2="+s_dbg3)  ////debug
    	  fmt.Println(s_tmp)
    	  fmt.Println(string(bs[:len]))
    	  fmt.Println(bs[:len])
    		return err
    	}
    	return nil
	}
}

func astSendCMD(con net.Conn,cmd string,bs []byte)  (string , error){
	
	//fmt.Println(cmd)
	con.Write([]byte(cmd))
	time.Sleep(time.Duration(80)*time.Millisecond)

    // set SetReadDeadline
    err := con.SetReadDeadline(time.Now().Add(5 * time.Second))

    if err != nil {
         fmt.Println("SetReadDeadline failed:", err)
    }

	len, err := con.Read(bs)
     if err != nil {
        if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
            fmt.Println("read timeout:", err)
            // time out
        } else {
            fmt.Println("read error:", err)
            // some error else, do something else, for example create new conn
        }
     }
	//if cmd=="cat /sys/devices/platform/videoip/State"{
	//fmt.Println("1st:",string(bs[:len]))
	//}

	con.Write([]byte("\r\n"))

	if cmd=="cat /sys/devices/platform/videoip/State"{
	   len, err = con.Read(bs)
	   fmt.Println("1st:",string(bs[:len]))
	   time.Sleep(time.Duration(20)*time.Millisecond)
	   len, err = con.Read(bs)
	   fmt.Println("2st:",string(bs[:len]))
	}else{
	   len, err = con.Read(bs)
	   len, err = con.Read(bs)
	}
	


	//if cmd=="astparam g switchpin"{
	/*if cmd=="cat /proc/uptime"{
	fmt.Println("astSendCMD:"+string(bs[:len]))  //debug
	}
	if cmd=="astparam r ethaddr"{
	fmt.Println("RO:"+string(bs[:len]))  //debug
	}
	if cmd=="lmparam g MY_MAC"{
	fmt.Println("astSendCMD:"+string(bs[:len]))  //debug
	}*/
	//fmt.Println("astSendCMD:")
	s := strings.Replace(string(bs[:len]), "/ # ", "",-1)
	return strings.Replace(s, "\r\n", "",-1), err
}

func astSend(con net.Conn,cmd string,bs []byte)  (string){
	if astLogin(con,bs) != nil {
		return "ERR_FMT"
	} 
	s, err := astSendCMD(con,cmd,bs)
	if err != nil {
		return "ERR_FMT"
	} else {
		return s
	}
}

func ast_node_list() {
	/////////////send query command
	go query_list()
	/////////////server and receiver
	addr, err := net.ResolveUDPAddr("udp4","225.1.0.0:3334")
	if err != nil {
		log.Fatal(err)
	}
	
	// Open up a connection	
	conn, err := net.ListenMulticastUDP("udp4", nil,addr)
	if err != nil {
		log.Fatal(err)
	}
	conn.SetReadBuffer(100000)
	// Loop forever reading from the socket
	var ss []byte
	for {
		buffer := make([]byte, 8192)
		//time.Sleep(time.Duration(50)*time.Millisecond)
		numBytes, src, err := conn.ReadFromUDP(buffer)
		if err != nil {
			log.Fatal("ReadFromUDP failed:", err)
		}
		numBytes=numBytes
		
		myreply := (*reply)(unsafe.Pointer((*reflect.SliceHeader)(unsafe.Pointer(&buffer)).Data))
		
		if(myreply.device_type == 1){
			ast_node_array[ast_node_num].Is_host = "t"
		}		else{
			ast_node_array[ast_node_num].Is_host = "r"
		}
		ast_node_array[ast_node_num].IP = src.IP.String()
		ss=myreply.device_status[:]
		ast_node_array[ast_node_num].Status = string(ss[:bytes.Index(ss,[]byte("\x00"))])
		ss=myreply.device_name[:]
		ast_node_array[ast_node_num].Host_name = string(ss[:bytes.Index(ss,[]byte("\x00"))])
		fmt.Println(ast_node_array[ast_node_num].Host_name+"-",ast_node_num+1)
		ast_node_num++
		
	}	
	conn.Close()
}

func query_list(){
	dataChangeFlag := false
	addr, err := net.ResolveUDPAddr("udp4","225.1.0.0:3333")
	if err != nil {
	    fmt.Println("udp error 1 \n")
		log.Fatal(err)
	}
	
	conn, err := net.DialUDP("udp4",aliasUDPAddr,addr)  //node_list alias IP
	if err != nil {
		    fmt.Println("udp error 2 \n")
		log.Fatal(err)
	}
	query := make([]byte, 8)
	bs := make([]byte, 1024)
	for {
		fmt.Println("start udp scan \n")
		conn.Write(query)	
		ast_node_num = 0; //reset list
		time.Sleep(time.Duration(5)*time.Second)
        fmt.Println("start to check1")
		ast_busy = true
		for _, map_data := range ast_info_map {// check who is offline
			find_node := false
			for i := 0; i<ast_node_num; i++	{
				//show search device all
				//fmt.Println("dev-"+ ast_node_array[i].Device_mac);
				if map_data.IP == ast_node_array[i].IP	{
					find_node = true
					if map_data.Alive == "n" {
						map_data.Alive = "y"
						map_data.aliveCnt = 0
						add_event("Mapping",map_data.Host_name+" is online")
					}
					break
				}
			}
			if find_node  == false && map_data.Alive == "y" {

				if map_data.aliveCnt>1{
				check_connection_lost(map_data.Device_mac)
				add_event("Mapping",map_data.Host_name+" is offline")
				map_data.Alive = "n"
				}else{
				map_data.aliveCnt = map_data.aliveCnt +1
				fmt.Println(map_data.Host_name + " alive loss-",map_data.aliveCnt);
				}
			}
		}
		fmt.Println("start to check2")
		for i := 0; i<ast_node_num; i++ {
		    fmt.Println("ip = " + ast_node_array[i].IP)
			con, err := aliasDialer.Dial("tcp",ast_node_array[i].IP+":24")
			fmt.Println("after connect")
			if err != nil {
				fmt.Println(err);
			    fmt.Println("connect error 1")
				//con.Close()
				continue
				//log.Fatal(err)
			}
		    fmt.Println("before login")
			if astLogin(con,bs) != nil {
			    fmt.Println("connect error 2")
				//fmt.Println("if astLogin(con,bs) != nil")  //debug	
				//log.Fatal(err)
				if(con != nil){
				   con.Close()
				}
				continue
			} 
		  fmt.Println("get mac 1")
		  tmp_MAC, _ := astSendCMD(con,"lmparam g MY_MAC",bs)
		  fmt.Println("get mac 2 : " + tmp_MAC)
		  //fmt.Println("MAC-",tmp_MAC);

		  fmt.Println("start to check3")
		  if tmp_MAC != "" && strings.Count(tmp_MAC,"") == 13	&& !strings.ContainsAny(tmp_MAC,"MY"){

		         fmt.Println("old device found")

		     	var s_temp string
		     	var s_pin_56 string


				if map_tmp, ok := ast_info_map[tmp_MAC]; ok {

					if ast_node_array[i].Is_host == "t"  {   //snapshot
					s_temp, _ := astSendCMD(con,"cat /sys/devices/platform/videoip/State",bs)

                    fmt.Println("old device group ID =  " + map_tmp.Device_id)
                    fmt.Println("send preview 1 "+ tmp_MAC+ pi4Addr)

					_, _ = astSendCMD(con,"echo jpg 320 240 x x > /dev/videoip",bs)
                    _, _ = astSendCMD(con,"cat /dev/videoip > " + tmp_MAC + ".jpg",bs)
	                _, _ = astSendCMD(con,"tftp -p -l " + tmp_MAC + ".jpg " + pi4Addr ,bs)

                    fmt.Println("send preview 2 "+ tmp_MAC+ pi4Addr)

					if strings.Contains(s_temp,"DETECTING_MODE"){
						    fmt.Println("DETECTING_MODE "+ tmp_MAC+ pi4Addr)


					}else if strings.Contains(s_temp,"OPERATING"){
							fmt.Println("OPERATING " + tmp_MAC+ pi4Addr)


					}else if strings.Contains(s_temp,"operating"){
							fmt.Println("operating "+ tmp_MAC+ pi4Addr)


					}else{
					 fmt.Println(tmp_MAC+"->err:"+s_temp)
					}

					}

                    fmt.Println("device now ip = " + ast_node_array[i].IP)
					if map_tmp.IP != ast_node_array[i].IP{
						map_tmp.IP = ast_node_array[i].IP
						dataChangeFlag =true
					}

					if map_tmp.Host_type == "t" {
						s_temp, _ = astSendCMD(con,"lmparam g CH_SELECT",bs)
						fmt.Println("old TX device group ID =  " + map_tmp.Device_id + "now group id = " + s_temp)
						if s_temp != "" && strings.Count(s_temp,"") == 5 {
                        	//check ID is number
                        	if _, err := strconv.Atoi(s_temp); err == nil {
                        	  if map_tmp.Device_id != s_temp{
                                 fmt.Println("Reset TX channel to group id " + map_tmp.Device_id)
                                 s_temp, _ = astSendCMD(con,"e e_reconnect::"+map_tmp.Device_id,bs)
                                 s_temp = map_tmp.Device_id
                              }else{
                                fmt.Println("TX channel the same")
                              }
                        	} else {
                        	 fmt.Println("TX not a number\n")
                        	}
                        }else{
                          fmt.Println("TX ID size not correct \n")
                        }

					}else {
						s_temp, _ = astSendCMD(con,"lmparam g CH_SELECT_V",bs)
						fmt.Println("old RX device group ID =  " + map_tmp.Device_id + "now group id = " + s_temp)
						if s_temp != "" && strings.Count(s_temp,"") == 5 {
							//check ID is number
                            if _, err := strconv.Atoi(s_temp); err == nil {
                               if map_tmp.Device_id != s_temp{
                                  fmt.Println("Reset RX channel to group id " + map_tmp.Device_id)
                                  s_temp, _ = astSendCMD(con,"e e_reconnect::" + map_tmp.Device_id + "::z",bs)
                                  s_temp = map_tmp.Device_id
                              }else{
                                fmt.Println("RX channel the same")
                              }
                            }else{
                             fmt.Println("RX not a number\n")
                            }
						}else{
						  fmt.Println("RX ID size not correct \n")
						}
					}

					//check ID length
					if s_temp != "" && strings.Count(s_temp,"") == 5 {
						//check ID is number
						if _, err := strconv.Atoi(s_temp); err == nil {
							if map_tmp.Device_id != s_temp{
							map_tmp.Device_id = s_temp
							fmt.Println("ID Update")
							dataChangeFlag =true
							}
						} else {
							fmt.Println("not a number\n")
						}
					}
					if strings.Contains(map_tmp.Device_id,"not defined"){
						map_tmp.Device_id = map_tmp.Device_pin
						fmt.Println("ID UNDEFINE Update")
						dataChangeFlag =true
					}

					/*if len(map_tmp.RHost_name) > 0 {
						if map_tmp.RHost_name != ast_node_array[i].Host_name {
							fmt.Println(tmp_MAC+":user define host name not same as aspeed host name-"+map_tmp.RHost_name+"-"+ast_node_array[i].Host_name)
							s_temp, _ = astSendCMD(con,"echo "+map_tmp.RHost_name+" > /etc/hostname",bs)
						}
					} else {

						//fmt.Println(map_tmp.Host_name)
						//fmt.Println(ast_node_array[i].Host_name)
						if map_tmp.Host_name != ast_node_array[i].Host_name{
						if map_tmp.Host_type=="t"{
						//hostnameTmp := "ast3-gateway"+map_tmp.Device_id
						s_temp, _ = astSendCMD(con,"echo "+ map_tmp.Host_name+" > /etc/hostname",bs)
						//map_tmp.Host_name = hostnameTmp
						}else{
						map_tmp.Host_name = ast_node_array[i].Host_name

						}
						//
						fmt.Println("Host_name Update")
						dataChangeFlag =true
						}
					}*/
                 //   s_temp, _ = astSendCMD(con,"e e_reconnect",bs)
                    s_pin_56, _ = astSendCMD(con,"cat /sys/devices/platform/leds/leds/button_link/N_Pin",bs)
					s_temp, _ = astSendCMD(con,"lmparam g CH_SELECT",bs)
					if strings.Contains(s_temp,"not defined") {
						s_temp = "na"
					}

                    s_temp = s_temp + s_pin_56
                    if(len(s_temp) >= 6){
                       s_temp = s_temp[0:6]
                    }

					if s_temp != "" && strings.Count(s_temp,"") == 7 {

						if _, err := strconv.Atoi(s_temp); err == nil {
							if map_tmp.Device_pin != s_temp{
                              map_tmp.Device_pin = s_temp
                              fmt.Println("Pin Update : ", map_tmp.Device_pin)
                              dataChangeFlag =true
                             }
						} else {
							fmt.Println("Pin not a number")
						}
					}

					map_tmp.Alive = "y"

					s_temp, err = astSendCMD(con,"cat /proc/uptime",bs)
					fmt.Println("live time = "+ tmp_MAC + ":" + s_temp)
					if s_temp == ""{
						fmt.Println(tmp_MAC+": time null")
					}else{
				     	map_tmp.Uptime = strings.Split(s_temp, " ")[0]
				        dataChangeFlag =true
					}

				}else {// new device found

				//JK
				//Current R/W mac add ":"
				fmt.Println("new device found")
				RW_MAC := insertNth(tmp_MAC,2)

				//Get RO mac
				s_temp, _ = astSendCMD(con,"astparam r ethaddr",bs)

				s_temp = strings.ToUpper(s_temp)
				RW_MAC = strings.ToUpper(RW_MAC)

				//aspeed defult macsddress
				if s_temp == "00:15:25:08:01:11" || s_temp == "00:15:25:35:00:11"{
					fmt.Println("RO is default change RW to RO")
					s_temp, _ = astSendCMD(con,"astparam w ethaddr "+RW_MAC ,bs)
					s_temp, _ = astSendCMD(con,"astparam save ro" ,bs)
				}else if s_temp!= RW_MAC{
					fmt.Println("RW and RO are different.")
					s_temp, _ = astSendCMD(con,"astparam s ethaddr "+s_temp ,bs)
				    s_temp, _ = astSendCMD(con,"astparam save",bs)
					s_temp, _ = astSendCMD(con,"reboot",bs)
					fmt.Println("reboot")
				}else{
				  fmt.Println("new")
				  tmp := Ast_info{}
				  tmp.IP = ast_node_array[i].IP
				  tmp.Host_type = ast_node_array[i].Is_host
				  tmp.Host_name = ast_node_array[i].Host_name
				  fmt.Println(ast_node_array[i].Host_name)
				  tmp.Device_mac = tmp_MAC
				  tmp.OSD = "0";

				  s_pin_56, _ = astSendCMD(con,"cat /sys/devices/platform/leds/leds/button_link/N_Pin",bs)
                  s_temp, _ = astSendCMD(con,"lmparam g CH_SELECT",bs)

                  if strings.Contains(s_temp,"not defined") {
                    	s_temp = "na"
                  }

                  s_temp = s_temp + s_pin_56
                  s_temp = s_temp[0:6]
			      tmp.Device_pin = s_temp

					fmt.Println("Wilson - Host_type")
					if tmp.Host_type == "t" {
						s_temp, _ = astSendCMD(con,"astparam g ch_select",bs)
					}	else {
						s_temp, _ = astSendCMD(con,"astparam g ch_select_v",bs)
					}
					fmt.Println("Wilson - Device_id")
					tmp.Device_id = s_temp
					if strings.Contains(tmp.Device_id,"not defined") {
						tmp.Device_id = tmp.Device_pin
					}
					fmt.Println("Wilson - Alive")
					tmp.Alive = "y"
					s_temp, err = astSendCMD(con,"astparam g coder",bs)
					if strings.Contains(s_temp,"not defined") {
						s_temp = "na"
					}
					tmp.FW_ver = s_temp
					fmt.Println("Wilson - uptime")
					s_temp, err = astSendCMD(con,"cat /proc/uptime",bs)
					tmp.Uptime = strings.Split(s_temp, " ")[0]
					fmt.Println("Wilson - Host_type")
					if tmp.Host_type=="t"{
					fmt.Println("Wilson - type tx")
					hostnameTmp := "ast3-gateway"+tmp.Device_id
					fmt.Println("Wilson - type tx1" + hostnameTmp)
				//	s_temp, _ = astSendCMD(con,"echo "+hostnameTmp+" > /etc/hostname",bs)
					fmt.Println("Wilson - type tx2")
					tmp.Host_name = hostnameTmp

				   }else{
				   	fmt.Println("Wilson - type rx")
					hostnameTmp := "ast3-client"+tmp.Device_mac
				//	s_temp, _ = astSendCMD(con,"echo "+hostnameTmp+" > /etc/hostname",bs)
					tmp.Host_name = hostnameTmp
						fmt.Println("Wilson - type rx2")
				    }
				   	fmt.Println("Wilson - tmp_MAC")
					ast_info_map[tmp_MAC] = &tmp
					dataChangeFlag =true
				  }
				}
			}
			con.Close()
		}

		ast_busy = false
		time.Sleep(time.Duration(2)*time.Second)
		if dataChangeFlag==true{
			fmt.Println("Update node list")
			device_info_save()
			dataChangeFlag =false
		}
		fmt.Println("END ")
	}
	defer conn.Close()
}

func ast_setup_info(mac string) (Ast_setup_info) {	
	var setup_tmp Ast_setup_info
	if ast_info_map[mac] == nil {
		setup_tmp.FW = "NOT_FIND"
		return setup_tmp
	}	
	con, err := aliasDialer.Dial("tcp",ast_info_map[mac].IP+":24")
	bs := make([]byte, 1024)	
	if err != nil {
		fmt.Println(err);
		//con.Close()
		setup_tmp.FW = "DIAL_ERR"
		return setup_tmp
	}
	if astLogin(con,bs) != nil {	
		con.Close()
		setup_tmp.FW = "LOGIN_ERR"
		return setup_tmp
	} 	
	
	s_tmp, _ := astSendCMD(con,"cat /etc/version",bs)	
	i := strings.LastIndexByte(s_tmp,'A')
	if i != -1 {
		setup_tmp.FW = s_tmp[i:]
	} else {
		setup_tmp.FW = s_tmp
	}
	s_tmp, _ = astSendCMD(con,"astparam g s0_baudrate",bs)	
	if strings.Contains(s_tmp,"not defined")	{
		setup_tmp.RS232 = "115200-8n1"
	}	else {
		setup_tmp.RS232 = s_tmp			
	}

	if ast_info_map[mac].Host_type == "r" { //rx only command
		setup_tmp.TX_NET_PROFILE = "x"
		setup_tmp.TX_SRC = "x"
		s_tmp, _ = astSendCMD(con,"astparam g display",bs)
		if s_tmp == "0" {
			setup_tmp.RX_TV = "off"
		} else {
			setup_tmp.RX_TV = "on"
		}
		s_tmp, _ = astSendCMD(con,"cat /sys/devices/platform/display/vw_rotate",bs)
		if strings.Contains(s_tmp,"Setting: Normal") {
			setup_tmp.RX_Rotate = "0"
		} else if strings.Contains(s_tmp,"Setting: Rotate 180") {
			setup_tmp.RX_Rotate = "3"
		} else {
			setup_tmp.RX_Rotate = "6"
		}
		s_tmp, _ = astSendCMD(con,"astparam g v_output_timing_convert",bs)
		if strings.Contains(s_tmp,"not defined")	{
			setup_tmp.RX_SCALE = "00000000"
		} else {
			setup_tmp.RX_SCALE = s_tmp
		}
		
	} else {   //tx only command
		setup_tmp.RX_TV = "x"
		setup_tmp.RX_SCALE = "x"
		setup_tmp.RX_Rotate = "x"
		s_tmp, _ = astSendCMD(con,"astparam g profile",bs)	
		if strings.Contains(s_tmp,"not defined")	{
			setup_tmp.TX_NET_PROFILE = "auto"
		}	else {
			setup_tmp.TX_NET_PROFILE = s_tmp			
		}
		s_tmp, _ = astSendCMD(con,"cat /sys/devices/platform/videoip/State",bs)
		if strings.Contains(s_tmp,"operating") || strings.Contains(s_tmp,"OPERATING")	{
			setup_tmp.TX_SRC = "on"
		}	else {
			setup_tmp.TX_SRC = "off"
		}	
	}		
	return setup_tmp
}

func ast_del_node(mac string) (string) {	
	
	event_device_del(mac)
	ast_del_node_preset(mac)
	delete(ast_info_map,mac)	
	device_info_save()
	device_info_load("keep")
	return "OK"
}

func device_info_save(){
  node_slice := make([]Ast_info, 0, len(ast_info_map))
  for _, tx := range ast_info_map {
      node_slice = append(node_slice, *tx)
  } 
	json_node_list ,_:= json.Marshal(node_slice)
	_ = ioutil.WriteFile("node_info.json",json_node_list, 0644)
}



func device_info_load(alive string){
	file, _ := ioutil.ReadFile("node_info.json")
	var node_slice []Ast_info
	_ = json.Unmarshal([]byte(file), &node_slice)
	for i := 0; i < len(node_slice); i++ {
		ast_info_map[node_slice[i].Device_mac] = &node_slice[i]
		if alive == "clear"{
		ast_info_map[node_slice[i].Device_mac].Alive = "n"
		ast_info_map[node_slice[i].Device_mac].videoStatus = "n"
		}
	}
}


func osd_info_save(){
	
	/*node_slice := make([]ast_osd_info, 0, 1)
	osd_tmp := ast_osd_info{"16","0","0","","16","0xFF000000","0","0","0",""}
	
	osd_tmp.type = false
	osd_tmp.alpha ="16"
	osd_tmp.mask ="0"
	osd_tmp.offtime ="0"
	osd_tmp.srt =""
	osd_tmp.size ="16"
	osd_tmp.color ="0xFF000000"
	osd_tmp.pos ="0"
	osd_tmp.x ="0"
	osd_tmp.y ="0"
	osd_tmp.pic =""
	node_slice = append(node_slice, osd_tmp)*/
	node_slice := make([]ast_osd_info, 0, len(ast_osd_map))
  for _, tx := range ast_osd_map {
      node_slice = append(node_slice, tx)
  } 
	json_node_list ,_:= json.Marshal(node_slice)
	_ = ioutil.WriteFile("osd_info.json",json_node_list, 0644)
}

func osd_info_load(){
	file, _ := ioutil.ReadFile("osd_info.json")
	var node_slice []ast_osd_info
	_ = json.Unmarshal([]byte(file), &node_slice)
	for i := 0; i < len(node_slice); i++ {
		ast_osd_map[i] = node_slice[i]
		//fmt.Println(ast_osd_map[i].OSD_string)
	}
}

func ast_lmparam_g(target_addr ,param string) (string) {
	// connect to this socket
	conn, err := aliasDialer.Dial("tcp",target_addr)
	if err != nil {
		return "DAIL_ERR"
	}
	
	bs := make([]byte, 1024)
	s := astSend(conn,"lmparam g "+param,bs)
	conn.Close()
	return s
}

func ast_set_vw(mac,vwh,vwv,vwp,vwl,vwr,vwu,vwb string) (string) {
	// connect to this socket
	var x1,y1,x2,y2 int
	
	int_tmp,_ := strconv.Atoi(vwh)	
	int_p,_ := strconv.Atoi(vwp)
	x2 = 10000/int_tmp
	x1 = (int_p-1) % int_tmp
	x1 = x1 * x2
	x2 = x2 * (((int_p - 1) % int_tmp) + 1 )

	y1 = (int_p-1) / int_tmp
	y2 = y1+1
	int_tmp,_ = strconv.Atoi(vwv)	
	int_tmp = 10000/int_tmp
	y1 = y1 * int_tmp
	y2 = y2 * int_tmp
	
	int_tmp,_ = strconv.Atoi(vwu)	
	y1 = y1 + (int_tmp*10)
	if y1 < 0  || y1 >= y2 {
		return "U_OUT"
	}
	int_tmp,_ = strconv.Atoi(vwb)	
	y2 = y2 + (int_tmp*10)
	if y2 > 10000  || y2 <= y1 {
		return "B_OUT"
	}
	int_tmp,_ = strconv.Atoi(vwl)	
	x1 = x1 + (int_tmp*10)
	if x1 < 0  || x1 >= x2 {
		return "L_OUT"
	}
	int_tmp,_ = strconv.Atoi(vwr)	
	x2 = x2 + (int_tmp*10)
	if x2  > 10000  || x2 <= x1 {
		return "R_OUT"
	}
	//fmt.Println(x1,y1,x2,y2)
	conn, err := aliasDialer.Dial("tcp",ast_info_map[mac].IP+":24")
	if err != nil {
		return "DAIL_ERR"
	}
	bs := make([]byte, 1024)
	if astLogin(conn,bs) != nil {	
		conn.Close()
		return "LOGIN_ERR"
	} 		
	_,err = astSendCMD(conn,"astparam s vw_h "+vwh,bs)
	if err != nil {
		conn.Close()
		return "ERR_FMT"
	}
	_,err = astSendCMD(conn,"astparam s vw_v "+vwv,bs)
	if err != nil {
		conn.Close()
		return "ERR_FMT"
	}
	_,err = astSendCMD(conn,"astparam s vw_p "+vwp,bs)
	if err != nil {
		conn.Close()
		return "ERR_FMT"
	} 
	_,err = astSendCMD(conn,"astparam s vw_l "+vwl,bs)
	if err != nil {
		conn.Close()
		return "ERR_FMT"
	} 
	_,err = astSendCMD(conn,"astparam s vw_r "+vwr,bs)
	if err != nil {
		conn.Close()
		return "ERR_FMT"
	} 
	_,err = astSendCMD(conn,"astparam s vw_u "+vwu,bs)
	if err != nil {
		conn.Close()
		return "ERR_FMT"
	} 
	_,err = astSendCMD(conn,"astparam s vw_b "+vwb,bs)
	if err != nil {
		conn.Close()
		return "ERR_FMT"
	} 
	_,err = astSendCMD(conn,"astparam save",bs)
	if err != nil {
		conn.Close()
		return "ERR_FMT"
	} 
	_,err = astSendCMD(conn,"e e_vw_enable_"+strconv.Itoa(x1)+"_"+strconv.Itoa(y1)+"_"+strconv.Itoa(x2)+"_"+strconv.Itoa(y2)+"_2",bs)
	//fmt.Println("e e_vw_enable_"+strconv.Itoa(x1)+"_"+strconv.Itoa(y1)+"_"+strconv.Itoa(x2)+"_"+strconv.Itoa(y2)+"_2")
	if err != nil {
		conn.Close()
		return "ERR_FMT"
	} 
	conn.Close()	
	
	return "OK"
}

func ast_set_ite_vw(mac, v_count, h_count, vw, ow, vh, oh, row, col string) (string) {

    conn, err := aliasDialer.Dial("tcp",ast_info_map[mac].IP+":24")

    if err != nil {
		return "DAIL_ERR"
	}
	bs := make([]byte, 1024)
	if astLogin(conn,bs) != nil {
		conn.Close()
		return "LOGIN_ERR"
	}
     fmt.Println("e e_vw_refresh_pos_idx_" + row + "_" + col)
	_,err = astSendCMD(conn,"e e_vw_refresh_pos_idx_" + row + "_" + col ,bs)
	if err != nil {
		conn.Close()
		return "ERR_FMT"
	}
    fmt.Println("e e_vw_moninfo_" + vw + "_" + ow + "_" + vh + "_" + oh)
	_,err = astSendCMD(conn,"e e_vw_moninfo_" + vw + "_" + ow + "_" + vh + "_" + oh ,bs)
    if err != nil {
    	conn.Close()
    	return "ERR_FMT"
    }
   fmt.Println("e e_vw_stretch_type_2")
    _,err = astSendCMD(conn,"e e_vw_stretch_type_2" ,bs)
    if err != nil {
        conn.Close()
        return "ERR_FMT"
    }
    fmt.Println("e e_vw_rotate_0")
    _,err = astSendCMD(conn,"e e_vw_rotate_0" ,bs)
    if err != nil {
       conn.Close()
       return "ERR_FMT"
    }
    fmt.Println("e e_vw_enable_" + v_count + "_" + h_count + "_" + row + "_" + col)
	_,err = astSendCMD(conn,"e e_vw_enable_" + v_count + "_" + h_count + "_" + row + "_" + col,bs)
	if err != nil {
		conn.Close()
		return "ERR_FMT"
	}
    fmt.Println("e e_vw_pos_layout_" + v_count + "_" + h_count)
	_,err = astSendCMD(conn,"e e_vw_pos_layout_" + v_count + "_" + h_count ,bs)
    if err != nil {
       conn.Close()
       return "ERR_FMT"
    }

	conn.Close()

	return "OK"
}

func ast_set_osd(index,name,alpha,mask,offtime,str,fsize,fcolor,pos,x,y,pic string, osdtype bool) (string) {
	// connect to this socket
	
	i, _ := strconv.Atoi(index)

	if i<=8{
	ast_osd_map[i-1].OSD_name = name
	ast_osd_map[i-1].OSD_type = osdtype
	ast_osd_map[i-1].OSD_alpha = alpha
	ast_osd_map[i-1].OSD_mask = mask
	ast_osd_map[i-1].OSD_offtime = offtime
	ast_osd_map[i-1].OSD_string = str
	ast_osd_map[i-1].OSD_fontSize = fsize
	ast_osd_map[i-1].OSD_fontColor = fcolor
	ast_osd_map[i-1].OSD_position = pos
	ast_osd_map[i-1].OSD_px = x
	ast_osd_map[i-1].OSD_py = y
	ast_osd_map[i-1].OSD_imgName = pic
	osd_info_save()
	
	fmt.Println(index,str);
	return "OK"
	}
	return "ERR_FMT"	
}

func ast_osd_apply(mac,index string) (string) {
	// connect to this socket
	
	i, _ := strconv.Atoi(index)

	if i>8{
	return "DAIL_ERR"
	}

	
	conn, err := aliasDialer.Dial("tcp",ast_info_map[mac].IP+":24")
	if err != nil {
		return "DAIL_ERR"
	}
	bs := make([]byte, 1024)
	if astLogin(conn,bs) != nil {	
		conn.Close()
		return "LOGIN_ERR"
	} 	
	//update osd index status
	ast_info_map[mac].OSD = index; 
	device_info_save()

	if ast_osd_map[i-1].OSD_type==true{	
		
		_,err = astSendCMD(conn,"osd_pic_off.sh 0",bs)
		if err != nil {
			conn.Close()
			return "ERR_FMT"
		}
		_,err = astSendCMD(conn,"e e_osd_on_str::0::240::"+ast_osd_map[i-1].OSD_alpha+"::"+ast_osd_map[i-1].OSD_mask+"::"+ast_osd_map[i-1].OSD_offtime+"::"+ast_osd_map[i-1].OSD_string+"::"+ast_osd_map[i-1].OSD_fontSize+"::"+ast_osd_map[i-1].OSD_fontColor,bs)
		if err != nil { 
			conn.Close()
			return "ERR_FMT"
		}

	}else{  //pic


      fmt.Println("11111111111 =  "+ "e e_osd_on_str::0::240::"+ast_osd_map[i-1].OSD_alpha+"::"+ast_osd_map[i-1].OSD_mask+"::"+"0"+"::"+""+"::"+ast_osd_map[i-1].OSD_fontSize+"::"+ast_osd_map[i-1].OSD_fontColor )


		_,err = astSendCMD(conn,"e e_osd_on_str::0::240::"+ast_osd_map[i-1].OSD_alpha+"::"+ast_osd_map[i-1].OSD_mask+"::"+"0"+"::"+""+"::"+ast_osd_map[i-1].OSD_fontSize+"::"+ast_osd_map[i-1].OSD_fontColor,bs)
		if err != nil { 
			conn.Close()
		    fmt.Println("Set image OSD error 1 ~~~~~~~~~~~~~~~~~~~~~")
			return "ERR_FMT"
		}
        fmt.Println("22222 =  "+ "tftp -g -r /osd/"+ ast_osd_map[i-1].OSD_imgName + " " + pi4Addr)
		_,err = astSendCMD(conn,"tftp -g -r /osd/"+ ast_osd_map[i-1].OSD_imgName + " " + pi4Addr ,bs)
        	if err != nil {
        	conn.Close()
        	fmt.Println("Set image OSD error 2 ~~~~~~~~~~~~~~~~~~~~~")
        	return "ERR_FMT"
        }
        fmt.Println("333333 =  "+ "mv "+ ast_osd_map[i-1].OSD_imgName + " share")
        _,err = astSendCMD(conn,"mv "+ ast_osd_map[i-1].OSD_imgName + " share",bs)
           if err != nil {
            conn.Close()
            fmt.Println("Set image OSD error 3 ~~~~~~~~~~~~~~~~~~~~~" , err)
            return "ERR_FMT"
        }
         fmt.Println("4444444 =  "+ "osd_pic_on.sh "+ast_osd_map[i-1].OSD_imgName+" "+ast_osd_map[i-1].OSD_offtime+" "+ast_osd_map[i-1].OSD_mask+" "+ast_osd_map[i-1].OSD_alpha)
		_,err = astSendCMD(conn,"osd_pic_on.sh "+ast_osd_map[i-1].OSD_imgName+" "+ast_osd_map[i-1].OSD_offtime+" "+ast_osd_map[i-1].OSD_mask+" "+ast_osd_map[i-1].OSD_alpha,bs)
		if err != nil {
			conn.Close()
			return "ERR_FMT"
		}
	}
	
	//position
	_,err = astSendCMD(conn,"osd_position.sh "+ast_osd_map[i-1].OSD_position+" "+ast_osd_map[i-1].OSD_px+" "+ast_osd_map[i-1].OSD_py,bs)
	if err != nil {
		conn.Close()
		return "ERR_FMT"
	}
	event := "Enable "+ ast_info_map[mac].Host_name + " OSD"
	add_event("OSD",event)

	conn.Close()		
	return "OK"
}

func ast_osd_close(mac,index string) (string) {
	// connect to this socket
	i, _ := strconv.Atoi(index)
	if i>8{
	return "DAIL_ERR"
	}

	ast_info_map[mac].OSD = "0"; 
	device_info_save()

	conn, err := aliasDialer.Dial("tcp",ast_info_map[mac].IP+":24")
	if err != nil {
		return "DAIL_ERR"
	}
	bs := make([]byte, 1024)
	if astLogin(conn,bs) != nil {	
		conn.Close()
		return "LOGIN_ERR"
	} 	
	if ast_osd_map[i-1].OSD_type==true{	
		
		_,err = astSendCMD(conn,"e e_osd_on_str::0::240::"+ast_osd_map[i-1].OSD_alpha+"::"+ast_osd_map[i-1].OSD_mask+"::"+"0"+"::"+""+"::"+ast_osd_map[i-1].OSD_fontSize+"::"+ast_osd_map[i-1].OSD_fontColor,bs)
		if err != nil { 
			conn.Close()
			return "ERR_FMT"
		}
	}else{  //pic
		_,err = astSendCMD(conn,"osd_pic_off.sh 0",bs)
		if err != nil {
			conn.Close()
			return "ERR_FMT"
		}
	}
	event := "Disable "+ ast_info_map[mac].Host_name + " OSD"
	add_event("OSD",event)
	
	conn.Close()		
	return "OK"
}


func ast_set_osd_s(mac,alpha,mask,offtime,str,fsize,fcolor string) (string) {
	// connect to this socket
	
	conn, err := aliasDialer.Dial("tcp",ast_info_map[mac].IP+":24")
	if err != nil {
		return "DAIL_ERR"
	}
	bs := make([]byte, 1024)
	if astLogin(conn,bs) != nil {	
		conn.Close()
		return "LOGIN_ERR"
	} 		
	_,err = astSendCMD(conn,"e e_osd_on_str::0::480::"+alpha+"::"+mask+"::"+offtime+"::"+str+"::"+fsize+"::"+fcolor,bs)
	if err != nil { 
		conn.Close()
		return "ERR_FMT"
	}
	conn.Close()		
	return "OK"
}

func ast_set_osd_pos(mac,pos,x,y string) (string) {
	// connect to this socket
	
	conn, err := aliasDialer.Dial("tcp",ast_info_map[mac].IP+":24")
	if err != nil {
		return "DAIL_ERR"
	}
	bs := make([]byte, 1024)
	if astLogin(conn,bs) != nil {	
		conn.Close()
		return "LOGIN_ERR"
	} 		
	_,err = astSendCMD(conn,"osd_position.sh "+pos+" "+x+" "+y,bs)
	if err != nil {
		conn.Close()
		return "ERR_FMT"
	}
	conn.Close()		
	return "OK"
}

func ast_set_osd_pic(mac,pic,offtime,mask,alpha string) (string) {
	// connect to this socket
	
	conn, err := aliasDialer.Dial("tcp",ast_info_map[mac].IP+":24")
	if err != nil {
		return "DAIL_ERR"
	}
	bs := make([]byte, 1024)
	if astLogin(conn,bs) != nil {	
		conn.Close()
		return "LOGIN_ERR"
	}
	_,err = astSendCMD(conn,"tftp "+pi4Addr+" -c get /osd/"+pic+" /share/"+pic,bs)
	if err != nil {
		conn.Close()
		return "ERR_FMT"
	}
	_,err = astSendCMD(conn,"osd_pic_on.sh "+pic+" "+offtime+" "+mask+" "+alpha,bs)
	if err != nil {
		conn.Close()
		return "ERR_FMT"
	}
	conn.Close()		
	return "OK"
}


func ast_vw_info(mac string) (Ast_vw_info) {
	// connect to this socket
	var vw_tmp Ast_vw_info
	conn, err := aliasDialer.Dial("tcp",ast_info_map[mac].IP+":24")
	if err != nil {
		vw_tmp.VWH = "DAIL_ERR"
		return vw_tmp
	}	
	bs := make([]byte, 1024)
	if astLogin(conn,bs) != nil {	
		conn.Close()
		vw_tmp.VWH = "LOGIN_ERR"
		return vw_tmp
	} 	
	
	s_tmp,_ := astSendCMD(conn,"astparam g vw_h",bs)	
	if strings.Contains(s_tmp,"not defined")	{
		vw_tmp.VWH = "1"
		vw_tmp.VWV = "1"
		vw_tmp.VWP = "1"
		vw_tmp.VWL = "0"
		vw_tmp.VWR = "0"
		vw_tmp.VWU = "0"
		vw_tmp.VWB = "0"
	} else {
		vw_tmp.VWH = s_tmp
		vw_tmp.VWV,_ = astSendCMD(conn,"astparam g vw_v",bs)
		vw_tmp.VWP,_ = astSendCMD(conn,"astparam g vw_p",bs)
		vw_tmp.VWL,_ = astSendCMD(conn,"astparam g vw_l",bs)
		vw_tmp.VWR,_ = astSendCMD(conn,"astparam g vw_r",bs)
		vw_tmp.VWU,_ = astSendCMD(conn,"astparam g vw_u",bs)
		vw_tmp.VWB,_ = astSendCMD(conn,"astparam g vw_b",bs)
	}	
	conn.Close()
	return vw_tmp
}



func ast_astparam_g(target_addr ,param string) (string) {
	// connect to this socket
	conn, err := aliasDialer.Dial("tcp",target_addr)
	if err != nil {
		return "DAIL_ERR"
	}	
	bs := make([]byte, 1024)
	s := astSend(conn,"astparam g "+param,bs)
	conn.Close()
	return s
}


func ast_sendcmd(target_addr ,value string) (string) {
	// connect to this socket
	conn, err := aliasDialer.Dial("tcp",target_addr)
	if err != nil {
		return "DAIL_ERR"
	}
	bs := make([]byte, 1024)
	s := astSend(conn,value,bs)
	conn.Close()
	

	//fmt.Println("mute",target_addr,value)

	event := "none"
	
	if s != "ERR_FMT" {	
	
	var dev string
	for _, map_data := range ast_info_map {
			tmpIP := map_data.IP+":24"	
			if  tmpIP== target_addr	{
						dev = map_data.Host_name
					break
				}
			}
		//mute on/off
		if strings.Contains(value,"screen_off"){
		if strings.Contains(value,"0"){
		event = " mute off"	
		}else{
		event = " mute on"
		}
			add_event("Mapping",dev + event)
			fmt.Println(dev + event)
		}
		//video rotate
		if strings.Contains(value,"vw_rotate"){
		if strings.Contains(value,"3"){
		event = " video rotate 180"	
		}else if strings.Contains(value,"6"){
		event = " video rotate 270"
		}else{
		event = " video rotate 0"
		}
			add_event("Setup",dev + event)
			fmt.Println(dev + event)
		}
		
		if strings.Contains(value,"CM_MUTELOOPOUTh"){
		if value == "echo 0 > /sys/devices/platform/ast1500_led.2/leds:button_link/ CM_MUTELOOPOUTh"{
			add_event("Setup",dev + " loop-out off")
			//fmt.Println(dev + " loop off")		
		}else{
			add_event("Setup",dev + " loop-out on")
			//fmt.Println(dev + " loop on")
		}
		}

		if strings.Contains(value,"reboot"){
			add_event("Setup",dev + " reboot")
		}
		if strings.Contains(value,"reset_to_default"){
			add_event("Setup",dev + " factory reset")
		}
		
		return "OK"
	}
	return s
}

func ast_sendcmd_g(target_addr ,value string) (string) {
	// connect to this socket
	conn, err := aliasDialer.Dial("tcp",target_addr)
	if err != nil {
		return "DAIL_ERR"
	}
	bs := make([]byte, 1024)
	s := astSend(conn,value,bs)
	conn.Close()	
	return s
}

func ast_astparam_s(target_addr ,parameter ,value string) (string) {
	// connect to this socket
	conn, err := aliasDialer.Dial("tcp",target_addr)
	if err != nil {
		return "DAIL_ERR"
	}
	bs := make([]byte, 1024)
	s := astSend(conn,"astparam s "+parameter + " " + value,bs)
	if s != "ERR_FMT" {
		conn.Close()	
		ast_sendcmd(target_addr,"astparam save")
		
		var dev string
		for _, map_data := range ast_info_map {
				tmpIP := map_data.IP+":24"	
				if  tmpIP== target_addr	{
						dev = map_data.Host_name
						break
					}
				}
		
		if strings.Contains(parameter,"s0_baudrate"){
			event := dev + " baud rate to " +value
			add_event("Setup",event)
			fmt.Println(event)
		
		}else if strings.Contains(parameter,"v_output_timing_convert"){
			var res string
			if value == "80000010"{
			res = "Full HD 1080p60"
			}else if value == "8000001F"{
			res = "Full HD 1080p50"
			}else if value == "8000005F"{
			res = "Ultra HD 2160p30"
			}else if value == "8000005E"{
			res = "Ultra HD 2160p25"
			}else if value == "80000004"{
			res = "HD 720p60"
			}else{
			res = "Pass-Through"
			}
			event := dev + " scaling to " +res
			add_event("Setup",event)
			fmt.Println(event)
			
		}else if strings.Contains(parameter,"profile"){
			event := dev + " speed " + value
			add_event("Setup",event)
			fmt.Println(event)
		
		}else{
		//fmt.Println(event)
		}
		
		
		return "OK"
	}
	conn.Close()	
	return s
}

func ast_set_name(mac ,value string) (string) {
	// connect to this socket
	if ast_info_map[mac] == nil {
		return "NOT_FIND"
	}
	if ast_info_map[mac].Alive == "n"	{
		return "NOT_FIND"
	}
	//s := ast_sendcmd(ast_info_map[mac].IP+":24","echo "+value+" > /etc/hostname")
	//s := ast_sendcmd(ast_info_map[mac].IP+":24","astparam s hostname_id "+value)
	//s2 := ast_sendcmd(ast_info_map[mac].IP+":24","e e_chg_hostname")
	//s3 := ast_sendcmd(ast_info_map[mac].IP+":24","astparam save")
	//if s != "ERR_FMT" && s2 != "ERR_FMT" && s3 != "ERR_FMT"{
		
		event := mac + " rename from "+ast_info_map[mac].Host_name + " to "+value
		add_event("Setup",event)
		ast_info_map[mac].RHost_name = value
		ast_info_map[mac].Host_name = value
		device_info_save()
	//}
	return "OK"
}


func astswitch(target_addr ,switch_id ,switch_type string) (string) {

	switchMode := false //false : switch tx ; true : switch rx
	var id_temp string
	dev_ip := "" 
	dev_id := ""
	bs := make([]byte, 1024)
	// connect to this socket
	conn, err := aliasDialer.Dial("tcp",target_addr)
	if err != nil {
		return "DAIL_ERR"
	}
	tmpIP := strings.Replace(target_addr, ":24", "",-1)
	for _, map_data := range ast_info_map {
		if map_data.IP == tmpIP	{
			map_data.Device_id = switch_id
			if map_data.Host_type=="t"{				 
			switchMode = false
			}else{
			switchMode = true
			}
			dev_ip = map_data.Host_name
			break
		}
	}  
	successFlag := false
	for i := 0; i < 5; i++ {	
		_ = astSend(conn,"e e_reconnect::"+switch_id + "::" + switch_type,bs)
		time.Sleep(time.Duration(30)*time.Millisecond)

		for j := 0; j < 3; j++ {
			if switchMode == false {
			id_temp, _ = astSendCMD(conn,"astparam g ch_select",bs)
			}else{
			id_temp, _ = astSendCMD(conn,"astparam g ch_select_v",bs)
			}	
			if id_temp != "" && strings.Count(id_temp,"") == 5 {
			   //check ID is number
			   if _, err := strconv.Atoi(id_temp); err == nil {
					fmt.Println("ID:"+id_temp)
					if switch_id == id_temp{
					   successFlag = true
					}
					break
				}else{		
					fmt.Println("ID not a number\n")
					continue
				}
			}
		}
		if successFlag ==true{
		   device_info_save()
		   break;
		}
	}
	conn.Close()

	if switchMode==true{
		for _, map_data := range ast_info_map {
			if map_data.Device_id == switch_id	{
				if map_data.Host_type == "t"{
				dev_id = map_data.Host_name					
				break
				}
			}
		}
	}
	if switchMode==false{
	add_event("Setup",dev_ip + " switch group ID to "+ switch_id)
	fmt.Println("Assign "+dev_ip+" switch group ID to "+switch_id)
	}else{
	add_event("Mapping","Assign "+dev_ip+" to "+dev_id)
	fmt.Println("Assign "+dev_ip+" to "+dev_id)
	}
	
	return "OK"

}

func astlinkoff(target_addr ,switch_type string) (string) {
	// connect to this socket
	conn, err := aliasDialer.Dial("tcp",target_addr)
	if err != nil {
		return "DAIL_ERR"
	}	
	bs := make([]byte, 1024)
	s := astSend(conn,"e e_stop_link::" + switch_type,bs)
	conn.Close()
	if s != "ERR_FMT" {
		add_event("Mapping","link off "+target_addr)
		return "OK"
	}
	return s
}

func http_post_debug(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r) 
	resp, err := http.Post("http://"+pi4Addr+":8080/api/set_vw", "application/x-www-form-urlencoded", strings.NewReader("mac="+vars["mac"]+"&vwh=2&vwv=2&vwp=3&vwl=10&vwr=-10&vwu=10&vwb=-10"))
	if err != nil {
	    log.Fatalln(err)
	}	
	defer resp.Body.Close()
	body,_ := ioutil.ReadAll(resp.Body)
	fmt.Println(string(body))
}



func pi4_savePng(name, b64 string) (string){

	b64_ := strings.Replace(b64, " ", "+",-1)
	unbased, err := base64.StdEncoding.DecodeString(b64_)
	if err != nil {
    	panic("Cannot decode b64")
		panic(b64)
		return "DAIL_ERR"
	}
	
	r := bytes.NewReader(unbased)
	im, err := png.Decode(r)
	if err != nil {
    	panic("Bad png")
		return "DAIL_ERR"
	}
	
	path := "./www/snapshot/osd/"+name+".png"
	//"./www/snapshot/osd/start.png"
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0777)
	if err != nil {
    	panic("Cannot open file")
		return "DAIL_ERR"
	}
	png.Encode(f, im)
	return "OK"


}

func pi4_saveJpg(name, b64 string) (string){

	b64_ := strings.Replace(b64, " ", "+",-1)
	//fmt.Println(b64_)
	unbased, err := base64.StdEncoding.DecodeString(b64_)
	if err != nil {
    	panic("Cannot decode b64")
		return "DAIL_ERR"
	}
	
	r := bytes.NewReader(unbased)
	im, err := jpeg.Decode(r)
	if err != nil {
    	panic("Bad jpeg")
		return "DAIL_ERR"
	}
	
	path := "./www/snapshot/osd/"+name+".jpg"
	//"./www/snapshot/osd/start.jpg"
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0777)
	if err != nil {
    	panic("Cannot open file")
		return "DAIL_ERR"
	}
	var opt jpeg.Options
	opt.Quality = 99

	jpeg.Encode(f, im, &opt)
	return "OK"


}
