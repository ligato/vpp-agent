package then

import (
	"strings"
	"github.com/onsi/gomega"
)

// ContainsItems can be used to verify if the provided item(s) is present in the table. It could be an agent label, an interface or
// a header
func ContainsItems(data string, item... string) {
	for _, header := range item {
		itemExists := strings.Contains(data, header)
		gomega.Expect(itemExists).To(gomega.BeTrue())
	}
}

// DoesNotContainItems can be used to verify if the provided item(s) is missing in the table. It could be an agent label, an interface or
// a header
func DoesNotContainItems(data string, item... string) {
	for _, header := range item {
		itemExists := strings.Contains(data, header)
		gomega.Expect(itemExists).To(gomega.BeFalse())
	}
}
