package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	dns "github.com/alibabacloud-go/alidns-20150109/v2/client"
	openapi "github.com/alibabacloud-go/darabonba-openapi/client"
	util "github.com/alibabacloud-go/tea-utils/service"
	"github.com/alibabacloud-go/tea/tea"
)

// edit it before run your own code

var AccessKeyId = ""
var AccessKeySecret = ""
var SubDomain = "xm"
var Domain = "onns.xyz"

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

func getCurrentIp() (ip string) {
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

func main() {
	var (
		client        *dns.Client
		currentHostIP string
		err           error
		rr            = tea.String(SubDomain)
		t             = tea.String("A")
	)
	currentHostIP = getCurrentIp()
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
	oldRecord := &dns.DescribeDomainRecordsResponseBodyDomainRecordsRecord{}
	for _, record := range resp.Body.DomainRecords.Record {
		if tea.BoolValue(util.EqualString(record.RR, rr)) && tea.BoolValue(util.EqualString(record.Type, t)) {
			oldRecord = record
			break
		}
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
	log.Printf("%s.%s change ip from %+v to %+v.\n", SubDomain, Domain, tea.StringValue(oldRecord.Value), currentHostIP)
	_, err = client.UpdateDomainRecord(req)
	if err != nil {
		panic(err)
	}
}
