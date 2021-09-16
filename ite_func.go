package main
import (
	"fmt"
    "log"
    "net"
    "bytes"
	"encoding/hex"
    "unsafe"
    "reflect"
    "time"    
    "strings"
    "strconv"
    "encoding/json"
    "io/ioutil"
    //"github.com/gorilla/mux"
	
)

const RX_DEVICE_TYPE byte = 0
const TX_DEVICE_TYPE byte = 1 

//#### ITE set device command ####

const(
CMD_SET_DHCP_MODE = 1 
ACK_SET_DHCP_MODE = 2
CMD_SET_IP_ADDRESS = 3 
ACK_SET_IP_ADDRESS = 4
CMD_SET_GROUPID = 5 
ACK_SET_GROUPID = 6
CMD_SET_UART_BAUDRATE = 7 
ACK_SET_UART_BAUDRATE = 8
CMD_SET_MACADDRESS = 9 
ACK_SET_MACADDRESS = 10
CMD_SET_DEVICE_NAME = 29 
ACK_SET_DEVICE_NAME = 30
CMD_SET_FACTORY_RESET = 35 
ACK_SET_FACTORY_RESET = 36
CMD_SET_STREAMING_MODE = 47 
ACK_SET_STREAMING_MODE = 48
CMD_SET_SCREEN_MODE = 257 //RX
ACK_SET_SCREEN_MODE = 258
CMD_SET_GROUP_VIDEO_BITRATE = 257 //TX
ACK_SET_GROUP_VIDEO_BITRATE = 258
CMD_SET_VIDEO_OUT_MODE = 261 //TX
ACK_SET_VIDEO_OUT_MODE = 262
CMD_SET_DOWN_SCALE_MODE = 263 
ACK_SET_DOWN_SCALE_MODE = 264
CMD_SET_OSD_MODE = 265 
ACK_SET_OSD_MODE = 266
CMD_SET_LOW_DELAY_MODE = 273 
ACK_SET_LOW_DELAY_MODE = 274
CMD_SET_MULTICAST_MODE = 277 
ACK_SET_MULTICAST_MODE = 278
CMD_SET_REBOOT = 240 
ACK_SET_REBOOT = 241
CMD_SET_MCU_COMMAND  = 498
ACK_SET_MCU_COMMAND  = 499


)



//#### ITE get device command ####
const CMD_GET_SOURCE_HDCP int = 15  //RX/TX
const ACK_GET_SOURCE_HDCP int = 16
const CMD_GET_VIDEO_LOCK int = 17  //RX/TX
const ACK_GET_VIDEO_LOCK int = 18
const CMD_GET_GROUP_ID int = 19  //RX/TX
const ACK_GET_GROUP_ID int = 20
const CMD_GET_DHCP_MODE int = 21  //RX/TX
const ACK_GET_DHCP_MODE int = 22
const CMD_GET_UART_BAUDRATE int = 23  //RX/TX
const ACK_GET_UART_BAUDRATE int = 24
const CMD_GET_STATICIP_CONFIG int = 25  //RX/TX
const ACK_GET_STATICIP_CONFIG int = 26
const CMD_GET_MAC_ADDRESS int = 27   //RX/TX
const ACK_GET_MAC_ADDRESS int = 28
const CMD_GET_IP_CONFIG int = 41  //RX/TX
const ACK_GET_IP_CONFIG int = 42
const CMD_GET_TX_FW_VERSION int = 271  //TX
const ACK_GET_TX_FW_VERSION int = 272
const CMD_GET_OSD_MODE int = 267  //
const ACK_GET_OSD_MODE int = 268
const CMD_GET_INFO_DISPLAY_MODE int = 267  //RX
const ACK_GET_INFO_DISPLAY_MODE int = 268
const CMD_GET_RX_FW_VERSION int = 269  //RX
const ACK_GET_RX_FW_VERSION int = 270
const CMD_GET_DOWN_SCALE_MODE int = 269  //TX
const ACK_GET_DOWN_SCALE_MODE int = 270
const CMD_GET_SCREEN_MODE int = 261  //RX
const ACK_GET_SCREEN_MODE int = 262
const CMD_GET_SOURCE_IP int = 263  //RX
const ACK_GET_SOURCE_IP int = 264
const CMD_GET_MULTICAST_MODE int = 279  //RX
const ACK_GET_MULTICAST_MODE int = 280
const CMD_GET_LOW_DELAY_MODE int = 275  //TX
const ACK_GET_LOW_DELAY_MODE int = 276
const CMD_GET_MCU_COMMAND int = 500  //TR
const ACK_GET_MCU_COMMAND int = 501


type IPTV_COMMAND_SET struct
{
    commandId int
    commandBuffer []byte
    commandSize int 
	busyFlag bool
    //unsigned char   commandResult;
} 

