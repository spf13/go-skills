package claims

import (
	"testing"
	"time"
)

// go/SKILL.md "time Channels Are Always Unbuffered (Go 1.27)": len and cap of timer
// and ticker channels are 0.
func TestTimerAndTickerChannelsAreUnbuffered(t *testing.T) {
	timer := time.NewTimer(time.Hour)
	defer timer.Stop()
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()
	if c := cap(timer.C); c != 0 {
		t.Errorf("cap(timer.C) = %d, want 0", c)
	}
	if c := cap(ticker.C); c != 0 {
		t.Errorf("cap(ticker.C) = %d, want 0", c)
	}
}
