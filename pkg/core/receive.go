package core

import (
	"Chamael/pkg/protobuf"
	"Chamael/pkg/utils"
	"io"
	"log"
	"net"
	"time"

	"google.golang.org/protobuf/proto"
)

// MakeReceiveChannel returns a channel receiving messages
func MakeReceiveChannel(port string) chan *protobuf.Message {
	var addr *net.TCPAddr
	var lis *net.TCPListener
	var err1, err2 error
	retry := true
	//Retry to make listener
	for retry {
		addr, err1 = net.ResolveTCPAddr("tcp4", ":"+port)
		lis, err2 = net.ListenTCP("tcp4", addr)
		if err1 != nil || err2 != nil {
			//log.Println(err1)
			//log.Println(err2)
			time.Sleep(1000)
			retry = true
		} else {
			retry = false
		}
	}
	log.Println("create listener", addr, "success")
	//Make the receive channel and the handle func
	var conn *net.TCPConn
	var err3 error
	receiveChannel := make(chan *protobuf.Message, MAXMESSAGE)
	go func() {

		/////filename := fmt.Sprintf("/home/hyx/Chamael/log/(Received)%s.log", lis.Addr())
		/////file, _ := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		/////fileLogger := log.New(file, "[MessageLogger] ", log.Ldate|log.Ltime|log.Lmicroseconds)
		for {
			//The handle func run forever
			conn, err3 = lis.AcceptTCP()
			conn.SetKeepAlive(true)
			if err3 != nil {
				log.Fatalln(err3, "In receive.go::go func(),AcceptTCP failed")
			}
			//Once connect to a node, make a sub-handle func to handle this connection
			go func(conn *net.TCPConn, channel chan *protobuf.Message) {
				/*
					filename := fmt.Sprintf("/home/hyx/Chamael/log/%s.log", conn.LocalAddr())
					file, _ := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
					fileLogger := log.New(file, "[FileLogger] ", log.Ldate|log.Ltime)
					fileLogger.Println("Logging to a file")
				*/
				for {
					//Receive bytes
					lengthBuf := make([]byte, 4)
					_, err1 := io.ReadFull(conn, lengthBuf)
					length := utils.BytesToInt(lengthBuf)
					buf := make([]byte, length)
					_, err2 := io.ReadFull(conn, buf)

					if err1 != nil || err2 != nil {
						log.Printf("The receive channel of %s (from %s) has break down", conn.LocalAddr(), conn.RemoteAddr())
						return
					}

					//Do Unmarshal
					var m protobuf.Message
					err3 := proto.Unmarshal(buf, &m)
					/////fileLogger.Println(m)
					if err3 != nil {
						log.Fatalln(err3, "In receive.go::go func(),Unmarshal failed")
					}
					//Push protobuf.Message to receivechannel
					(channel) <- &m
				}

			}(conn, receiveChannel)
		}
	}()
	return receiveChannel
}
