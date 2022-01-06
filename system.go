package main
import (
    "fmt"
    "time"
	"net"
    "net/http"
    "strings"
    "encoding/json"
    "io/ioutil"
    "os/exec"
	"crypto/md5"
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
	"log"
	"os"
	"strconv"
	"github.com/gorilla/mux"
	"encoding/base64"
	"net/smtp"
	"net/mail"
	"crypto/tls"
	"errors"
	"encoding/binary"
	"bufio"
)

var TimezoneList = [...]string {
"Africa/Cairo",
"Africa/Johannesburg",
"Africa/Lagos",
"America/Anchorage",
"America/Argentina/Buenos_Aires",
"America/Chicago",
"America/Denver",
"America/Detroit",
"America/Noronha",
"America/Los_Angeles",
"America/Tijuana",
"America/Mexico_City",
"America/Montevideo",
"America/Panama",
"America/Phoenix",
"America/Costa_Rica",
"America/Puerto_Rico",
"America/Santiago",
"America/Santo_Domingo",
"America/St_Johns",
"America/Vancouver",
"America/Whitehorse",
"America/Danmarkshavn",
"Asia/Amman",
"Asia/Baghdad",
"Asia/Bangkok",
"Asia/Shanghai",
"Asia/Beirut",
"Asia/Qatar",
"Asia/Dubai",
"Asia/Dushanbe",
"Asia/Ho_Chi_Minh",
"Asia/Hong_Kong",
"Asia/Jakarta",
"Asia/Jerusalem",
"Asia/Kabul",
"Asia/Karachi",
"Asia/Kolkata",
"Australia/Perth",
"Asia/Pyongyang",
"Asia/Aden",
"Asia/Seoul",
"Asia/Singapore",
"Asia/Taipei",//44
"Asia/Tbilisi",
"Asia/Tehran",
"Asia/Thimphu",
"Asia/Tokyo",
"Asia/Yangon",
"Asia/Urumqi",
"Europe/Amsterdam",
"Europe/Athens",
"Europe/Berlin",
"Europe/Brussels",
"Europe/Budapest",
"Europe/Copenhagen",
"Europe/Dublin",
"Europe/Helsinki",
"Europe/Istanbul",
"Europe/Kaliningrad",
"Europe/Lisbon",
"Europe/London",
"Europe/Madrid",
"Europe/Moscow",
"Europe/Paris",
"Europe/Rome",
"Europe/Stockholm",
"Europe/Vienna",
"Pacific/Pitcairn",
"Australia/Adelaide",
"Pacific/Niue",
"Pacific/Auckland",
"Pacific/Rarotonga",
"Pacific/Norfolk",
"Australia/Melbourne",
"Pacific/Pago_Pago",
"Pacific/Honolulu",
"Pacific/Efate",
"Australia/Sydney",
"Pacific/Marquesas",
"Pacific/Niue",//use for UTC options - start
"Pacific/Honolulu",
"Pacific/Marquesas",
"America/Anchorage",
"America/Los_Angeles",
"America/Denver",
"America/Chicago",
"America/Detroit",
"America/Puerto_Rico",
"America/St_Johns",
"America/Argentina/Buenos_Aires",
"America/Noronha",
"Atlantic/Azores",
"Europe/London",//UTC 0
"Africa/Lagos",
"Africa/Cairo",
"Asia/Aden",
"Asia/Tehran",
"Asia/Dubai",
"Asia/Kabul",
"Asia/Dushanbe",
"Asia/Kolkata",
"Asia/Kathmandu",
"Asia/Thimphu",
"Asia/Yangon",
"Asia/Bangkok",
"Asia/Taipei",
"Australia/Eucla",
"Asia/Pyongyang",
"Australia/Adelaide",
"Australia/Sydney",
"Australia/Lord_Howe",
"Pacific/Norfolk",
"Pacific/Auckland",//use for UTC options - end
}

const (
    file  = "/proc/net/route"
    line  = 1    // line containing the gateway addr. (first line: 0)
    sep   = "\t" // field separator
    field = 2    // field containing hex gateway address (first field: 0)
)


type systemConfig struct {
	UI string `json:"ui"`
	Timezone string `json:"tz"`
	IP_mode string `json:"ip_mode"`
	TIME string  `json:"time"`
	IP string `json:"ip"`
	MASK string `json:"mask"`
	GATEWAY string `json:"gw"`
}



type Event_log struct {
	Type string `json:"type"`
	Time string `json:"time"`
	Event_body string `json:"event"`
}

type emialGroup struct {
	INDEX string `json:"index"`
	NAME string `json:"name"`
	EMAIL string `json:"email"`
}

type emialList struct {
	Mail []emialGroup `json:"mail"`
}

type event_node struct {
	EMAIL string `json:"email_index"`
	MAC string  `json:"mac"`
	TYPE string  `json:"type"`
}

type LoginAuth struct {
    username, password string
}

