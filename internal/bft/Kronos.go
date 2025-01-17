package bft

import (
	"Chamael/internal/party"
	"Chamael/pkg/core"
	"Chamael/pkg/crypto"
	"Chamael/pkg/protobuf"
	"Chamael/pkg/txs"
	"Chamael/pkg/utils"
	"bytes"
	"fmt"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/pairing/bn256"
	"go.dedis.ch/kyber/v3/sign/bls"
	"time"
)

// 按输入分片分类交易
func CategorizeTransactionsByInputShard(transactions []string) map[int][]string {
	inputShardCategories := make(map[int][]string)

	for _, tx := range transactions {
		// 提取交易详情
		//fmt.Println(tx)
		details, err := txs.ExtractTransactionDetails(tx)
		if err != nil {
			fmt.Printf("Skipping invalid transaction: %v\n ", err)
			fmt.Println(tx)
			continue
		}

		// 将交易分配到每个输入分片对应的类别
		for _, shard := range details.InputShard {
			inputShardCategories[shard] = append(inputShardCategories[shard], tx)
		}
	}

	return inputShardCategories
}

// 按输出分片分类交易
func CategorizeTransactionsByOutputShard(transactions []string) (map[int][]string, []string) {
	crossShardTransactions := make(map[int][]string) // 按输出分片存储跨片交易
	innerShardTransactions := []string{}             // 存储片内交易

	for _, tx := range transactions {
		// 提取交易详情
		details, err := txs.ExtractTransactionDetails(tx)
		if err != nil {
			fmt.Printf("Skipping invalid transaction: %v\n ", err)
			fmt.Println(tx)
			continue
		}

		// 判断是否是跨片交易
		isCrossShard := false
		for _, shard := range details.InputShard {
			if shard != details.OutputShard {
				isCrossShard = true
				break
			}
		}

		if isCrossShard {
			// 按输出分片分类
			crossShardTransactions[details.OutputShard] = append(crossShardTransactions[details.OutputShard], tx)
		} else {
			// 片内交易
			innerShardTransactions = append(innerShardTransactions, tx)
		}
	}

	return crossShardTransactions, innerShardTransactions
}

/*
	有两个东西需要全程非阻塞监听,把内容放到对应Channel里：
	1:别的分片发来的 完全未处理的 本分片为输入分片的交易,也就是TXs_Inform。
		收到后什么都不用干,让它们躺在TXsInformChannel里就行了.
		主进程在每个时期(开始后不久)完成TXs_Inform的发送之后 统一收菜
	2:别的分片发来的 部分处理了的 本分片为输出分片的交易,也就是InputBFT_Result。
		收到后验证树根聚合签名,验证默克尔树(注意有两个验证)
		对它们进行处理,放到交易池中——若交易池中没有该交易则添加,然后把对应的输入分片标记为true
		再把交易池中输入分片已经处理完了(输入分片元组全为true)的交易拿出来组装成InputResultTobeDoneChannel中
		主进程在每个时期需要调用片内交易之前.统一收菜
*/
func TXs_Inform_Handler(p *party.HonestParty, e uint32, TXsInformChannel chan []string) {
	var l []int
	var Result []string
	seen := make(map[int]bool)
	for {
		m := <-p.GetMessage("TXs_Inform", utils.Uint32ToBytes(e))
		payload := (core.Decapsulation("TXs_Inform", m)).(*protobuf.TXs_Inform)
		if !seen[int(m.Sender)] {
			l = append(l, int(m.Sender))
			seen[int(m.Sender)] = true
			Result = append(Result, payload.Txs...)
		}

		if len(l) >= int(p.N)-1 {
			TXsInformChannel <- Result
			return
		}
	}
}

