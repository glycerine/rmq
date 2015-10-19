package main

import (
	"crypto/tls"
	"fmt"
	"os"
	"shared"
	"strconv"
	"tutorial"

	"git.apache.org/thrift.git/lib/go/thrift"
)

type CalculatorHandler struct {
	log map[int]*shared.SharedStruct
}

func NewCalculatorHandler() *CalculatorHandler {
	return &CalculatorHandler{log: make(map[int]*shared.SharedStruct)}
}

func (p *CalculatorHandler) Ping() (err error) {
	fmt.Print("ping()\n")
	return nil
}

func (p *CalculatorHandler) Add(num1 int32, num2 int32) (retval17 int32, err error) {
	fmt.Print("add(", num1, ",", num2, ")\n")
	return num1 + num2, nil
}

func (p *CalculatorHandler) Calculate(logid int32, w *tutorial.Work) (val int32, err error) {
	fmt.Print("calculate(", logid, ", {", w.Op, ",", w.Num1, ",", w.Num2, "})\n")
	switch w.Op {
	case tutorial.Operation_ADD:
		val = w.Num1 + w.Num2
		break
	case tutorial.Operation_SUBTRACT:
		val = w.Num1 - w.Num2
		break
	case tutorial.Operation_MULTIPLY:
		val = w.Num1 * w.Num2
		break
	case tutorial.Operation_DIVIDE:
		if w.Num2 == 0 {
			ouch := tutorial.NewInvalidOperation()
			ouch.WhatOp = int32(w.Op)
			ouch.Why = "Cannot divide by 0"
			err = ouch
			return
		}
		val = w.Num1 / w.Num2
		break
	default:
		ouch := tutorial.NewInvalidOperation()
		ouch.WhatOp = int32(w.Op)
		ouch.Why = "Unknown operation"
		err = ouch
		return
	}
	entry := shared.NewSharedStruct()
	entry.Key = logid
	entry.Value = strconv.Itoa(int(val))
	k := int(logid)
	/*
	   oldvalue, exists := p.log[k]
	   if exists {
	     fmt.Print("Replacing ", oldvalue, " with ", entry, " for key ", k, "\n")
	   } else {
	     fmt.Print("Adding ", entry, " for key ", k, "\n")
	   }
	*/
	p.log[k] = entry
	return val, err
}

func (p *CalculatorHandler) GetStruct(key int32) (*shared.SharedStruct, error) {
	fmt.Print("getStruct(", key, ")\n")
	v, _ := p.log[int(key)]
	return v, nil
}

func (p *CalculatorHandler) Zip() (err error) {
	fmt.Print("zip()\n")
	return nil
}

func main() {

}

func startServer() {
	//	protocol := flag.String("P", "binary", "Specify the protocol (binary, compact, json, simplejson)")
	//	framed := flag.Bool("framed", false, "Use framed transport")
	//	buffered := flag.Bool("buffered", false, "Use buffered transport")
	//	addr := flag.String("addr", "localhost:9090", "Address to listen to")
	//	secure := flag.Bool("secure", false, "Use tls secure transport")

	var protocolFactory thrift.TProtocolFactory
	switch *protocol {
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
	if *buffered {
		transportFactory = thrift.NewTBufferedTransportFactory(8192)
	} else {
		transportFactory = thrift.NewTTransportFactory()
	}

	if *framed {
		transportFactory = thrift.NewTFramedTransportFactory(transportFactory)
	}

	// always run server here
	if err := runServer(transportFactory, protocolFactory, *addr, *secure); err != nil {
		fmt.Println("error running server:", err)
	}
}

func runServer(transportFactory thrift.TTransportFactory, protocolFactory thrift.TProtocolFactory, addr string, secure bool) error {
	var transport thrift.TServerTransport
	var err error
	if secure {
		cfg := new(tls.Config)
		if cert, err := tls.LoadX509KeyPair("keys/server.crt", "keys/server.key"); err == nil {
			cfg.Certificates = append(cfg.Certificates, cert)
		} else {
			return err
		}
		transport, err = thrift.NewTSSLServerSocket(addr, cfg)
	} else {
		transport, err = thrift.NewTServerSocket(addr)
	}

	if err != nil {
		return err
	}
	fmt.Printf("%T\n", transport)
	handler := NewCalculatorHandler()
	processor := tutorial.NewCalculatorProcessor(handler)
	server := thrift.NewTSimpleServer4(processor, transport, transportFactory, protocolFactory)

	fmt.Println("Starting the simple server... on ", addr)
	return server.Serve()
}