type videowall_pair struct {
    NAME string `json:"name"`
	MAC string `json:"mac"`
	POS string `json:"pos"`
	VS_SHIFT string `json:"vs_shift"`
	VE_SHIFT string `json:"ve_shift"`
	HS_SHIFT string `json:"hs_shift"`
	HE_SHIFT string `json:"he_shift"`
}

type videowall_preset struct {
	NAME string `json:"name"`
	INDEX string `json:"index"`
	TX string `json:"tx_mac"`
	ROW string `json:"row"`
	COL string `json:"col"`


	RX []videowall_pair `json:"rx_list"`
}


var event_connect_map =make(map[string]*event_node)
var event_videolost_map =make(map[string]*event_node)


var event_num , event_index int = 0 , 0
var event_array[520] Event_log
var pi4_mac string
var pi4_password string ="admin"
var eventEmail_map[8] emialGroup
var vw_preset[8] videowall_preset
var fw_file string
var sys_config systemConfig

func encodeRFC2047(String string) string{
	// use mail's rfc2047 to encode any string
	addr := mail.Address{String, ""}
	return strings.Trim(addr.String(), " <>")
}



func api_modify_vw_Preset(w http.ResponseWriter, r *http.Request) {

	var p videowall_preset

	w.Header().Set("Access-Control-Allow-Origin", "*")//delete this line when FW is publishing **************************
	w.WriteHeader(http.StatusOK)

	err := json.NewDecoder(r.Body).Decode(&p)
	if err!=nil{
		fmt.Println("error")
		fmt.Fprintf(w,"{\"result\":\"%v\"}", "DAIL_ERR")
	}else{
	fmt.Println(p)
	sys_save_vw_Preset(p.INDEX,p)
	fmt.Fprintf(w,"{\"result\":\"%v\"}", "OK")
	}
	w.(http.Flusher).Flush()
}



func sys_save_vw_Preset(index string,p videowall_preset){

k, _ := strconv.Atoi(index)
if k>8{
	return
}
path := "./preset/vw_preset"+ index + "_info.json"
vw_preset[k-1] = p
json_preset ,_:= json.Marshal(vw_preset[k-1])
_ = ioutil.WriteFile(path,json_preset, 0644)

}

func sys_load_vw_Preset(){

var path string
	for i := 0; i < 8; i++ {
		index := strconv.Itoa(i+1)
		path = "./preset/vw_preset"+ index + "_info.json"
		file, err := ioutil.ReadFile(path)
		if err == nil {
			//var vw_slice videowall_preset
			_ = json.Unmarshal([]byte(file), &vw_preset[i])
			//fmt.Println(vw_preset[i].NAME)
		}else{
		vw_preset[i].NAME = "preset"+index
		fmt.Println(err)
		}
	}
}


func api_list_vw_Preset(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")//delete this line when FW is publishing **************************

	node_slice := make([]videowall_preset, 0, 8)
	for i := 0; i < 8; i++ {
	 node_slice = append(node_slice, vw_preset[i])
	}
	json_node_list ,_:= json.Marshal(node_slice)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(json_node_list)
	w.(http.Flusher).Flush()
}




func get_pi4_ipconfig(){



	ipCnt := 0
	/*mgmtInterface, err := net.InterfaceByName("eth0")
    if err != nil {
        fmt.Println("Unable to find interface")
        os.Exit(-1)
    }*/

	addrs, err := net.InterfaceAddrs()//mgmtInterface.Addrs()
	if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

	var tmp_ip,tmp_mask string

	for _, addr := range addrs {

		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
        var ip net.IP
        var mask net.IPMask
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
		cleanMask := fmt.Sprintf("%d.%d.%d.%d", mask[0], mask[1], mask[2], mask[3])
        //fmt.Println(ip, cleanMask,ipCnt)
		ipCnt = ipCnt+1

		if ipCnt>2 {
		continue
		}
		tmp_ip = ip.String()
		tmp_mask = cleanMask

		}
    }

	if sys_config.IP_mode == "STATIC"{

		 if tmp_ip!=sys_config.IP{

		 fmt.Println("static IP not found",tmp_ip,sys_config.IP)
		 cmd := exec.Command("sudo","ip","adderss","add",sys_config.IP+"/24","dev eth0:0")
		 out,err := cmd.Output()
		 if err != nil {
			fmt.Println(err)
		 }
		 fmt.Println(string(out))
		 }


	}else{

		sys_config.IP = tmp_ip
		sys_config.MASK = tmp_mask

	}


	//
	file, err := os.Open(file)
    if err != nil {
        log.Fatal(err)
    }
    defer file.Close()

    scanner := bufio.NewScanner(file)
    for scanner.Scan() {

        // jump to line containing the agteway address
        for i := 0; i < line; i++ {
            scanner.Scan()
        }

        // get field containing gateway address
        tokens := strings.Split(scanner.Text(), sep)

        fmt.Println("field = ", field)
        fmt.Println("len(tokens) = ", len(tokens))

        if(len(tokens) >= 2){
            fmt.Println("enter ~~~~~~~~~~~~~")
            gatewayHex := "0x" + tokens[field]

            // cast hex address to uint32
            d, _ := strconv.ParseInt(gatewayHex, 0, 64)
            d32 := uint32(d)

            // make net.IP address from uint32
            ipd32 := make(net.IP, 4)
            binary.LittleEndian.PutUint32(ipd32, d32)
                  //fmt.Printf("%T --> %[1]v\n", ipd32)

            // format net.IP to dotted ipV4 string
            gw := net.IP(ipd32).String()
          	sys_config.GATEWAY = gw
            //fmt.Printf("%T --> %[1]v\n", ip)

            // exit scanner
        }


        break
    }



}

