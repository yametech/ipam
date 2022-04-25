package main

import (
	"github.com/containernetworking/cni/pkg/skel"
	"github.com/containernetworking/cni/pkg/types"
	"github.com/containernetworking/cni/pkg/version"
	"github.com/containernetworking/plugins/pkg/utils/buildversion"
	"github.com/yametech/global-ipam/pkg/cni"
	"github.com/yametech/global-ipam/pkg/cni/client"
	"github.com/yametech/global-ipam/pkg/config"
)

func main() {
	skel.PluginMain(addCmd, chekCmd, delCmd, version.All, buildversion.BuildString(cni.GLOBAL_IPAM))
}

func chekCmd(args *skel.CmdArgs) error { return nil }

func addCmd(args *skel.CmdArgs) error {
	netCfg, ver, err := config.LoadConfig(args.StdinData, args.Args)
	if err != nil {
		return err
	}
	res, err := client.New().Allocate(netCfg.Namespace, netCfg.PodName, netCfg)
	if err != nil {
		return err
	}
	return types.PrintResult(res, ver)
}

func delCmd(args *skel.CmdArgs) error {
	netCfg, ver, err := config.LoadConfig(args.StdinData, args.Args)
	if err != nil {
		return err
	}
	_ = ver
	return client.New().Release(netCfg.Namespace, netCfg.PodName)
}
