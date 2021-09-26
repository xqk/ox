package onet

import (
	"fmt"
	"net"
	"net/http"
	"time"
)

//
// Ping
// @Description: 调用http://ip:port/health 接口, 失败返回错误
// @param host
// @param port
// @return error
//
func Ping(host string, port int) error {
	client := http.Client{
		Timeout: time.Millisecond * 100,
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err := client.Get(fmt.Sprintf("http://%s:%d/health", host, port))

	if err != nil {
		return err
	}

	_ = resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("returned status %d", resp.StatusCode)
	}
	return nil
}

//
// Dial
// @Description: dial 指定的端口，失败返回错误
// @param addr
// @param timeout
// @return error
//
func Dial(addr string, timeout time.Duration) error {
	conn, err := net.DialTimeout("tcp", addr, timeout)
	if err != nil {
		return err
	}

	conn.Close()
	return nil
}
