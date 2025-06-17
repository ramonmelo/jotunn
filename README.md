# Jötunn

[![Go Reference](https://pkg.go.dev/badge/github.com/LinharesAron/jotunn.svg)](https://pkg.go.dev/github.com/LinharesAron/jotunn)

🔥 **Jötunn – From the blood of giants, only ruin will remain.**  
A fast, intelligent, and adaptive brute-force engine written in Go.  
Crafted for red teamers, pentesters, and hackers who wield ancient power with modern tools.

---

## 📦 Install

You can install the latest version using Go:

```bash
go install github.com/LinharesAron/jotunn/cmd/jotunn@latest
```

---

## 🚀 Features

- ⚔️ **Concurrent brute-force** using goroutines.
- 🚦 **Rate limit detection** with exponential backoff and cooldown.
- ♻️ **Retry queue** for rescheduling blocked combos.
- 🎯 **Keyword-based validation** for success or failure detection.
- 🌐 **Proxy support**.
- 📩 **Custom headers** with fallback to default `User-Agent` and `Content-Type`.
- 📊 **Progress bar** that stays clean and persistent.
- 🧾 **Log file support**.
- ✅ **Wordlist cleaner** by ignoring empty lines automatically.

---

## ⚙️ Usage

```bash
jotunn \
  --url https://target.com/login \
  --method POST \
  --users wordlists/users.txt \
  --passwords wordlists/passwords.txt \
  --payload "username=^USER^&password=^PASS^" \
  --success "Welcome back" \
  --threads 10 \
  --proxy http://127.0.0.1:8080 \
  --log-file jotunn.log \
  --header "X-Custom-Header: value" \
  --throttle-status-codes 429
```

---

## 📌 Required Flags

| Flag                    | Description                             |
|-------------------------|-----------------------------------------|
| `--url`                 | Target login URL                        |
| `--payload`             | HTTP payload with `^USER^` and `^PASS^` |
| `--success` or `--fail` | Keyword to detect success/failure       |

### 📚 Available Flags

| Flag                        | Description                                                                 |
|-----------------------------|-----------------------------------------------------------------------------|
| `--url, -u`                 | Target login URL (**required**)                                             |
| `--method, -m`              | HTTP method (`GET` or `POST`) – default is `POST`                           |
| `--users, -U`               | Path to username list – default is `wordlists/users.txt`                    |
| `--passwords, -P`           | Path to password list – default is `wordlists/passwords.txt`                |
| `--payload, -d`             | Payload format with ^USER^ and ^PASS^ placeholders                          |
| `--success, -s`             | Keyword in the response indicating a successful login                       |
| `--fail, -f`                | Keyword indicating a failed login attempt                                   |
| `--threads, -t`             | Number of threads – default is `10`                                         |
| `--threshold, -T`           | Request per minute threshold – default is `100`                             |
| `--proxy`                   | HTTP or SOCKS5 proxy to route brute-force requests through                  |
| `--tor`                     | Enable Tor mode using proxy `socks5://127.0.0.1:9050`                       |
| `--no-limit`                | Disable any throttling logic (faster, but risk of block)                    |
| `--csrfsource`              | URL to fetch the CSRF token before login                                    |
| `--csrffield`               | HTML input name that holds the CSRF token                                   |
| `--log-file`                | Path to save the output logs                                                |
| `--throttle-status-codes`   | Status codes to treat as throttling (default `[429]`)                       |

---

## 📁 Wordlists

Wordlists are plain text files containing lists of usernames and passwords. These lists are used in conjunction with the `^USER^` and `^PASS^` placeholders in the payload. The tool will iterate through each combination of username and password from the provided files and send them in the payload.

```bash
--users <path>.txt     # Path to the username wordlist file (default: `wordlists/users.txt`)
--passwords <path>.txt # Path to the password wordlist file (default: `wordlists/passwords.txt`)
```

---

## 🧠 Headers

You can customize HTTP headers using the `--header` flag.
This allows you to set any headers required by the target application, such as authentication tokens or content types.
The flag can be used multiple times to set multiple headers.

```bash
--header "<Header-Name>: <value>"         # Custom header to include in requests
--header "Content-Type: application/json" # Example of setting Content-Type
--header "X-Auth: abc123"                 # Example of setting an authentication header
```

### ✅ Default headers (applied only if not overridden)

- `User-Agent: Jotunn/1.0`
- `Content-Type: application/x-www-form-urlencoded`

---

## 🌐 Proxy Support

Use `--proxy` to route requests through a proxy:

```bash
--proxy <URL>                   # URL of the proxy server
--proxy http://127.0.0.1:8080   # Example of an HTTP proxy
--proxy socks5://127.0.0.1:9050 # Example of a SOCKS5 proxy
```

---

## 🚦 Throttle Types

Many applications implement rate limiting to prevent brute-force attacks. Jötunn comes with multiple Throttle strategies to adapt the attack pace and avoid detection or blocking.

### Available Throttlers

You can control which throttling logic will be used with the following flags:

```bash
--no-limit               # Disables all throttling. Fastest but risky
--throttle-status-codes  # List of HTTP status codes considered throttling (default: 429)
--tor                    # Enable Tor mode (requires Tor and ControlPort access)
```

### 🧊 StandardThrottler (default)

This is the default strategy used by Jötunn. It:

- Monitors request rate (RPM).
- Automatically lowers the threshold by 10% on each block.
- Waits and applies exponential backoff: starting with 5 minutes, doubling until 50 minutes max.
- Automatically detects recovery and resumes.

**Behavior:**

```conf
  Trigger → Cooldown → Resume → If blocked again → Lower RPM → Longer cooldown
```

> You can customize which status codes trigger this behavior with `--throttle-status-codes`.

### 🧅 TorThrottle – Evade with a new identity

When `--tor` is enabled, Jötunn will:

- Route traffic through the Tor Network (via `127.0.0.1:9050`).
- Use the ControlPort (`9051`) to request a new identity/IP when throttled.
- Pause all workers while waiting for the new IP to be active.
- Resume only when the IP has changed or timeout occurs.

**Requirements:**

Ensure Tor is installed and that the following lines are present in your `torrc` configuration file:

```conf
  ControlPort 9051
  CookieAuthentication 0
```

> You can usually find your `torrc` in `/etc/tor/torrc` or `/usr/local/etc/tor/torrc`.

**Behavior:**

```conf
  Trigger → Request new IP → Wait for IP change → Resume
```

> 💡 This strategy is ideal for hardened targets or CTFs that aggressively block brute-force attempts.

### ☠️ NoLimitThrottle

When you use `--no-limit`, Jötunn disables all request pacing and retries.

- Useful for fast testing or internal environments.
- Dangerous against real targets – likely to trigger defenses or get IP banned.
- No backoff, no retries, no detection of `429`/`403` – it just goes.

---

## 🔐 CSRF Token Support

Some login forms include a `CSRF Token` as a hidden field to prevent automated or cross-site submissions. Jötunn supports extracting this token before sending the login attempt.

### How CSRF Token Extraction Works

If the flag `--csrffield` is provided, Jötunn will:

1. Perform a **GET** request to the target page (by default the same URL as `--url` unless `--csrfsource` is defined).
2. Parse the HTML to extract the value of the `CSRF Token` using the provided field name.
3. Replace the `^CSRF^` placeholder in the payload with the extracted token.
4. Proceed with the brute-force attempt using the updated payload.

### Flags

```bash
--csrffield <string>  # Name of the CSRF field to extract (e.g. "csrf_token")
--csrfsource <URL>    # URL where the CSRF token will be retrieved (defaults to --url if not provided)
```

### Payload Usage

Your payload should include the special token `^CSRF^`, which will be dynamically replaced.
Example:

```bash
--payload "username=^USER^&password=^PASS^&csrf_token=^CSRF^"
```

### Example

```bash
jotunn \
  --url https://target.com/login \
  --csrffield "csrf_token" \
  --payload "username=^USER^&password=^PASS^&csrf_token=^CSRF^" \
  --users users.txt \
  --passwords passwords.txt \
  --fail "Invalid credentials"
```

> 💡 If CSRF extraction fails due to a rate-limit (429), the request will be retried according to the throttling logic. Otherwise, it will be ignored.

---

## 📝 Logging

All logs include timestamps, status indicators, and are written both to the terminal and (if configured) to the log file.

You can save all output to a log file:

```bash
--log-file <path>/<file>.log  # Path to save the output logs
```

---

## ⚠️ Disclaimer

This tool is intended for **authorized testing and research only**.  
Use it ethically and responsibly. The author is not responsible for misuse.
