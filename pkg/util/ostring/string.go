package ostring

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"net"
	"strconv"
	"strings"
)

//
// Addr2Hex
// @Description:将地址字符串转换为十六进制字符串，仅支持ipv4
// @param str
// @return string
// @return error
//
func Addr2Hex(str string) (string, error) {
	ipStr, portStr, err := net.SplitHostPort(str)
	if err != nil {
		return "", err
	}

	ip := net.ParseIP(ipStr).To4()
	port, err := strconv.ParseUint(portStr, 10, 16)
	if err != nil {
		return "", nil
	}

	buf := make([]byte, 2)
	binary.BigEndian.PutUint16(buf, uint16(port))
	ip = append(ip, buf...)

	return hex.EncodeToString(ip), nil
}

//
// Hex2Addr
// @Description:将十六进制字符串转换为地址
// @param str
// @return string
// @return error
//
func Hex2Addr(str string) (string, error) {
	buf, err := hex.DecodeString(str)
	if err != nil {
		return "", err
	}
	if len(buf) < 4 {
		return "", fmt.Errorf("bad hex string length")
	}
	return fmt.Sprintf("%s:%d", net.IP(buf[:4]).String(), binary.BigEndian.Uint16(buf[4:])), nil
}

// Strings 字符串数组
type Strings []string

//
// KickEmpty
// @Description:从ss中剔除空元素
// @param ss
// @return Strings
//
func KickEmpty(ss []string) Strings {
	var ret = make([]string, 0)
	for _, str := range ss {
		if str != "" {
			ret = append(ret, str)
		}
	}
	return Strings(ret)
}

//
// AnyBlank
// @Description: 如果ss中有空元素则返回true
// @param ss
// @return bool
//
func AnyBlank(ss []string) bool {
	for _, str := range ss {
		if str == "" {
			return true
		}
	}

	return false
}

//
// HeadT
// @Description: 返回数组的第一个元素和其他元素数组
// @receiver ss
// @return string
// @return Strings
//
func (ss Strings) HeadT() (string, Strings) {
	if len(ss) > 0 {
		return ss[0], Strings(ss[1:])
	}

	return "", Strings{}
}

//
// Head
// @Description: 返回数组的第一个元素
// @receiver ss
// @return string
//
func (ss Strings) Head() string {
	if len(ss) > 0 {
		return ss[0]
	}
	return ""
}

//
// Head2
// @Description: 返回数组的第一个元素和第二个元素
// @receiver ss
// @return h0
// @return h1
//
func (ss Strings) Head2() (h0, h1 string) {
	if len(ss) > 0 {
		h0 = ss[0]
	}
	if len(ss) > 1 {
		h1 = ss[1]
	}
	return
}

//
// Head3
// @Description: 返回数组的第一、第二和第三个元素
// @receiver ss
// @return h0
// @return h1
// @return h2
//
func (ss Strings) Head3() (h0, h1, h2 string) {
	if len(ss) > 0 {
		h0 = ss[0]
	}
	if len(ss) > 1 {
		h1 = ss[1]
	}
	if len(ss) > 2 {
		h2 = ss[2]
	}
	return
}

//
// Head4
// @Description: 返回数组的第一、第二、第三和第四个元素
// @receiver ss
// @return h0
// @return h1
// @return h2
// @return h3
//
func (ss Strings) Head4() (h0, h1, h2, h3 string) {
	if len(ss) > 0 {
		h0 = ss[0]
	}
	if len(ss) > 1 {
		h1 = ss[1]
	}
	if len(ss) > 2 {
		h2 = ss[2]
	}
	if len(ss) > 3 {
		h3 = ss[3]
	}
	return
}

//
// Split
// @Description: 分隔
// @param raw
// @param sep
// @return Strings
//
func Split(raw string, sep string) Strings {
	return Strings(strings.Split(raw, sep))
}