func (a *LoginAuth) Start(server *smtp.ServerInfo) (string, []byte, error) {
    return "LOGIN", []byte{}, nil
}

func NewLoginAuth(username, password string) smtp.Auth {
    return &LoginAuth{username, password}
}

func (a *LoginAuth) Next(fromServer []byte, more bool) ([]byte, error) {
    if more {
        switch string(fromServer) {
        case "Username:":
            return []byte(a.username), nil
        case "Password:":
            return []byte(a.password), nil
        default:
            return nil, errors.New("Unknown fromServer")
        }
    }
    return nil, nil
}



func SendGMail(Msg []byte,index string){
    fmt.Println("SendGMail~~~~~~~~~~~")

	k, _ := strconv.Atoi(index)
	if k>8{
	return
	}

	var mailTo []string
	mailTo = strings.Split(eventEmail_map[k-1].EMAIL, ";")
	mailTo = mailTo[:len(mailTo)-1]
	fmt.Println(mailTo,len(mailTo))

	if len(mailTo)==0{
		return
	}

    //addr := "mse.gomax-electronics.com.tw:25"
	c, err := smtp.Dial("mse.gomax-electronics.com.tw:25")
    host, _, _ := net.SplitHostPort("mse.gomax-electronics.com.tw:25")
	if err != nil {
        fmt.Println("call dial")
        return
    }
    defer c.Close()
	if ok, _ := c.Extension("STARTTLS"); ok {
        config := &tls.Config{ServerName: host, InsecureSkipVerify: true}
        if err = c.StartTLS(config); err != nil {
            fmt.Println("call start tls")
            return
        }
    }

	auth := NewLoginAuth("pi4control@gomax-electronics.com.tw","Iym#g052")

	if ok, _ := c.Extension("AUTH"); ok {
            if err = c.Auth(auth); err != nil {
                fmt.Println("check auth with err:", err)
                return
            }
    }

	from := "pi4control@gomax-electronics.com.tw"

	if err = c.Mail(from); err != nil {
	    fmt.Println("mail err:", err)
        return
    }

	//to := []string{"pokemongoplayer20160817@gmail.com","kdjljfsjf@gmail.com"}
	//fmt.Println(to)
	//mailTo
    for _, addr := range mailTo {
        if err = c.Rcpt(addr); err != nil {
			 fmt.Println("range err:", err)
            return
        }
    }
    w, err := c.Data()
    if err != nil {
		fmt.Println("Data err:", err)
        return
    }


	//msg := []byte("astpi4 test")

	header := make(map[string]string)
    header["Subject"] = "Aspeed System notification"
    header["MIME-Version"] = "1.0"
    header["Content-Type"] = "text/plain; charset=\"utf-8\""
    header["Content-Transfer-Encoding"] = "base64"
    message := ""
    fmt.Println("before Enter header while k = " , k )
    for k, v := range header {
        fmt.Println("log = %s: %s\r\n", k, v)
        message += fmt.Sprintf("%s: %s\r\n", k, v)
    }
    message += "\r\n" + base64.StdEncoding.EncodeToString(Msg)


    _, err = w.Write([]byte(message))

	if err != nil {
		fmt.Println("write err:", err)
        return
    }
    err = w.Close()
    if err != nil {
        return
    }
    c.Quit()


}

