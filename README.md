# ipam

```
kubectl apply -f deploy/yamecloud.io_ip.yaml
kubectl apply -f deploy/rbac.yaml
kubectl apply -f deploy/global-ipam.yaml
```

## on linux system

```
mkdir -p /opt/cni/bin/ && cd  /opt/cni/bin/

## install plugin
git clone https://github.com/containernetworking/plugins.git
cd plugins
## install plugin

```

## install cnitool

```
git clone https://github.com/containernetworking/cni.git
cd cni/cnitool/
go build -o /usr/local/bin/cnitool cnitool.go

cnitool

# output
cnitool: Add, check, or remove network interfaces from a network namespace
cnitool add   <net> <netns>
cnitool check <net> <netns>
cnitool del   <net> <netns>


git clone https://github.com/yametech/global-ipam.git
cd global-ipam
rm -rf /opt/cni/bin/global-ipam && go build -o /opt/cni/bin/global-ipam cmd/cni/main.go
```

## install macvlan & global-ipam config

```
cat >/etc/cni/net.d/10-macvlan-global-ipam.conf  << "EOF"
{
    "name": "macvlan-global-ipam",
    "type": "macvlan",
    "cniVersion": "0.4.0",
    "master": "eth0",
    "args": {
        "cni": {
            "ips": [ "10.211.55.20", "2001:db8:1::11"]
        }
    },
    "ipam": {
        "type": "global-ipam",
        "ranges": [
            [{
                "subnet": "10.211.55.0/24",
                "rangeStart": "10.211.55.20",
                "rangeEnd": "10.211.55.40",
                "gateway": "10.211.55.1"
            }]
        ],
        "routes": [
            { "dst": "0.0.0.0/0" }
        ]
    }
}
EOF


export CNI_PATH=/opt/cni/bin/
# delete ns a
ip netns delete a

# if not exists create

ip netns add a
CNI_ARGS="K8S_POD_NAME=abc;K8S_POD_NAMESPACE=xyz" cnitool add macvlan-global-ipam /var/run/netns/a
CNI_ARGS="K8S_POD_NAME=abc;K8S_POD_NAMESPACE=xyz" cnitool del macvlan-global-ipam /var/run/netns/a

```

# check ns ip addr

```
ip netns exec a ip addr
```

