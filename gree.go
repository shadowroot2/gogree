package gogree

import (
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/forgoer/openssl"
	"github.com/goccy/go-json"
	"github.com/rs/zerolog"
	"net"
	"strconv"
	"strings"
	"time"
)

// NewGree create an object of Gree
func NewGree(logger *zerolog.Logger, host string, cid string, secKey string) *Gree {
	return &Gree{
		debug:    false,
		log:      logger,
		host:     host,
		port:     7000,
		cid:      cid,
		secKey:   secKey,
		tryLimit: 3,
		alias: map[string][]string{
			"Pow":        {"off", "on"},
			"Mod":        {"auto", "cool", "dry", "fan", "heat"},
			"TemUn":      {"celsius", "fahrenheit"},
			"WdSpd":      {"auto", "low", "medium-low", "medium", "medium-high", "high"},
			"Air":        {"off", "on"},
			"Blo":        {"off", "on"},
			"Health":     {"off", "on"},
			"SwhSlp":     {"off", "on"},
			"Lig":        {"off", "on"},
			"SwingLfRig": {"default", "full swing", "pos 1", "pos 2", "pos 3", "pos 4", "pos 5"},
			"SwUpDn": {
				"default",
				"full swing",
				"upmost position",
				"middle-up position",
				"middle position",
				"middle-low position",
				"lowest position",
				"downmost region",
				"middle-low region",
				"middle region",
				"middle-up region",
				"upmost region",
			},
			"Quiet": {"off", "on"},
			"Tur":   {"off", "on"},
			"SvSt":  {"off", "on"},
			"StHt":  {"off", "on"},
		},
	}
}

// SetDebug debug set
func (g *Gree) SetDebug(debug bool) {
	g.debug = debug
}

// SetHost host set
func (g *Gree) SetHost(host string) {
	g.host = host
}

// SetPort port set
func (g *Gree) SetPort(port uint16) {
	g.port = port
}

// SetCid cid set
func (g *Gree) SetCid(cid string) {
	g.cid = cid
}

// SetSecKey secKey set
func (g *Gree) SetSecKey(secKey string) {
	g.secKey = secKey
}

// Scan function
func (g *Gree) Scan() error {

	// Request
	request := NewRequest("scan")

	// Send request
	response, err := g.sendRequest(request, false)
	if err != nil {
		err = errors.New("Can not send request: " + err.Error())
		return err
	}

	// Setting CID
	if response.Pack.Mac != response.Cid {
		return errors.New("can not find CID")
	}

	if g.debug {
		// JSON total
		jsonStr, _ := json.Marshal(response)
		fmt.Println("Decrypted JSON:")
		fmt.Println(string(jsonStr))
	}

	g.SetCid(response.Pack.Mac)
	g.log.Info().Str("CID", response.Pack.Mac).Msg("found")

	return nil
}

// GetBindKey function
func (g *Gree) GetBindKey() (key string, err error) {

	request := NewRequest("bind")
	request.I = 1
	request.Tcid = g.cid

	// Send request
	response, err := g.sendRequest(request, true)
	if err != nil {
		err = errors.New("can not get bind key from response:" + err.Error())
		return "", err
	}

	// Check response value
	if response.Pack.T != "bindok" || response.Pack.Key == "" {
		return "", errors.New("can not get sec-key")
	}

	// Remember the key
	g.SetSecKey(response.Pack.Key)
	g.log.Info().Str("bind key", response.Pack.Key).Msg("found bind key")

	return g.secKey, nil
}

// On switch on AC
func (g *Gree) On() (map[string]any, error) {
	request := NewRequest("cmd")
	request.toPack = &RequestPack{
		Opt: []string{"Pow"},
		P:   []int{1},
	}

	// Sending request
	response, err := g.sendRequest(request, true)
	if err != nil {
		err = errors.New("can switch on AC " + err.Error())
		return nil, err
	}

	// Check response value
	if len(response.Pack.Val) == 0 {
		return nil, errors.New("can not get Pow value from response")
	}

	// Value
	value := response.Pack.Val[0]
	if len(g.alias["Pow"]) < value {
		return nil, errors.New("can not get alias for Pow value " + strconv.Itoa(value))
	}

	return map[string]any{"Pow": g.alias["Pow"][value]}, nil
}

// Off switch off AC
func (g *Gree) Off() (map[string]any, error) {

	request := NewRequest("cmd")
	request.toPack = &RequestPack{
		Opt: []string{"Pow"},
		P:   []int{0},
	}

	// Sending request
	response, err := g.sendRequest(request, true)
	if err != nil {
		err = errors.New("can not switch off AC: " + err.Error())
		return nil, err
	}

	// Check response value
	if len(response.Pack.Val) == 0 {
		return nil, errors.New("can not get Pow value from response")
	}

	// Value
	value := response.Pack.Val[0]
	if len(g.alias["Pow"]) < value {
		return nil, errors.New("can not get alias for Pow value " + strconv.Itoa(value))
	}

	return map[string]any{"Pow": g.alias["Pow"][value]}, nil
}