func SendMail(Msg,index string){

    fmt.Println("SendMail~~~~~~~~~~~")
	k, _ := strconv.Atoi(index)
	if k>8{
	return
	}

	mailServer := "pi4control@gomax-electronics.com.tw"
	smtpServer := "mse.gomax-electronics.com.tw"//"smtp.gmail.com"
	//host mail
	auth := smtp.PlainAuth(
		"",
		mailServer,//"XXX@gmail.com",//account
		"Iym#g052",                    //password
		smtpServer,
	)
	from := mail.Address{"ststem", mailServer}
	mailTo := strings.Split(eventEmail_map[k-1].EMAIL, ";")
	/*to := []string {
		"pokemongoplayer20160817@gmail.com",
		"pokemongoplayer20160822@gmail.com",
		}*/
	//to := mail.Address{"reciever", "pokemongoplayer20160817@gmail.com"}

	title := "ast event"
	header := make(map[string]string)
	header["From"] = from.String()
	header["To"] = "administrator"//to[0]//to.String()
	header["Subject"] = encodeRFC2047(title)
	header["MIME-Version"] = "1.0"
	header["Content-Type"] = "text/plain; charset=\"utf-8\""
	header["Content-Transfer-Encoding"] = "base64"

	message := ""
	for k, v := range header {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + base64.StdEncoding.EncodeToString([]byte(Msg))
	err := smtp.SendMail(
		smtpServer + ":25",
		auth,
		from.Address,
		mailTo,//[]string{to.Address},
		[]byte(message),
		//[]byte("This is the email body."),
	)
	if err != nil {
		log.Fatal(err)
	}

}

func chmodApp(){
	cmd := exec.Command("sudo","chmod","-R","777","/home/pi/app/")
	out,err := cmd.Output()
    if err != nil {
        fmt.Println(err)
    }
	fmt.Println(string(out))
}

func runShell(){

	command := `sh update.sh`
    //cmd := exec.Command("/bin/sh", "-c", command)
    cmd := exec.Command("sh", "update.sh")
    output, err := cmd.Output()
    if err != nil {
        fmt.Printf("Execute Shell:%s failed with error:%s", command, err.Error())
        return
    }
    fmt.Println("Execute Shell:%s finished with output:\n%s", command, string(output))

	/*
	cmd := exec.Command("unzip","-o","/home/pi/golang/controller/tmp/fw.zip","-d","/home/pi/app/","&")
	out,err := cmd.Output()
    if err != nil {
        fmt.Println(err)
    }
	fmt.Println(string(out))
	//fmt.Println("Execute Shell:%s", cmd)*/
}

func sync(){
    cmd := exec.Command("sync")
	out,err := cmd.Output()
    if err != nil {
        fmt.Println(err)
    }
	fmt.Println(string(out))
    //fmt.Println("Execute Shell:%s", cmd)
}

func fw_update(total,index, md5Data , data64 string)(string){

	b64_ := strings.Replace(data64, " ", "+",-1)

	if index=="1"{//clear
	fw_file ="";
	}
	fw_file = fw_file + b64_

	m_data := []byte(b64_)
	md5_Data := md5.Sum(m_data)

	md5str1 := fmt.Sprintf("%X", md5_Data) //byte to hex
	fmt.Println("Ori:"+md5Data)
	fmt.Println("data:"+md5str1)
	if md5Data!=md5str1{
	return "DAIL_ERR"
	}

	if total == index{

	unbased, err := base64.StdEncoding.DecodeString(fw_file)
	if err != nil {
    	panic("Cannot decode b64")
		return "DAIL_ERR"
	}
	_ = ioutil.WriteFile("./tmp/fw.zip",unbased, 0644)
	fmt.Println("wait for save firmware...")
	time.Sleep(3000 * time.Millisecond)
	sync()
	time.Sleep(2000 * time.Millisecond)
	fmt.Println("start update...")
	go runShell()
	time.Sleep(3000 * time.Millisecond)
	sync()
	time.Sleep(1000 * time.Millisecond)
	chmodApp()
	}
	return "OK";
}

func api_fw_update(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Access-Control-Allow-Origin", "*")//delete this line when FW is publishing **************************
	w.WriteHeader(http.StatusOK)

	fmt.Fprintf(w,"{\"result\":\"%v\"}", fw_update(r.FormValue("total"),r.FormValue("index"),r.FormValue("md5"),r.FormValue("base64")))
	w.(http.Flusher).Flush()

}

func event_device_del(mac string){

	delete(event_connect_map,mac)
	delete(event_videolost_map,mac)
	saveEvent("connection")
	saveEvent("videolost")
}


func check_connection_lost(mac string){

	if event_connect_map[mac]!=nil{
	Msg :=  ast_info_map[mac].Host_name + " connection loss"
	var Array_Msg[]byte = []byte(Msg)
	SendGMail(Array_Msg,event_connect_map[mac].EMAIL)
	fmt.Println(mac+" connection lost")
	}
	//send email
}

func check_video_lost(mac string){

	if event_videolost_map[mac]!=nil{

	Msg :=  ast_info_map[mac].Host_name + " video loss"
	var Array_Msg[]byte = []byte(Msg)
	SendGMail(Array_Msg,event_videolost_map[mac].EMAIL)
	fmt.Println(mac+" video lost")
	}
	//send email
}

func api_del_connection_lost(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w,"{\"result\":\"%v\"}",del_connection_lost("connection",vars["mac"]))
	w.(http.Flusher).Flush()
}

func api_del_video_lost(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w,"{\"result\":\"%v\"}",del_connection_lost("videolost",vars["mac"]))
	w.(http.Flusher).Flush()
}

func del_connection_lost(eventType,mac string) (string) {

	if eventType=="connection"{
	fmt.Println("delete connection lost event "+mac)
	delete(event_connect_map,mac)
	saveEvent("connection")
	}else{
	fmt.Println("delete video lost event "+mac)
	delete(event_videolost_map,mac)
	saveEvent("videolost")
	}


	return "OK"
}

