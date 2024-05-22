package starling_node

import "time"

type ScenarioSpamRouteRequests struct {
	EmptyScenario
	delay time.Duration
}

func SpamRouteRequestScenario(delay time.Duration) *ScenarioSpamRouteRequests {
	return &ScenarioSpamRouteRequests{
		delay: delay,
	}
}

func (s *ScenarioSpamRouteRequests) OnStart(node *Node) {
	node.logf("scenario:spam_rreq:start:%d", node.ID())

	node.sim.DelayBy(func() {
		s.broadcast_route_request(node)
	}, s.delay)
}

func (s *ScenarioSpamRouteRequests) broadcast_route_request(node *Node) {
	node.proto.BroadcastRouteRequest()

	node.sim.DelayBy(func() {
		s.broadcast_route_request(node)
	}, s.delay)
}