type REPLY_INFO struct {

	trx bool //true: t false:r
	ack bool
	mac string
	osd_mode byte
	ip string
	netmask string
	gateway string
	ip_s string      //static
	netmask_s string
	gateway_s string
	baudrate int
	ipMode bool
	hdcp bool
	videoStatus bool
	fw_version string 
	screenMode bool     
	encoder_version string
	lowDelay bool
	FHD_DSM byte 
	HD_DSM byte 
	MulticastMode bool
	SourceIP string
}

type reply_ite struct {
	cmdHeader [8] byte
	deviceIp  [4] byte
	devicePort [2]byte
	cmdSync byte
	cmdId [2] byte
	dataLength [2]byte
	cmdChecksum byte
	deviceName[32] byte
	deviceIp2  [4] byte
	devicePort2 [2]byte
	deviceId [2]byte
	deviceType byte
	deviceState byte
	deviceStreamType byte
	deviceStreamingMode byte	 
}

type Ite_node struct {	
	IP string `json:"ip"`
	Is_host string `json:"is_host"`
	Status string  `json:"status"`
	Name string  `json:"name"`
	Device_id string  `json:"id"`
	StreamingMode bool  `json:"dsm"`
	
}

type Ite_info struct {	

	Name string  `json:"name"`
	Host_type string `json:"type"`	
	Device_mac string  `json:"mac"`
	Device_id string  `json:"group_id"`
	IP string `json:"ip"`
	Netmask string `json:"mask"`
	Gateway string `json:"gateway"`
	IpMode bool `json:"dhcp_status"`
	Alive string `json:"lan_status"`
	HDCP_status bool `json:"hdcp_status"`
	StreamingMode bool `json:"streaming_mode"`
	UartBaudrate int `json:"uart_baudrate"`
	FW_ver string  `json:"fw_version"`	
	EncoderVersion string  `json:"encoder_version"`	
	VideoStatus bool `json:"video_lock_status"`
	MulticastMode bool `json:"multicast_mode"`
	LowDelayMode bool `json:"low_delay_mode"`
	VideoOutMode byte `json:"video_out_mode"`
	OSD_mode byte `json:"osd_mode"`
	ScreenMode bool `json:"screen_mode"`
	SourceIP string `json:"source_ip"`
	FHD_DSM byte `json:"fhd_down_scale_mode"`
	HD_DSM byte `json:"hd_down_scale_mode"`
	aliveCnt int  `json:"aliveCount"`     
}

var ite_node_array[64] Ite_node
var ite_node_num int
var ite_info_map =make(map[string]*Ite_info)  

//ITE
var tCommandSet IPTV_COMMAND_SET  
var ri REPLY_INFO

func ite_node_list() {
	/////////////send query command
	go tcp_server()
	go ite_query_list()
	/////////////server and receiver
	addr, err := net.ResolveUDPAddr("udp4","225.1.0.0:9003")//
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
		
		myreply := (*reply_ite)(unsafe.Pointer((*reflect.SliceHeader)(unsafe.Pointer(&buffer)).Data))
		
		if(myreply.deviceType == RX_DEVICE_TYPE){
			ite_node_array[ite_node_num].Is_host = "r"
		}else{
			ite_node_array[ite_node_num].Is_host = "t"
		}
		if myreply.deviceStreamingMode == 0{
		   ite_node_array[ite_node_num].StreamingMode = false //Unicast
		}else{
		   ite_node_array[ite_node_num].StreamingMode = true //Multicast
		}
		

		ite_node_array[ite_node_num].IP = src.IP.String()
		ss = myreply.deviceName[:]
		ite_node_array[ite_node_num].Name = string(ss[:bytes.Index(ss,[]byte("\x00"))])

		var deviceID int = ((int)(myreply.deviceId[0])<<8) + (int)(myreply.deviceId[1])
		ite_node_array[ite_node_num].Device_id = strconv.Itoa(deviceID) 
		//fmt.Println("found device",ite_node_array[ite_node_num].Device_id,ite_node_array[ite_node_num].Is_host,ite_node_array[ite_node_num].IP);
		ite_node_num++
	}	
	conn.Close()
}