func api_get_connection_lost(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")//delete this line when FW is publishing **************************
	node_slice := make([]event_node, 0, len(event_connect_map))
  for _, tx := range event_connect_map {
      node_slice = append(node_slice, *tx)
  }
	json_node_list ,_:= json.Marshal(node_slice)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(json_node_list)
	w.(http.Flusher).Flush()
}

func api_get_video_lost(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")//delete this line when FW is publishing **************************
	node_slice := make([]event_node, 0, len(event_videolost_map))
  for _, tx := range event_videolost_map {
      node_slice = append(node_slice, *tx)
  }
	json_node_list ,_:= json.Marshal(node_slice)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(json_node_list)
	w.(http.Flusher).Flush()
}

func api_update_connection_lost(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")//delete this line when FW is publishing **************************
	w.WriteHeader(http.StatusOK)

	var p event_node
	err := json.NewDecoder(r.Body).Decode(&p)
	if err!=nil{
		fmt.Println("error")
		fmt.Fprintf(w,"{\"result\":\"%v\"}", "DAIL_ERR")
	}else{
	fmt.Fprintf(w,"{\"result\":\"%v\"}", updateEvent("connection",p))
	}
	w.(http.Flusher).Flush()
}

func api_update_video_lost(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")//delete this line when FW is publishing **************************
	w.WriteHeader(http.StatusOK)

	var p event_node
	err := json.NewDecoder(r.Body).Decode(&p)
	if err!=nil{
		fmt.Println("error")
		fmt.Fprintf(w,"{\"result\":\"%v\"}", "DAIL_ERR")
	}else{
	fmt.Fprintf(w,"{\"result\":\"%v\"}", updateEvent("videolost",p))
	}
	w.(http.Flusher).Flush()
}

func loadEvent(){

	file, err := ioutil.ReadFile("./event/connectionEvent.json")
	//fmt.Println(err)
	if err == nil {
		var node_slice []event_node
		_ = json.Unmarshal([]byte(file), &node_slice)
		for i := 0; i < len(node_slice); i++ {
		event_connect_map[node_slice[i].MAC] = &node_slice[i]
		}
	}
	file, err = ioutil.ReadFile("./event/videolostEvent.json")
	//fmt.Println(err)
	if err == nil {
		var video_slice []event_node
		_ = json.Unmarshal([]byte(file), &video_slice)
		for i := 0; i < len(video_slice); i++ {
		event_videolost_map[video_slice[i].MAC] = &video_slice[i]
		}
	}
}

func saveEvent(event string){

  if(event=="connection"){
	  node_slice := make([]event_node, 0, len(event_connect_map))
	  for _, tx := range event_connect_map {
		  node_slice = append(node_slice, *tx)
	  }
		event_list ,_:= json.Marshal(node_slice)
		_ = ioutil.WriteFile("./event/connectionEvent.json",event_list, 0644)
		}else{
	  node_slice := make([]event_node, 0, len(event_videolost_map))
	  for _, tx := range event_videolost_map {
		  node_slice = append(node_slice, *tx)
	  }
		event_list ,_:= json.Marshal(node_slice)
		_ = ioutil.WriteFile("./event/videolostEvent.json",event_list, 0644)

	}
}

func updateEvent (event string, m_list event_node)(string){

	tmp := event_node{}
	tmp.EMAIL = m_list.EMAIL
	tmp.MAC = m_list.MAC
	tmp.TYPE = m_list.TYPE

	fmt.Println("mail:"+tmp.EMAIL)
	fmt.Println("mac:"+tmp.MAC)
	fmt.Println("type:"+tmp.TYPE)

	if(event=="connection"){
		if map_tmp, ok := event_connect_map[tmp.MAC]; ok {
		fmt.Println("modify")
		map_tmp.EMAIL = m_list.EMAIL
		map_tmp.MAC = m_list.MAC
		map_tmp.TYPE = m_list.TYPE
		}else{
		fmt.Println("new")
		event_connect_map[tmp.MAC] = &tmp
		}
		saveEvent("connection")
		/*
		event_slice := make([]event_node, 0, len(event_connect_map))
		for _, tx := range event_connect_map {
		event_slice = append(event_slice, *tx)
		}
		event_list ,_:= json.Marshal(event_slice)
		_ = ioutil.WriteFile("./event/connectionEvent.json",event_list, 0644)*/
	}else{
		if map_tmp, ok := event_videolost_map[tmp.MAC]; ok {
		fmt.Println("modify")
		map_tmp.EMAIL = m_list.EMAIL
		map_tmp.MAC = m_list.MAC
		map_tmp.TYPE = m_list.TYPE
		}else{
		fmt.Println("new event")
		event_videolost_map[tmp.MAC] = &tmp
		}
		saveEvent("videolost")
	}

	return "OK"

}

