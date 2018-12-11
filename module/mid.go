package module

import (
	"BeanGithub/crawler/errors"
	"fmt"
	"net"
	"strconv"
	"strings"
)

// MID 组件ID。
type MID string

// midTemplate 组件ID模板。
var midTemplate = "%s%d|%s"

// GenMID 根据给定参数生成组件ID。
func GenMID(moduleType Type, sn uint64, addr net.Addr) (MID, error) {
	if !LegalType(moduleType) {
		errMsg := fmt.Sprintf("illegal module type: %s", moduleType)
		return "", errors.NewIllegalParameterError(errMsg)
	}
	letter := legalTypeLetterMap[moduleType]
	var midStr string
	if addr == nil {
		midStr = fmt.Sprintf(midTemplate, letter, sn, "")
		midStr = midStr[:len(midStr)-1]
	} else {
		midStr = fmt.Sprintf(midTemplate, letter, sn, addr.String())
	}
	return MID(midStr), nil
}

// LegalMID 判断组件ID是否合法。
func LegalMID(mid MID) bool {
	if _, err := SplitMID(mid); err == nil {
		return true
	}
	return false
}

// SplitMID 分解组件ID。
// 第二个结果值表示是否分解成功。
// 若分解成功，第一个结果值长度为3。
// 依次包含类型字母、序列号、组件网络地址（如果有的话）。
func SplitMID(mid MID) ([]string, error) {
	var ok bool
	var letter, snStr, addr string

	midStr := string(mid)
	if len(midStr) <= 1 {
		return nil, errors.NewIllegalParameterError("insufficient MID")
	}
	letter = midStr[:1]
	if _, ok = legalLetterTypeMap[letter]; !ok {
		return nil, errors.NewIllegalParameterError(
			fmt.Sprintf("illegal module type letter: %s", letter))
	}
	snAndAddr := midStr[1:]
	index := strings.LastIndex(snAndAddr, "|")
	if index < 0 {
		snStr = snAndAddr
		if !legalSN(snStr) {
			return nil, errors.NewIllegalParameterError(
				fmt.Sprintf("illegal module SN: %s", snStr))
		}
	} else {
		snStr = snAndAddr[:index]
		if !legalSN(snStr) {
			return nil, errors.NewIllegalParameterError(
				fmt.Sprintf("illegal module SN: %s", snStr))
		}
		addr = snAndAddr[index+1:]
		index = strings.LastIndex(addr, ":")
		if index <= 0 {
			return nil, errors.NewIllegalParameterError(
				fmt.Sprintf("illegal module address: %s", addr))
		}
		ipStr := addr[:index]
		if ip := net.ParseIP(ipStr); ip == nil {
			return nil, errors.NewIllegalParameterError(
				fmt.Sprintf("illegal module IP: %s", ipStr))
		}
		portStr := addr[index+1:]
		if _, err := strconv.ParseUint(portStr, 10, 64); err != nil {
			return nil, errors.NewIllegalParameterError(
				fmt.Sprintf("illegal module port: %s", portStr))
		}
	}
	return []string{letter, snStr, addr}, nil
}

// legalSN 判断序列号的合法性。
func legalSN(snStr string) bool {
	_, err := strconv.ParseUint(snStr, 10, 64)
	if err != nil {
		return false
	}
	return true
}
