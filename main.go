package main

//go:generate go run builder/builder.go

import (
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/auth/basic"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/config"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/region"
	dnssdk "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/dns/v2"
	dnsmodel "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/dns/v2/model"
)

func main() {
	// 解析命令行参数
	ak := flag.String("ak", "", "华为云 Access Key (必须)")
	sk := flag.String("sk", "", "华为云 Secret Key (必须)")
	projectId := flag.String("projectId", "", "华为云项目ID (必须) 需与区域名称对应")
	regionName := flag.String("region", "", "区域名称 (必须)，如cn-east-3")
	zoneId := flag.String("zoneId", "", "域名ID (必须)")
	recordSetId := flag.String("recordSetId", "", "记录集ID (必须)")
	publicIP := flag.String("ip", "", "公网IP地址 (可选，不提供则自动获取)")
	ipType := flag.String("ipType", "v4", "IP记录类型 (可选)，默认为v4（即IPv4），可以是v6（即IPv6）")
	desc := flag.String("desc", "", "备注信息 (可选)，用于描述此记录用途。")
	skipTLS := flag.Bool("skipTLS", false, "跳过TLS证书验证 (可选)，适用于嵌入式设备")

	// 覆盖默认的 Usage 函数
	flag.Usage = func() {
		fmt.Println("华为云DDNS工具 使用说明:")
		fmt.Println("参数:")
		fmt.Println("  -ak")
		fmt.Println("      华为云 Access Key (必须)")
		fmt.Println("  -sk")
		fmt.Println("      华为云 Secret Key (必须)")
		fmt.Println("  -projectId")
		fmt.Println("      华为云项目ID (必须) 需与区域名称对应")
		fmt.Println("  -region")
		fmt.Println("      区域名称 (必须)，如cn-east-3")
		fmt.Println("  -zoneId")
		fmt.Println("      域名ID (必须)")
		fmt.Println("  -recordSetId")
		fmt.Println("      记录集ID (必须)")
		fmt.Println("  -ip")
		fmt.Println("      公网IP地址 (可选，不提供则自动获取)")
		fmt.Println("  -ipType")
		fmt.Println("      IP记录类型 (可选)，默认为v4（即IPv4），可以是v6（即IPv6）")
		fmt.Println("  -desc")
		fmt.Println("      备注信息 (可选)，用于描述此记录用途。")
		fmt.Println("  -skipTLS")
		fmt.Println("      跳过TLS证书验证 (可选)，适用于嵌入式设备")
		fmt.Println("")
		fmt.Println("author: xm0625(402276694@qq.com), have fun!")
	}

	flag.Parse()

	// 打印备注信息（如果有）
	if *desc != "" {
		fmt.Printf("备注信息: %s\n", *desc)
	}

	// 检查必要参数
	if *ak == "" || *sk == "" || *projectId == "" || *regionName == "" || *zoneId == "" || *recordSetId == "" {
		fmt.Println("错误: 缺少必要参数 请确认命令行参数值已使用引号包裹 如：-region \"cn-east-3\"")
		flag.Usage()
		os.Exit(1)
	}

	// 验证IP类型参数
	if *ipType != "v4" && *ipType != "v6" {
		fmt.Println("错误: ipType 参数必须是 \"v4\" 或 \"v6\"")
		os.Exit(1)
	}

	// 获取公网IP (如果未提供)
	currentIP := *publicIP
	if currentIP == "" {
		ip, err := getPublicIP(*ipType, *skipTLS)
		if err != nil {
			fmt.Printf("获取公网IP失败: %v\n", err)
			os.Exit(1)
		}
		currentIP = ip
		fmt.Printf("自动获取公网%s地址: %s\n", getIPTypeName(*ipType), currentIP)
	} else {
		// 验证手动输入的IP格式
		if !isValidIP(currentIP, *ipType) {
			fmt.Printf("错误: 提供的IP地址 %s 与指定的类型 %s 不匹配\n", currentIP, *ipType)
			os.Exit(1)
		}
		fmt.Printf("使用指定公网%s地址: %s\n", getIPTypeName(*ipType), currentIP)
	}

	// 创建华为云DNS客户端
	client := createDNSClient(*ak, *sk, *projectId, *regionName, *skipTLS)
	if client == nil {
		fmt.Println("创建DNS客户端失败")
		os.Exit(1)
	}

	// 更新DNS记录
	err := updateDNSRecord(client, *zoneId, *recordSetId, currentIP)
	if err != nil {
		fmt.Printf("更新DNS记录失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("DNS记录更新成功!")
}

// 获取IP类型名称
func getIPTypeName(ipType string) string {
	if ipType == "v6" {
		return "IPv6"
	}
	return "IPv4"
}

// 验证IP地址格式
func isValidIP(ip, ipType string) bool {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}

	if ipType == "v4" {
		return parsedIP.To4() != nil
	}

	// IPv6地址验证
	return parsedIP.To4() == nil && parsedIP.To16() != nil
}

// 创建华为云DNS客户端
func createDNSClient(ak, sk, projectId, regionName string, skipTLS bool) *dnssdk.DnsClient {
	auth, err := basic.NewCredentialsBuilder().
		WithAk(ak).
		WithSk(sk).
		WithProjectId(projectId).
		SafeBuild()

	if err != nil {
		fmt.Printf("NewCredentialsBuilder 创建认证信息失败: %v\n", err)
		return nil
	}

	reg := region.NewRegion(regionName, "https://dns."+regionName+".myhuaweicloud.com")

	builder := dnssdk.DnsClientBuilder().
		WithRegion(reg).
		WithCredential(auth)

	// 如果需要跳过TLS验证
	if skipTLS {
		httpConfig := config.DefaultHttpConfig().
			WithIgnoreSSLVerification(true)
		builder = builder.WithHttpConfig(httpConfig)
	}

	dnsClient, err := builder.SafeBuild()

	if err != nil {
		fmt.Printf("DnsClientBuilder 创建DNS客户端失败: %v\n", err)
		return nil
	}

	return dnssdk.NewDnsClient(dnsClient)
}

// 更新DNS记录
func updateDNSRecord(client *dnssdk.DnsClient, zoneId, recordSetId, newIP string) error {
	records := []string{newIP}
	request := &dnsmodel.UpdateRecordSetRequest{
		ZoneId:      zoneId,
		RecordsetId: recordSetId,
		Body: &dnsmodel.UpdateRecordSetReq{
			Records: &records,
		},
	}

	_, err := client.UpdateRecordSet(request)
	return err
}

// 获取公网IP (根据类型)
func getPublicIP(ipType string, skipTLS bool) (string, error) {
	if ipType == "v6" {
		return getIPv6Address(skipTLS)
	}
	return getIPv4Address(skipTLS)
}

// 获取IPv4地址
func getIPv4Address(skipTLS bool) (string, error) {
	services := []string{
		"http://cip.cc",
		"http://ip.3322.net",
		"http://4.ipw.cn",
		"http://v4.ip.zxinc.org/info.php?type=json",
	}

	for _, service := range services {
		ip, err := fetchIPFromService(service, "v4", skipTLS)
		if err == nil && ip != "" {
			return ip, nil
		}
	}

	return "", fmt.Errorf("所有IPv4查询服务均失败")
}

// 获取IPv6地址
func getIPv6Address(skipTLS bool) (string, error) {
	services := []string{
		"http://6.ipw.cn",
		"http://v6.ident.me",
		"http://ipv6.icanhazip.com",
		"http://v6.ip.zxinc.org/info.php?type=json",
	}

	for _, service := range services {
		ip, err := fetchIPFromService(service, "v6", skipTLS)
		if err == nil && ip != "" {
			return ip, nil
		}
	}

	return "", fmt.Errorf("所有IPv6查询服务均失败")
}

// 从指定服务获取IP
func fetchIPFromService(url, ipType string, skipTLS bool) (string, error) {
	client := &http.Client{Timeout: 5 * time.Second}

	// 如果需要跳过TLS验证
	if skipTLS {
		client.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}

	resp, err := client.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("服务返回非200状态码: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	content := strings.TrimSpace(string(body))

	// 使用正则表达式提取IP地址
	var re *regexp.Regexp
	if ipType == "v4" {
		re = regexp.MustCompile(`\b(?:\d{1,3}\.){3}\d{1,3}\b`)
	} else {
		// IPv6地址正则表达式（简化版）
		re = regexp.MustCompile(`([0-9a-fA-F]{1,4}:){7}[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,7}:|([0-9a-fA-F]{1,4}:){1,6}:[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,5}(:[0-9a-fA-F]{1,4}){1,2}|([0-9a-fA-F]{1,4}:){1,4}(:[0-9a-fA-F]{1,4}){1,3}|([0-9a-fA-F]{1,4}:){1,3}(:[0-9a-fA-F]{1,4}){1,4}|([0-9a-fA-F]{1,4}:){1,2}(:[0-9a-fA-F]{1,4}){1,5}|[0-9a-fA-F]{1,4}:((:[0-9a-fA-F]{1,4}){1,6})|:((:[0-9a-fA-F]{1,4}){1,7}|:)|fe80:(:[0-9a-fA-F]{0,4}){0,4}%[0-9a-zA-Z]+|::(ffff(:0{1,4})?:)?((25[0-5]|(2[0-4]|1?[0-9])?[0-9])\.){3}(25[0-5]|(2[0-4]|1?[0-9])?[0-9])|([0-9a-fA-F]{1,4}:){1,4}:((25[0-5]|(2[0-4]|1?[0-9])?[0-9])\.){3}(25[0-5]|(2[0-4]|1?[0-9])?[0-9])`)
	}

	ip := re.FindString(content)
	if ip == "" {
		return "", fmt.Errorf("未找到IP地址")
	}

	return ip, nil
}
