package core

import (
	"Chamael/pkg/protobuf"
	"Chamael/pkg/utils"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"google.golang.org/protobuf/proto"
)

// MAXMESSAGE is the size of channels
var MAXMESSAGE = 1024

// MakeSendChannel returns a channel to send messages to hostIP
func MakeSendChannel(hostIP string, hostPort string, dirname string, Debug bool) chan *protobuf.Message {
	var addr *net.TCPAddr
	var conn *net.TCPConn
	var err1, err2 error
	var fileLogger *log.Logger
	//Retry to connet to node
	retry := true
	for retry {
		addr, err1 = net.ResolveTCPAddr("tcp4", hostIP+":"+hostPort)
		conn, err2 = net.DialTCP("tcp4", nil, addr)
		if err1 != nil {
			retry = true
			time.Sleep(1000)
			continue
		}
		if err2 != nil {
			retry = true
			time.Sleep(1000)
			continue
		}
		retry = false

		conn.SetKeepAlive(true)
	}
	//Make the send channel and the handle func
	sendChannel := make(chan *protobuf.Message, MAXMESSAGE)

	go func(conn *net.TCPConn, channel chan *protobuf.Message) {
		if Debug == true {
			filename := fmt.Sprintf("%s/(Send)%s.log", dirname, conn.RemoteAddr())
			file, _ := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
			fileLogger = log.New(file, "[MessageLogger] ", log.Ldate|log.Ltime|log.Lmicroseconds)
		}
		for {
			//Pop protobuf.Message form sendchannel

			m := <-(channel)
			if Debug == true {
				fileLogger.Println(m)
			}
			//Do Marshal
			byt, err1 := proto.Marshal(m)
			if err1 != nil {
				log.Fatalln(err1)
			}
			//Send bytes

			length := len(byt)
			_, err2 := conn.Write(utils.IntToBytes(length))
			_, err3 := conn.Write(byt)
			if err2 != nil || err3 != nil {
				log.Fatalln("The send channel has break down!", err2)
			}
		}
	}(conn, sendChannel)

	return sendChannel
}
