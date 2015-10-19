package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/glycerine/turnpike"
)

var client *turnpike.Client

func server_main() {
	turnpike.Debug()
	s := turnpike.NewBasicWebsocketServer("turnpike.examples")
	server := &http.Server{
		Handler: s,
		Addr:    ":8000",
	}
	client, _ = s.GetLocalClient("turnpike.examples", nil)
	if err := client.BasicRegister("alarm.set", alarmSet); err != nil {
		panic(err)
	}
	log.Println("turnpike server starting on port 8000")
	log.Fatal(server.ListenAndServe())
}

var myCount int64 = 0

// takes one argument, the (integer) number of seconds to set the alarm for
func alarmSet(args []interface{}, kwargs map[string]interface{}) (result *turnpike.CallResult) {
	fmt.Printf("args[0] is '%#v' of type %T\n", args[0], args[0]) // getting a float64. arg.
	_, ok := args[0].(float64)
	if !ok {
		fmt.Printf("first arg to alarmSet() was not an float64.\n")
		return &turnpike.CallResult{Err: turnpike.URI("rpc-example.invalid-argument")}
	}
	/*	go func() {
			time.Sleep(time.Duration(duration) * time.Second)
			client.Publish("alarm.ring", nil, nil)
		}()
	*/
	myCount++
	return &turnpike.CallResult{Args: []interface{}{myCount}}
}
