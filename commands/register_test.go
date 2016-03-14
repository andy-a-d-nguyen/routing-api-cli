package commands_test

import (
	"github.com/cloudfoundry-incubator/routing-api-cli/commands"
	"github.com/cloudfoundry-incubator/routing-api/fake_routing_api"
	"github.com/cloudfoundry-incubator/routing-api/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe(".Register", func() {
	var (
		client *fake_routing_api.FakeClient
	)

	BeforeEach(func() {
		client = &fake_routing_api.FakeClient{}
	})

	It("registers routes", func() {
		routes := []models.Route{{}}
		commands.Register(client, routes)
		Expect(client.UpsertRoutesCallCount()).To(Equal(1))
		Expect(client.UpsertRoutesArgsForCall(0)).To(Equal(routes))
	})

})
