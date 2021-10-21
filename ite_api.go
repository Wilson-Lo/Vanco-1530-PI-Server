package main
import (
	"fmt"
	"net/http"
	"encoding/json"
	"github.com/gorilla/mux"
	"strconv"
)


func ite_api_setDeviceName(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")//delete this line when FW is publishing **************************
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w,"{\"result\":\"%v\"}", ite_Set_DeviceName(r.FormValue("mac"),r.FormValue("name")))
	w.(http.Flusher).Flush()
}

func ite_api_setMacAddress(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")//delete this line when FW is publishing **************************
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w,"{\"result\":\"%v\"}", ite_Set_Macaddress(r.FormValue("current_mac"),r.FormValue("new_mac")))
	w.(http.Flusher).Flush()
}

func ite_api_setGroupID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")//delete this line when FW is publishing **************************
	w.WriteHeader(http.StatusOK)
	gid, _ := strconv.Atoi(r.FormValue("id"))
	fmt.Fprintf(w,"{\"result\":\"%v\"}", ite_Set_Group_ID(r.FormValue("mac"),gid))
	w.(http.Flusher).Flush()
}

func ite_api_setStaticIP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")//delete this line when FW is publishing **************************
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w,"{\"result\":\"%v\"}", ite_Set_Static_IP_Config(r.FormValue("mac"),r.FormValue("ip"),r.FormValue("mask"),r.FormValue("gateway")))
	w.(http.Flusher).Flush()
}

func ite_api_setDhcpMode(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")//delete this line when FW is publishing **************************
	w.WriteHeader(http.StatusOK)
	mode := true
	if r.FormValue("dhcp_status") == "true"{
	 mode =true
	}else{
	mode = false
	}
	fmt.Fprintf(w,"{\"result\":\"%v\"}", ite_Set_DHCP_Mode(r.FormValue("mac"),mode))
	w.(http.Flusher).Flush()
}

func ite_api_setBaudRate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")//delete this line when FW is publishing **************************
	w.WriteHeader(http.StatusOK)
	baudRate, _ := strconv.Atoi(r.FormValue("baud_rate"))
	fmt.Fprintf(w,"{\"result\":\"%v\"}", ite_Set_Baud_Rate(r.FormValue("mac"),baudRate))
	w.(http.Flusher).Flush()
}

func ite_api_setStreamingMode(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")//delete this line when FW is publishing **************************
	w.WriteHeader(http.StatusOK)
	mode := true
	if r.FormValue("streaming_mode") == "1"{
	 mode =true
	}else{
	mode = false
	}
	fmt.Fprintf(w,"{\"result\":\"%v\"}", ite_Set_StreamingMode(r.FormValue("mac"),mode))
	w.(http.Flusher).Flush()
}

func ite_api_setScreenMode(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")//delete this line when FW is publishing **************************
	w.WriteHeader(http.StatusOK)
	mode := true
	if r.FormValue("screen_mode") == "1"{
	 mode =true
	}else{
	mode = false
	}
	fmt.Fprintf(w,"{\"result\":\"%v\"}", ite_Set_ScreenMode(r.FormValue("mac"),mode))
	w.(http.Flusher).Flush()
}

func ite_api_setOsdMode(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")//delete this line when FW is publishing **************************
	w.WriteHeader(http.StatusOK)
	mode, _ := strconv.Atoi(r.FormValue("osd_mode"))
	var b_mode byte = (byte)(mode)

	fmt.Fprintf(w,"{\"result\":\"%v\"}", ite_Set_Osd_Mode(r.FormValue("mac"),b_mode))
	w.(http.Flusher).Flush()
}

func ite_api_setVideoBitRate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")//delete this line when FW is publishing **************************
	w.WriteHeader(http.StatusOK)
	group, _ := strconv.Atoi(r.FormValue("video_group"))
	var b_group byte = (byte)(group)
	bitRate, _ := strconv.Atoi(r.FormValue("bit_rate"))


	fmt.Fprintf(w,"{\"result\":\"%v\"}", ite_Set_Group_Video_Bitrate(r.FormValue("mac"),b_group,bitRate))
	w.(http.Flusher).Flush()
}

func ite_api_setVideoOutMode(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")//delete this line when FW is publishing **************************
	w.WriteHeader(http.StatusOK)
	mode, _ := strconv.Atoi(r.FormValue("video_out_mode"))
	var b_mode byte = (byte)(mode)
	fmt.Fprintf(w,"{\"result\":\"%v\"}", ite_Set_Video_Out_Mode(r.FormValue("mac"),b_mode))
	w.(http.Flusher).Flush()
}

func ite_api_setVideoLowDelayMode(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")//delete this line when FW is publishing **************************
	w.WriteHeader(http.StatusOK)
	mode := true
	if r.FormValue("low_delay_mode") == "1"{
	 mode =true
	}else{
	mode = false
	}
	fmt.Fprintf(w,"{\"result\":\"%v\"}", ite_Set_Low_Delay_Mode(r.FormValue("mac"),mode))
	w.(http.Flusher).Flush()
}

func ite_api_setMulticastMode(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")//delete this line when FW is publishing **************************
	w.WriteHeader(http.StatusOK)
	mode := true
	if r.FormValue("multicast_mode") == "1"{
	 mode =true
	}else{
	mode = false
	}
	fmt.Fprintf(w,"{\"result\":\"%v\"}", ite_Set_Multicast_Mode(r.FormValue("mac"),mode))
	w.(http.Flusher).Flush()
}

func ite_api_setDownScale(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")//delete this line when FW is publishing **************************
	w.WriteHeader(http.StatusOK)
	group, _ := strconv.Atoi(r.FormValue("group"))
	var b_group byte = (byte)(group)
	action, _ := strconv.Atoi(r.FormValue("action"))
	var b_action byte = (byte)(action)

	fmt.Fprintf(w,"{\"result\":\"%v\"}", ite_Set_DownScaler_Mode(r.FormValue("mac"),b_group,b_action))
	w.(http.Flusher).Flush()
}

func ite_api_setMcuCommand(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")//delete this line when FW is publishing **************************
	w.WriteHeader(http.StatusOK)

	id, _ := strconv.Atoi(r.FormValue("cmdID"))
	var b_data byte
	if r.FormValue("data")=="true"{
	   b_data = 1
	}else{
		b_data = 0
	}
	
	fmt.Fprintf(w,"{\"result\":\"%v\"}", ite_Set_MCU_Command(r.FormValue("mac"),id,b_data))
	w.(http.Flusher).Flush()
}

func ite_api_factoryReset(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r) 
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w,"{\"result\":\"%v\"}",ite_Set_FactoryReset(vars["mac"]))
	w.(http.Flusher).Flush()
}

func ite_api_Reboot(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r) 
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w,"{\"result\":\"%v\"}",ite_Set_Reboot(vars["mac"]))
	w.(http.Flusher).Flush()
}

func ite_api_node_info(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")//delete this line when FW is publishing **************************
	node_slice := make([]Ite_info, 0, len(ite_info_map))
  for _, tx := range ite_info_map {
      node_slice = append(node_slice, *tx)
  } 
	json_node_list ,_:= json.Marshal(node_slice)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(json_node_list) 
	w.(http.Flusher).Flush()
}

















