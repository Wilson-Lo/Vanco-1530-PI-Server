package main
import (
    "io"
	"fmt"
	"net/http"
	"encoding/json"
	"github.com/gorilla/mux"
	"os"
	"strconv"
	"strings"
	"io/ioutil"
)

type ackStatus int
const(
    unknow = 0
    ACK_OK = 1
	ACK_GetSystemPlatform = 2
	ACK_GetSystemModel =3
	ACK_GetAutoTimeInfo = 4
	ACK_GetExtenderVendor = 5  
	ACK_GetExtenderLists = 6
)


type json_thirdpart struct {	
	jsonrPc string `json:"jsonrpc"`
	id string `json:"id"`
	method string `json:"method"`
	//params string `json:"params"`
}


type ExtenderList struct {	
	M_vendor string `json:"vendor"`
	M_deviceName string `json:"deviceName"`
	M_macAddress string `json:"macAddress"`
	M_ipAddress string `json:"ipAddress"`
	M_netmask string `json:"netmask"`
	M_gateway string `json:"gateway"`
	M_dhcpStatus bool `json:"dhcpStatus"`
	M_displayMode int `json:"displayMode"`
	M_videoOutputMode int `json:"videoOutputMode"`
	M_rxScreenMode int `json:"rxScreenMode"`
	M_groupId int `json:"groupId"`
	M_dipGroupId int `json:"dipGroupId"`
	M_deviceType string `json:"deviceType"`
	M_hdmiMuteFlag bool `json:"hdmiMuteFlag"`
	M_audioMuteFlag bool `json:"audioMuteFlag"`
	M_ledFlag bool `json:"ledFlag"`
	M_cecFlag bool `json:"cecFlag"`
	M_cecDetectionFlag bool `json:"cecDetectionFlag"`
	M_learningType string `json:"learningType"`
	M_screenshotpath string `json:"screenshotpath"`
	M_sinkConnection bool `json:"sinkConnection"`
	M_osdIndex int `json:"osdIndex"`
	M_osdMute bool `json:"osdMute"`
	M_osdResolution string `json:"osdResolution"`
	M_rxCurrentResolution string `json:"rxCurrentResolution"`
	M_clientSource int `json:"clientSource"`
	M_edidMode int `json:"edidMode"`
	M_learningEDIDFrom string `json:"learningEDIDFrom"`
	M_homeFlag bool `json:"homeFlag"`
	M_speedMode string `json:"speedMode"`
	M_rotate int `json:"rotate"`
	M_multicastAddress string `json:"multicastAddress"`
	M_txDeviceIpAddress string `json:"txDeviceIpAddress"`
	M_remotReceivePort int `json:"remotReceivePort"`
	M_fullHDDownScaleMode int `json:"fullHDDownScaleMode"`
	M_hdDOwnScaleMode int `json:"hdDOwnScaleMode"`
	M_upScaleMode int `json:"upScaleMode"`
	M_baudRate int `json:"baudRate"`
	M_kBitRate int `json:"kBitRate"`
	M_txVersion string `json:"txVersion"`
	M_rxVersion string `json:"rxVersion"`
	M_encoderVersion string `json:"encoderVersion"`
	M_controlVersion string `json:"controlVersion"`
	M_status int `json:"status"`
	M_hdcpStatus bool `json:"hdcpStatus"`
	M_lanStatus bool `json:"lanStatus"`
	M_videoLockStatus bool `json:"videoLockStatus"`
	M_alive int `json:"alive"`
	M_retry int `json:"retry"`
	M_portNo int `json:"portNo"`
	M_factoryDefault bool `json:"factoryDefault"`
	M_infoReady int `json:"infoReady"`
	M_lastErrorMessage string `json:"lastErrorMessage"`
	M_recycleTxIndex int `json:"recycleTxIndex"`
	M_verifyFactory bool `json:"verifyFactory"`
}


type ExtenderLists struct {	
	ID string `json:"id"`
	JSONRPC string `json:"jsonrpc"`
	RESULT []ExtenderList  `json:"result"`
}



func api_node_info(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")//delete this line when FW is publishing **************************
	node_slice := make([]Ast_info, 0, len(ast_info_map))
  for _, tx := range ast_info_map {
      node_slice = append(node_slice, *tx)
  } 
	json_node_list ,_:= json.Marshal(node_slice)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(json_node_list) 
	w.(http.Flusher).Flush()
}