func parseResponse(data []byte){

 var CmdID int = ((int)(data[15])<<8) + (int)(data[16])

 setCmdFlag := false

 //fmt.Println(data)

 switch CmdID{
	case ACK_GET_MAC_ADDRESS:
	ri.mac = strings.ToUpper(hex.EncodeToString(data[21:27]))
	ri.mac = strings.Replace(ri.mac, ":", "",-1)

	case ACK_GET_INFO_DISPLAY_MODE://0: auto 1: on 2:off
	ri.osd_mode = data[21] //or 20 not check
	case ACK_GET_IP_CONFIG:
	ri.ip = net.IP(data[20:24]).String()
	ri.netmask = net.IP(data[24:28]).String()
	ri.gateway = net.IP(data[28:32]).String()
	case ACK_GET_STATICIP_CONFIG:
	if data[20]==0x00{
	ri.ip_s = net.IP(data[21:25]).String()
	ri.netmask_s = net.IP(data[25:29]).String()
	ri.gateway_s = net.IP(data[29:33]).String()
	fmt.Println(ri.ip_s,ri.netmask_s,ri.gateway_s)
	}
	case ACK_GET_UART_BAUDRATE:
	if data[20]==0x00{
	ri.baudrate = ((int)(data[21])<<24) + ((int)(data[22])<<16) + ((int)(data[23])<<8) + (int)(data[24])
	//fmt.Println("Baud rate",ri.baudrate)
	}
	case ACK_GET_DHCP_MODE:
	if data[20]==0x00{
	if data[21]==0x01{
	ri.ipMode = true
	}else{
	ri.ipMode = false
	}
	}
	case ACK_GET_SOURCE_HDCP:
	if data[20]==0x00{
	if data[21]==0x01{
	ri.hdcp = true
	}else{
	ri.hdcp = false
	}
	}
	case ACK_GET_VIDEO_LOCK:
	if data[20]==0x00{
	if data[21]==0x01{
	ri.videoStatus = true
	}else{
	ri.videoStatus = false
	}
	}
	case ACK_GET_TX_FW_VERSION:
		ri.fw_version = string(data[20:(bytes.Index(data[20:],[]byte("\x00"))+20)])
		ri.encoder_version = string(data[52:(bytes.Index(data[52:],[]byte("\x00"))+52)])
	case ACK_GET_RX_FW_VERSION: //ACK_GET_DOWN_SCALE_MODE
		if ri.trx==false{
			ri.fw_version = string(data[20:(bytes.Index(data[20:],[]byte("\x00"))+20)])
			ri.encoder_version = string(data[52:(bytes.Index(data[52:],[]byte("\x00"))+52)])
		}else{
		    if data[20]==0x00{
				ri.FHD_DSM = data[21] 
				ri.HD_DSM = data[22]
				
				}
		}
	//fmt.Println(ri.fw_version,ri.encoder_version)
	case ACK_GET_SCREEN_MODE: 
	if ri.trx==false{
		if data[20]==0x00{
			if data[21]==0x01{
				ri.screenMode = true
				}else{
				ri.screenMode = false
			}
		}
	}else{ //case ACK_SET_VIDEO_OUT_MODE:
		setCmdFlag = true
	}
	case ACK_GET_LOW_DELAY_MODE:	
	if data[21]==0x01{
		ri.lowDelay = true
	}else{
	    ri.lowDelay =false
	}
	case ACK_GET_MULTICAST_MODE:
	if data[20]==0x01{
		ri.MulticastMode = true
	}else{
	    ri.MulticastMode =false
	}
	case ACK_GET_MCU_COMMAND:





	case ACK_GET_SOURCE_IP:

	if ri.trx==false{
	ri.SourceIP = net.IP(data[20:24]).String()
	//fmt.Println(ri.SourceIP)
	}else{//case ACK_SET_DOWN_SCALE_MODE:
	setCmdFlag = true
	}		
	//############ setting
	case ACK_SET_DHCP_MODE:
		setCmdFlag = true
	case ACK_SET_IP_ADDRESS:
		setCmdFlag = true
	case ACK_SET_GROUPID:
		setCmdFlag = true
	case ACK_SET_UART_BAUDRATE:
		setCmdFlag = true
	case ACK_SET_MACADDRESS:
		setCmdFlag = true
	case ACK_SET_DEVICE_NAME:
		setCmdFlag = true
	case ACK_SET_STREAMING_MODE:
		setCmdFlag = true
	case ACK_SET_SCREEN_MODE://ACK_SET_GROUP_VIDEO_BITRATE
		setCmdFlag = true
	case ACK_SET_OSD_MODE:
		setCmdFlag = true
	case ACK_SET_LOW_DELAY_MODE:
		setCmdFlag = true
	case ACK_SET_MULTICAST_MODE:
		setCmdFlag = true
	case ACK_SET_REBOOT:
		setCmdFlag = true
	case ACK_SET_MCU_COMMAND:
		fmt.Println("mcu command")
		setCmdFlag = true

	default:
	break
	} 
	//for setting type command
	if setCmdFlag == true {
		if data[20]==0x00{
			ri.ack = true
			fmt.Println("success")
		}else{
			ri.ack = false
			fmt.Println("fail")
		}
	}
}

