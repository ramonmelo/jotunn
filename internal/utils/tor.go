package utils

import (
	"io"
	"net"
	"strings"
	"time"

	"github.com/LinharesAron/jotunn/internal/httpclient"
)

func CheckTorControl() bool {
	conn, err := net.DialTimeout("tcp", "127.0.0.1:9051", 2*time.Second)
	if err != nil {
		return false
	}
	defer conn.Close()

	_, err = conn.Write([]byte("AUTHENTICATE\r\n"))
	if err != nil {
		return false
	}

	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		return false
	}

	response := string(buf[:n])
	return strings.HasPrefix(response, "250")
}

func RetrieveTorIP() (string, error) {
	client := httpclient.Get()

	resp, err := client.Get("http://checkip.amazonaws.com/")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	ip, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(ip)), nil
}