func api_node_list(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")//delete this line when FW is publishing **************************
	node_slice := ast_node_array[:ast_node_num]
	json_node_list ,_:= json.Marshal(node_slice)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(json_node_list)
	w.(http.Flusher).Flush()
}

func api_listPreset(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")//delete this line when FW is publishing **************************
	var preset_array Ast_preset_array
	preset_array.PRESET = nil
	for i := 0; i < 8; i++ {
	preset_array.PRESET = append(preset_array.PRESET,ast_preset_map[i])
	}
	preset_slice := make([]Ast_preset_array, 0, 1)	
	preset_slice = append(preset_slice, preset_array)
	json_node_list ,_:= json.Marshal(preset_slice)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(json_node_list) 
	w.(http.Flusher).Flush()
}

func api_osd_list(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")//delete this line when FW is publishing **************************
	node_slice := ast_osd_map[:]
	json_node_list ,_:= json.Marshal(node_slice)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(json_node_list)
	w.(http.Flusher).Flush()
}

func api_savePreset(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")//delete this line when FW is publishing **************************
	w.WriteHeader(http.StatusOK)
	r.ParseForm()	
	fmt.Fprintf(w,"{\"result\":\"%v\"}", ast_savePreset(r.FormValue("index")))
	w.(http.Flusher).Flush()
}

func api_loadPreset(w http.ResponseWriter, r *http.Request) {
	
	w.Header().Set("Access-Control-Allow-Origin", "*")//delete this line when FW is publishing **************************
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w,"{\"result\":\"OK\"}")
	w.(http.Flusher).Flush()
	index := r.FormValue("index")
	event := "Load preset "+ index
	add_event("Mapping",event)
}

func api_del_node(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r) 
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w,"{\"result\":\"%v\"}",ast_del_node(vars["mac"]))
	w.(http.Flusher).Flush()
}

func api_setup_info(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")//delete this line when FW is publishing **************************
	vars := mux.Vars(r) 
	w.Header().Set("Content-Type", "application/json")
	json_setup_info ,_:= json.Marshal(ast_setup_info(vars["mac"]))
	w.Write(json_setup_info) 
	w.(http.Flusher).Flush()
}

func api_save_node(w http.ResponseWriter, r *http.Request) {
	device_info_save()
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "{\"result\":\"OK\"}")
	w.(http.Flusher).Flush()
}


func api_lmparam_g(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r) 
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w,"{\"result\":\"%v\"}", ast_lmparam_g(vars["ip"]+":24",vars["param"]))
	w.(http.Flusher).Flush()
}

func api_set_vw(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")//delete this line when FW is publishing **************************
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w,"{\"result\":\"%v\"}", ast_set_vw(r.FormValue("mac"),r.FormValue("vwh"),r.FormValue("vwv"),r.FormValue("vwp"),r.FormValue("vwl"),r.FormValue("vwr"),r.FormValue("vwu"),r.FormValue("vwb")))
	w.(http.Flusher).Flush()
}

func api_set_ite_vw(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")//delete this line when FW is publishing **************************
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w,"{\"result\":\"%v\"}", ast_set_ite_vw(r.FormValue("mac"),r.FormValue("v_count"),r.FormValue("h_count"),r.FormValue("vw"),r.FormValue("ow"),r.FormValue("vh"),r.FormValue("oh"),r.FormValue("row"),r.FormValue("col")))
	w.(http.Flusher).Flush()
}

func api_set_osd_s(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")//delete this line when FW is publishing **************************
	w.WriteHeader(http.StatusOK)
	r.ParseForm()
	osdtype := r.FormValue("type") == "true"
	
	fmt.Fprintf(w,"{\"result\":\"%v\"}", ast_set_osd(r.FormValue("index"),r.FormValue("name"),r.FormValue("alpha"),r.FormValue("mask"),r.FormValue("offtime"),r.FormValue("str"),r.FormValue("fsize"),r.FormValue("fcolor"),r.FormValue("pos"),r.FormValue("x"),r.FormValue("y"),r.FormValue("pic"),osdtype))
	w.(http.Flusher).Flush()//index,alpha,mask,offtime,str,fsize,fcolor,pos,x,y,pic string
}

