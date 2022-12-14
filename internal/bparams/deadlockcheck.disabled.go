//+build !deadlock

package bparams

import (
	"github.com/algorand/go-deadlock"
)
const Deadlock = false
func init() {
	deadlock.Opts.Disable = true
}
