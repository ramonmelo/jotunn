package throttle

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/LinharesAron/jotunn/internal/httpclient"
	"github.com/LinharesAron/jotunn/internal/logger"
	"github.com/LinharesAron/jotunn/internal/ui"
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
		cooldown: 1 * time.Minute,
		codes:    throttleCodes,
	}
	t.cond = sync.NewCond(&t.mu)

	ip, err := GetCurrentIp()
	if err == nil {
		t.currentIp = ip
		ui.GetUI().SendIpProgressEvent(t.currentIp)
		logger.Info("[TorThrottle] Start with success, current IP %s", t.currentIp)
	}
	return t
}

func (t *TorThrottler) Wait() {
	t.mu.Lock()
	defer t.mu.Unlock()
	for t.blocked {
		t.cond.Wait()
	}

	t.reqCount++
}

func (t *TorThrottler) MarkRecovered() {}

func (t *TorThrottler) Trigger() {
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
	var err error

	for range 5 {
		newip, err = utils.RetrieveTorIP()
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

	httpclient.Reset()
	ui.GetUI().SendIpProgressEvent(t.currentIp)
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
