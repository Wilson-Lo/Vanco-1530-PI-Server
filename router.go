package main
import (
	"time"
	"fmt"
	"net/http"
	"github.com/gorilla/mux"
	"net"  
    "strings"
    "strconv"
	"encoding/hex"
)


var aliasDialer net.Dialer
var aliasUDPAddr *net.UDPAddr
var ite_aliasDialer net.Dialer
var ite_aliasUDPAddr *net.UDPAddr

var pi4Addr string

//ITE
var pi4Addr_ite string
var _pi4Addr_ite [4]byte

func IPToByte4(ipnr string) ([4]byte,string) {
    bits := strings.Split(ipnr, ".")
	var ip4 [4]byte
	if len(bits)!=4{
	
	return ip4,"ERR"
	}

	
    b0, _ := strconv.Atoi(bits[0])
    b1, _ := strconv.Atoi(bits[1])
    b2, _ := strconv.Atoi(bits[2])
    b3, _ := strconv.Atoi(bits[3])

	ip4[0] = byte(b0)
	ip4[1] = byte(b1)
	ip4[2] = byte(b2)
	ip4[3] = byte(b3)
    return ip4,"OK"
}

func MacaddressToByte6(mac string) []byte {//not include ':'
    
	b_mac , _ := hex.DecodeString(mac )
    return b_mac
}

