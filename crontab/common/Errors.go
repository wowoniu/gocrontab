package common

import "errors"

var (
	ERR_LOCK_ALREADY_REQUIRED = errors.New("锁已被占用")

	ERR_NO_NET_INTERFACE_FOUND = errors.New("没有可用的网卡")
)