func api_osd_apply(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")//delete this line when FW is publishing **************************
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w,"{\"result\":\"%v\"}", ast_osd_apply(r.FormValue("mac"),r.FormValue("index")))
	w.(http.Flusher).Flush()
}

func api_osd_close(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")//delete this line when FW is publishing **************************
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w,"{\"result\":\"%v\"}", ast_osd_close(r.FormValue("mac"),r.FormValue("index")))
	w.(http.Flusher).Flush()//index,alpha,mask,offtime,str,fsize,fcolor,pos,x,y,pic string
}

func api_set_osd_pos(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")//delete this line when FW is publishing **************************
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w,"{\"result\":\"%v\"}", ast_set_osd_pos(r.FormValue("mac"),r.FormValue("pos"),r.FormValue("x"),r.FormValue("y")))
	w.(http.Flusher).Flush()
}

func api_set_osd_pic(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")//delete this line when FW is publishing **************************
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w,"{\"result\":\"%v\"}", ast_set_osd_pic(r.FormValue("mac"),r.FormValue("pic"),r.FormValue("offtime"),r.FormValue("mask"),r.FormValue("alpha")))
	w.(http.Flusher).Flush()
}


func api_vw_info(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r) 
	w.Header().Set("Access-Control-Allow-Origin", "*")//delete this line when FW is publishing **************************	
	w.Header().Set("Content-Type", "application/json")
	json_vw_info ,_:= json.Marshal(ast_vw_info(vars["mac"]))
	w.Write(json_vw_info) 
	w.(http.Flusher).Flush()
}


func api_astparam_g(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")//delete this line when FW is publishing **************************
	vars := mux.Vars(r) 
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w,"{\"result\":\"%v\"}", ast_astparam_g(vars["ip"]+":24",vars["param"]))
	w.(http.Flusher).Flush()
}

func api_ast_sendcmd(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")//delete this line when FW is publishing **************************
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w,"{\"result\":\"%v\"}", ast_sendcmd(r.FormValue("ip")+":24",r.FormValue("value")))
	w.(http.Flusher).Flush()
}

func api_astparam_s(w http.ResponseWriter, r *http.Request) {

	//fmt.Println("get param setting\n")
	w.Header().Set("Access-Control-Allow-Origin", "*")//delete this line when FW is publishing **************************
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w,"{\"result\":\"%v\"}", ast_astparam_s(r.FormValue("ip")+":24",r.FormValue("param"),r.FormValue("value")))
	w.(http.Flusher).Flush()

}

func api_ast_set_name(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")//delete this line when FW is publishing **************************
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w,"{\"result\":\"%v\"}", ast_set_name(r.FormValue("mac"),r.FormValue("value")))
	w.(http.Flusher).Flush()
}

func api_astswitch(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")//delete this line when FW is publishing **************************
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w,"{\"result\":\"%v\"}", astswitch(r.FormValue("ip")+":24",r.FormValue("switch_id"),r.FormValue("switch_type")))
	w.(http.Flusher).Flush()
}

//sub-streaming
func api_mjpg(w http.ResponseWriter, r *http.Request) {

	 resp, err := http.Get("http://" + r.FormValue("ip") + ":8080/?action=stream&w=320&h=240&fps=25&bw=8000&as=20")
     if err != nil{
     }

     for {
     	w.Header().Set("Content-Type", "multipart/x-mixed-replace; boundary=boundarydonotcross")
        io.Copy(w, resp.Body)
        w.(http.Flusher).Flush()
     }
}

func api_modifyPreset(w http.ResponseWriter, r *http.Request) {
	
	var p Ast_preset 
	
	w.Header().Set("Access-Control-Allow-Origin", "*")//delete this line when FW is publishing **************************
	w.WriteHeader(http.StatusOK)
	
	err := json.NewDecoder(r.Body).Decode(&p)
	if err!=nil{
		fmt.Println("error")
		fmt.Fprintf(w,"{\"result\":\"%v\"}", "DAIL_ERR")
	}else{
	fmt.Println(p)
	ast_modifyPreset(p.INDEX,p.SETTING)
	fmt.Fprintf(w,"{\"result\":\"%v\"}", "OK")
	}
	w.(http.Flusher).Flush()
}


