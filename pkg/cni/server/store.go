package server

import (
	"context"
	"fmt"
	"net"

	"github.com/yametech/global-ipam/pkg/allocator"
	v1 "github.com/yametech/global-ipam/pkg/apis/yamecloud/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

var _ allocator.Store = &Server{}

func (s Server) LastReservedIP(ctx context.Context) (net.IP, error) {
	ipList, err := s.Interface.Resource(IP).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	runtimeIpList := &v1.IpList{}
	if err := runtime.DefaultUnstructuredConverter.
		FromUnstructured(ipList.UnstructuredContent(), runtimeIpList); err != nil {
		return nil, err
	}
	
	return net.ParseIP(runtimeIpList.Reuse()), nil
}

func (s Server) Reserve(ctx context.Context, namespace, pod string, requestedIp string) error {
	netIP, netIPnet, err := net.ParseCIDR(requestedIp)
	if err != nil {
		return err
	}
	ip := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "yamecloud.io/v1",
			"kind":       "Ip",
			"metadata": map[string]interface{}{
				"name": fmt.Sprintf("ip-%d", v1.IPStringToUInt32(netIP.String())),
				"labels": map[string]interface{}{
					"namespace": namespace,
					"pod":       pod,
				},
			},
			"spec": v1.IPSpec{
				Ip:   netIP.String(),
				Mask: string(netIPnet.Mask),
			},
		},
	}
	_, err = s.Interface.Resource(IP).Create(ctx, ip, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	return nil
}
