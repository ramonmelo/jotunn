package main

import (
	"os"
	"time"

	"github.com/LinharesAron/jotunn/internal/config"
	"github.com/LinharesAron/jotunn/internal/core"
	"github.com/LinharesAron/jotunn/internal/httpclient"
	"github.com/LinharesAron/jotunn/internal/io"
	"github.com/LinharesAron/jotunn/internal/logger"
	"github.com/LinharesAron/jotunn/internal/throttle"
	"github.com/LinharesAron/jotunn/internal/ui"
	"github.com/LinharesAron/jotunn/internal/utils"
	"github.com/LinharesAron/jotunn/internal/worker"
)

func main() {
	ui.Init()

	logger.Info("🔥 Jötunn – From the blood of giants, only ruin will remains 🔥")

	cfg := config.Load()

	logger.Init(cfg.LogFile)

	logger.Info("Starting attack on: %s", cfg.URL)
	logger.Info("Method: %s | Threads: %d", cfg.Method, cfg.Threads)
	logger.Info("Users: %s | Passwords: %s\n", cfg.UserList, cfg.PassList)

	err := httpclient.Init(cfg.Proxy, true)
	if err != nil {
		logger.Error("Failed to initialize HTTP client: %s", err)
		os.Exit(1)
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
			os.Exit(1)
		}
	}

	start := time.Now()
	logger.Info("[~] Loading wordlists and initializing...")

	users, err := io.ReadLines(cfg.UserList)
	if err != nil {
		logger.Error("[!] Failed to read users file: %v", err)
		os.Exit(1)
	}

	passwords, err := io.ReadLines(cfg.PassList)
	if err != nil {
		logger.Error("[!] Failed to read passwords file: %v", err)
		os.Exit(1)
	}

	logger.InitProgressTracker(len(users) * len(passwords))
	logger.Info("[~] Starting the BruteForce...")

	dispatcher := core.NewDispatcher(cfg.Threads, 3, 10000)
	throttle := throttle.New(cfg)

	handler := worker.NewAttack(cfg, throttle)

	dispatcher.StartWorkersHandler(cfg.Threads, handler)
	dispatcher.StartRetryHandler(handler)

	dispatcher.DistributeToWorkers(users, passwords)

	dispatcher.CloseWorkers()
	dispatcher.WaitWorkers()

	dispatcher.CloseRetries()
	dispatcher.WaitRetries()

	duration := time.Since(start)
	logger.Info("✅ Done in %s", duration)
}