func api_astlinkoff(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w,"{\"result\":\"%v\"}", astlinkoff(r.FormValue("ip")+":24",r.FormValue("switch_id")))
	w.(http.Flusher).Flush()
}

func api_uploadImg(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")//delete this line when FW is publishing **************************
	w.WriteHeader(http.StatusOK)
	r.ParseForm()
	png_jpg := r.FormValue("type") == "true"

	if png_jpg==false {
	
	fmt.Fprintf(w,"{\"result\":\"%v\"}", pi4_savePng(r.FormValue("name"),r.FormValue("data")))
	
	}else{ 
	fmt.Fprintf(w,"{\"result\":\"%v\"}", pi4_saveJpg(r.FormValue("name"),r.FormValue("data")))
	}
	
	w.(http.Flusher).Flush()
}

func listAll(path string, curHier int) {
    readerInfos ,err := ioutil.ReadDir(path)
    if err != nil {
        fmt.Println(err)
        return
    }
    for _, info := range readerInfos { 
        if info.IsDir() {
            for tmpheir := curHier; tmpheir >0; tmpheir-- {
                fmt.Printf("|\t")
            }
            fmt.Println(info.Name(),"\\")
            listAll(path + "\\" + info.Name(),curHier+1)
        }else {
            for tmpheir := curHier ;tmpheir >0; tmpheir-- {
                fmt.Printf("|\t")
            }
            fmt.Println(info.Name())
        }
    }
}

func pi4_del_image(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")//delete this line when FW is publishing **************************
	vars := mux.Vars(r) 
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	path := "./www/snapshot/osd/"+vars["name"]
	e := os.Remove(path) 
	if e != nil { 
		fmt.Fprintf(w,"{\"result\":\"DAIL_ERR\"}")
    }else{
	fmt.Fprintf(w,"{\"result\":\"OK\"}")
	}
	w.(http.Flusher).Flush()
	mypath := "./www/snapshot/osd/"
	listAll(mypath,0);
}

func func_getMacaddress(params map[string]interface{})(string){

macaddress := params["macAddress"].(string)
macaddress = strings.Replace(macaddress, ":", "",-1)
return macaddress
}

func func_getMacaddressAndValue(params map[string]interface{})(string, string){

macaddress := params["macAddress"].(string)
value := int(params["value"].(float64))
macaddress = strings.Replace(macaddress, ":", "",-1)
values := strconv.Itoa(value)   
return macaddress,values
}

func func_getMacaddressAndBoolean(params map[string]interface{})(string, bool){

macaddress := params["macAddress"].(string)
value := params["value"].(bool)
macaddress = strings.Replace(macaddress, ":", "",-1)
return macaddress,value
}

func func_getMacaddressAndString(params map[string]interface{})(string, string){

macaddress := params["macAddress"].(string)
value := params["value"].(string)
macaddress = strings.Replace(macaddress, ":", "",-1)
return macaddress,value
}