func InpufBFT_Result_Handler(p *party.HonestParty, e uint32, InputResultTobeDoneChannel chan []string, txPool *TransactionPool) {
	suite := bn256.NewSuite()
	var l []int
	seen := make(map[int]bool)
	for {
		m := <-p.GetMessage("InputBFT_Result", utils.Uint32ToBytes(e))
		payload := (core.Decapsulation("InputBFT_Result", m)).(*protobuf.InputBFT_Result)
		AggPK := utils.BytesToPoint(payload.Aggpk)
		err := bls.Verify(suite, AggPK, payload.Root, payload.Aggsig)
		if err != nil {
			fmt.Println("AggSig(root) verification failed:", err)
			return
		}
		/*
			result := crypto.VerifyMerkleTreeProof(payload.Root, payload.Path, payload.Indicator, payload.Txs)
			if result == false {
				fmt.Println("MerkleTree verification failed")
				return
			}
		*/
		if !seen[int(m.Sender)] {
			l = append(l, int(m.Sender))
			seen[int(m.Sender)] = true

			for _, tx := range payload.Txs {
				err := txPool.AddTransaction(tx, int((m.Sender-m.Sender%p.N)/p.N))
				if err != nil {
					fmt.Println("Failed to add transaction to pool:", err)
				}
			}
		}
		//fmt.Println(l, p.PID)
		if len(l) >= int(p.M)-1 {
			break
		}
	}

	//fmt.Println(p.PID)
	//txPool.PrintTxPoolDetail()

	completedTransactions := txPool.CheckAndRemoveTransactions()
	//fmt.Println(completedTransactions)
	// 将完成的交易发送到 InputResultTobeDoneChannel
	//if len(completedTransactions) > 0 {
	InputResultTobeDoneChannel <- completedTransactions
	//}

	return
}