func tcp_server(){

bs := make([]byte, 128)
li, err := net.Listen("tcp", ":9001")
if err != nil {
		log.Fatalln(err)
	}
defer li.Close()

for{
	conn, err := li.Accept()
	
	if err != nil {
	log.Println(err)
	continue
	}
	for i := 0; i < 15; i++ {
	time.Sleep(time.Duration(20)*time.Millisecond)
	len, _ := conn.Read(bs)
	if len>0 {
	 break
    	}
	}
	parseResponse(bs)
	tCommandSet.busyFlag = false
	//fmt.Fprintln(conn, "bye")
	conn.Close()
	}
	
}

func ite_query_list(){

	
	dataChangeFlag := false
	addr, err := net.ResolveUDPAddr("udp4","255.255.255.255:9002")
	if err != nil {
		log.Fatal(err)
	}
	
	conn, err := net.DialUDP("udp4",ite_aliasUDPAddr,addr)  //node_list alias IP
	if err != nil {
		log.Fatal(err)
	}
	query := make([]byte, 31)
	query = []byte {0x49,0x50,0x54,0x56,0x5f,0x43,0x4d,0x44,_pi4Addr_ite[0],_pi4Addr_ite[1],_pi4Addr_ite[2],_pi4Addr_ite[3],0x23,0x2b,0x74,0x00,0xfe,0x00,0x0b,0x09,0x00,0x00,0x00,0x00,0x00,0x00,0x00,0x00,0x00,0x00,0x00}

	defer func() {
		if err := recover(); err != nil {
			fmt.Println("tcp error",err)
		}
	}()

	for {
		conn.Write(query)	
		ite_node_num = 0; //reset list
		time.Sleep(time.Duration(5)*time.Second)
		// check who is offline
		for _, map_data := range ite_info_map {
			find_node := false
			for i := 0; i<ite_node_num; i++	{
				if map_data.IP == ite_node_array[i].IP	{
					find_node = true
					if map_data.Alive == "n" {

						map_data.IP = ite_node_array[i].IP
						map_data.Host_type = ite_node_array[i].Is_host
						map_data.Name = ite_node_array[i].Name	
						//recover group id
						fmt.Println(map_data.Device_mac, "online")
						//fmt.Println("logID:",map_data.Device_id,"iteID:", ite_node_array[i].Device_id)

						if  map_data.Device_id!=ite_node_array[i].Device_id{
						gid, _ := strconv.Atoi(map_data.Device_id)
						_=ite_Set_Group_ID_By_IP(map_data.IP,gid-1)
						_=ite_Set_Group_ID_By_IP(map_data.IP,gid)
						//fmt.Println(map_data.IP," - recover group id - ",map_data.Device_id)
						}

						map_data.StreamingMode = ite_node_array[i].StreamingMode 
						map_data.Alive = "y"
						map_data.aliveCnt = 0
						//get
						ite_getInfo(map_data,map_data.IP)
						add_event("Mapping",map_data.Name+" is online")
						
					}
					break
				}
			}
		    if find_node  == false && map_data.Alive == "y" {
				if map_data.aliveCnt>1{
				check_connection_lost(map_data.Device_mac)
				add_event("Mapping",map_data.Name+" is offline")
				map_data.Alive = "n"
				}else{
				map_data.aliveCnt = map_data.aliveCnt +1
				fmt.Println(map_data.Name + " loss - ",map_data.aliveCnt);
				}
			}
	   }
	   for i := 0; i<ite_node_num; i++ {
		if ite_SendCMD(ite_node_array[i].IP,CMD_GET_MAC_ADDRESS,[]byte{0})==false{
			
			for _, map_data := range ite_info_map {
				if map_data.IP == ite_node_array[i].IP	{
					map_data.Alive = "n"
					fmt.Println("device:",map_data.Device_mac,map_data.IP,"loss")
					break
				 }
			  }
			
			continue			
			}
		ite_waitForDone()
		//fmt.Println(ri.mac,ite_node_array[i].IP)					
		if map_tmp, ok := ite_info_map[ri.mac]; ok {


		if map_tmp.IP != ite_node_array[i].IP{
			map_tmp.IP = ite_node_array[i].IP
			fmt.Println("IP Update")
			dataChangeFlag =true
		}
		if map_tmp.Name != ite_node_array[i].Name{
			map_tmp.Name = ite_node_array[i].Name
			fmt.Println("device name Update")
			dataChangeFlag =true
		}
		if map_tmp.Device_id != ite_node_array[i].Device_id{
			map_tmp.Device_id = ite_node_array[i].Device_id
			fmt.Println("device id Update")
			dataChangeFlag =true
		}

	    //get tx or rx status
		if ite_node_array[i].Is_host == "t"{
		    
			ri.trx =true
			

		}else{
			ri.trx =false
			_= ite_SendCMD(map_tmp.IP,CMD_GET_SOURCE_IP,[]byte{0})
			ite_waitForDone()
			map_tmp.SourceIP = ri.SourceIP  

		}		
		_= ite_SendCMD(map_tmp.IP,CMD_GET_VIDEO_LOCK,[]byte{0})
		ite_waitForDone()
		map_tmp.VideoStatus = ri.videoStatus
		if map_tmp.VideoStatus == true{
		_= ite_SendCMD(map_tmp.IP,CMD_GET_SOURCE_HDCP,[]byte{0})
		   ite_waitForDone()
		map_tmp.HDCP_status = ri.hdcp
		}

		}else{//new device
		
		 tmp := Ite_info{}
		 tmp.IP = ite_node_array[i].IP
		 tmp.Host_type = ite_node_array[i].Is_host
		 tmp.Name = ite_node_array[i].Name	
		 tmp.Device_mac = ri.mac
		 tmp.Device_id = ite_node_array[i].Device_id
		 tmp.Alive = "y"
		 tmp.StreamingMode = ite_node_array[i].StreamingMode 
		 //tmp.OSD_mode = 0;
		  //get
		 ite_getInfo(&tmp,tmp.IP)
		 ite_info_map[tmp.Device_mac] = &tmp
		 fmt.Println("new",tmp.Device_mac,tmp.Host_type);
		 dataChangeFlag = true
		}
	   }
	   time.Sleep(time.Duration(2)*time.Second)
	   if dataChangeFlag==true{
		fmt.Println("Update node list")
		ite_device_info_save()
		dataChangeFlag =false
		}
	}
	defer conn.Close()
	//fmt.Println("end");
	
}

