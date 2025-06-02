package main

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/LinharesAron/jotunn/internal/config"
	"github.com/LinharesAron/jotunn/internal/core"
	"github.com/LinharesAron/jotunn/internal/httpclient"
	"github.com/LinharesAron/jotunn/internal/io"
	"github.com/LinharesAron/jotunn/internal/logger"
	"github.com/LinharesAron/jotunn/internal/throttle"
	"github.com/LinharesAron/jotunn/internal/tracker"
	"github.com/LinharesAron/jotunn/internal/ui"
	"github.com/LinharesAron/jotunn/internal/utils"
	"github.com/LinharesAron/jotunn/internal/worker"
)

func run() bool {
	ui.Init()
	defer ui.Stop()

	logger.Info("🔥 Jötunn – From the blood of giants, only ruin will remains 🔥")

	cfg := config.Load()
	if cfg == nil {
		return waitToExit(false)
	}

	logger.Init(cfg.LogFile)

	logger.Info("Starting attack on: %s", cfg.URL)
	logger.Info("Method: %s | Threads: %d | Threshold: %d", cfg.Method, cfg.Threads, cfg.Threshold)
	logger.Info("Users: %s | Passwords: %s\n", cfg.UserList, cfg.PassList)

	err := httpclient.Init(cfg.Proxy, true)
	if err != nil {
		logger.Error("Failed to initialize HTTP client: %s", err)
		return waitToExit(false)
	}

	if cfg.UseTor {
		ok := utils.CheckTorControl()
		if !ok {
			logger.Error(`[!] Tor mode enabled, but Tor ControlPort is not available at 127.0.0.1:9051
Please make sure Tor is running with ControlPort enabled.

You can add the following to your torrc file:

    ControlPort 9051
    CookieAuthentication 0
`)
			return waitToExit(false)
		}
	}

	start := time.Now()
	logger.Info("[~] Loading wordlists and initializing...")

	users, err := io.ReadLines(cfg.UserList)
	if err != nil {
		logger.Error("[!] Failed to read users file: %v", err)
		return waitToExit(false)
	}

	passwords, err := io.ReadLines(cfg.PassList)
	if err != nil {
		logger.Error("[!] Failed to read passwords file: %v", err)
		return waitToExit(false)
	}

	attemptFile := fmt.Sprintf("jotunn_%s_%s.attempts", hashList(users), hashList(passwords))
	tracker.InitTracker(cfg.BasePath).
		StartAttempts(attemptFile).
		StartCredential()
	defer tracker.Get().CloseAll()

	attemps := tracker.FilterUnseen(tracker.Get().Attempts, users, passwords)

	if len(attemps) == 0 {
		logger.Warn("[!] All combinations from this list have already been tried. Try a different list or delete the Jötunn attempt file at %s", cfg.BasePath)
		return waitToExit(false)
	}

	ui.GetUI().SendTotalProgressEvent(len(attemps))
	dispatcher := core.NewDispatcher(cfg.Threads, 3, 10000)
	throttle := throttle.New(cfg)

	handler := worker.NewAttack(cfg, throttle)

	dispatcher.StartWorkersHandler(cfg.Threads, handler)
	dispatcher.StartRetryHandler(handler)

	logger.Info("[~] Starting the BruteForce...")
	dispatcher.DistributeToWorkers(attemps)

	dispatcher.CloseWorkers()
	dispatcher.WaitWorkers()

	dispatcher.CloseRetries()
	dispatcher.WaitRetries()

	duration := time.Since(start)
	logger.Info("✅ Done in %s", duration)

	return true
}

func waitToExit(b bool) bool {
	time.Sleep(2 * time.Second)
	return b
}

func main() {
	if ok := run(); !ok {
		os.Exit(1)
	}
}

func hashList(list []string) string {
	h := sha1.Sum([]byte(strings.Join(list, ",")))
	return hex.EncodeToString(h[:])
}
