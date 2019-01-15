package common

import "errors"

var (
	ERROR_LOCK_ALREADY_EXIST = errors.New("锁已经被占用")
)
