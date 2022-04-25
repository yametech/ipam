package server

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/yametech/global-ipam/pkg/cni"
	"github.com/yametech/global-ipam/pkg/common"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)


type Server struct {
	*http.Server
	dynamic.Interface
}

func NewServer(ctx context.Context) (*Server, error) {
	route := gin.Default()
	s := &Server{
		Server: &http.Server{
			Handler: route,
		},
	}
	go func() {
		for range ctx.Done() {
			if err := s.Shutdown(ctx); err != nil {
				fmt.Printf("shutdown server error: %v", err)
			}
		}
	}()

	config, err := createRestConfig()
	if err != nil {
		return nil, err
	}

	k8sclient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	s.Interface = k8sclient

	{
		route.POST("/allocate", s.Allocate)
		route.POST("/release", s.Release)
	}

	return s, nil
}

func (s *Server) Start() error {
	if err := os.Remove(cni.UNIX_SOCK_PATH); err != nil {
		fmt.Printf("remove unix socket error: %v\r\n", err)
	}

	defer func() {
		if err := os.Remove(cni.UNIX_SOCK_PATH); err != nil {
			fmt.Printf("remove unix socket error: %v\r\n", err)
		}
	}()

	unixListener, err := net.Listen("unix", cni.UNIX_SOCK_PATH)
	if err != nil {
		return err
	}

	return s.Serve(unixListener)
}

func createRestConfig() (*rest.Config, error) {
	if common.InCluster {
		return createCfgFromCluster()
	}
	return createCfgFromPath()
}

func createCfgFromPath() (*rest.Config, error) {
	cfg, err := clientcmd.BuildConfigFromFlags("", *common.KubeConfig)
	if err != nil {
		return nil, err
	}
	return applyDefaultRateLimiter(cfg, 2), nil
}

func createCfgFromCluster() (*rest.Config, error) {
	cfg, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	return applyDefaultRateLimiter(cfg, 2), nil
}

func applyDefaultRateLimiter(config *rest.Config, flowRate int) *rest.Config {
	if flowRate < 0 {
		flowRate = 1
	}
	// here we magnify the default qps and burst in client-go
	config.QPS = rest.DefaultQPS * float32(flowRate)
	config.Burst = rest.DefaultBurst * flowRate

	return config
}