func ite_getInfo(info *Ite_info,ip string){

	
	_= ite_SendCMD(ip,CMD_GET_UART_BAUDRATE,[]byte{0})
	ite_waitForDone()	
	info.UartBaudrate = ri.baudrate
	_= ite_SendCMD(ip,CMD_GET_DHCP_MODE,[]byte{0})
	ite_waitForDone()	
	info.IpMode = ri.ipMode
	//fmt.Println("IP MODE:",ri.ipMode)

	_= ite_SendCMD(ip,CMD_GET_IP_CONFIG,[]byte{0})
	ite_waitForDone()	
	info.IP = ri.ip
	info.Netmask = ri.netmask
	info.Gateway = ri.gateway
	//fmt.Println("IP config:",ri.ip,ri.netmask,ri.gateway)
	_= ite_SendCMD(ip,CMD_GET_MULTICAST_MODE,[]byte{0})
	ite_waitForDone()
	info.MulticastMode = ri.MulticastMode
	if info.Host_type == "t"{
	ri.trx = true
	_= ite_SendCMD(ip,CMD_GET_TX_FW_VERSION,[]byte{0})
	ite_waitForDone()
	info.FW_ver = ri.fw_version
	info.EncoderVersion = ri.encoder_version
	_= ite_SendCMD(ip,CMD_GET_LOW_DELAY_MODE,[]byte{0})	
	ite_waitForDone()
	info.LowDelayMode = ri.lowDelay
	_= ite_SendCMD(ip,CMD_GET_DOWN_SCALE_MODE,[]byte{0}) 
	ite_waitForDone()
	info.FHD_DSM = ri.FHD_DSM
	info.HD_DSM = ri.HD_DSM

	}else{
	ri.trx = false
	_= ite_SendCMD(ip,CMD_GET_RX_FW_VERSION,[]byte{0})
	ite_waitForDone()
	info.FW_ver = ri.fw_version
	info.EncoderVersion = ri.encoder_version
	_= ite_SendCMD(ip,CMD_GET_SCREEN_MODE,[]byte{0})
	ite_waitForDone()
	info.ScreenMode = ri.screenMode
	_= ite_SendCMD(ip,CMD_GET_INFO_DISPLAY_MODE,[]byte{0})
	ite_waitForDone()
	info.OSD_mode = ri.osd_mode
	//fmt.Println("Screen mode:",info.ScreenMode)
	//fmt.Println("OSD mode:",info.OSD_mode)
	}	
	//fmt.Println(info.UartBaudrate)
	//fmt.Println(info.FW_ver,info.EncoderVersion )
	
}

func ite_waitForDone(){

for i := 0; i < 300; i++ {
	time.Sleep(time.Duration(10)*time.Millisecond)
	if tCommandSet.busyFlag ==false{
	break
	}
  }
}

func ite_SendCMD(trx_ip string, cmdId int,param []byte)  (bool){
	
	ite_commandCtrl(cmdId,param)
	ipaddress_port := trx_ip+":9001" 
	con, err := ite_aliasDialer.Dial("tcp",ipaddress_port)
	if err != nil {
		fmt.Println(err);
		return false;
	}
	tCommandSet.busyFlag = true
	con.Write(tCommandSet.commandBuffer[:tCommandSet.commandSize])
	time.Sleep(time.Duration(20)*time.Millisecond)
	con.Close()
	return true
}

