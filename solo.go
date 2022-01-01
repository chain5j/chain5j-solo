// Package solo
//
// @author: xwc1125
package solo

import (
	"context"
	"fmt"
	"github.com/chain5j/chain5j-pkg/util/dateutil"
	"github.com/chain5j/chain5j-protocol/models"
	"github.com/chain5j/chain5j-protocol/protocol"
	"github.com/chain5j/logger"
	"strings"
	"time"
)

var (
	_ protocol.Consensus = new(solo)
)

type solo struct {
	log    logger.Logger
	ctx    context.Context
	cancel context.CancelFunc
	nodeId models.NodeID // 节点ID

	config     protocol.Config
	nodeKey    protocol.NodeKey
	apis       protocol.APIs
	selfConfig *SoloConfig
}

func NewConsensus(rootCtx context.Context, opts ...option) (protocol.Consensus, error) {
	ctx, cancel := context.WithCancel(rootCtx)
	c := &solo{
		log:    logger.New("solo"),
		ctx:    ctx,
		cancel: cancel,
	}
	var err error
	if err := apply(c, opts...); err != nil {
		c.log.Error("solo apply is error", "err", err)
		return nil, err
	}
	c.nodeId, err = c.nodeKey.ID()
	if err != nil {
		c.log.Error("get peerId err", "err", err)
		return nil, err
	}

	chainConfig := c.config.ChainConfig()
	if cName := chainConfig.Consensus.Name; !strings.EqualFold("solo", cName) {
		c.log.Error("consensus is diff", "consensus", cName)
		return nil, fmt.Errorf("consensus is diff: %s", cName)
	}

	selfConfig := new(SoloConfig)
	err = chainConfig.Consensus.Data.ToStruct(selfConfig)
	if err != nil {
		c.log.Error("decode consensus data err", "err", err)
		return nil, err
	}

	c.selfConfig = selfConfig

	// 注册API
	//c.apis.RegisterAPI([]protocol.API{{
	//	Namespace: "solo",
	//	Version:   "1.0",
	//	Service:   &API{},
	//	Public:    true,
	//}})
	return c, nil
}

func (s *solo) Start() error {
	return nil
}

func (s *solo) Stop() error {
	return nil
}

func (s *solo) Begin() error {
	return nil
}

func (s *solo) VerifyHeader(blockReader protocol.BlockReader, header *models.Header) error {
	return s.verifyHeader(blockReader, header, nil)
}

func (s *solo) VerifyHeaders(blockReader protocol.BlockReader, headers []*models.Header, seals []bool) (chan<- struct{}, <-chan error) {
	//TODO implement me
	panic("implement me")
}

// verifyHeader 验证header
func (s *solo) verifyHeader(blockReader protocol.BlockReader, header *models.Header, parents []*models.Header) error {
	chainConfig := s.config.ChainConfig()
	// 判断区块高度是否为创世区块
	if header.Height == chainConfig.GenesisHeight {
		return ErrUnknownBlock
	}
	// 判断header的时间戳
	// 为了避免两台服务器的时间不同步，那么一般会增加一定时间的容错
	// 比如5秒的容错时间
	faultTolerantTime := uint64(5000)
	if header.Timestamp > uint64(dateutil.CurrentTime())+faultTolerantTime {
		return ErrFutureBlock
	}

	// todo 从header中提取共识所需的参数数据
	// 区块需要携带的共识需要的内容。然后进行验证是否所传递的内容是符合当前共识的

	// 校验级联字段
	return s.verifyCascadingFields(blockReader, header, nil)
}

// verifyCascadingFields 校验级联字段
func (s *solo) verifyCascadingFields(blockReader protocol.BlockReader, header *models.Header, parents []*models.Header) error {
	chainConfig := s.config.ChainConfig()
	// 判断区块高度是否为创世区块
	number := header.Height
	if number == chainConfig.GenesisHeight {
		return ErrUnknownBlock
	}

	var parent *models.Header
	if len(parents) > 0 {
		parent = parents[len(parents)-1]
	} else {
		parent = blockReader.GetHeader(header.ParentHash, number-1)
	}
	// 判断父区块信息
	if parent == nil || parent.Height != number-1 || parent.Hash() != header.ParentHash {
		return ErrUnknownAncestor
	}
	// 判断时间戳
	// 保证两个区块之间的时间间隔不能太近
	period := uint64(0)
	if period == 0 {
		period = 1
	}
	if parent.Timestamp+period > header.Timestamp {
		return ErrInvalidTimestamp
	}
	// 验证区块的签名地址为Validator
	if err := s.verifySigner(blockReader, header, parents); err != nil {
		return err
	}

	// todo 可验证共识中携带的内容的合法性
	return nil
}

