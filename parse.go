package replitgo

import (
	api "github.com/LordPos/protocol-go"
	
)

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