func ite_commandCtrl(cmdId int,param []byte){
//hearder + ip + port + sync + cmdID + cmdSize + checksum + data + checksum
length := 8+4+2+1+2+2+1+len(param)+1 // 21+ N
dataLength := len(param)+1 //(N+1)
var packetChecksum byte = 0
tCommandSet.commandBuffer = make([]byte, length)
_ = copy(tCommandSet.commandBuffer,[]byte{0x49,0x50,0x54,0x56,0x5f,0x43,0x4d,0x44,_pi4Addr_ite[0],_pi4Addr_ite[1],_pi4Addr_ite[2],_pi4Addr_ite[3],0x23,0x29,0x74})//,0x01,0x13,0x00,0x02,0x16	

tCommandSet.commandBuffer[15] = (byte)((cmdId>>8)&0xFF)
tCommandSet.commandBuffer[16] = (byte)(cmdId&0xFF)
tCommandSet.commandBuffer[17] = (byte)((dataLength>>8)&0xFF)
tCommandSet.commandBuffer[18] = (byte)(dataLength&0xFF)
packetChecksum = tCommandSet.commandBuffer[15]+tCommandSet.commandBuffer[16]+tCommandSet.commandBuffer[17]+tCommandSet.commandBuffer[18]
tCommandSet.commandBuffer[19] = packetChecksum
var dataChecksum byte = 0
for i := 0; i < (dataLength-1); i++ {
        tCommandSet.commandBuffer[20+i] = param[i] 
		dataChecksum = dataChecksum + param[i]
    }

tCommandSet.commandBuffer[length-1] = dataChecksum
tCommandSet.commandSize = length 

}

func ite_Set_DHCP_Mode(mac string,dhcp bool)(string){
	if ite_info_map[mac] != nil {
		if dhcp==true{
		_= ite_SendCMD(ite_info_map[mac].IP,CMD_SET_DHCP_MODE,[]byte{1})	
		}else{
		_= ite_SendCMD(ite_info_map[mac].IP,CMD_SET_DHCP_MODE,[]byte{0})	
		}
		ite_waitForDone()
		if ri.ack == true{
		ite_info_map[mac].IpMode = dhcp
		return "OK"
		}else{
		return "ERR"
		} 
	}else{
		return "ERR"
	}
}

func ite_Set_Static_IP_Config(mac,ip,mask,gateway string)(string){
	if ite_info_map[mac] != nil {

		config := make([]byte, 12)
		var tmp [4]byte
		tmp,_ = IPToByte4(ip) 
		for i := 0; i < 4; i++ {
		config[i] = tmp[i]
		}
		tmp,_ = IPToByte4(mask) 
		for i := 0; i < 4; i++ {
		config[i+4] = tmp[i]
		}
		tmp,_ = IPToByte4(gateway) 
		for i := 0; i < 4; i++ {
		config[i+8] = tmp[i]
		}

		_= ite_SendCMD(ite_info_map[mac].IP,CMD_SET_IP_ADDRESS,config)	
		ite_waitForDone()
		if ri.ack == true{
		return "OK"
		}else{
		return "ERR"
		} 
	}else{
		return "ERR"
	}
}

func ite_Set_Group_ID_By_IP(ip string,id int)(string){

	if id>1023{
	return "ERR"
	}

	//0-1023
	gid := make([]byte, 2)
	gid[0] = (byte)((id>>8)&0xFF)
	gid[1] = (byte)(id&0xFF)
	_= ite_SendCMD(ip,CMD_SET_GROUPID,gid)	
	ite_waitForDone()
	return "OK"
}

func ite_Set_Group_ID(mac string,id int)(string){

	if id>1023{
	return "ERR"
	}
	if ite_info_map[mac] != nil {
		//0-1023
		gid := make([]byte, 2)
		gid[0] = (byte)((id>>8)&0xFF)
		gid[1] = (byte)(id&0xFF)
		_= ite_SendCMD(ite_info_map[mac].IP,CMD_SET_GROUPID,gid)	
		ite_waitForDone()
		if ri.ack == true{
		ite_info_map[mac].Device_id = strconv.Itoa(id)   
		return "OK"
		}else{
		return "ERR"
		} 
	}else{
		return "ERR"
	}
}

