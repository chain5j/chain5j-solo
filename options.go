// Package solo
//
// @author: xwc1125
package solo

import (
	"fmt"
	"github.com/chain5j/chain5j-protocol/protocol"
)

type option func(f *solo) error

func apply(f *solo, opts ...option) error {
	for _, opt := range opts {
		if opt == nil {
			continue
		}
		if err := opt(f); err != nil {
			return fmt.Errorf("option apply err:%v", err)
		}
	}
	return nil
}

func WithConfig(config protocol.Config) option {
	return func(f *solo) error {
		f.config = config
		return nil
	}
}

func WithNodeKey(nodeKey protocol.NodeKey) option {
	return func(f *solo) error {
		f.nodeKey = nodeKey
		return nil
	}
}

func WithAPIs(apis protocol.APIs) option {
	return func(f *solo) error {
		f.apis = apis
		return nil
	}
}