// Status get current status values of AC
func (g *Gree) Status() (map[string]any, error) {

	request := NewRequest("status")
	request.toPack = &RequestPack{
		Cols: []string{
			"Pow",
			"Mod",
			"SetTem",
			"WdSpd",
			"Air",
			"Blo",
			"Health",
			"SwhSlp",
			"Lig",
			"SwingLfRig",
			"SwUpDn",
			"Quiet",
			"Tur",
			"StHt",
			"TemUn",
			"HeatCoolType",
			"TemRec",
			"SvSt",
		},
	}

	// Sending request
	response, err := g.sendRequest(request, true)
	if err != nil {
		err = errors.New("can not send get status: " + err.Error())
		return nil, err
	}

	// Formatting status
	if len(response.Pack.Cols) == 0 || len(response.Pack.Dat) != len(response.Pack.Cols) {
		return nil, errors.New("can not get status from response")
	}

	status := make(map[string]any)
	for i, col := range response.Pack.Cols {

		// Value from response
		v := response.Pack.Dat[i]

		// Alias value
		if len(g.alias[col]) > 0 && len(g.alias[col][0]) >= v {
			status[col] = g.alias[col][v]
		} else {
			status[col] = v
		}
	}

	return status, nil
}

// GetPower current Pow status
func (g *Gree) GetPower() (map[string]any, error) {

	request := NewRequest("status")
	request.toPack = &RequestPack{
		Cols: []string{"Pow"},
	}

	// Sending request
	response, err := g.sendRequest(request, true)
	if err != nil {
		err = errors.New("Can not get status of power" + err.Error())
		return nil, err
	}

	// Check response value
	if len(response.Pack.Dat) == 0 {

		return nil, errors.New("can not get Pow value from response")
	}

	value := response.Pack.Dat[0]
	if len(g.alias["Pow"]) < value {
		return nil, errors.New("can not get alias for Pow value " + strconv.Itoa(value))
	}

	return map[string]any{"Pow": g.alias["Pow"][value]}, nil
}

// GetMode get current AC mode
func (g *Gree) GetMode() (map[string]any, error) {

	request := NewRequest("status")
	request.toPack = &RequestPack{
		Cols: []string{"Mod"},
	}

	// Sending request
	response, err := g.sendRequest(request, true)
	if err != nil {
		err = errors.New("can not get mode" + err.Error())
		return nil, err
	}

	// Check response value
	if len(response.Pack.Dat) == 0 {
		return nil, errors.New("can not get Pow value from response")
	}

	// Value
	value := response.Pack.Dat[0]
	if len(g.alias["Mod"]) < value {
		return nil, errors.New("can not get alias for Mod value " + strconv.Itoa(value))
	}

	return map[string]any{"Mod": g.alias["Mod"][value]}, nil
}

// SetMode set work-mode
func (g *Gree) SetMode(mode string) (map[string]any, error) {

	// Numeric value if Mod
	mods := ArrayFlip(g.alias["Mod"])
	_, ok := mods[mode]
	if !ok {
		return nil, errors.New("can not find mod " + mode)
	}

	request := NewRequest("cmd")
	request.toPack = &RequestPack{
		Opt: []string{"Mod"},
		P:   []int{mods[mode]},
	}

	// Sending request
	response, err := g.sendRequest(request, true)
	if err != nil {
		err = errors.New("can not set mode" + err.Error())
		return nil, err
	}

	// Check response value
	if len(response.Pack.Val) == 0 {
		return nil, errors.New("can not get Mod value from response")
	}

	// Value
	value := response.Pack.Val[0]
	if len(g.alias["Mod"]) < value {
		return nil, errors.New("can not get alias for Mod value " + strconv.Itoa(value))
	}

	return map[string]any{"Mod": g.alias["Mod"][value]}, nil
}

// GetVentSpeed get current vent speed
func (g *Gree) GetVentSpeed() (map[string]any, error) {

	request := NewRequest("status")
	request.toPack = &RequestPack{
		Cols: []string{"WdSpd"},
	}

	// Sending request
	response, err := g.sendRequest(request, true)
	if err != nil {
		err = errors.New("can not get vent speed" + err.Error())
		return nil, err
	}

	// Check response value
	if response.Pack.Dat == nil || len(response.Pack.Dat) == 0 {
		err = errors.New("can not get vent speed from response")
		return nil, err
	}

	// Value
	value := response.Pack.Dat[0]

	// Check for alias
	if len(g.alias["WdSpd"]) < value {
		return nil, errors.New("alias " + strconv.Itoa(value) + "for WdSpd not found")
	}

	return map[string]any{"WdSpd": g.alias["WdSpd"][value]}, nil
}