func ite_Set_Baud_Rate(mac string,baudrate int)(string){

	if ite_info_map[mac] != nil {
		arr_baudrate := make([]byte, 4)
		arr_baudrate[0] = (byte)((baudrate>>24)&0xFF)
		arr_baudrate[1] = (byte)((baudrate>>16)&0xFF)
		arr_baudrate[2] = (byte)((baudrate>>8)&0xFF)
		arr_baudrate[3] = (byte)(baudrate&0xFF)

		_= ite_SendCMD(ite_info_map[mac].IP,CMD_SET_UART_BAUDRATE,arr_baudrate)	
		ite_waitForDone()
		if ri.ack == true{
		ite_info_map[mac].UartBaudrate = baudrate
		return "OK"
		}else{
		return "ERR"
		} 
	}else{
		return "ERR"
	}
}

func ite_Set_Macaddress(mac ,newmac string)(string){


	if ite_info_map[mac] != nil {
		
		b_mac := MacaddressToByte6(newmac)
		_= ite_SendCMD(ite_info_map[mac].IP,CMD_SET_MACADDRESS,b_mac)	
		ite_waitForDone()
		if ri.ack == true{
		_= ite_SendCMD(ite_info_map[mac].IP,CMD_SET_REBOOT,[]byte{0})			
		ite_waitForDone()
		//move data to new mac and delete old
		ite_del_node(mac)
		return "OK"
		}else{
		return "ERR"
		} 
	}else{
		return "ERR"
	}
}

func ite_Set_DeviceName(mac ,name string)(string){

	if ite_info_map[mac] != nil {
		
		b_name := make([]byte, 32)
		_ = copy(b_name,[]byte(name))
		_= ite_SendCMD(ite_info_map[mac].IP,CMD_SET_DEVICE_NAME,b_name)	
		ite_waitForDone()
		if ri.ack == true{
		ite_info_map[mac].Name = name

		//event := mac + " rename from "+ite_info_map[mac].Name + " to "+value
		//add_event("Setup",event)
		//ite_info_map[mac].RHost_name = value
		//ite_info_map[mac].Name = value
		//device_info_save()

		return "OK"
		}else{
		return "ERR"
		} 
	}else{
		return "ERR"
	}
}

func ite_Set_StreamingMode(mac string, mode bool)(string){

	if ite_info_map[mac] != nil {
		if mode==true{
		_= ite_SendCMD(ite_info_map[mac].IP,CMD_SET_STREAMING_MODE,[]byte{1})	
		}else{
		_= ite_SendCMD(ite_info_map[mac].IP,CMD_SET_STREAMING_MODE,[]byte{0})	
		}
		ite_waitForDone()
		if ri.ack == true{
		ite_info_map[mac].StreamingMode = mode
		return "OK"
		}else{
		return "ERR"
		} 
	}else{
		return "ERR"
	}
}

func ite_Set_ScreenMode(mac string, mode bool)(string){

	if ite_info_map[mac] != nil {
		if mode==true{
		_= ite_SendCMD(ite_info_map[mac].IP,CMD_SET_SCREEN_MODE,[]byte{1})	
		}else{
		_= ite_SendCMD(ite_info_map[mac].IP,CMD_SET_SCREEN_MODE,[]byte{0})	
		}
		ite_waitForDone()
		if ri.ack == true{
		ite_info_map[mac].ScreenMode = mode
		return "OK"
		}else{
		return "ERR"
		} 
	}else{
		return "ERR"
	}
}

func ite_Set_Osd_Mode(mac string, mode byte)(string){

	if ite_info_map[mac] != nil {

		_= ite_SendCMD(ite_info_map[mac].IP,CMD_SET_OSD_MODE,[]byte{mode})	
		ite_waitForDone()
		if ri.ack == true{
		ite_info_map[mac].OSD_mode = mode
		return "OK"
		}else{
		return "ERR"
		} 
	}else{
		return "ERR"
	}
}

func ite_Set_Group_Video_Bitrate(mac string, group byte,bitRate int)(string){

	if ite_info_map[mac] != nil {
	b_param := make([]byte, 3)
	b_param[0] = group
	b_param[1] = (byte)((bitRate>>8)&0xFF)
	b_param[2] = (byte)(bitRate&0xFF)

	_= ite_SendCMD(ite_info_map[mac].IP,CMD_SET_GROUP_VIDEO_BITRATE,b_param)	
	ite_waitForDone()
	if ri.ack == true{
	//ite_info_map[mac].OSD_mode = mode
	return "OK"
	}else{
	    return "ERR"
	} 
	}else{
		return "ERR"
	}
}

func ite_Set_Video_Out_Mode(mac string, mode byte)(string){ //0-1-2-3

	if ite_info_map[mac] != nil {

		_= ite_SendCMD(ite_info_map[mac].IP,CMD_SET_VIDEO_OUT_MODE,[]byte{mode})	
		ite_waitForDone()
		if ri.ack == true{
		ite_info_map[mac].VideoOutMode = mode
		return "OK"
		}else{
		return "ERR"
		} 
	}else{
		return "ERR"
	}
}

