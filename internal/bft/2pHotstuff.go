package bft

import (
	"Chamael/internal/party"
	"Chamael/pkg/core"
	"Chamael/pkg/protobuf"
	"Chamael/pkg/utils"
	"fmt"
	"strings"

	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/pairing/bn256"
	"go.dedis.ch/kyber/v3/sign/bls"
)

// 收集足量的New_View消息后广播Prepare消息
func Prepare_BroadCast(p *party.HonestParty, e uint32, txs []string, isGlobal bool) {
	var l []int
	seen := make(map[int]bool)

	var threshold int
	if isGlobal {
		F := (int(p.N*p.M) - 1) / 3
		threshold = 2*F + 1
	} else {
		threshold = 2*int(p.F) + 1
	}

	for {
		if (len(l) >= threshold) || (e == 1) {
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
	if isGlobal {
		p.Broadcast(PrepareMessage)
	} else {
		p.Intra_Broadcast(PrepareMessage)
	}

}

// 收集足量的Prepare_Vote消息,验证AggSig1(txs||vote1||epoch)后广播Precommit消息
func Precommit_BroadCast(p *party.HonestParty, e uint32, txs []string, isGlobal bool) {
	suite := bn256.NewSuite()
	var l []int
	seen := make(map[int]bool)
	var signatures [][]byte
	var pubkeys []kyber.Point

	var threshold int
	if isGlobal {
		F := (int(p.N*p.M) - 1) / 3
		threshold = 2*F + 1
	} else {
		threshold = 2*int(p.F) + 1
	}

	for {
		m := <-p.GetMessage("Prepare_Vote", utils.Uint32ToBytes(e))
		payload := (core.Decapsulation("Prepare_Vote", m)).(*protobuf.Prepare_Vote)
		if !seen[int(m.Sender)] {
			l = append(l, int(m.Sender))
			seen[int(m.Sender)] = true
			signatures = append(signatures, payload.Sig)
			pubkeys = append(pubkeys, p.PK[m.Sender])
		}
		if len(l) >= threshold {
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
	if isGlobal {
		p.Broadcast(PrecommitMessage)
	} else {
		p.Intra_Broadcast(PrecommitMessage)
	}
}

// 收集足量的Precommit_Vote消息,验证AggSig2(vote2||epoch)后广播Commit消息
func Commit_BroadCast(p *party.HonestParty, e uint32, txs []string, outputChannel chan []string, isGlobal bool) {
	suite := bn256.NewSuite()
	var l []int
	seen := make(map[int]bool)
	var signatures [][]byte
	var pubkeys []kyber.Point

	var threshold int
	if isGlobal {
		F := (int(p.N*p.M) - 1) / 3
		threshold = 2*F + 1
	} else {
		threshold = 2*int(p.F) + 1
	}

	for {
		m := <-p.GetMessage("Precommit_Vote", utils.Uint32ToBytes(e))
		payload := (core.Decapsulation("Precommit_Vote", m)).(*protobuf.Precommit_Vote)
		if !seen[int(m.Sender)] {
			l = append(l, int(m.Sender))
			seen[int(m.Sender)] = true
			signatures = append(signatures, payload.Sig)
			pubkeys = append(pubkeys, p.PK[m.Sender])
		}
		if len(l) >= threshold {
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
	if isGlobal {
		p.Broadcast(CommitMessage)
	} else {
		p.Intra_Broadcast(CommitMessage)
	}
	outputChannel <- txs
}

// isGlobal: true 全局共识, false 片内共识
func HotStuffProcess(p *party.HonestParty, epoch int, inputChannel chan []string, outputChannel chan []string, isGlobal bool) {
	suite := bn256.NewSuite()
	e := uint32(epoch)
	var txs []string //处理自己作为Leader时提议的交易集合;从inputchannel来,所以是[]String
	var Txs []byte   //处理自己作为普通参与者时接收的交易集合;只供验签使用,所以用[]byte

	var gotPrepare bool = false // 判断是否收到Prepare消息，防止Leader在收到Precommit/Commit消息后，没有收到Prepare消息，导致Txs为空

	// 判断是否是Leader
	var is_leader bool = false
	if isGlobal { // 全局共识，选择 PID = (e-1)%(N*M)
		if (e-1)%(p.N*p.M) == p.PID {
			is_leader = true
			txs = <-inputChannel
		}
	} else { // 片内共识，选择 SID = (e-1)%N
		if (e-1)%p.N == p.SID {
			is_leader = true
			txs = <-inputChannel
		}
	}

	if is_leader == true { //自己作为领导者时
		//收集足量的New_View消息后广播Prepare消息
		Prepare_BroadCast(p, e, txs, isGlobal)
		//收集足量的Prepare_Vote消息,验证AggSig1(txs||vote1||epoch)后广播Precommit消息
		Precommit_BroadCast(p, e, txs, isGlobal)
		//收集足量的Precommit_Vote消息,验证AggSig2(vote2||epoch)后广播Commit消息并把Txs放入输出通道
		Commit_BroadCast(p, e, txs, outputChannel, isGlobal)

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
				gotPrepare = true
			//收到Precommit消息,验证aggsig1(txs||vote1||epoch),签sig2(vote2||epoch)并回复Precommit_Vote消息
			case m := <-p.GetMessage("Precommit", utils.Uint32ToBytes(e)):
				payload := (core.Decapsulation("Precommit", m)).(*protobuf.Precommit)

				if !gotPrepare {
					mPrepare := <-p.GetMessage("Prepare", utils.Uint32ToBytes(e))
					payloadPrepare := (core.Decapsulation("Prepare", mPrepare)).(*protobuf.Prepare)
					txs = payloadPrepare.Txs
					Txs = []byte(strings.Join(txs, ""))
					gotPrepare = true
				}

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

				if !gotPrepare {
					mPrepare := <-p.GetMessage("Prepare", utils.Uint32ToBytes(e))
					payloadPrepare := (core.Decapsulation("Prepare", mPrepare)).(*protobuf.Prepare)
					txs = payloadPrepare.Txs
					gotPrepare = true
				}

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
				if isGlobal {
					p.Broadcast(New_ViewMessage)
				} else {
					p.Intra_Broadcast(New_ViewMessage)
				}
				outputChannel <- txs
				break Loop
			}
		}
	}

}
