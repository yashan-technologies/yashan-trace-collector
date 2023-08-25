package datadef

import (
	"fmt"
	"strings"
)

const (
	DESC_DEFAULT                    = "请查看YTC产品文档或日志定位原因"
	DESC_NO_SAR_COMMAND             = "找不到sar命令，请检查是否安装sar"
	DESC_NO_PERMISSION              = "%s不存在或无权限，请检查文件权限"
	DESC_UBUNTU_FIREWALLD           = "查看Ubuntu系统防火墙状态需要root权限"
	DESC_KYLIN_DMESG                = "执行dmesg命令失败，请查看日志定位具体原因"
	DESC_READ_CODEDUMP_PATH         = "读取coredump文件夹：%s失败，请检查文件权限或查看日志定位具体原因"
	DESC_GET_COREDUMP_PATH          = "获取系统coredump路径失败，请检查日志查看原因或手动修改YTC配置文件设置coredump路径"
	DESC_GET_DATABASE_PARAMETER     = "获取数据库%s参数失败，请检查数据库状态或查看YTC日志定位原因"
	DESC_SKIP_COLLECT_DATABASE_INFO = "数据库连接失败，将跳过本检查项"
	DESC_YASDB_PROCESS_STATUS       = "获取数据库进程信息失败，请检查数据库进程是否存在"
	DESC_GET_DATABASE_VIEW          = "查询数据库%s视图失败，请检查数据库状态或查看YTC日志定位原因"
)

const (
	key_command_not_found = "command not found"
)

func GenDefaultDesc() string {
	return DESC_DEFAULT
}

func GenNoPermissionDesc(str string) string {
	return fmt.Sprintf(DESC_NO_PERMISSION, str)
}

func GenHostWorkloadDesc(e error) string {
	if strings.Contains(e.Error(), key_command_not_found) {
		return DESC_NO_SAR_COMMAND
	}
	return DESC_DEFAULT
}

func GenUbuntuFirewalldDesc() string {
	return DESC_UBUNTU_FIREWALLD
}

func GenKylinDmesgDesc() string {
	return DESC_KYLIN_DMESG
}

func GenGetCoreDumpPathDesc() string {
	return DESC_GET_COREDUMP_PATH
}

func GenReadCoreDumpPathDesc(path string) string {
	return fmt.Sprintf(DESC_READ_CODEDUMP_PATH, path)
}

func GenGetDatabaseParameterDesc(parameter string) string {
	return fmt.Sprintf(DESC_GET_DATABASE_PARAMETER, parameter)
}

func GenSkipCollectDatabaseInfoDesc() string {
	return DESC_SKIP_COLLECT_DATABASE_INFO
}

func GenYasdbProcessStatusDesc() string {
	return DESC_YASDB_PROCESS_STATUS
}

func GenGetDatabaseViewDesc(view string) string {
	return fmt.Sprintf(DESC_GET_DATABASE_VIEW, view)
}
