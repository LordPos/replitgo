package replitgo

import (
	api "github.com/LordPos/protocol-go"
	
)


func (c *channel) Exec(args []string) (string, error) {
	e := api.Command_Exec{
			Exec: &api.Exec{
				Args : args,
				Blocking : true,
		},
	}
	return c.getOutput(&e)
	
}