func KronosProcess(p *party.HonestParty, epoch int, itx_inputChannel chan []string, ctx_inputChannel chan []string, outputChannel chan []string, timeChannel chan time.Time) {
	txPool := NewTransactionPool()
	var TXsInformChannel = make(chan []string, 1024)
	var InputResultTobeDoneChannel = make(chan []string, 1024)
	suite := bn256.NewSuite()
	for e := uint32(1); e <= uint32(epoch); e++ {
		//go TXs_Inform_Handler(p, e-1, TXsInformChannel)
		//go InpufBFT_Result_Handler(p, e-1, InputResultTobeDoneChannel, txPool)
		var txs_in []string            //放入片内共识的交易整体
		var txs_ctx_in []string        //别的分片发来的,本分片为输入分片的交易;是放入片内共识交易的跨片部分
		var txs_itx []string           //从inputchannel来,本分片的片内交易;是放入片内共识交易的片内部分
		var txs_pool_finished []string //从缓冲池来,输入分片已经处理完的,本分片作为输出分片的交易;是放入片内共识交易的片内部分

		var txs_ctx map[int][]string //从inputchannel来,按输入分片分类后的跨片交易;是TXs_Inform的内容

		var txs_out []string          //从片内共识里拿取的交易整体
		var txs_ctx2 map[int][]string //从片内共识来,按输出分片分类后的跨片交易
		var txs_itx2 []string         //从片内共识来,进行分类后的片内交易

		//var Txs_ctx map[int][]byte //分类后的交易转化为字节形式;建树&发送&签名&验签用的都是字节形式
		//var Txs []byte
		var is_coordinator bool

		if (e+1)%p.N == p.SID {
			is_coordinator = true
		} else {
			is_coordinator = false
		}

		if e > 1 {
			InpufBFT_Result_Handler(p, e-1, InputResultTobeDoneChannel, txPool)
			txs_pool_finished = <-InputResultTobeDoneChannel
			txs_in = append(txs_in, txs_pool_finished...)
		}

		//获取新跨片交易,把跨片交易按输入分片分类后发给对应分片
		ctx := <-ctx_inputChannel
		txs_ctx = CategorizeTransactionsByInputShard(ctx)
		for i := uint32(0); i < p.M; i++ {
			TXsInformMesssage := core.Encapsulation("TXs_Inform", utils.Uint32ToBytes(e), p.PID, &protobuf.TXs_Inform{
				Txs: txs_ctx[int(i)],
			})
			p.Shard_Broadcast(TXsInformMesssage, i)
		}

		//把片内和跨片交易放入片内共识,并获取结果、进行分类(通道获取交易&片内共识全是阻塞的)
		txs_itx = <-itx_inputChannel
		txs_in = append(txs_in, txs_itx...)
		TXs_Inform_Handler(p, e, TXsInformChannel)
		txs_ctx_in = <-TXsInformChannel
		txs_in = append(txs_in, txs_ctx_in...)

		//fmt.Println("here!2", p.PID, e)
		/*
			select {
			case txs_ctx_in = <-TXsInformChannel:
				txs_in = append(txs_in, txs_ctx_in...)
			default:
				fmt.Println("No data in TXsInformChannel, skipping...")
			}

			select {
			case txs_pool_finished = <-InputResultTobeDoneChannel:
				txs_in = append(txs_in, txs_pool_finished...)
			default:
				fmt.Println("No data in InputResultTobeDoneChannel, skipping...")
			}
		*/

		inputChannel := make(chan []string, 1024)
		receiveChannel := make(chan []string, 1024)
		//fmt.Println("txs_in:", txs_in, "\n\n\n\n\n\n\n\n\n")
		inputChannel <- txs_in

		HotStuffProcess(p, int(e), inputChannel, receiveChannel)
		txs_out = <-receiveChannel
		txs_ctx2, txs_itx2 = CategorizeTransactionsByOutputShard(txs_out)
		//fmt.Println("txs_out:", p.PID, txs_out, "\n\n\n")

		//对于片内交易和输出分片为自己的交易,直接输出,作为吞吐量计算
		timeChannel <- time.Now()
		outputChannel <- txs_itx2
		outputChannel <- txs_ctx2[int(p.Snumber)]
		txs_ctx2[int(p.Snumber)] = nil

		//对于跨片交易,建立默克尔树,并对树根签名
		mktree, _ := crypto.NewMerkleTree(utils.MapToSlice(txs_ctx2))
		//if is_coordinator == true {
		//	fmt.Println("txs_out:", p.PID, txs_out, "\n\n\n\n\n\n\n\n\n")
		//}
		//fmt.Println("txs_ctx2", txs_ctx2, "\n\n\n\n\n\n\n\n\n")
		//fmt.Println("mktree", mktree, "\n\n\n\n\n\n\n\n\n")
		Root := mktree.GetMerkleTreeRoot()
		sigRoot, _ := bls.Sign(suite, p.SK, Root)

		/*
			如果自己是跨片协调者:
				1:片内广播Sig_Inform表明身份
				2:监听收集片内其他节点对于树根的签名,收齐后聚合
				3:分别向各个分片广播InputBFT_Result
			如果自己不是跨片交易协调者:
				监听Sig_Inform消息,收到后发送Sigmsg给协调者
		*/
		if is_coordinator == true {
			SigInformMessage := core.Encapsulation("Sig_Inform", utils.Uint32ToBytes(e), p.PID, &protobuf.Sig_Inform{
				None: make([]byte, 0),
			})
			p.Intra_Broadcast(SigInformMessage)

			var l []int
			seen := make(map[int]bool)
			var signatures [][]byte
			var pubkeys []kyber.Point
			for {
				m := <-p.GetMessage("Sigmsg", utils.Uint32ToBytes(e))
				payload := (core.Decapsulation("Sigmsg", m)).(*protobuf.Sigmsg)

				//fmt.Println(payload.Root, Root, p.PID, m.Sender)

				if !bytes.Equal(payload.Root, Root) {
					fmt.Println("Invalid Mktree Root(Unequal Root)")
					return
				}
				if !seen[int(m.Sender)] {
					l = append(l, int(m.Sender))
					seen[int(m.Sender)] = true
					signatures = append(signatures, payload.Sig)
					pubkeys = append(pubkeys, p.PK[m.Sender])
				}

				if len(l) >= int(p.N)-1 {
					break
				}

				signatures = append(signatures, sigRoot)
				pubkeys = append(pubkeys, p.PK[p.PID])
			}
			aggSig, _ := bls.AggregateSignatures(suite, signatures...)
			aggPubKey := bls.AggregatePublicKeys(suite, pubkeys...)
			err := bls.Verify(suite, aggPubKey, Root, aggSig)
			if err != nil {
				fmt.Println("Invalid Mktree Root(Invalid aggSig)", err)
				return
			}

			for i := uint32(0); i < p.M; i++ {
				path, indicator := mktree.GetMerkleTreeProof(int(i))
				TXsInformMesssage := core.Encapsulation("InputBFT_Result", utils.Uint32ToBytes(e), p.PID, &protobuf.InputBFT_Result{
					Txs:       txs_ctx2[int(i)],
					Root:      Root,
					Path:      path,
					Indicator: indicator,
					Aggsig:    aggSig,
					Aggpk:     utils.PointToBytes(aggPubKey),
				})
				p.Shard_Broadcast(TXsInformMesssage, i)
			}

		} else {
			m := <-p.GetMessage("Sig_Inform", utils.Uint32ToBytes(e))
			SigMessage := core.Encapsulation("Sigmsg", utils.Uint32ToBytes(e), p.PID, &protobuf.Sigmsg{
				Root: Root,
				Sig:  sigRoot,
			})
			p.Send(SigMessage, m.Sender)
		}
	}
	time.Sleep(time.Second * 15)
}