func api_thirdPart(w http.ResponseWriter, r *http.Request) {
	
	var ack int
	
	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
      fmt.Println(err)
    }
	var jsonObj map[string]interface{}
	json.Unmarshal(bodyBytes, &jsonObj)
	classMethod := jsonObj["method"].(string)
	uid := jsonObj["id"].(string)

	fmt.Println("third part API")
	ack = 1

	switch {
	
	//Matrix
	//------------------------------------------------------------------
	case classMethod == "APIAutoAVSetting.SetExtenderGroupIdCommand":
	params := jsonObj["params"].(map[string]interface{})
	mac,index := func_getMacaddressAndValue(params)
	fmt.Println("out:",mac,index)
	if ast_info_map[mac] != nil {
	ip := ast_info_map[mac].IP +":24"
	astswitch(ip,index,"z")
	ack = 0
	}
	//------------------------------------------------------------------
	case classMethod == "APIAutoAVSetting.DeleteExtenderConfig":
	mac := jsonObj["params"].(string)
	mac = strings.Replace(mac, ":", "",-1)
	fmt.Println("Delete Extender:",mac)
	ast_del_node(mac)
	ack = 0
	//------------------------------------------------------------------
	case classMethod == "APIAutoAVSetting.SetExtenderMapping":
	params := jsonObj["params"].([]interface{}) 

	for _, p := range params {
	param := p.(map[string]interface{})
	index := int(param["index"].(float64))
	mapping := param["mappings"].(map[string]interface{})
	groupId := int(mapping["groupId"].(float64))
	groupId_s := strconv.Itoa(groupId)   
	deviceName := mapping["deviceName"].(string)
	deviceType := mapping["deviceType"].(string)
	mac := mapping["macAddress"].(string)
	fmt.Println("out:",index,deviceName,deviceType,mac,groupId_s)
	}
	//------------------------------------------------------------------
	case classMethod == "APIAutoAVSetting.SetCurrentExtenderMapping":
	params := jsonObj["params"].(map[string]interface{})
	value := int(params["value"].(float64))
	values := strconv.Itoa(value)   
	fmt.Println("SetCurrentExtenderMapping:",values)
	//------------------------------------------------------------------
	case classMethod == "APIAutoAVSetting.SetExtenderClientSourceCommand":
	params := jsonObj["params"].(map[string]interface{})
	mac,groupId := func_getMacaddressAndValue(params)
	fmt.Println("out:",mac,groupId)
		if ast_info_map[mac] != nil {
			ip := ast_info_map[mac].IP +":24"
			astswitch(ip,groupId,"z")
			
		}
	ack = 0 
	//------------------------------------------------------------------
	case classMethod == "APIAutoAVSetting.SetExtenderCECCommand":
	params := jsonObj["params"].(map[string]interface{})
	mac,value := func_getMacaddressAndBoolean(params)
	fmt.Println("out:",mac,value)
	//------------------------------------------------------------------ mute
	case classMethod == "APIAutoAVSetting.SetExtenderHDMIMuteCommand":
	params := jsonObj["params"].(map[string]interface{})
	mac,value := func_getMacaddressAndBoolean(params)
	if ast_info_map[mac] != nil {
	ip := ast_info_map[mac].IP +":24"
	var cmd string
	if value == true{
	cmd = "echo 1 > /sys/devices/platform/display/screen_off"
	}else{
	cmd = "echo 0 > /sys/devices/platform/display/screen_off"
	_ = ast_sendcmd(ip,cmd)
	ack = 0

	}
	fmt.Println("cmd:",ip,cmd)

	}	
	//------------------------------------------------------------------
	//#################### OSD ####################//
	//------------------------------------------------------------------ OSD Mapping
	case classMethod == "APIAutoAVSetting.SetExtenderOSDMappingCommand":
	params := jsonObj["params"].(map[string]interface{})
	mac,value := func_getMacaddressAndValue(params)
	if ast_info_map[mac] != nil {
	ast_osd_apply(mac,value)
	}
	//------------------------------------------------------------------ Auto OSD Config
	case classMethod == "APIAutoAVSetting.SetExtenderAutoOSDConfig":
	params := jsonObj["params"].(map[string]interface{})
	setting := params["setting"].(map[string]interface{})
	backgroudColor := int(setting["backgroudColor"].(float64))
	textColor := setting["textColor"].(string)
	timeout := int(setting["timeout"].(float64))
	//transparency := int(setting["transparency"].(float64))
	fmt.Println("OSD Config",backgroudColor,textColor,timeout)
	//#################### Setup ####################//
	//------------------------------------------------------------------ LED flash
	case classMethod == "APIAutoAVSetting.SetExtenderLEDFlashCommand": //cat /sys/devices/platform/ast1500_led.2/leds:button_link/N_Led
	params := jsonObj["params"].(map[string]interface{})
	mac,value := func_getMacaddressAndBoolean(params)
	fmt.Println("LED flash:",mac,value)
	if ast_info_map[mac] != nil {
	ip := ast_info_map[mac].IP +":24"		
	var cmd string
	if value{
	cmd = "cat /sys/devices/platform/ast1500_led.2/leds:button_link/N_Led"
	}else{
	cmd = "echo 2 > /sys/devices/platform/ast1500_led.2/leds:button_link/N_Led"
	}
	_ = ast_sendcmd(ip,cmd)
	}
	ack = 0
	//------------------------------------------------------------------ Device Name
	case classMethod == "APIAutoAVSetting.SetExtenderDeviceNameCommand": 
	params := jsonObj["params"].(map[string]interface{})
	mac,value := func_getMacaddressAndString(params)
	fmt.Println("Device Name:",mac,value)
	if ast_info_map[mac] != nil {
		ast_set_name(mac,value)
	}
	ack = 0
	//------------------------------------------------------------------ Speed
	case classMethod == "APIAutoAVSetting.SetExtenderSpeedCommand": 
	params := jsonObj["params"].(map[string]interface{})
	mac,value := func_getMacaddressAndString(params)
	if ast_info_map[mac] != nil {
		ip := ast_info_map[mac].IP +":24"		
		_ = ast_astparam_s(ip,"profile",value)
		//auto,10M, 50M, 100M, 150M, 200M
		
	}


	//------------------------------------------------------------------ UpScale
	case classMethod == "APIAutoAVSetting.SetExtenderUpScaleCommand": 
	params := jsonObj["params"].(map[string]interface{})
	mac,value := func_getMacaddressAndString(params)
	fmt.Println("UpScale:",mac,value)
	if ast_info_map[mac] != nil {
		ip := ast_info_map[mac].IP +":24"		
		_ = ast_astparam_s(ip,"v_output_timing_convert",value)
	}
	/*
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
			res = "Pass-Through"//00000000
			}*/
	//------------------------------------------------------------------ Video loop Out
	case classMethod == "APIAutoAVSetting.SetExtenderVideoOutCommand": 
	params := jsonObj["params"].(map[string]interface{})
	mac,value := func_getMacaddressAndValue(params)
	fmt.Println("loop out:",mac,value)
	if ast_info_map[mac] != nil {
		ip := ast_info_map[mac].IP +":24"	
		var cmd string
		if value == "0" {
			cmd = "echo 0 > /sys/devices/platform/ast1500_led.2/leds:button_link/ CM_MUTELOOPOUTh"
		}else{
			cmd = "echo 1 > /sys/devices/platform/ast1500_led.2/leds:button_link/ CM_MUTELOOPOUTh"
		}
		_ = ast_sendcmd(ip,cmd)
	}
	//------------------------------------------------------------------ reboot
	case classMethod == "APIAutoAVSetting.SetExtenderRebootCommand": 
	params := jsonObj["params"].(map[string]interface{})
	mac := func_getMacaddress(params)
	fmt.Println("Reboot:",mac)
	if ast_info_map[mac] != nil {
		ip := ast_info_map[mac].IP +":24"	
		_ = ast_sendcmd(ip,"reboot")
	}

	//------------------------------------------------------------------ Factory Reset
	case classMethod == "APIAutoAVSetting.SetExtenderFactoryResetCommand": 
	params := jsonObj["params"].(map[string]interface{})
	mac := func_getMacaddress(params)
	fmt.Println("Factory Reset:",mac)
		if ast_info_map[mac] != nil {
		ip := ast_info_map[mac].IP +":24"	
		_ = ast_sendcmd(ip,"reset_to_default.sh")
	}
	ack = 0
	//------------------------------------------------------------------ BaudRate
	case classMethod == "APIAutoAVSetting.SetExtenderBaudRateCommand": 
	params := jsonObj["params"].(map[string]interface{})
	mac,value := func_getMacaddressAndValue(params)
	fmt.Println("BaudRate:",mac,value)
	if ast_info_map[mac] != nil { 
		ip := ast_info_map[mac].IP +":24"	
		var v_baudrate string
		v_baudrate = value+"-8n1"	
		_ = ast_astparam_s(ip,"s0_baudrate",v_baudrate)
	}
	ack = 0
	//------------------------------------------------------------------ Rotate
	case classMethod == "APIAutoAVSetting.SetExtenderRotateCommand": 
	params := jsonObj["params"].(map[string]interface{})
	mac,value := func_getMacaddressAndValue(params)
	fmt.Println("Rotate:",mac,value)
		if ast_info_map[mac] != nil {
		ip := ast_info_map[mac].IP +":24"	
		var cmd string
		if value== "270"{
		cmd = "vw_rotate 6"
		}else if value=="180"{
		cmd = "vw_rotate 3"
		}else{
		cmd = "vw_rotate 0"
		}
		_ = ast_sendcmd(ip,cmd)
	}
	//#################### System ####################//
	//------------------------------------------------------------------ 
	case classMethod == "APIAutoAVSetting.GetSystemPlatform": 
	fmt.Println("GetSystemPlatform")
	ack = ACK_GetSystemPlatform
	case classMethod == "APIAutoAVSetting.GetSystemModel": 
	fmt.Println("ACK_GetSystemModel")
	ack = ACK_GetSystemModel
	case classMethod == "APIAutoAVSetting.GetAutoTimeInfo": 
	fmt.Println("ACK_GetAutoTimeInfo")
	ack = ACK_GetAutoTimeInfo
	case classMethod == "APIAutoAVSetting.GetExtenderVendor": 
	fmt.Println("ACK_GetExtenderVendor")
	ack = ACK_GetExtenderVendor
	case classMethod == "APIAutoAVSetting.GetExtenderLists": 
	
	ack = ACK_GetExtenderLists
	

	default: 
	fmt.Println("not supprot")
	}

	//fmt.Fprintf(w,"{\"result\":\"OK\"}")
	w.Header().Set("Access-Control-Allow-Origin", "*")//delete this line when FW is publishing **************************

	ackList := make([]map[string]interface{}, 0)


	switch {
	case ack == 0:
	ackList = append(ackList, map[string]interface{}{
              "id": uid,
			  "jsonrpc": "2.0",
			  "result": 1,			  
              })
	case ack == ACK_GetSystemPlatform:
	mac := insertNth(pi4_mac,2) 
	ackList = append(ackList, map[string]interface{}{
		"id": uid,
		"jsonrpc": "2.0",
		"result": map[string]interface{}{
        "macAddress":mac,
        "value":"pi",
		}})
	case ack == ACK_GetSystemModel:
	mac := insertNth(pi4_mac,2) 
	ackList = append(ackList, map[string]interface{}{
		"id": uid,
		"jsonrpc": "2.0",
		"result": map[string]interface{}{
        "macAddress":mac,
        "value":"SmartController",
		}})
	case ack == ACK_GetAutoTimeInfo: 
	sys_config.TIME = GetCurrentTime()
	sys_config.TIME = strings.Replace(sys_config.TIME, "\n", "",-1)
	k, _ := strconv.Atoi(sys_config.Timezone)
	if k>0 {
	k = k-1
	}
	ackList = append(ackList, map[string]interface{}{
		"id": uid,
		"jsonrpc": "2.0",
		"result": map[string]interface{}{
        "timezone":TimezoneList[k], //sys_config.Timezone
        "unixTime":sys_config.TIME,
		}})
	case ack == ACK_GetExtenderVendor:
	mac := insertNth(pi4_mac,2) 
	ackList = append(ackList, map[string]interface{}{
		"id": uid,
		"jsonrpc": "2.0",
		"result": map[string]interface{}{
        "macAddress":mac,
        "value":"aspeedITE",
		}})
	case ack == ACK_GetExtenderLists:
	fmt.Println("ACK_GetExtenderLists")
	
	var extenderNode ExtenderLists
	extenderNode.ID = uid
	extenderNode.JSONRPC = "2.0"
	

    for _, tx := range ast_info_map {
      
	  var tmp ExtenderList
	  tmp.M_vendor = "AST"
	  tmp.M_deviceName = tx.Host_name
	  tmp.M_macAddress = insertNth(tx.Device_mac,2) 
	  tmp.M_ipAddress = tx.IP
	  tmp.M_netmask = "255.255.0.0"
	  tmp.M_groupId , _ = strconv.Atoi(tx.Device_id)
	  tmp.M_deviceType = tx.Host_type
	  tmp.M_learningType = "static"
	  

	  extenderNode.RESULT = append(extenderNode.RESULT, tmp)
	}

	json_mac ,_:= json.Marshal(extenderNode)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(json_mac)
	w.(http.Flusher).Flush()

	return 
	default:
	ackList = append(ackList, map[string]interface{}{
              "id": uid,
			  "jsonrpc": "2.0",
			  "result": 0,			  
              })
	}
	json_mac ,_:= json.Marshal(ackList)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(json_mac)
	w.(http.Flusher).Flush()
	
} 

