## YTC(yashan trace collector) | 崖山一键收集工具

---

YTC(崖山一键收集工具)，是一款针对崖山数据库信息与日志收集的工具。

## 用户定位

- 崖山DBA

## 产品定位

- 轻量的独立工具
- 开箱即用

## 场景建议

- 数据库出现故障时
- 数据库出现性能问题时
- 其他任何想要快速收集相关信息时

## 核心功能

### 基础信息的收集

- 崖山数据库基本信息(版本号/commit号等)
- 所在服务器的基本信息(操作系统/硬件配置/防火墙等)
- 所在服务器的负载情况(网络流量/CPU占用/IO负载/内存/磁盘容量等)
- ...

### 故障信息的收集

- 数据库状态及进程检查
- ADR日志
- coredump文件
- 数据库日志(run.log/alert.log等)
- 操作系统日志(Dmesg/message_log等)
- ...

### 性能数据的收集

- AWR报告
- 慢日志
- ...

### 灵活可配的收集策略

- 支持自定义时间周期，绝对灵活的时间周期选择
- 支持自定义收集模块，三大模块均可灵活搭配选择
- 支持自定义路径数据收集，目录或文件均可批量选择
- ...

### 丰富健全的数据管理

- 支持自定义收集数据的存放路径
- 多种收集数据展示形式(txt/md/html)
- ...

## 使用方法

### 工具帮助信息

```
bash # ./ytcctl --help
Usage: ytcctl <command>

Ytcctl is used to manage the yashan trace collector.

Flags:
  -h, --help                          Show context-sensitive help.
  -v, --version                       Show version.
  -c, --config="./config/ytc.toml"    Configuration file.

Commands:
  collect    The collect command is used to gather trace data.


Run "ytcctl <command> --help" for more information on a command.
```

### 最佳实践

```shell
./ytcctl collect
```

>更多使用方法详见产品文档 (工具包路径/docs/ytc.pdf)