func ite_Set_Low_Delay_Mode(mac string, mode bool)(string){

	if ite_info_map[mac] != nil {
		if mode==true{
		_= ite_SendCMD(ite_info_map[mac].IP,CMD_SET_LOW_DELAY_MODE,[]byte{1})	
		}else{
		_= ite_SendCMD(ite_info_map[mac].IP,CMD_SET_LOW_DELAY_MODE,[]byte{0})	
		}
		ite_waitForDone()
		if ri.ack == true{
		ite_info_map[mac].LowDelayMode = mode
		return "OK"
		}else{
		return "ERR"
		} 
	}else{
		return "ERR"
	}
}

func ite_Set_Multicast_Mode(mac string, mode bool)(string){

	if ite_info_map[mac] != nil {
		if mode==true{
		_= ite_SendCMD(ite_info_map[mac].IP,CMD_SET_MULTICAST_MODE,[]byte{1})	
		}else{
		_= ite_SendCMD(ite_info_map[mac].IP,CMD_SET_MULTICAST_MODE,[]byte{0})	
		}
		ite_waitForDone()
		if ri.ack == true{
		ite_info_map[mac].MulticastMode = mode
		return "OK"
		}else{
		return "ERR"
		} 
	}else{
		return "ERR"
	}
}

func ite_Set_DownScaler_Mode(mac string, source ,target byte)(string){

	if ite_info_map[mac] != nil {
		
		b_param := make([]byte, 2)
		b_param[0] = source
		b_param[1] = target
		ri.trx = true //this is for TX command
		_= ite_SendCMD(ite_info_map[mac].IP,CMD_SET_DOWN_SCALE_MODE,b_param)	
		ite_waitForDone()
		if ri.ack == true{
		if source==0{
		ite_info_map[mac].FHD_DSM = target
		}else{
		ite_info_map[mac].HD_DSM = target
		}

		return "OK"
		}else{
		return "ERR"
		} 
	}else{
		return "ERR"
	}
}

func ite_Set_FactoryReset(mac string)(string){
	if ite_info_map[mac] != nil {
		_= ite_SendCMD(ite_info_map[mac].IP,CMD_SET_FACTORY_RESET,[]byte{0})			
		ite_waitForDone()
		if ri.ack == true{
		return "OK"
		}else{
		return "ERR"
		} 
	}else{
		return "ERR"
	}
}

func ite_Set_Reboot(mac string)(string){
	if ite_info_map[mac] != nil {
		_= ite_SendCMD(ite_info_map[mac].IP,CMD_SET_REBOOT,[]byte{0})			
		ite_waitForDone()
		if ri.ack == true{
		ite_info_map[mac].Alive = "n"
		return "OK"
		}else{
		return "ERR"
		} 
	}else{
		return "ERR"
	}
}

func ite_Set_MCU_Command(mac string,cmdID int,data byte)(string){
	
	if ite_info_map[mac] != nil {
		b_param := make([]byte, 7)
		b_param[0] = 0
		b_param[1] = 0
		b_param[2] = 0
		b_param[3] = 3
		b_param[4] = (byte)((cmdID>>8)&0xFF)
		b_param[5] = (byte)(cmdID&0xFF)
		b_param[6] = data
		fmt.Println("data:",b_param)
		_= ite_SendCMD(ite_info_map[mac].IP,CMD_SET_MCU_COMMAND,b_param)			
		ite_waitForDone()
		if ri.ack == true{
		return "OK"
		}else{
		return "ERR"
		} 
	}else{
		return "ERR"
	}
}

func ite_del_node(mac string) (string) {	
	
	//event_device_del(mac)
	//ast_del_node_preset(mac)
	delete(ite_info_map,mac)	
	ite_device_info_save()
	ite_device_info_load("keep")
	return "OK"
}

func ite_device_info_save(){  
	node_slice := make([]Ite_info, 0, len(ite_info_map))
  for _, tx := range ite_info_map {
      node_slice = append(node_slice, *tx)
  } 
	json_node_list ,_:= json.Marshal(node_slice)
	_ = ioutil.WriteFile("ite_node_info.json",json_node_list, 0644)
}

func ite_device_info_load(alive string){
	file, _ := ioutil.ReadFile("ite_node_info.json")
	var node_slice []Ite_info
	_ = json.Unmarshal([]byte(file), &node_slice)
	for i := 0; i < len(node_slice); i++ {
		ite_info_map[node_slice[i].Device_mac] = &node_slice[i]
		if alive == "clear"{
		ite_info_map[node_slice[i].Device_mac].Alive = "n"
		ite_info_map[node_slice[i].Device_mac].VideoStatus = false
		}
	}
}