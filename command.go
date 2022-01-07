package main

type CommandObject struct {
	Etag   string      `json:"etag"`
	Time   string      `json:"time"`
	Body   string      `json:"body"`
	Sign   string      `json:"sign"`
	To     string      `json:"to"`
	Extra  string      `json:"extra"`
	Method string      `json:"method"`
}

type DeviceInfoObject struct {
    Mac   string      `json:"mac"`
	IP   string      `json:"ip"`
	Gateway   interface{} `json:"gateway"`
	Mask   string      `json:"mask"`
}

type SwitchChannelObject struct{
    IP   string      `json:"ip"`
    Channel   string      `json:"channel"`
    Type   string      `json:"type"`
}
