# 华为云 DDNS 客户端

一个用于自动更新华为云 DNS 记录的命令行工具，支持 IPv4 和 IPv6 地址的动态 DNS 更新。该工具可以自动检测您的公网 IP 地址，并将其更新到华为云 DNS 记录中。

## 功能特点

- 支持 IPv4 和 IPv6 地址的自动检测和更新
- 多个公网 IP 检测服务，提高可靠性
- 支持手动指定 IP 地址
- 跨平台支持：
  - Linux (AMD64, ARMv7, ARMv8, MIPS, MIPSEL)
  - Windows (AMD64)
- 支持 UPX 压缩，减小程序体积

## 使用方法

### 命令行参数

```bash
  -ak string
        华为云 Access Key (必须)
  -sk string
        华为云 Secret Key (必须)
  -projectId string
        华为云项目ID (必须) 需与区域名称对应
  -region string
        区域名称 (必须)，如cn-east-3
  -zoneId string
        域名ID (必须)
  -recordSetId string
        记录集ID (必须)
  -ip string
        公网IP地址 (可选，不提供则自动获取)
  -ipType string
        IP记录类型 (可选)，默认为v4（即IPv4），可以是v6（即IPv6）
  -desc string
        备注信息 (可选)，用于描述此记录用途
  -skipTLS
        跳过TLS证书验证 (可选)，适用于嵌入式设备
```

### 示例

1. 自动获取 IPv4 地址并更新 DNS 记录：
```bash
./hwddns -ak "YOUR_AK" -sk "YOUR_SK" -projectId "YOUR_PROJECT_ID" -region "cn-east-3" -zoneId "YOUR_ZONE_ID" -recordSetId "YOUR_RECORD_SET_ID"
```

2. 自动获取 IPv6 地址并更新 DNS 记录：
```bash
./hwddns -ak "YOUR_AK" -sk "YOUR_SK" -projectId "YOUR_PROJECT_ID" -region "cn-east-3" -zoneId "YOUR_ZONE_ID" -recordSetId "YOUR_RECORD_SET_ID" -ipType "v6"
```

3. 手动指定 IP 地址：
```bash
./hwddns -ak "YOUR_AK" -sk "YOUR_SK" -projectId "YOUR_PROJECT_ID" -region "cn-east-3" -zoneId "YOUR_ZONE_ID" -recordSetId "YOUR_RECORD_SET_ID" -ip "1.2.3.4"
```

4. 在嵌入式设备上跳过 TLS 证书验证：
```bash
./hwddns -ak "YOUR_AK" -sk "YOUR_SK" -projectId "YOUR_PROJECT_ID" -region "cn-east-3" -zoneId "YOUR_ZONE_ID" -recordSetId "YOUR_RECORD_SET_ID" -skipTLS
```

## 编译说明

项目使用 Go 语言开发，支持多架构交叉编译。编译工具会自动为不同目标平台生成可执行文件。

### 编译要求

- Go 1.16 或更高版本
- （可选）UPX 压缩工具

### 编译步骤

1. 克隆代码仓库
2. 进入项目目录
3. 运行编译命令：
```bash
go generate
```

编译后的文件将保存在 `build` 目录下，包括：
- hwddns-armv7：适用于 ARMv7 设备
- hwddns-armv8：适用于 ARMv8/ARM64 设备
- hwddns-mips-softfpu：适用于 MIPS 设备
- hwddns-mipsel-softfpu：适用于 MIPSEL 设备
- hwddns-windows-amd64.exe：适用于 Windows 64位系统
- hwddns-linux-amd64：适用于 Linux 64位系统

如果安装了 UPX，还会生成对应的压缩版本（文件名带有 -upx 后缀）。

## 依赖说明

- github.com/huaweicloud/huaweicloud-sdk-go-v3：华为云 SDK
- 标准库：net/http、regexp、flag 等

## 注意事项

1. 请妥善保管您的 Access Key 和 Secret Key
2. 确保提供的项目 ID 与区域名称相对应
3. 建议使用 crontab 等工具定期执行以实现动态 DNS 更新
4. IPv6 地址检测需要环境支持 IPv6 网络
5. `-skipTLS` 参数会跳过 TLS 证书验证，仅在嵌入式设备等特殊环境下使用，存在安全风险
6. 请确认命令行参数值已使用引号包裹 如：`-region "cn-east-3"`

## 许可证

MIT
