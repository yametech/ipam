package main

import (
	"fmt"
	"net"
	"strconv"
	"strings"
)

var IpList = []string{
	"10.200.11.180",
	"10.200.11.181",
	"10.200.11.182",
	"10.200.11.184",
	"10.200.11.190",
}

func UInt32ToIP(intIP uint32) string {
	var bytes [4]byte
	bytes[0] = byte(intIP & 0xFF)
	bytes[1] = byte((intIP >> 8) & 0xFF)
	bytes[2] = byte((intIP >> 16) & 0xFF)
	bytes[3] = byte((intIP >> 24) & 0xFF)
	return net.IPv4(bytes[3], bytes[2], bytes[1], bytes[0]).String()
}

func IPStringToUInt32(ip string) uint32 {
	bits := strings.Split(ip, ".")
	b0, _ := strconv.Atoi(bits[0])
	b1, _ := strconv.Atoi(bits[1])
	b2, _ := strconv.Atoi(bits[2])
	b3, _ := strconv.Atoi(bits[3])
	var sum uint32
	sum += uint32(b0) << 24
	sum += uint32(b1) << 16
	sum += uint32(b2) << 8
	sum += uint32(b3)
	return sum
}

func find(ip string) bool {
	for _, ipList := range IpList {
		if ipList == ip {
			return true
		}
	}
	return false
}

func reuse(start, end string) string {
	startIp := IPStringToUInt32(start)
	endIp := IPStringToUInt32(end)
	for startIp <= endIp {
		if !find(UInt32ToIP(startIp)) {
			return UInt32ToIP(startIp)
		}
		startIp++
	}
	return ""
}

func main() {
	fmt.Println(reuse(IpList[0], IpList[len(IpList)-1]))
}
