package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	typesVer "github.com/containernetworking/cni/pkg/types/100"
	"github.com/gin-gonic/gin"
	"github.com/yametech/global-ipam/pkg/allocator"
	v1 "github.com/yametech/global-ipam/pkg/apis/yamecloud/v1"
	"github.com/yametech/global-ipam/pkg/cni"
	"github.com/yametech/global-ipam/pkg/config"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var IP = schema.GroupVersionResource{Group: "yamecloud.io", Version: "v1", Resource: "ip"}

func (s *Server) Release(g *gin.Context) {
	ns := g.PostForm("namespace")
	pod := g.PostForm("pod")
	defaultResponse := &cni.ReleaseResponse{IsRelease: true}
	ips, err := s.Interface.Resource(IP).List(g.Request.Context(), metav1.ListOptions{
		LabelSelector: fmt.Sprintf("namespace=%s,pod=%s", ns, pod),
	})
	if err != nil {
		if errors.IsNotFound(err) {
			g.JSON(http.StatusOK, defaultResponse)
			return
		}
		defaultResponse.Error = err
		defaultResponse.IsRelease = false
		g.JSON(http.StatusInternalServerError, defaultResponse)
	}

	for _, ip := range ips.Items {
		err := s.Interface.Resource(IP).Delete(g.Request.Context(), ip.GetName(), metav1.DeleteOptions{})
		if err != nil {
			defaultResponse.Error = err
			defaultResponse.IsRelease = false
			g.JSON(http.StatusInternalServerError, defaultResponse)
		}
	}

	g.JSON(http.StatusOK, defaultResponse)
}

func (s *Server) Allocate(g *gin.Context) {
	ns := g.PostForm("namespace")
	pod := g.PostForm("pod")
	cfg := g.PostForm("cfg")

	response := &cni.AllocateResponse{Reserved: false}

	netCfg := &config.Net{}
	if err := json.Unmarshal([]byte(cfg), netCfg); err != nil {
		response.Error = err
		g.JSON(http.StatusBadRequest, response)
		return
	}

RLLOCATE:
	result, err := s.allocateIp(g.Request.Context(), ns, pod, netCfg.Ipam.Ranges)
	if err != nil {
		if errors.IsAlreadyExists(err) {
			goto RLLOCATE
		}
		response.Error = err
		g.JSON(http.StatusInternalServerError, response)
		return
	}
	result.Routes = netCfg.Ipam.Routes
	response.Result = result

	g.JSON(http.StatusOK, response)
}

func (s *Server) allocateIp(ctx context.Context, ns, pod string, rss []allocator.RangeSet) (*typesVer.Result, error) {
	ips, err := s.Interface.Resource(IP).List(ctx, metav1.ListOptions{})
	if err != nil {
		if !errors.IsNotFound(err) {
			goto IGNORE
		}
		return nil, err
	}

IGNORE:
	ipList := v1.IpList{}
	if ips != nil {
		if err := runtime.DefaultUnstructuredConverter.FromUnstructured(ips.UnstructuredContent(), &ipList); err != nil {
			return nil, err
		}
	}
	result := &typesVer.Result{}
	for _, rs := range rss {
		ipConf, err := allocator.NewIPAllocator(&rs, s, ipList.Ips()).Allocate(ns, pod)
		if err != nil {
			return nil, err
		}
		result.IPs = append(result.IPs, ipConf)
	}
	return result, nil
}
