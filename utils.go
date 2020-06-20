package replitgo

import (
	"net/http"
	"encoding/json"
	"io/ioutil"
	"strings"
	

)
//GetJSON : Gets the JSON data for a repl, including repl ID, URL, and some other useful info.
func GetJSON(user string, repl string) (map[string]interface{}, error){

	resp, err := http.Get("https://repl.it/data/repls/@"+user+"/"+repl)
	if err != nil{
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil{
		return nil, err
	}
	var finaljson map[string]interface{}
	json.Unmarshal(body, &finaljson) 
	return finaljson, nil
}

//GetToken : Given a repl ID and an API key, get a one-time token for that repl.
func GetToken(id string, key string) (string, error){
	
	resp, err := http.Post("https://repl.it/api/v0/repls/"+id+"/token", "application/json", strings.NewReader("{ \"apiKey\":\""+key+"\" }"))
	if err != nil{
		return "", err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil{
		return "", err

	}
	return strings.Trim(string(body),"\""),nil
}

//GetURL Gets the websocket connection URL for a given token, host, and port.
//Normally the host will be eval.repl.it and the port will be 80
func GetURL(token, host, port string, secure bool) string{
	k := ""
	if secure { k = "s" }
	return "ws" + k + "://" + host + ":" + port + "/wsv2/" + token
}