func NewRouter() *mux.Router {    
	r := mux.NewRouter()

	r.HandleFunc("/api/video_wall_preset",api_list_vw_Preset).Methods("GET")
	r.HandleFunc("/api/video_wall_preset",api_modify_vw_Preset).Methods("POST")
	r.HandleFunc("/api/system_config",sys_get_config).Methods("GET")
	r.HandleFunc("/api/system_config",sys_set_config).Methods("POST")
	r.HandleFunc("/api/fw_update",api_fw_update).Methods("POST")
	r.HandleFunc("/api/connection_lost",api_get_connection_lost).Methods("GET")
	r.HandleFunc("/api/connection_lost/delete/{mac}",api_del_connection_lost).Methods("GET")
	r.HandleFunc("/api/connection_lost/update",api_update_connection_lost).Methods("POST")
	r.HandleFunc("/api/video_lost",api_get_video_lost).Methods("GET")
	r.HandleFunc("/api/video_lost/update",api_update_video_lost).Methods("POST")
	r.HandleFunc("/api/video_lost/delete/{mac}",api_del_video_lost).Methods("GET")
	r.HandleFunc("/api/get_mac",api_get_mac).Methods("GET")
	r.HandleFunc("/api/change_password",system_changePassword).Methods("POST")
	r.HandleFunc("/api/debug",api_debug).Methods("POST")
	r.HandleFunc("/api/login",api_login).Methods("POST")
	r.HandleFunc("/api/list_mail",api_listEmail).Methods("GET")
	r.HandleFunc("/api/modify_mail",api_modifyEmail).Methods("POST")
	r.HandleFunc("/api/list_preset",api_listPreset).Methods("GET")
	r.HandleFunc("/api/load_preset",api_loadPreset).Methods("POST")
	r.HandleFunc("/api/save_preset",api_savePreset).Methods("POST")
	r.HandleFunc("/api/modify_preset",api_modifyPreset).Methods("POST")
	r.HandleFunc("/api/node_info",api_node_info).Methods("GET")
	r.HandleFunc("/api/astswitch",api_astswitch).Methods("POST")
	r.HandleFunc("/api/history_list",api_history_list).Methods("GET")	
	r.HandleFunc("/api/history_list/delete",api_history_delete).Methods("GET")	
	r.HandleFunc("/api/astparam_g/{ip}/{param}",api_astparam_g).Methods("GET")
	r.HandleFunc("/api/astparam_s",api_astparam_s).Methods("POST")	
	r.HandleFunc("/api/ast_sendcmd",api_ast_sendcmd).Methods("POST")	
	r.HandleFunc("/api/setup_info/{mac}",api_setup_info).Methods("GET")
	r.HandleFunc("/api/node_list",api_node_list).Methods("GET")
	r.HandleFunc("/api/set_hostname",api_ast_set_name).Methods("POST")		
	r.HandleFunc("/api/lmparam_g/{ip}/{param}",api_lmparam_g).Methods("GET")
	r.HandleFunc("/api/vw_info/{mac}",api_vw_info).Methods("GET")
	r.HandleFunc("/api/set_vw",api_set_vw).Methods("POST")
	r.HandleFunc("/api/set_ite_vw",api_set_ite_vw).Methods("POST") // for bzb set ite video wall
	r.HandleFunc("/api/del_node/{mac}",api_del_node).Methods("GET")
	r.HandleFunc("/api/save_node",api_save_node).Methods("GET")
	r.HandleFunc("/api/astlinkoff",api_astlinkoff).Methods("POST")
	r.HandleFunc("/api/osd_list",api_osd_list).Methods("GET")
	r.HandleFunc("/api/set_osd_s",api_set_osd_s).Methods("POST")
	r.HandleFunc("/api/osd_apply",api_osd_apply).Methods("POST")
	r.HandleFunc("/api/osd_close",api_osd_close).Methods("POST")//JK
	r.HandleFunc("/api/set_osd_pos",api_set_osd_pos).Methods("POST")
	r.HandleFunc("/api/set_osd_pic",api_set_osd_pic).Methods("POST")
	r.HandleFunc("/api/set_ip",api_pi4_set_ip).Methods("POST")
	r.HandleFunc("/api/upload_image",api_uploadImg).Methods("POST")
	r.HandleFunc("/api/del_image/{name}",pi4_del_image).Methods("GET")
	r.HandleFunc("/api/set_dhcp",api_pi4_set_dhcp).Methods("GET")
	r.HandleFunc("/api/mjpg", api_mjpg).Methods("GET")
	r.HandleFunc("/api/post_test/{mac}",http_post_debug).Methods("GET")  //DEBUG
	//api third party
	r.HandleFunc("/api",api_thirdPart).Methods("POST")
	//test PIP
	r.HandleFunc("/api/pip", api_pip).Methods("GET")

	r.PathPrefix("/").Handler(http.StripPrefix("/", http.FileServer(http.Dir("./www/"))))




   get_ite_ip := false 
  i,_:= net.InterfaceByName("eth0")
  addrs, _ := i.Addrs()
  for _, a := range addrs {

    fmt.Println(a.String())
  	if(strings.Contains(a.String(),"169.254"))	{
  		pi4Addr = strings.Replace(a.String(), "/16", "",-1)
		pi4Addr = strings.Replace(pi4Addr , "/24", "",-1)
    	aliasDialer = net.Dialer{LocalAddr: &net.TCPAddr{IP:a.(*net.IPNet).IP,Port: 0,},Timeout: 3*time.Second}
    	aliasUDPAddr = &net.UDPAddr{IP: a.(*net.IPNet).IP}  //node_list alias IP
		fmt.Println("pi3: AST " + pi4Addr)
		//break
  	}else if(strings.Contains(a.String(),"/64")){
		//
	}else{
	   fmt.Println("pi3: " , a.String())
	   pi4Addr_ite = strings.Replace(a.String(), "/24", "",-1)
	   _pi4Addr_ite,_ = IPToByte4(pi4Addr_ite) //get pi3 to _pi4Addr_ite[4]
	   fmt.Println("pi3: ITE " , _pi4Addr_ite)
	   ite_aliasDialer = net.Dialer{LocalAddr: &net.TCPAddr{IP:a.(*net.IPNet).IP,Port: 0,},Timeout: 3*time.Second}
	   ite_aliasUDPAddr = &net.UDPAddr{IP: a.(*net.IPNet).IP}  //node_list alias IP
	   get_ite_ip = true
	}

  }
  if get_ite_ip == false{
	fmt.Println("Not get normal IP")
    pi4Addr_ite = pi4Addr
	_pi4Addr_ite,_ = IPToByte4(pi4Addr_ite) //get pi3 to _pi4Addr_ite[4]
	ite_aliasDialer = aliasDialer
	ite_aliasUDPAddr = aliasUDPAddr
	}

  
	return r
}

