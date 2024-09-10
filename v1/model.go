package v1

import "github.com/rs/zerolog"

type Base64Str string

type Gree struct {
	debug    bool
	log      *zerolog.Logger
	host     string
	port     uint16
	cid      string
	secKey   string
	tryLimit int
	alias    map[string][]string
}

type RequestPack struct {
	T    string      `json:"t"`
	P    []int       `json:"p,omitempty"`
	Cols []string    `json:"cols,omitempty"`
	Mac  string      `json:"mac"`
	Uid  uint        `json:"uid,omitempty"`
	Opt  []string    `json:"opt,omitempty"`
	Val  map[int]any `json:"val,omitempty"`
}

type Request struct {
	T      string     `json:"t"`
	I      uint8      `json:"i"`
	Uid    uint8      `json:"uid"`
	Cid    string     `json:"cid"`
	Tcid   string     `json:"tcid"`
	Pack   *Base64Str `json:"pack,omitempty"`
	toPack *RequestPack
}

func NewRequest(t string) Request {
	return Request{T: t}
}

type EncodedResponse struct {
	T    string     `json:"t"`
	I    uint8      `json:"i"`
	UID  uint8      `json:"uid"`
	Cid  string     `json:"cid"`
	Tcid string     `json:"tcid"`
	Pack *Base64Str `json:"pack"`
}

type Response struct {
	T    string        `json:"t"`
	I    uint8         `json:"i,omitempty"`
	UID  uint8         `json:"uid,omitempty"`
	Cid  string        `json:"cid"`
	Tcid string        `json:"tcid,omitempty"`
	Pack *ResponsePack `json:"pack,omitempty"`
}

type ResponsePack struct {
	T       string   `json:"t"`
	Cid     string   `json:"cid,omitempty"`
	Bc      string   `json:"bc,omitempty"`
	Brand   string   `json:"brand,omitempty"`
	Catalog string   `json:"catalog,omitempty"`
	Mac     string   `json:"mac"`
	Key     string   `json:"key,omitempty"`
	Mid     string   `json:"mid,omitempty"`
	Model   string   `json:"model,omitempty"`
	Name    string   `json:"name,omitempty"`
	Series  string   `json:"series,omitempty"`
	Vender  string   `json:"vender,omitempty"`
	Ver     string   `json:"ver,omitempty"`
	Lock    uint8    `json:"lock,omitempty"`
	R       uint8    `json:"r,omitempty"`
	Opt     []string `json:"opt,omitempty"`
	P       []int    `json:"p,omitempty"`
	Cols    []string `json:"cols,omitempty"`
	Dat     []int    `json:"dat,omitempty"`
	Val     []int    `json:"val,omitempty"`
}
