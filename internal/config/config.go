package config

import (
	"os"
	"slices"
	"strings"

	"github.com/LinharesAron/jotunn/internal/logger"
	"github.com/spf13/pflag"
)

type AttackConfig struct {
	URL       string
	Method    string
	UserList  string
	PassList  string
	Threads   int
	Threshold int

	Payload string
	Headers map[string]string

	CSRFField     string
	CSRFSourceURL string

	SuccessKeyword string
	FailKeyword    string

	LogFile    string
	Proxy      string
	UseTor     bool
	UseNoLimit bool

	ThrottleCodes []int
}

func (cfg *AttackConfig) Keyword() (bool, string) {
	if cfg.SuccessKeyword != "" {
		return true, cfg.SuccessKeyword
	}
	return false, cfg.FailKeyword
}

func (cfg *AttackConfig) IsThrottlingStatus(statusCode int) bool {
	return slices.Contains(cfg.ThrottleCodes, statusCode)
}

func Load() *AttackConfig {
	cfg := &AttackConfig{}

	var headerList []string

	pflag.StringVarP(&cfg.URL, "url", "u", "", "Target login URL (required)")
	pflag.StringVarP(&cfg.Method, "method", "m", "POST", "HTTP method (GET or POST)")
	pflag.StringVarP(&cfg.UserList, "users", "U", "wordlists/users.txt", "Path to usernames list")
	pflag.StringVarP(&cfg.PassList, "passwords", "P", "wordlists/passwords.txt", "Path to passwords list")
	pflag.IntVarP(&cfg.Threads, "threads", "t", 10, "Number of concurrent threads")
	pflag.IntVarP(&cfg.Threshold, "threshold", "T", 5000, "Number of request per minute")
	pflag.StringVarP(&cfg.Payload, "payload", "d", "", "Payload format with ^USER^ and ^PASS^ placeholders")
	pflag.StringArrayVar(&headerList, "header", []string{}, "Additional headers (can be repeated)")

	pflag.StringVar(&cfg.CSRFField, "csrffield", "", "Name of the CSRF field to extract from HTML (e.g., csrf_token). Your payload must contain ^CSRF^ for the token to be replaced.")
	pflag.StringVar(&cfg.CSRFSourceURL, "csrfsource", "", "Optional URL to fetch the CSRF token from (defaults to --url if not set)")

	pflag.StringVarP(&cfg.SuccessKeyword, "success", "s", "", "Success message if login completed")
	pflag.StringVarP(&cfg.FailKeyword, "fail", "f", "", "Fail message if login failed (override by the --success flag)")

	pflag.IntSliceVar(&cfg.ThrottleCodes, "throttle-status-codes", []int{429}, "List of HTTP status codes considered for throttling")

	pflag.StringVar(&cfg.LogFile, "log-file", "", "Path where the log file will be writen")
	pflag.StringVar(&cfg.Proxy, "proxy", "", "Proxy to use for requests (e.g. http://127.0.0.1:8080)")

	pflag.BoolVar(&cfg.UseTor, "tor", false, "Route all traffic through Tor (overrides --proxy with socks5://127.0.0.1:9050)")
	pflag.BoolVar(&cfg.UseNoLimit, "no-limit", false, "Disables rate limiting. Use with caution: may trigger throttling on the target server.")

	pflag.Parse()

	if cfg.URL == "" {
		logger.Error("[!] Missing required --url")
		pflag.Usage()
		os.Exit(1)
	}

	if cfg.Payload == "" {
		logger.Error("[!] Payload required --payload")
		pflag.Usage()
		os.Exit(1)
	}

	if cfg.SuccessKeyword == "" && cfg.FailKeyword == "" {
		logger.Error("[!] Success keyword or Fail Keyword required --success or --fail")
		pflag.Usage()
		os.Exit(1)
	}

	if cfg.CSRFField != "" && cfg.CSRFSourceURL == "" {
		cfg.CSRFSourceURL = cfg.URL
		logger.Info("[~] No --csrfsource provided, defaulting to --url: %s", cfg.URL)
	}

	if cfg.UseTor {
		cfg.Proxy = "socks5://127.0.0.1:9050"
		logger.Info("[~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~]")
		logger.Info("[~ Tor mode enabled – using SOCKS5 proxy on 127.0.0.1:9050 ~]")
		logger.Info("[~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~]")
	}

	cfg.Headers = make(map[string]string)

	if _, ok := cfg.Headers["User-Agent"]; !ok {
		cfg.Headers["User-Agent"] = "Jotunn/1.0"
	}

	if _, ok := cfg.Headers["Content-Type"]; !ok {
		cfg.Headers["Content-Type"] = "application/x-www-form-urlencoded"
	}

	for _, h := range headerList {
		parts := strings.SplitN(h, ":", 2)
		if len(parts) == 2 {
			cfg.Headers[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
	}

	return cfg
}