func loadEmail(){

 for i := 0; i < 8; i++ {
	index := strconv.Itoa(i+1)
	path := "./email/group"+index+".dat"
	content, err := ioutil.ReadFile(path)
	if err != nil {
      eventEmail_map[i].INDEX = index
	  eventEmail_map[i].NAME = "group"+index

	  json_email ,_:= json.Marshal(eventEmail_map[i])
	  _ = ioutil.WriteFile(path,json_email, 0644)
      fmt.Println("create new email group"+index)
	  continue
    }
    _ = json.Unmarshal([]byte(content), &eventEmail_map[i])
	}
}


func api_modifyEmail(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")//delete this line when FW is publishing **************************
	w.WriteHeader(http.StatusOK)

	var p emialGroup
	err := json.NewDecoder(r.Body).Decode(&p)
	if err!=nil{
		fmt.Println("error")
		fmt.Fprintf(w,"{\"result\":\"%v\"}", "DAIL_ERR")
	}else{

	k, _ := strconv.Atoi(p.INDEX)
	if k>8{
	fmt.Fprintf(w,"{\"result\":\"DAIL_ERR\"}")
	}else{
	fmt.Fprintf(w,"{\"result\":\"%v\"}", "OK")
	eventEmail_map[k-1] = p
	path := "./email/group"+p.INDEX+".dat"
	json_email ,_:= json.Marshal(eventEmail_map[k-1])
	_ = ioutil.WriteFile(path,json_email, 0644)
	fmt.Println("modify group"+p.INDEX+",name:"+eventEmail_map[k-1].NAME)
	}
	}
	w.(http.Flusher).Flush()


}

func api_listEmail(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")//delete this line when FW is publishing **************************
	var email_array emialList
	email_array.Mail = nil
	for i := 0; i < 8; i++ {
	email_array.Mail = append(email_array.Mail,eventEmail_map[i])
	}
	//email_slice := make([]Ast_preset_array, 0, 1)
	//preset_slice = append(preset_slice, preset_array)
	json_node_list ,_:= json.Marshal(email_array)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(json_node_list)
	w.(http.Flusher).Flush()
}


func add_event(event_type, event_body string) {

	if event_num == 500 && event_index ==500{
	event_index = 0
	//event_num--
	}


	event_array[event_index] = Event_log{event_type,time.Now().Format("200601021504"),event_body}
	if event_num < 500{
	event_num++
	}
	event_index++
	history_list_save()
}

func api_history_list(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")//delete this line when FW is publishing **************************
	//event_list_save()
	event_slice := event_array[:event_num]
	json_event_list ,_:= json.Marshal(event_slice)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(json_event_list)
}

func api_history_delete(w http.ResponseWriter, r *http.Request) {
	event_num = 0
	event_index = 0
	history_list_save()
	add_event("History","To delete all history")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w,"{\"result\":\"OK\"}",)
	w.(http.Flusher).Flush()
}

func history_list_save(){
	event_slice := event_array[0:event_num]
	json_event_list ,_:= json.Marshal(event_slice)
	_ = ioutil.WriteFile("event_list.json",json_event_list, 0644)
}

func history_list_load(){
	file, _ := ioutil.ReadFile("event_list.json")
	var event_slice []Event_log
	_ = json.Unmarshal([]byte(file), &event_slice)
	for i := 0; i < len(event_slice); i++ {
		event_array[i] = event_slice[i]
		//fmt.Println(event_array[i])
	}
	event_index = len(event_slice)
	event_num = len(event_slice)

}

func api_pi4_set_ip(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")//delete this line when FW is publishing **************************
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w,"{\"result\":\"%v\"}", pi4_set_ip(r.FormValue("ip"),r.FormValue("mask"),r.FormValue("gw")))
	w.(http.Flusher).Flush()
}

func api_pi4_set_dhcp(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")//delete this line when FW is publishing **************************
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w,"{\"result\":\"%v\"}", pi4_set_dhcp())
	w.(http.Flusher).Flush()
}

func pi4_set_dhcp() (string) {

	cmd := exec.Command("sudo","sed","-i","10c iface eth0 inet manual","/etc/network/interfaces")
	_,_ = cmd.Output()
	cmd  = exec.Command("sudo","sed","-i","11c up dhclient -nw eth0","/etc/network/interfaces")
	_,_ = cmd.Output()
	param := "12c #address"
	cmd  = exec.Command("sudo","sed","-i",param,"/etc/network/interfaces")
	_,_ = cmd.Output()
	param = "13c #netmask"
	cmd  = exec.Command("sudo","sed","-i",param,"/etc/network/interfaces")
	_,_ = cmd.Output()
	param = "14c #gateway"
	cmd  = exec.Command("sudo","sed","-i",param,"/etc/network/interfaces")
	_,_ = cmd.Output()
	param = "15c #dns-nameservers 8.8.8.8"
	cmd  = exec.Command("sudo","sed","-i",param,"/etc/network/interfaces")
	_,_ = cmd.Output()
    //removed dhcp
    /*cmd3 := exec.Command("sudo","dhclient","-r","eth0")
    _,err3 := cmd3.Output()
    if err3 != nil {
        fmt.Println(err3)
    }
    //enable dhcp
    _ = exec.Command("sudo","dhclient","-nw","eth0")
    _,err4 := cmd3.Output()
    if err4 != nil {
        fmt.Println(err4)
    }*/
    return "OK"
}

