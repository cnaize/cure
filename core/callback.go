package core

import (
	"errors"

	"github.com/sansecio/yargo/scanner"
)

var ErrMatchFound = errors.New("match found")

var DefaultCallback scanner.ScanCallback = ScanCallback{}

type ScanCallback struct {
}

func (c ScanCallback) RuleMatching(rule *scanner.MatchRule) (bool, error) {
	return true, ErrMatchFound
}
