package bft

import (
	"Chamael/internal/party"
	"Chamael/pkg/core"
	"Chamael/pkg/protobuf"
	"Chamael/pkg/utils"
	"fmt"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/pairing/bn256"
	"go.dedis.ch/kyber/v3/sign/bls"
	"strings"
)

//收集足量的New_View消息后广播Prepare消息
func Prepare_BroadCast(p *party.HonestParty, e uint32, txs []string) {
	var l []int
	seen := make(map[int]bool)
	for {
		//if (len(l) >= int(p.F)*2+1) || (e == 1) {
		if (len(l) >= int(p.N)-1) || (e == 1) {
			fmt.Println("New View ", e, "start")
			break
		}
		m := <-p.GetMessage("New_View", utils.Uint32ToBytes(e))
		if !seen[int(m.Sender)] {
			l = append(l, int(m.Sender))
			seen[int(m.Sender)] = true
		}
	}
	PrepareMessage := core.Encapsulation("Prepare", utils.Uint32ToBytes(e), p.PID, &protobuf.Prepare{
		Txs: txs,
	})
	p.Intra_Broadcast(PrepareMessage)

}

//收集足量的Prepare_Vote消息,验证AggSig1(txs||vote1||epoch)后广播Precommit消息
func Precommit_BroadCast(p *party.HonestParty, e uint32, txs []string) {
	suite := bn256.NewSuite()
	var l []int
	seen := make(map[int]bool)
	var signatures [][]byte
	var pubkeys []kyber.Point
	for {
		m := <-p.GetMessage("Prepare_Vote", utils.Uint32ToBytes(e))
		payload := (core.Decapsulation("Prepare_Vote", m)).(*protobuf.Prepare_Vote)
		if !seen[int(m.Sender)] {
			l = append(l, int(m.Sender))
			seen[int(m.Sender)] = true
			signatures = append(signatures, payload.Sig)
			pubkeys = append(pubkeys, p.PK[m.Sender])
		}
		//if len(l) >= int(p.F)*2+1 {

		if len(l) >= int(p.N)-1 {
			break
		}
	}
	aggSig, _ := bls.AggregateSignatures(suite, signatures...)
	aggPubKey := bls.AggregatePublicKeys(suite, pubkeys...)
	local := utils.MessageEncap([][]byte{[]byte(strings.Join(txs, "")), utils.Uint32ToBytes(1), utils.Uint32ToBytes(e)})
	err := bls.Verify(suite, aggPubKey, local, aggSig)
	if err != nil {
		fmt.Println("AggSig1(txs||vote1||epoch) verification failed(Malicious Participator):", err)
		return
	}

	PrecommitMessage := core.Encapsulation("Precommit", utils.Uint32ToBytes(e), p.PID, &protobuf.Precommit{
		Aggsig: aggSig,
		Aggpk:  utils.PointToBytes(aggPubKey),
	})
	p.Intra_Broadcast(PrecommitMessage)
}

//收集足量的Precommit_Vote消息,验证AggSig2(vote2||epoch)后广播Commit消息
func Commit_BroadCast(p *party.HonestParty, e uint32, txs []string, outputChannel chan []string) {
	suite := bn256.NewSuite()
	var l []int
	seen := make(map[int]bool)
	var signatures [][]byte
	var pubkeys []kyber.Point
	for {
		m := <-p.GetMessage("Precommit_Vote", utils.Uint32ToBytes(e))
		payload := (core.Decapsulation("Precommit_Vote", m)).(*protobuf.Precommit_Vote)
		if !seen[int(m.Sender)] {
			l = append(l, int(m.Sender))
			seen[int(m.Sender)] = true
			signatures = append(signatures, payload.Sig)
			pubkeys = append(pubkeys, p.PK[m.Sender])
		}
		//if len(l) >= int(p.F)*2+1 {
		if len(l) >= int(p.N)-1 {
			break
		}
	}
	aggSig, _ := bls.AggregateSignatures(suite, signatures...)
	aggPubKey := bls.AggregatePublicKeys(suite, pubkeys...)
	local := utils.MessageEncap([][]byte{utils.Uint32ToBytes(1), utils.Uint32ToBytes(e)})
	err := bls.Verify(suite, aggPubKey, local, aggSig)
	if err != nil {
		fmt.Println("AggSig2(vote2||epoch) verification failed(Malicious Participator):", err)
		return
	}

	CommitMessage := core.Encapsulation("Commit", utils.Uint32ToBytes(e), p.PID, &protobuf.Commit{
		Aggsig: aggSig,
		Aggpk:  utils.PointToBytes(aggPubKey),
	})
	p.Intra_Broadcast(CommitMessage)
	outputChannel <- txs
}

