package throttle

import (
	"bufio"
	"fmt"
	"net"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/LinharesAron/jotunn/internal/core"
	"github.com/LinharesAron/jotunn/internal/logger"
	"github.com/LinharesAron/jotunn/internal/types"
	"github.com/LinharesAron/jotunn/internal/utils"
)

type TorThrottler struct {
	mu       sync.Mutex
	cond     *sync.Cond
	blocked  bool
	cooldown time.Duration
	codes    []int

	currentIp string
	reqCount  int
}

func NewTorThrottler(throttleCodes []int) *TorThrottler {
	t := &TorThrottler{
		cooldown: 2 * time.Minute,
		codes:    throttleCodes,
	}
	t.cond = sync.NewCond(&t.mu)

	ip, err := GetCurrentIp()
	if err == nil {
		t.currentIp = ip
		logger.Info("[TorThrottle] Start with success, current IP %s", t.currentIp)
	}
	return t
}

func (t *TorThrottler) RegisterRequest() {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.reqCount++
}

func (t *TorThrottler) WaitIfBlocked() {
	t.mu.Lock()
	defer t.mu.Unlock()
	for t.blocked {
		t.cond.Wait()
	}
}

func (t *TorThrottler) WaitCadence() {}

func (t *TorThrottler) IsThrottling(statusCode int) bool {
	return slices.Contains(t.codes, statusCode)
}

func (t *TorThrottler) HandleThrottle(statusCode int, dispatcher *core.Dispatcher, attempt types.Attempt) bool {
	if !t.IsThrottling(statusCode) {
		return false
	}

	if err := dispatcher.Retry(attempt); err != nil {
		logger.Warn("[TorThrottle] Retry limit reached for %s:%s – ignoring attempt → %s", attempt.Username, attempt.Password, err)
	}

	t.trigger()
	return true
}

func (t *TorThrottler) MarkRecovered() {}

func (t *TorThrottler) trigger() {
	t.mu.Lock()
	if t.blocked {
		t.mu.Unlock()
		return
	}
	t.blocked = true
	t.mu.Unlock()

	now := time.Now().Format("15:04:05")
	logger.Warn("[TorThrottle] [%s] Rate limit detected after %d attempts, pausing workers for %s", now, t.reqCount, t.cooldown)

	time.Sleep(t.cooldown)
	if t.resetTorIdentity() {
		go t.waitResetTorIdentity()
	}
}

func (t *TorThrottler) waitResetTorIdentity() {
	var newip string
	for range 5 {
		newip, err := utils.RetrieveTorIP()
		if err != nil {
			logger.Warn("[TorThrottle] [%s] Unable to retrieve IP, trying again...", time.Now().Format("15:04:05"))
		}

		if t.currentIp != "" && t.currentIp != newip {
			break
		}

		logger.Warn("[TorThrottle] [%s] IP unchanged (%s), retrying...", time.Now().Format("15:04:05"), t.currentIp)
		time.Sleep(30 * time.Second)
	}

	t.mu.Lock()
	t.blocked = false
	t.reqCount = 0
	t.currentIp = newip
	t.mu.Unlock()

	logger.Info("[TorThrottle] [%s] Cooldown complete, resuming operations with new %s", time.Now().Format("15:04:05"), t.currentIp)
	t.cond.Broadcast()
}

func (t *TorThrottler) resetTorIdentity() bool {
	conn, err := net.Dial("tcp", "127.0.0.1:9051")
	if err != nil {
		logger.Error("[TorThrottle] Unable to connect to Tor ControlPort: %v", err)
		return false
	}
	defer conn.Close()

	reader := bufio.NewReader(conn)
	fmt.Fprintf(conn, "AUTHENTICATE \"\"\r\n")
	authResp, _ := reader.ReadString('\n')
	if !strings.HasPrefix(authResp, "250") {
		logger.Error("[TorThrottle] Failed to authenticate to Tor ControlPort")
		return false
	}

	fmt.Fprintf(conn, "SIGNAL NEWNYM\r\n")
	signalResp, _ := reader.ReadString('\n')
	if !strings.HasPrefix(signalResp, "250") {
		logger.Error("[TorThrottle] Failed to send NEWNYM signal")
		return false
	}

	fmt.Fprintf(conn, "QUIT\r\n")
	logger.Info("[TorThrottle] Sent NEWNYM to Tor — requested new IP")

	return true
}

func GetCurrentIp() (string, error) {
	ip, err := utils.RetrieveTorIP()
	return ip, err
}