func (s *solo) verifySigner(blockReader protocol.BlockReader, header *models.Header, parents []*models.Header) error {
	height := header.Height
	if height == s.config.ChainConfig().GenesisHeight {
		// todo 创世区块不支持校验
		return ErrUnknownBlock
	}

	// 从签名中获取签名者
	hashNoSign := header.HashNoSign()
	peerId, err := s.nodeKey.RecoverId(hashNoSign.Bytes(), header.Signature) // 此处可以加缓存
	if err != nil {
		return err
	}
	_ = peerId

	// 获取上一个区块的快照,快照主要包含了共识的一些配置内容（比如谁可以出块等）
	// 然后通过snapshot验证peerId是否为出块者
	//snap, err := s.snapshot(blockReader, height-1, header.ParentHash, parents)
	//if err != nil {
	//	return err
	//}
	return nil
}

// Prepare 预处理区块头。【此刻header还没有进行签名处理】
// 包括往header中添加：
//		- consensus的内容
//		- 修改timestamp
func (s *solo) Prepare(blockReader protocol.BlockReader, header *models.Header) error {
	height := header.Height
	parent := blockReader.GetHeader(header.ParentHash, height-1)
	if parent == nil {
		return ErrUnknownAncestor
	}

	// todo 获取上一个区块的共识内容的快照
	// 然后，将修改后的内容或原始需要往下传递的内容添加到header的consensus中
	header.Consensus = &models.Consensus{
		Name:      "solo",
		Consensus: nil,
	}
	period := uint64(0)
	if period == 0 {
		period = 1
	}
	// 设置header的时间戳
	header.Timestamp = parent.Timestamp + period
	if header.Timestamp < uint64(dateutil.CurrentTime()) {
		header.Timestamp = uint64(dateutil.CurrentTime())
	}
	return nil
}

// Finalize 将交易进行组装，形成最终的区块。【此刻header还没有进行签名处理】
// 可能会涉及到一些和区块相关的其他操作，如挖矿的区块奖励计算
func (s *solo) Finalize(blockReader protocol.BlockReader, header *models.Header, txs models.Transactions) (*models.Block, error) {
	return models.NewBlock(header, txs, nil), nil
}

// Seal 对区块进行签名，并将签名后没有共识的区块进行共识处理
// 最终共识处理完的结果，通过chan将结果响应给packer
func (s *solo) Seal(ctx context.Context, blockReader protocol.BlockReader, block *models.Block, results chan<- *models.Block) error {
	t1 := time.Now()

	header := block.Header()
	height := header.Height

	// todo 通过共识内容的快照，判断当前节点是否有权限进行出块
	// 如果没有出块能力，那么直接返回错误。
	{
		t2 := time.Now()

		if s.config.ConsensusConfig().IsMetrics(1) {
			s.log.Debug("seal-2) get validator end", "nodeId", s.nodeId, "elapsed", dateutil.PrettyDuration(time.Since(t2)))
		}
	}

	// 如果有出块权限
	// 获取父区块header
	var parent *models.Header
	{
		t2 := time.Now()
		parent = blockReader.GetHeader(header.ParentHash, height-1)
		if parent == nil {
			s.log.Warn("seal-3) get header by parent err", "height", height-1, "bHash", header.ParentHash)
			return ErrUnknownAncestor
		}
		if s.config.ConsensusConfig().IsMetrics(1) {
			s.log.Debug("seal-3) get header by parent end", "height", height-1, "bHash", header.ParentHash, "elapsed", dateutil.PrettyDuration(time.Since(t2)))
		}
	}

	// 对区块进行签名，如果时间戳不对，可修改时间戳
	{
		t2 := time.Now()
		// 签名区块
		signResult, err := s.nodeKey.Sign(header.HashNoSign().Bytes())
		if err != nil {
			if s.config.ConsensusConfig().IsMetrics(2) {
				s.log.Error("seal-4_1) sign header err", "err", err)
			}
			return err
		}
		if s.config.ConsensusConfig().IsMetrics(2) {
			s.log.Debug("seal-4_1) sign header end", "elapsed", dateutil.PrettyDuration(time.Since(t2)))
		}
		header.Signature = signResult
	}
	var blockSigned *models.Block
	{
		t2 := time.Now()
		blockSigned = block.WithSeal(header)
		if s.config.ConsensusConfig().IsMetrics(1) {
			s.log.Debug("seal-4) update block end", "parentHeight", parent.Height, "parentHash", parent.Hash(), "blockHeight", blockSigned.Height(), "blockHash", blockSigned.Hash(), "elapsed", dateutil.PrettyDuration(time.Since(t2)))
		}
	}

	// 需要进行真实的共识处理了。
	//delay := time.Unix(int64(blockSigned.Header().Timestamp), 0).Sub(time.Now())
	//select {
	//case <-time.After(delay):
	//case <-ctx.Done():
	//	s.log.Warn("context hash canceled")
	//	results <- nil
	//	return nil
	//}
	if s.config.ConsensusConfig().IsMetrics(1) {
		s.log.Debug("seal-*) consensus seal end", "blockHeight", blockSigned.Height(), "blockHash", blockSigned.Hash(), "elapsed", dateutil.PrettyDuration(time.Since(t1)))
	}
	results <- blockSigned
	return nil
}
