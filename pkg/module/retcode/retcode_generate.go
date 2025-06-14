// Code generated by fluteNAS. DO NOT EDIT.

package retcode

var StatusOK = func(data any) *RetCode { return &RetCode{Code: 0, Message: "request success", Data: data}}
var StatusDirNotExist = func(data any) *RetCode { return &RetCode{Code: 1000, Message: "directory path not exist", Data: data}}
var StatusDirEmpty = func(data any) *RetCode { return &RetCode{Code: 1001, Message: "directory path is empty", Data: data}}
var StatusParamInvalid = func(data any) *RetCode { return &RetCode{Code: 1002, Message: "parameter %s invalid", Data: data}}
var StatusUmountDiskFailed = func(data any) *RetCode { return &RetCode{Code: 2000, Message: "umount disk on path %s failed, maybe you can umount manually in terminal first.", Data: data}}
var StatusError = func(data any) *RetCode { return &RetCode{Code: 9999, Message: "request failed", Data: data}}
