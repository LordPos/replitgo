package replitgo

import(
	"net/http"
	"golang.org/x/net/websocket"
	"math/rand"
	"math"
	"github.com/martinlindhe/base36"
	"encoding/json"
	"io/ioutil"
	api "github.com/LordPos/protocol-go"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"errors"
	"strings"
	"fmt"
	
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
type Exec = api.Exec
type Chat = api.ChatMessage
type File = api.File
type Move = api.Move
type Eval = api.Command_Eval
type FileOp struct{
	Op string
	File File
}
func makeBody(b interface{}) api.IsCommand_Body {
	if c,ok := b.(Exec); ok{
		return &api.Command_Exec{Exec : &c}
	}
	if c,ok := b.(Chat); ok{
		return &api.Command_ChatMessage{ChatMessage : &c}
	}
	if c,ok := b.(Move); ok{
		return &api.Command_Move{Move : &c}
	}
	if c,ok := b.(Eval); ok{
		return &c
	}
	if c,ok := b.(FileOp); ok{
		switch c.Op{
		case "read":
			return &api.Command_Read{Read : &c.File}
		case "write":
			return &api.Command_Write{Write : &c.File}
		case "remove":
			return &api.Command_Remove{Remove : &c.File}
		case "tryremove":
			return &api.Command_TryRemove{TryRemove : &c.File}
		case "mkdir":
			return &api.Command_Mkdir{Mkdir : &c.File}
		case "readdir":
			return &api.Command_Readdir{Readdir : &c.File}
		}
	}
	return nil
}

type channel struct {
	id int32
	service string
	name string
	ws *websocket.Conn
}

func (c *channel) Send(data interface{}) ([]api.Command,error) {
	var cmd api.Command
	cmd.Session = 0
	cmd.Channel = c.id
	cmd.Body = makeBody(data)
	fmt.Println(cmd)
	ndata,_ := proto.Marshal(&cmd)
	websocket.Message.Send(c.ws,ndata)

	var got []api.Command
	var res api.Command
	for {
		var b []byte
		websocket.Message.Receive(c.ws, &b)
		proto.Unmarshal(b, &res)
		if res.Channel == c.id {
			got = append(got, res)
		}
		if res.GetState() == 0{
			break
		}

	}
	if res.GetError() == "" {
		return got, nil
	} else {
		return nil, errors.New(res.GetError())
	}
}

func (c *channel) Run(data interface{}) {
	var cmd api.Command
	cmd.Session = 0
	dat := makeBody(data)
	cmd.Channel = c.id
	cmd.Body = dat
	ndata,_ := proto.Marshal(&cmd)
	websocket.Message.Send(c.ws,ndata)

}

func (c *channel) GetOutput(data interface{}) (string, error){
	got, err := c.Send(data)
	if err == nil{
		s := ""
		for _,res := range got{
			s = s + res.GetOutput()
		}
		return s, nil
	} else{
		return "", err
	}
}
//Implement this later
func (c *channel) GetJSON(data map[string]interface{}) {//([]map[string]interface{}, error){
	return

}

type Client struct{
	Token string
	Repl string
	URL string
	ws *websocket.Conn
	channels []int32
}

func (c *Client) Init(token,repl,url string) error{
	c.Token = token
	c.Repl = repl
	c.URL = url
	var err error
	c.ws, err = websocket.Dial(c.URL, "", "https://example.com")
	if err!= nil{
		return err
	}
	c.channels = []int32{}

	return nil
}

func (c *Client) Open(service, name string) channel{
	var cmd api.Command
	cmd.Channel = 0
	cmd.Session = 0
	var oc api.OpenChannel
	oc.Service = service
	oc.Name = name
	oc.Id = 0
	cmd.Body = &api.Command_OpenChan{OpenChan : &oc} 
	cmd.Ref = base36.Encode(uint64(rand.Float32() * float32(math.Pow(10,16))))

	data,_ := proto.Marshal(&cmd)
	websocket.Message.Send(c.ws, data)
	res := &api.OpenChannelRes{}
	for{		
		var data []byte
		websocket.Message.Receive(c.ws, &data)
		proto.Unmarshal(data, res)
		if res.Id > 0{ break }
		
	}
	fmt.Println("here")
	c.channels = append(c.channels, res.Id)

	return channel{id : res.Id, service : service, name : name, ws : c.ws}

}

func (c *Client) Close(){
	for channel := range c.channels{
		var cmd api.Command
		cmd.Channel = 0
		cmd.GetCloseChan().Id = int32(channel)
		cmd.GetCloseChan().Action = 1
		cmd.Ref = base36.Encode(uint64(rand.Float32() * float32(math.Pow(10,16))))
		data := cmd.String()
		websocket.Message.Send(c.ws, data)
		for{
			var res api.Command
			var ndata []byte
			websocket.Message.Receive(c.ws, &ndata)
			jsonpb.UnmarshalString(string(ndata), &res)
			if res.GetCloseChanRes().Id == int32(channel) {break}
		}
	}
	c.ws.Close()
}





