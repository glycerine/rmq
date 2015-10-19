package main

import (
	"fmt"
	"log"

	"github.com/glycerine/turnpike"
)

var myCallNum int64 = 0

func client_main() {
	turnpike.Debug()
	c, err := turnpike.NewWebsocketClient(turnpike.JSON, "ws://localhost:8000/")
	if err != nil {
		log.Fatal(err)
	}
	_, err = c.JoinRealm("turnpike.examples", nil)
	if err != nil {
		log.Fatal(err)
	}

	/*

		quit := make(chan bool)
		c.Subscribe("alarm.ring", func([]interface{}, map[string]interface{}) {
			fmt.Println("The alarm rang!")
			c.Close()
			quit <- true
		})
			fmt.Print("Enter the timer duration: ")
			scanner := bufio.NewScanner(os.Stdin)
			scanner.Scan()
			if err := scanner.Err(); err != nil {
				log.Fatalln("reading stdin:", err)
			}
			text := scanner.Text()
			if duration, err := strconv.Atoi(text); err != nil {
				log.Fatalln("invalid integer input:", err)
			} else {
	*/

	myCallNum++

	res, err := c.Call("alarm.set", []interface{}{myCallNum}, nil)
	if err != nil {
		fmt.Printf("cli: error setting alarm: '%s'. call result = '%s'.\n", err, res)
	} else {
		fmt.Printf("cli: successfully called alarm.set().\n")
	}

	//	}
	//<-quit
}
