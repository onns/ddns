package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os/exec"
	"strings"

	dns "github.com/alibabacloud-go/alidns-20150109/v2/client"
	openapi "github.com/alibabacloud-go/darabonba-openapi/client"
	util "github.com/alibabacloud-go/tea-utils/service"
	"github.com/alibabacloud-go/tea/tea"
)

// edit it before run your own code

var (
	AccessKeyId     = ""
	AccessKeySecret = ""
	SubDomain       = ""
	Domain          = "onns.xyz"
)

// edit over

func CreateClient() (_result *dns.Client, _err error) {
	config := &openapi.Config{
		AccessKeyId:     &AccessKeyId,
		AccessKeySecret: &AccessKeySecret,
	}
	config.Endpoint = tea.String("alidns.cn-shanghai.aliyuncs.com")
	_result = &dns.Client{}
	_result, _err = dns.NewClient(config)
	return
}

func getLocalIp() (res string) {
	// 获取所有网络接口
	interfaces, err := net.Interfaces()
	if err != nil {
		fmt.Println("获取网络接口失败:", err)
		return
	}

	// 遍历每个网络接口
	for _, iface := range interfaces {
		// 排除非物理接口和回环接口
		if iface.Flags&net.FlagUp != 0 && iface.Flags&net.FlagLoopback == 0 {
			addrs, err := iface.Addrs()
			if err != nil {
				fmt.Println("获取接口地址失败:", err)
				continue
			}
			// 遍历接口的地址
			for _, addr := range addrs {
				switch v := addr.(type) {
				case *net.IPNet:
					// 检查是否为 IPv4 地址，并且不是回环地址
					if v.IP.To4() != nil && !v.IP.IsLoopback() {
						res = v.IP.String()
						return
					}
				}
			}
		}
	}
	return
}

func getWANIp() (ip string) {
	responseClient, err := http.Get("http://members.3322.org/dyndns/getip")
	if err != nil {
		fmt.Printf("get ip err: %+v \n", err)
		panic(err)
	}
	defer responseClient.Body.Close()
	body, _ := ioutil.ReadAll(responseClient.Body)
	ip = fmt.Sprintf("%s", string(body))
	// html has '\n' in the end of the file, remove it.
	ip = strings.Trim(ip, "\n")
	return
}

func getCurrentIp() (ip string) {
	cmd := exec.Command("/usr/sbin/ipconfig", "getifaddr", "en0")
	stdout, err := cmd.Output()
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	ip = string(stdout)
	ip = strings.Trim(ip, "\n")
	return
}

func main() {
	sb := flag.String("sd", "", "sub domain")
	wan := flag.Bool("w", false, "wan or lan ip")
	ip := flag.String("ip", "", "specific ip, empty for LAN ip default")
	flag.Parse()
	var (
		client        *dns.Client
		currentHostIP string
		err           error
		rr            = tea.String(SubDomain)
		t             = tea.String("A")
	)
	if len(*sb) < 4 {
		panic("最少四个字符")
	}
	if *sb != "" {
		rr = sb
	}
	currentHostIP = getLocalIp()
	if *wan {
		currentHostIP = getWANIp()
	}
	if *ip != "" {
		currentHostIP = *ip
	}
	client, err = CreateClient()
	if err != nil {
		panic(err)
	}
	describeDomainRecordsRequest := &dns.DescribeDomainRecordsRequest{
		DomainName: tea.String(Domain),
		RRKeyWord:  rr,
	}
	resp, _err := client.DescribeDomainRecords(describeDomainRecordsRequest)
	if _err != nil {
	}
	var oldRecord *dns.DescribeDomainRecordsResponseBodyDomainRecordsRecord
	for _, record := range resp.Body.DomainRecords.Record {
		if tea.BoolValue(util.EqualString(record.RR, rr)) && tea.BoolValue(util.EqualString(record.Type, t)) {
			oldRecord = record
			break
		}
	}
	if oldRecord == nil {
		req := &dns.AddDomainRecordRequest{
			RR:         rr,
			DomainName: &Domain,
			Type:       tea.String("A"),
			Value:      &currentHostIP,
		}

		log.Printf("%s.%s change ip to %+v.\n", *rr, Domain, currentHostIP)
		_, err = client.AddDomainRecord(req)
		if err != nil {
			panic(err)
		}
		return
	}
	if tea.StringValue(oldRecord.Value) == currentHostIP {
		return
	}
	req := &dns.UpdateDomainRecordRequest{
		RecordId: oldRecord.RecordId,
		RR:       oldRecord.RR,
		Type:     oldRecord.Type,
		Value:    &currentHostIP,
	}
	log.Printf("%s.%s change ip from %+v to %+v.\n", *rr, Domain, tea.StringValue(oldRecord.Value), currentHostIP)
	_, err = client.UpdateDomainRecord(req)
	if err != nil {
		panic(err)
	}
}
