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

- ⚔️ **Concurrent brute-force** using goroutines
- 🚦 **Rate limit detection** with exponential backoff and cooldown
- ♻️ **Retry queue** for rescheduling blocked combos
- 🎯 **Keyword-based validation** for success or failure detection
- 🌐 **Proxy support** (`--proxy`)
- 📩 **Custom headers** with fallback to default `User-Agent` and `Content-Type`
- 📊 **Progress bar** that stays clean and persistent
- 🧾 **Log file support** (`--log-file`)
- ✅ **Wordlist cleaner** (ignores empty lines automatically)

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
  --ratelimit-status-codes 429
```

---

## 📌 Required Flags

| Flag        | Description                                 |
|-------------|---------------------------------------------|
| `--url`     | Target login URL                            |
| `--payload` | HTTP payload with `^USER^` and `^PASS^`      |
| `--success` or `--fail` | Keyword to detect success/failure |

---

## 📁 Wordlists

- `--users` → default: `wordlists/users.txt`
- `--passwords` → default: `wordlists/passwords.txt`

---

## 🧠 Headers

You can pass headers using `--header` multiple times:

```bash
--header "Content-Type: application/json" --header "X-Auth: abc123"
```

### ✅ Default headers (applied only if not overridden):

- `User-Agent: Jotunn/1.0`
- `Content-Type: application/x-www-form-urlencoded`

---

## 🌐 Proxy Support

Use `--proxy` to route requests through a proxy:

```bash
--proxy http://127.0.0.1:8080
```

---

## 📉 Rate Limit Handling

- Detects rate-limit status codes (default: `429`)
- Uses exponential backoff (log(n)) cooldown
- Retries combos that were blocked

Customize with:

```bash
--ratelimit-status-codes 429,403,503
```

---

## 📝 Logging

You can save all output to a log file:

```bash
--log-file jotunn.log
```
---

## 🔐 CSRF Token Support

If the target login form requires a **CSRF token**, you can configure Jötunn to fetch it automatically before each login attempt.

### 🧭 How it works

1. A `GET` request is sent to the page that contains the CSRF token (usually the login form).
2. The token is extracted from the HTML input field (by name).
3. It is injected into your payload, replacing the placeholder `^CSRF^`.

---

### ⚙️ Required flags

| Flag             | Description                                                                 |
|------------------|-----------------------------------------------------------------------------|
| `--csrffield`    | Name of the HTML input field that holds the CSRF token (e.g., `csrf_token`) |
| `--csrfsource`   | (Optional) URL to fetch the token from. Defaults to `--url` if not set      |

---

### ⚠️ Important

Your `--payload` must include the placeholder `^CSRF^` or the token will **not** be inserted.

```bash
--payload "username=^USER^&password=^PASS^&csrf_token=^CSRF^"
```

---

### ✅ Example

```bash
jotunn \
  --url https://target.com/login \
  --csrffield csrf_token \
  --payload "username=^USER^&password=^PASS^&csrf_token=^CSRF^" \
  --success "Welcome back"
```

---
## ⚠️ Disclaimer

This tool is intended for **authorized testing and research only**.  
Use it ethically and responsibly. The author is not responsible for misuse.