func HotStuffProcess(p *party.HonestParty, epoch int, inputChannel chan []string, outputChannel chan []string) {
	suite := bn256.NewSuite()
	e := uint32(epoch)
	var txs []string //处理自己作为Leader时提议的交易集合;从inputchannel来,所以是[]String
	var Txs []byte   //处理自己作为普通参与者时接收的交易集合;只供验签使用,所以用[]byte
	var is_leader bool
	if (e-1)%p.N == p.SID {
		is_leader = true
		txs = <-inputChannel
	} else {
		is_leader = false
	}
	if is_leader == true { //自己作为领导者时
		//收集足量的New_View消息后广播Prepare消息
		Prepare_BroadCast(p, e, txs)
		//收集足量的Prepare_Vote消息,验证AggSig1(txs||vote1||epoch)后广播Precommit消息
		Precommit_BroadCast(p, e, txs)
		//收集足量的Precommit_Vote消息,验证AggSig2(vote2||epoch)后广播Commit消息并把Txs放入输出通道
		Commit_BroadCast(p, e, txs, outputChannel)

	} else { //自己作为普通参与节点时
	Loop:
		for {
			select {
			//收到Prepare消息,签sig1(txs||vote1||epoch)并回复Prepare_Vote消息
			case m := <-p.GetMessage("Prepare", utils.Uint32ToBytes(e)):
				payload := (core.Decapsulation("Prepare", m)).(*protobuf.Prepare)
				txs = payload.Txs
				Txs = []byte(strings.Join(txs, ""))
				var vote uint32
				vote = 1
				smessage := utils.MessageEncap([][]byte{Txs, utils.Uint32ToBytes(vote), utils.Uint32ToBytes(e)})

				sigPrepare, _ := bls.Sign(suite, p.SK, smessage) //sign(txs||vote1||epoch)
				Prepare_VoteMessage := core.Encapsulation("Prepare_Vote", utils.Uint32ToBytes(e), p.PID, &protobuf.Prepare_Vote{
					Vote: vote,
					Sig:  sigPrepare,
				})
				p.Send(Prepare_VoteMessage, m.Sender)
			//收到Precommit消息,验证aggsig1(txs||vote1||epoch),签sig2(vote2||epoch)并回复Precommit_Vote消息
			case m := <-p.GetMessage("Precommit", utils.Uint32ToBytes(e)):
				payload := (core.Decapsulation("Precommit", m)).(*protobuf.Precommit)
				sver := utils.MessageEncap([][]byte{Txs, utils.Uint32ToBytes(1), utils.Uint32ToBytes(e)})
				AggPK := utils.BytesToPoint(payload.Aggpk)
				err := bls.Verify(suite, AggPK, sver, payload.Aggsig)
				if err != nil {
					fmt.Println("AggSig1(txs||vote1||epoch) verification failed(Malicious Leader):", err)
					return
				}

				var vote uint32
				vote = 1
				smessage := utils.MessageEncap([][]byte{utils.Uint32ToBytes(vote), utils.Uint32ToBytes(e)})

				sigPrepare, _ := bls.Sign(suite, p.SK, smessage) //sign(vote2||epoch)
				Precommit_VoteMessage := core.Encapsulation("Precommit_Vote", utils.Uint32ToBytes(e), p.PID, &protobuf.Precommit_Vote{
					Vote: vote,
					Sig:  sigPrepare,
				})
				p.Send(Precommit_VoteMessage, m.Sender)
			//收到Commit消息,验证aggsig2(vote2||epoch)并回复New_View消息;
			case m := <-p.GetMessage("Commit", utils.Uint32ToBytes(e)):
				payload := (core.Decapsulation("Commit", m)).(*protobuf.Commit)
				sver := utils.MessageEncap([][]byte{utils.Uint32ToBytes(1), utils.Uint32ToBytes(e)})
				AggPK := utils.BytesToPoint(payload.Aggpk)
				err := bls.Verify(suite, AggPK, sver, payload.Aggsig)
				if err != nil {
					fmt.Println("AggSig2(vote2||epoch) verification failed(Malicious Leader):", err)
					return
				}
				New_ViewMessage := core.Encapsulation("New_View", utils.Uint32ToBytes(e+1), p.PID, &protobuf.New_View{
					None: make([]byte, 0),
				})
				//p.Send(New_ViewMessage, m.Sender)
				p.Intra_Broadcast(New_ViewMessage)
				outputChannel <- txs
				break Loop
			}
		}
	}

}