// SetVentSpeed set vent speed
func (g *Gree) SetVentSpeed(speed string) (map[string]any, error) {

	// Numeric value of WdSpd
	speeds := ArrayFlip(g.alias["WdSpd"])
	iSpeed, ok := speeds[speed]
	if !ok {
		err := errors.New("incorrect speed value WdSpd" + speed)
		return nil, err
	}

	request := NewRequest("cmd")
	request.toPack = &RequestPack{
		Opt: []string{"WdSpd"},
		P:   []int{iSpeed},
	}

	// Sending request
	response, err := g.sendRequest(request, true)
	if err != nil {
		err = errors.New("can not set vent speed" + err.Error())
		return nil, err
	}

	// Check response value
	if len(response.Pack.Val) == 0 {
		return nil, errors.New("can not get WdSpd value from response")
	}

	// Value
	value := response.Pack.Val[0]
	if len(g.alias["WdSpd"]) < value {
		return nil, errors.New("alias " + strconv.Itoa(value) + " for WdSpd not found")
	}

	return map[string]any{"WdSpd": g.alias["WdSpd"][value]}, nil
}

// GetTemperature get current target temperature
func (g *Gree) GetTemperature() (map[string]any, error) {

	request := NewRequest("status")
	request.toPack = &RequestPack{
		Cols: []string{"SetTem", "Add0.5"},
	}

	// Sending request
	response, err := g.sendRequest(request, true)
	if err != nil {
		err = errors.New("can not get temperature" + err.Error())
		return nil, err
	}

	// Check response
	if len(response.Pack.Dat) < 2 {
		return nil, errors.New("can not get temperature from response")
	}

	// Values
	v1 := response.Pack.Dat[0]
	v2 := false
	if response.Pack.Dat[1] == 1 {
		v2 = true
	}

	// Status
	return map[string]any{"SetTem": v1, "Add0.5": v2}, nil
}

// SetTemperature temperature in Celsius
// Note: If you need set with 0.5 step, use add05 true or false
func (g *Gree) SetTemperature(temperature int, add05 bool) (map[string]any, error) {

	request := NewRequest("cmd")
	add05int := 0
	if add05 == true {
		add05int = 1
	}

	request.toPack = &RequestPack{
		Opt: []string{"SetTem", "Add0.5"},
		P:   []int{temperature, add05int},
	}

	// Sending request
	response, err := g.sendRequest(request, true)
	if err != nil {
		err = errors.New("can not get temperature" + err.Error())
		return nil, err
	}

	// Check response
	if len(response.Pack.Val) < 2 {
		return nil, errors.New("can not get temperature from response")
	}

	// Values
	v1 := response.Pack.Val[0]
	v2 := false
	if response.Pack.Val[1] == 1 {
		v2 = true
	}

	// Status
	return map[string]any{"SetTem": v1, "Add0.5": v2}, nil
}

// GetHealth status of Health function
func (g *Gree) GetHealth() (map[string]any, error) {

	request := NewRequest("status")
	request.toPack = &RequestPack{
		Cols: []string{"Health"},
	}

	// Sending request
	response, err := g.sendRequest(request, true)
	if err != nil {
		return nil, errors.New("Can not get Health value from response: " + err.Error())
	}

	// Check response value
	if len(response.Pack.Dat) == 0 {
		return nil, errors.New("can not get Health value from response")
	}

	// Value
	value := response.Pack.Dat[0]
	if len(g.alias["Health"]) < value {
		return nil, errors.New("can not get Health alias for " + strconv.Itoa(value))
	}

	return map[string]any{"Health": g.alias["Health"][value]}, nil
}

// SetHealth switch on/off Health function
func (g *Gree) SetHealth(enable bool) (map[string]any, error) {

	sw := 0
	if enable {
		sw = 1
	}

	request := NewRequest("cmd")
	request.toPack = &RequestPack{
		Opt: []string{"Health"},
		P:   []int{sw},
	}

	// Sending request
	response, err := g.sendRequest(request, true)
	if err != nil {
		err = errors.New("can not set Health value from response")
		return nil, err
	}

	// Check response value
	if len(response.Pack.Val) == 0 {
		return nil, errors.New("can not get Health value from response")
	}

	// Value
	value := response.Pack.Val[0]
	if len(g.alias["Health"]) < value {
		return nil, errors.New("can not get alias for Health value " + strconv.Itoa(value))
	}

	return map[string]any{"Health": g.alias["Health"][value]}, nil
}

