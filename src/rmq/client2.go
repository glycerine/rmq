package main

/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements. See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership. The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License. You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

import (
	"fmt"
	"os"

	"crypto/tls"
	"tutorial"

	"git.apache.org/thrift.git/lib/go/thrift"
)

func handleClient(client *tutorial.CalculatorClient) (err error) {
	client.Ping()
	fmt.Println("ping()")

	sum, _ := client.Add(1, 1)
	fmt.Print("1+1=", sum, "\n")

	work := tutorial.NewWork()
	work.Op = tutorial.Operation_DIVIDE
	work.Num1 = 1
	work.Num2 = 0
	quotient, err := client.Calculate(1, work)
	if err != nil {
		switch v := err.(type) {
		case *tutorial.InvalidOperation:
			fmt.Println("Invalid operation:", v)
		default:
			fmt.Println("Error during operation:", err)
		}
		return err
	} else {
		fmt.Println("Whoa we can divide by 0 with new value:", quotient)
	}

	work.Op = tutorial.Operation_SUBTRACT
	work.Num1 = 15
	work.Num2 = 10
	diff, err := client.Calculate(1, work)
	if err != nil {
		switch v := err.(type) {
		case *tutorial.InvalidOperation:
			fmt.Println("Invalid operation:", v)
		default:
			fmt.Println("Error during operation:", err)
		}
		return err
	} else {
		fmt.Print("15-10=", diff, "\n")
	}

	log, err := client.GetStruct(1)
	if err != nil {
		fmt.Println("Unable to get struct:", err)
		return err
	} else {
		fmt.Println("Check log:", log.Value)
	}
	return err
}

type RmqClient struct {
	transport        thrift.TTransport
	transportFactory thrift.TTransportFactory
	protocolFactory  thrift.TProtocolFactory
	addr             string
	secure           bool
}

func NewRmqClient(transportFactory thrift.TTransportFactory, protocolFactory thrift.TProtocolFactory, addr string, secure bool) (*RmqClient, error) {
	c := RmqClient{
		transportFactory: transportFactory,
		protocolFactory:  protocolFactory,
		secure:           secure,
		addr:             addr,
	}
	var err error
	if secure {
		cfg := new(tls.Config)
		cfg.InsecureSkipVerify = true
		c.transport, err = thrift.NewTSSLSocket(addr, cfg)
	} else {
		c.transport, err = thrift.NewTSocket(addr)
	}
	if err != nil {
		fmt.Println("Error opening socket:", err)
		return nil, err
	}
	if c.transport == nil {
		return nil, fmt.Errorf("Error opening socket, got nil transport. Is server available?")
	}
	c.transport = transportFactory.GetTransport(c.transport)
	if c.transport == nil {
		return nil, fmt.Errorf("Error from transportFactory.GetTransport(), got nil transport. Is server available?")
	}

	err = c.transport.Open()
	if err != nil {
		return nil, err
	}

	return &c, nil
}

func (c *RmqClient) doClientCall() error {
	return handleClient(tutorial.NewCalculatorClientFactory(c.transport, c.protocolFactory))
}

var myRmqClient *RmqClient

func StartClient() error {

	protocol := "compact" // flag.String("P", "binary", "Specify the protocol (binary, compact, json, simplejson)")
	framed := true        // flag.Bool("framed", false, "Use framed transport")
	buffered := true      // flag.Bool("buffered", false, "Use buffered transport"), non-blocking server requires this.
	addr := "localhost:9090"
	secure := false // flag.Bool("secure", false, "Use tls secure transport")

	var protocolFactory thrift.TProtocolFactory
	switch protocol {
	case "compact":
		protocolFactory = thrift.NewTCompactProtocolFactory()
	case "simplejson":
		protocolFactory = thrift.NewTSimpleJSONProtocolFactory()
	case "json":
		protocolFactory = thrift.NewTJSONProtocolFactory()
	case "binary", "":
		protocolFactory = thrift.NewTBinaryProtocolFactoryDefault()
	default:
		fmt.Fprint(os.Stderr, "Invalid protocol specified", protocol, "\n")
		os.Exit(1)
	}

	var transportFactory thrift.TTransportFactory
	if buffered {
		transportFactory = thrift.NewTBufferedTransportFactory(8192)
	} else {
		transportFactory = thrift.NewTTransportFactory()
	}

	if framed {
		transportFactory = thrift.NewTFramedTransportFactory(transportFactory)
	}

	var err error
	if myRmqClient == nil {
		myRmqClient, err = NewRmqClient(transportFactory, protocolFactory, addr, secure)
		if err != nil {
			myRmqClient = nil
			return err
		}
	}
	return myRmqClient.doClientCall()
}
