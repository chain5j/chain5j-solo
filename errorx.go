// Package solo
//
// @author: xwc1125
package solo

import "errors"

var (
	ErrUnknownBlock     = errors.New("unknown block")
	ErrFutureBlock      = errors.New("block in the future")
	ErrUnknownAncestor  = errors.New("unknown ancestor")
	ErrInvalidTimestamp = errors.New("invalid timestamp")
)
