package v1

import (
	"net"
	"sort"
	"strconv"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:resource:path=ip,scope=Cluster
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type Ip struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec IPSpec `json:"spec"`
}

type IPSpec struct {
	Ip   string `json:"ip"`
	Mask string `json:"mask"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type IpList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []Ip `json:"items"`
}

func (ipList *IpList) Ips() []string {
	ips := make([]string, 0)
	for _, ip := range ipList.Items {
		ips = append(ips, ip.Spec.Ip)
	}
	return ips
}

func (ipList *IpList) FindIp(ip string) bool {
	for _, item := range ipList.Items {
		if item.Spec.Ip == ip {
			return true
		}
	}
	return false
}

func (ipList *IpList) Reuse() string {
	ips := make([]string, 0)
	for _, i := range ipList.Items {
		ips = append(ips, i.Spec.Ip)
	}
	if len(ips) == 0 {
		return ""
	}
	sort.StringSlice(ips).Sort()
	reuse := func(start, end string) string {
		startIp := IPStringToUInt32(start)
		endIp := IPStringToUInt32(end)
		for startIp <= endIp {
			if !ipList.FindIp(UInt32ToIP(startIp)) {
				return UInt32ToIP(startIp - 1)
			}
			startIp++
		}
		return end
	}
	return UInt32ToIP(IPStringToUInt32(reuse(ips[0], ips[len(ips)-1])) - 1)
}

func init() {
	register(&Ip{}, &IpList{})
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