func pi4_set_ip(ip,mask,gw string) (string) {

    cmd := exec.Command("sudo","sed","-i","10c iface eth0 inet static","/etc/network/interfaces")
	_,_ = cmd.Output()
    cmd  = exec.Command("sudo","sed","-i","11c #dhcp","/etc/network/interfaces")
	_,_ = cmd.Output()
	param := "12c address "+ip
	cmd  = exec.Command("sudo","sed","-i",param,"/etc/network/interfaces")
	_,_ = cmd.Output()
	param = "13c netmask "+mask
	cmd  = exec.Command("sudo","sed","-i",param,"/etc/network/interfaces")
	_,_ = cmd.Output()
	param = "14c gateway "+gw
	cmd  = exec.Command("sudo","sed","-i",param,"/etc/network/interfaces")
	_,_ = cmd.Output()
	param = "15c dns-nameservers 8.8.8.8"
	cmd  = exec.Command("sudo","sed","-i",param,"/etc/network/interfaces")
	_,_ = cmd.Output()

    //static ip
    //cmd3 := exec.Command("sudo","ifconfig","eth0",ip,"netmask","255.255.255.0")//,netmask, mask
    //route add default gw, gateway
    //_,err3 := cmd3.Output()
    //if err3 != nil {
    //    fmt.Println(err3)
    //}


    return "OK"
}




func api_get_mac(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")//delete this line when FW is publishing **************************


	j_mac := map[string]interface{}{
              "mac": pi4_mac,
              }

	json_mac ,_:= json.Marshal(j_mac)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(json_mac)
}

func getMacAddrs() (macAddrs string) {
    netInterfaces, err := net.Interfaces()
    if err != nil {
        fmt.Printf("fail to get net interfaces: %v", err)
        return ""
    }

    for _, netInterface := range netInterfaces {
        macAddr := netInterface.HardwareAddr.String()
        if len(macAddr) == 0 {
            continue
        }

        macAddrs = macAddr
		macAddrs = strings.Replace(string(macAddrs[:]), ":", "",-1)
		macAddrs = strings.ToUpper(macAddrs)
		return macAddrs

    }
    return macAddrs
}

func system_check_login(s string)(string){

str := pi4_mac+"admin"+pi4_password
data := []byte(str)
md5Data := md5.Sum(data)
md5str1 := fmt.Sprintf("%X", md5Data) //byte to hex
fmt.Println(md5str1)

if md5str1 == s {

return "OK"
}
return "DAIL_ERR"
}

func system_changePassword(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")//delete this line when FW is publishing **************************
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w,"{\"result\":\"%v\"}", ChangePassword(r.FormValue("old"),r.FormValue("new")))
	w.(http.Flusher).Flush()
}

func api_login(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")//delete this line when FW is publishing **************************
    w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w,"{\"result\":\"%v\"}", system_check_login(r.FormValue("password")))

	w.(http.Flusher).Flush()
}

func LoadPassword(){

 content, err := ioutil.ReadFile("./tmp/account.dat")
 if err != nil {
      pi4_password = "admin"
	  SavePassword()
      fmt.Println("password file not exist create new one")
	  return
    }
  pi4_password = string(content)
  Strkey := pi4_mac+"1234"
  Dencrypted , _:= hex.DecodeString(pi4_password )
  var key[]byte = []byte(Strkey)
  decrypted := AesDecryptCBC(Dencrypted, key)
  pi4_password = string(decrypted)
  log.Println("password:", string(pi4_password))

}

func SavePassword()(string){


	aesKey := pi4_mac+"1234"
	var passwordByte []byte = []byte(pi4_password)
	var key[]byte = []byte(aesKey)
	encrypted := AesEncryptCBC(passwordByte, key)
	strEncrypted := hex.EncodeToString(encrypted)

	f, err := os.Create("./tmp/account.dat")
    if err != nil {
        fmt.Println(err)
        return "DAIL_ERR"
    }
	_, err = f.WriteString(strEncrypted)
    if err != nil {
        fmt.Println(err)
        f.Close()
        return "DAIL_ERR"
    }
	return "OK"
}

func sys_get_config(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")//delete this line when FW is publishing **************************

	sys_config.TIME = GetCurrentTime()
	get_pi4_ipconfig()

  	_json := map[string]interface{}{


	//i_timezone, _ := strconv.Atoi(sys_config.Timezone)
	//s_timezone := strconv.Itoa(i_timezone+1)

              "type": sys_config.UI,
			  "timezone": sys_config.Timezone,
			  "ip_mode": sys_config.IP_mode,
			  "time": sys_config.TIME,
			  "ip":sys_config.IP,
			  "mask":sys_config.MASK,
			  "gateway":sys_config.GATEWAY,


              }
	jsonString ,_:= json.Marshal(_json)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonString)
}

