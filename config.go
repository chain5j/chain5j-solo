// Package solo
//
// @author: xwc1125
package solo

type SoloConfig struct {
	Period uint64 `json:"period"` // Number of seconds between blocks to enforce
	Epoch  uint64 `json:"epoch"`  // 重置投票及检测点的长度
}