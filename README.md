# replitgo
A Go client for interacting with the [repl.it api](https://crosis.turbio.repl.co)

currently undocumented. 
## code sample
```go
package main

import (
	"fmt"
	"github.com/LordPos/replitgo"
	"bufio"
	"os"
	"strings"
)

func main(){
	key := "key goes here"
	j,_ := replitgo.GetJSON("user_name without the @","repl name")
	id,_ := j["id"].(string) 
	token, _ := replitgo.GetToken(id,key)
	var client replitgo.Client
	url := replitgo.GetURL(token, "eval.repl.it","80", false)
	client.Init(token, "w3", url)
	channel := client.Open("exec","execer")
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("> ")
		s,_ := reader.ReadString('\n')
		a,err := channel.Exec(strings.Split(strings.TrimSpace(s), " "))
		fmt.Println(a)
		if err != nil{
			fmt.Println(err)
		}
	}
}
```