func sys_set_config(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")//delete this line when FW is publishing **************************
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w,"{\"result\":\"%v\"}",sys_config_ctrl(r.FormValue("type"),r.FormValue("timezone"),r.FormValue("ip_mode"),r.FormValue("ip"),r.FormValue("mask"),r.FormValue("gateway")))
	w.(http.Flusher).Flush()
}

func SetTimezone(zone string){

	cmd := exec.Command("sudo","timedatectl","set-timezone",zone)  //America/Adak Asia/Taipei
	_,err := cmd.Output()
    if err != nil {
        fmt.Println(err)
    }
	fmt.Println("set timezone to ",zone)
}

func GetCurrentTime()(string){

	cmd := exec.Command("date","+%s")
	out,err := cmd.Output()
    if err != nil {
        fmt.Println(err)
		return "error"
    }else{

	return string(out)
	}

}

func sys_config_ctrl(mappingMode, timezone, ip_mode, ip, mask, gw string)(string){

	dataChanged := false
	if sys_config.UI!= mappingMode{
	sys_config.UI = mappingMode
	dataChanged = true;
	}

	if k, err := strconv.Atoi(timezone); err == nil {
		if sys_config.Timezone!= timezone{
		sys_config.Timezone = timezone
		SetTimezone(TimezoneList[k-1])

		dataChanged = true;
	}
	}else{
	fmt.Println("error timezone index")
	return "DAIL_ERR"
	}



	if sys_config.IP_mode != ip_mode{
	   sys_config.IP_mode = ip_mode
	   dataChanged = true;
	   if sys_config.IP_mode == "DHCP"{
	   pi4_set_dhcp()
	   fmt.Println("DHCP ON")
	   }else{
	   pi4_set_ip(ip,mask,gw)
	     fmt.Println("static ON")
	   }
	   fmt.Println("ip mode: ",ip_mode)
	}else{

	//ethernet settings
	if sys_config.IP_mode == "STATIC"{
		if sys_config.IP != ip || sys_config.MASK != mask || sys_config.GATEWAY != gw{
	     pi4_set_ip(ip,mask,gw)
	     fmt.Println("static changed")
		}
	}

	}
	if dataChanged==true {
	SaveSystemConfig()
	}
	return "OK"


}

func LoadSystemConfig(){

		content, err := ioutil.ReadFile("./tmp/config.dat")
		if err != nil {
		 sys_config.UI = "true"
		 sys_config.Timezone = "1"
		 sys_config.IP_mode = "DHCP"
		 sys_config.IP = ""
		 sys_config.MASK = ""
		 sys_config.GATEWAY = ""

		 SaveSystemConfig()
		 fmt.Println("create new system config")
		}else{
		_ = json.Unmarshal([]byte(content), &sys_config)
		fmt.Println(sys_config.UI,sys_config.Timezone,sys_config.IP_mode)
		}
}

func SaveSystemConfig(){
	path := "./tmp/config.dat"
	json_config ,_:= json.Marshal(sys_config)
	_ = ioutil.WriteFile(path,json_config, 0644)
}

func ChangePassword(old, new string)(string){

	str := pi4_mac+"admin"+pi4_password
	data := []byte(str)
	md5Data := md5.Sum(data)
	md5str1 := fmt.Sprintf("%X", md5Data) //byte to hex
	fmt.Println(md5str1)
	if md5str1 != old {
	return "DAIL_ERR"
	}
	fmt.Println("old password verify OK")
	pi4_password = new
	result := SavePassword()
	return result
}


// =================== CBC ======================
func AesEncryptCBC(origData []byte, key []byte) (encrypted []byte) {
	//
	block, _ := aes.NewCipher(key)
	blockSize := block.BlockSize()                              //
	origData = pkcs5Padding(origData, blockSize)                //
	blockMode := cipher.NewCBCEncrypter(block, key[:blockSize]) //
	encrypted = make([]byte, len(origData))                     //
	blockMode.CryptBlocks(encrypted, origData)                  //
	return encrypted
}
func AesDecryptCBC(encrypted []byte, key []byte) (decrypted []byte) {
	block, _ := aes.NewCipher(key)                              //
	blockSize := block.BlockSize()                              //
	blockMode := cipher.NewCBCDecrypter(block, key[:blockSize]) //
	decrypted = make([]byte, len(encrypted))                    //
	blockMode.CryptBlocks(decrypted, encrypted)                 //
	decrypted = pkcs5UnPadding(decrypted)                       //
	return decrypted
}
func pkcs5Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}
func pkcs5UnPadding(origData []byte) []byte {
	length := len(origData)
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}

func GetSystemConfig() systemConfig{
     return sys_config
}

func GetSystemFWVersion() string{
     return "FW 1.1"
}