package helpers_test

import (
	"errors"
	"time"

	"github.com/cloudfoundry-incubator/routing-api/db"
	fake_db "github.com/cloudfoundry-incubator/routing-api/db/fakes"
	"github.com/cloudfoundry-incubator/routing-api/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-golang/lager/lagertest"
)

var _ = Describe("Helpers", func() {
	Describe("RegisterRoutingAPI", func() {
		var (
			database fake_db.FakeDB
			route    db.Route
			logger   *lagertest.TestLogger

			timeChan chan time.Time
			ticker   *time.Ticker

			received chan struct{}
			quitChan chan bool
		)

		BeforeEach(func() {
			route = db.Route{
				Route:   "i dont care",
				Port:    3000,
				IP:      "i dont care even more",
				TTL:     120,
				LogGuid: "i care a little bit more now",
			}
			database = fake_db.FakeDB{}
			logger = lagertest.NewTestLogger("event-handler-test")

			timeChan = make(chan time.Time)
			ticker = &time.Ticker{C: timeChan}

			received = make(chan struct{})

			quitChan = make(chan bool)
		})

		Context("registration", func() {
			Context("with no errors", func() {
				BeforeEach(func() {
					database.SaveRouteStub = func(route db.Route) error {
						received <- struct{}{}
						return nil
					}
				})

				It("registers the route for a routing api on init", func() {
					go helpers.RegisterRoutingAPI(quitChan, &database, route, ticker, logger)
					<-received

					Expect(database.SaveRouteCallCount()).To(Equal(1))
					Expect(database.SaveRouteArgsForCall(0)).To(Equal(route))
				})

				It("registers on an interval", func() {
					go helpers.RegisterRoutingAPI(quitChan, &database, route, ticker, logger)
					<-received
					timeChan <- time.Now()
					<-received

					Expect(database.SaveRouteCallCount()).To(Equal(2))
					Expect(database.SaveRouteArgsForCall(1)).To(Equal(route))
					Expect(len(logger.Logs())).To(Equal(0))
				})
			})

			Context("when there are errors", func() {
				It("only logs the error once for each attempt", func() {
					database.SaveRouteStub = func(route db.Route) error {
						received <- struct{}{}
						return errors.New("beep boop, self destruct mode engaged")
					}

					go helpers.RegisterRoutingAPI(quitChan, &database, route, ticker, logger)
					<-received
					Expect(logger.Logs()[0].Data["error"]).To(ContainSubstring("beep boop, self destruct mode engaged"))
					Expect(len(logger.Logs())).To(Equal(1))
				})
			})
		})

		Context("unregistration", func() {
			It("unregisters the routing api when a quit message is received", func() {
				go func() {
					quitChan <- true
				}()

				helpers.RegisterRoutingAPI(quitChan, &database, route, ticker, logger)

				Expect(database.DeleteRouteCallCount()).To(Equal(1))
				Expect(database.DeleteRouteArgsForCall(0)).To(Equal(route))
				Expect(quitChan).To(BeClosed())
			})
		})
	})
})
