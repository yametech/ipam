package config

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestConfig(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Config Suite")
}

var _ = Describe("IPAM config", func() {
	It("Should parse config args", func() {
		input := `{
			"cniVersion": "0.4.0",
			"name": "mynet",
			"type": "macvlan",
			"master": "eth0",
			"args": {
				"cni": {
					"ips": [ "10.1.2.11", "11.11.11.11", "2001:db8:1::11"]
				}
			},
			"ipam": {
				"type": "host-local",
				"ranges": [
					[{
						"subnet": "10.1.2.0/24",
						"rangeStart": "10.1.2.9",
						"rangeEnd": "10.1.2.20",
						"gateway": "10.1.2.30"
					}],
					[{
						"subnet": "11.1.2.0/24",
						"rangeStart": "11.1.2.9",
						"rangeEnd": "11.1.2.20",
						"gateway": "11.1.2.30"
					}],
					[{
						"subnet": "2001:db8:1::/64"
					}]
				],
				"routes": [
					{ "dst": "0.0.0.0/0" },
					{ "dst": "192.168.0.0/16", "gw": "11.1.2.30" },
					{ "dst": "2001:db8:1::1/64" }
				]
			}
		}`

		envArgs := "IP=10.1.2.10;K8S_POD_NAME=abc;K8S_POD_NAMESPACE=xyz"

		conf, _, err := LoadConfig([]byte(input), envArgs)
		Expect(err).NotTo(HaveOccurred())
		_ = conf
		Expect(conf.Namespace).To(Equal("xyz"))
		Expect(conf.PodName).To(Equal("abc"))
		Expect(conf.CniVersion).To(Equal("0.4.0"))
		Expect(conf.Ipam.Ranges).ShouldNot(BeNil())
	})
})
