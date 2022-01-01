// Package solo
//
// @author: xwc1125
package solo

import (
	"github.com/chain5j/chain5j-pkg/database/kvstore"
	"github.com/chain5j/chain5j-protocol/models"
)

type Snapshot interface {
	Store(db kvstore.Database) error                  // 将快照存入数据库
	Copy() Snapshot                                   // 深度拷贝快照
	Apply(headers []*models.Header) (Snapshot, error) // 通过给定的header生成授权快照
}

var (
	_ Snapshot = new(snapshot)
)

type snapshot struct {
}

func (s *snapshot) Store(db kvstore.Database) error {
	panic("implement me")
}

func (s *snapshot) Copy() Snapshot {
	panic("implement me")
}

func (s *snapshot) Apply(headers []*models.Header) (Snapshot, error) {
	panic("implement me")
}
