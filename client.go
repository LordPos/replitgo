package replitgo

import(
	"golang.org/x/net/websocket"
	"math/rand"
	"math"
	"github.com/martinlindhe/base36"
	api "github.com/LordPos/protocol-go"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"errors"
	"fmt"
	
)

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
	ndata,_ := proto.Marshal(&cmd)
	websocket.Message.Send(c.ws,ndata)
	fmt.Println(" ")
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