// Sending request to AC
func (g *Gree) sendRequest(request Request, preformat bool) (*Response, error) {

	// Preformat standard request
	if preformat == true {
		if request.toPack == nil {
			requestPack := &RequestPack{
				T: request.T,
			}
			request.toPack = requestPack
		} else {
			request.toPack.T = request.T
			request.Cid = g.cid
		}
		request.T = "pack"
		request.toPack.Mac = g.cid
	}

	// Encrypt pack if exists pack key
	if request.toPack != nil {
		encoded, err := g.encrypt(request.toPack)
		if err != nil {
			return nil, err
		}
		request.Pack = &encoded
	}

	jsonRequest, err := json.Marshal(request)
	if err != nil {
		err = errors.New("Can not encode JSON request: " + err.Error())
		return nil, err
	}

	g.log.Info().Msg("request to UDP " + g.host + ":" + strconv.Itoa(int(g.port)) + "...")

	// Socket connection open
	jsonResponse := make([]byte, 1024)
	conn, err := net.Dial("udp", g.host+":"+strconv.Itoa(int(g.port)))
	if err != nil {
		err = errors.New("can not dial UDP: " + err.Error())
		return nil, err
	}
	if err = conn.SetDeadline(time.Now().Add(time.Second * 5)); err != nil {
		err = errors.New("can not set deadline: " + err.Error())
		return nil, err
	}

	// Socket connection close on exit
	defer conn.Close()

	// Try send
	encodedResponse := &EncodedResponse{}
	for i := 0; i < g.tryLimit; i++ {

		// 1s sleep before send
		time.Sleep(time.Second)

		// Debug
		if g.debug {
			fmt.Println("Request: " + string(jsonRequest))
		}

		// Send request
		if _, err = conn.Write(jsonRequest); err != nil {
			g.log.Warn().Msg("can not write to UDP: " + err.Error())
			continue
		}

		// Read response
		if _, err = conn.Read(jsonResponse); err != nil {
			g.log.Warn().Msg("cannot read from UPD: " + err.Error())
			continue
		}

		// Decrypting response
		if err = json.Unmarshal(jsonResponse, encodedResponse); err != nil {
			g.log.Warn().Msg("can not decode JSON response: " + err.Error())
			continue
		}

		// Debug
		if g.debug {
			fmt.Println("Response: " + string(jsonResponse))
		}
		break
	}

	// Check pack key
	if encodedResponse.T != "pack" {
		return nil, errors.New("can not connect to " + g.host + ":" + strconv.Itoa(int(g.port)))
	}

	// Response
	response := &Response{
		T:    encodedResponse.T,
		I:    encodedResponse.I,
		UID:  encodedResponse.UID,
		Cid:  encodedResponse.Cid,
		Tcid: encodedResponse.Tcid,
	}

	// Decoding pack
	if len(*encodedResponse.Pack) > 16 {
		pack := &ResponsePack{}
		if err = g.decrypt(encodedResponse.Pack, pack); err != nil {
			return nil, err
		}
		response.Pack = pack
	}

	return response, nil
}

// Encryption
func (g *Gree) encrypt(pack *RequestPack) (Base64Str, error) {

	encoded := Base64Str("")

	// JSON encoding
	jsonData, err := json.Marshal(pack)
	if err != nil {
		err = errors.New("cannot marshal data: " + err.Error())
		return encoded, err
	}

	if g.debug {
		fmt.Println("Pak JSON: " + string(jsonData))
	}

	// SSL AES-128-ECB encrypt
	sslData, err := openssl.AesECBEncrypt(jsonData, []byte(g.secKey), openssl.ZEROS_PADDING)
	if err != nil {
		err = errors.New("can not encrypt Pack data: " + err.Error())
		return encoded, err
	}

	// Base64 Encoding
	encoded = Base64Str(base64.StdEncoding.EncodeToString(sslData))

	return encoded, nil
}

// Decryption
func (g *Gree) decrypt(base64String *Base64Str, pack *ResponsePack) error {

	// Base64 decode
	base64Data, err := base64.StdEncoding.DecodeString(string(*base64String))
	if err != nil {
		err = errors.New("can not decode base64 data: " + err.Error())
		return err
	}

	// SSL AES-128-ECB decrypt
	sslData, err := openssl.AesECBDecrypt(base64Data, []byte(g.secKey), openssl.ZEROS_PADDING)
	if err != nil {
		err = errors.New("can not decrypt Pack data: " + err.Error())
		return err
	}

	// Searching the end of JSON
	i := strings.Index(string(sslData), "}")
	if i <= 1 {
		err = errors.New("JSON not found on current data")
		return err
	}

	if g.debug {
		fmt.Println("SSL decrypt data: " + string(sslData[:i+1]))
	}

	// JSON decode
	if err = json.Unmarshal(sslData[:i+1], &pack); err != nil {
		g.log.Warn().Msg("can not decode JSON data: " + err.Error())
		pack.Key = string(sslData)
	}

	return nil
}
