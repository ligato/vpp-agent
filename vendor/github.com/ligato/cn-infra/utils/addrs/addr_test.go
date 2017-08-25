package addrs

import (
	"testing"

	"github.com/onsi/gomega"
)

func TestMacIntToString(t *testing.T) {
	gomega.RegisterTestingT(t)
	res := MacIntToString(0)
	gomega.Expect(res).To(gomega.BeEquivalentTo("00:00:00:00:00:00"))

	res = MacIntToString(255)
	gomega.Expect(res).To(gomega.BeEquivalentTo("00:00:00:00:00:ff"))
}

func TestParseIPWithPrefix(t *testing.T) {
	gomega.RegisterTestingT(t)

	ip, isIpv6, err := ParseIPWithPrefix("127.0.0.1")
	gomega.Expect(err).To(gomega.BeNil())
	gomega.Expect(isIpv6).To(gomega.BeFalse())
	gomega.Expect(ip.IP.String()).To(gomega.BeEquivalentTo("127.0.0.1"))
	maskOnes, maskBits := ip.Mask.Size()
	gomega.Expect(maskOnes).To(gomega.BeEquivalentTo(32))
	gomega.Expect(maskBits).To(gomega.BeEquivalentTo(32))

	ip, isIpv6, err = ParseIPWithPrefix("192.168.2.100/24")
	gomega.Expect(err).To(gomega.BeNil())
	gomega.Expect(isIpv6).To(gomega.BeFalse())
	gomega.Expect(ip.IP.String()).To(gomega.BeEquivalentTo("192.168.2.100"))

	ip, isIpv6, err = ParseIPWithPrefix("2001:db9::54")
	gomega.Expect(err).To(gomega.BeNil())
	gomega.Expect(isIpv6).To(gomega.BeTrue())
	gomega.Expect(ip.IP.String()).To(gomega.BeEquivalentTo("2001:db9::54"))
	maskOnes, maskBits = ip.Mask.Size()
	gomega.Expect(maskOnes).To(gomega.BeEquivalentTo(128))
	gomega.Expect(maskBits).To(gomega.BeEquivalentTo(128))

	ip, isIpv6, err = ParseIPWithPrefix("2001:db8::68/120")
	gomega.Expect(err).To(gomega.BeNil())
	gomega.Expect(isIpv6).To(gomega.BeTrue())
	gomega.Expect(ip.IP.String()).To(gomega.BeEquivalentTo("2001:db8::68"))
	maskOnes, maskBits = ip.Mask.Size()
	gomega.Expect(maskOnes).To(gomega.BeEquivalentTo(120))
	gomega.Expect(maskBits).To(gomega.BeEquivalentTo(128))

	_, _, err = ParseIPWithPrefix("127.0.0.1/abcd")
	gomega.Expect(err).NotTo(gomega.BeNil())
}
