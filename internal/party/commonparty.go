package party

import (
	"Chamael/pkg/core"
	"Chamael/pkg/protobuf"
	"errors"
	"fmt"
	"os"
	"sync"
)

// CommonParty is a struct of normal consensus parties
type CommonParty struct {
	N                 uint32
	F                 uint32
	m                 uint32 //分片个数
	PID               uint32
	Snumber           uint32 //节点所在的分片编号
	SID               uint32 //节点在分片内的编号
	ipList            []string
	portList          []string
	sendChannels      []chan *protobuf.Message
	dispatcheChannels *sync.Map
	ShardList         []int //节点负责沟通的分片
	Debug             bool
}

// NewCommonParty return a new common party object
func NewCommonParty(N uint32, F uint32, m uint32, pid uint32, snum uint32, sid uint32, ipList []string, portList []string, ShardList []int) *CommonParty {
	p := CommonParty{
		N:            N,
		F:            F,
		m:            m, //分片个数
		PID:          pid,
		Snumber:      snum, //节点所在的分片编号
		SID:          sid,  //节点在分片内的编号
		ipList:       ipList,
		portList:     portList,
		sendChannels: make([]chan *protobuf.Message, N*m), //N改成N*m ！
		ShardList:    ShardList,
	}

	return &p
}

// InitReceiveChannel setup the listener and Init the receiveChannel
func (p *CommonParty) InitReceiveChannel() error {
	p.dispatcheChannels = core.MakeDispatcheChannels(core.MakeReceiveChannel(p.portList[p.PID], p.Debug), p.N)
	return nil
}

// InitSendChannel setup the sender and Init the sendChannel, please run this after initializing all party's receiveChannel
func (p *CommonParty) InitSendChannel() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	dirname := fmt.Sprintf(homeDir+"/Chamael/log/%s", p.ipList[p.PID]+":"+p.portList[p.PID])
	os.Mkdir(dirname, 0755)
	for i := uint32(0); i < p.N*p.m; i++ {
		p.sendChannels[i] = core.MakeSendChannel(p.ipList[i], p.portList[i], dirname, p.Debug)
	}
	return nil
}

// Send a message to party des
func (p *CommonParty) Send(m *protobuf.Message, des uint32) error {
	if !p.checkInit() {
		return errors.New("This party hasn't been initialized")
	}
	if des < p.N*p.m {
		p.sendChannels[des] <- m
		return nil
	}
	return errors.New("Destination id is too large")
}

// Broadcast a message to all parties
func (p *CommonParty) Broadcast(m *protobuf.Message) error {
	if !p.checkInit() {
		return errors.New("This party hasn't been initialized")
	}
	for i := uint32(0); i < p.N*p.m; i++ {
		err := p.Send(m, i)
		if err != nil {
			return err
		}
	}
	return nil
}

// Broadcast a message to parties in the same shard
func (p *CommonParty) Intra_Broadcast(m *protobuf.Message) error {
	if !p.checkInit() {
		return errors.New("This party hasn't been initialized")
	}
	for i := p.Snumber * p.N; i < (p.Snumber+1)*p.N; i++ {
		err := p.Send(m, i)
		if err != nil {
			return err
		}
	}
	return nil
}

// Broadcast a message to parties in a specified shard
func (p *CommonParty) Shard_Broadcast(m *protobuf.Message, des uint32) error {
	if !p.checkInit() {
		return errors.New("This party hasn't been initialized")
	}
	for i := des * p.N; i < (des+1)*p.N; i++ {
		err := p.Send(m, i)
		if err != nil {
			return err
		}
	}
	return nil
}

// GetMessage Try to get a message according to messageType, ID
func (p *CommonParty) GetMessage(messageType string, ID []byte) chan *protobuf.Message {
	value1, _ := p.dispatcheChannels.LoadOrStore(messageType, new(sync.Map))

	var value2 any
	if messageType == "Dec" {
		value2, _ = value1.(*sync.Map).LoadOrStore(string(ID), make(chan *protobuf.Message, 100000))
	} else {
		value2, _ = value1.(*sync.Map).LoadOrStore(string(ID), make(chan *protobuf.Message, 1024))
	}

	return value2.(chan *protobuf.Message)
}

func (p *CommonParty) checkInit() bool {
	if p.sendChannels == nil {
		return false
	}
	return true
}
