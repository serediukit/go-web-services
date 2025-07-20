package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"
)

// тут вы пишете код
// обращаю ваше внимание - в этом задании запрещены глобальные переменные
type BizServerImpl struct {
	UnimplementedBizServer
}

func (s *BizServerImpl) Check(context.Context, *Nothing) (*Nothing, error) {
	return &Nothing{}, nil
}

func (s *BizServerImpl) Add(context.Context, *Nothing) (*Nothing, error) {
	return &Nothing{}, nil
}

func (s *BizServerImpl) Test(context.Context, *Nothing) (*Nothing, error) {
	return &Nothing{}, nil
}

func NewBizServer() *BizServerImpl {
	return &BizServerImpl{}
}

type AdminServerImpl struct {
	UnimplementedAdminServer

	subs *EventSubs
}

func (s *AdminServerImpl) Logging(_ *Nothing, srv Admin_LoggingServer) error {
	id, events := s.subs.NewSub()
	defer s.subs.RemoveSub(id)

	for e := range events {
		if err := srv.Send(e); err != nil {
			return err
		}
	}

	return nil
}

func (s *AdminServerImpl) Statistics(si *StatInterval, srv Admin_StatisticsServer) error {
	id, events := s.subs.NewSub()
	defer s.subs.RemoveSub(id)

	statistics := newStatisticsCollector()

	t := time.NewTicker(time.Duration(si.IntervalSeconds) * time.Second)
	defer t.Stop()

	for {
		select {
		case e, ok := <-events:
			if ok {
				statistics.Update(e)
			} else {
				return nil
			}
		case <-t.C:
			if err := srv.Send(statistics.Collect()); err != nil {
				return err
			}
		}
	}
}

func NewAdminServer(s *EventSubs) *AdminServerImpl {
	return &AdminServerImpl{subs: s}
}

type EventSubs struct {
	id   int
	subs map[int]chan *Event
	mux  *sync.Mutex
}

func newEventSubs() *EventSubs {
	return &EventSubs{
		subs: map[int]chan *Event{},
		mux:  &sync.Mutex{},
	}
}

func (es *EventSubs) NewSub() (int, chan *Event) {
	es.mux.Lock()
	defer es.mux.Unlock()

	es.id++
	es.subs[es.id] = make(chan *Event)

	return es.id, es.subs[es.id]
}

func (es *EventSubs) RemoveSub(id int) {
	es.mux.Lock()
	defer es.mux.Unlock()

	if ch, ok := es.subs[id]; ok {
		close(ch)
		delete(es.subs, id)
	}
}

func (es *EventSubs) RemoveAll() {
	es.mux.Lock()
	defer es.mux.Unlock()

	for _, ch := range es.subs {
		close(ch)
	}
	es.subs = map[int]chan *Event{}
}

func (es *EventSubs) Notify(e *Event) {
	es.mux.Lock()
	defer es.mux.Unlock()

	for _, ch := range es.subs {
		ch <- e
	}
}

type StatisticsCollector struct {
	stat Stat
}

func newStatisticsCollector() *StatisticsCollector {
	s := &StatisticsCollector{}
	s.reset()

	return s
}

func (sc *StatisticsCollector) reset() {
	sc.stat = Stat{
		ByMethod:   map[string]uint64{},
		ByConsumer: map[string]uint64{},
	}
}

func (sc *StatisticsCollector) Update(e *Event) {
	sc.stat.ByMethod[e.Method]++
	sc.stat.ByConsumer[e.Consumer]++
}

func (sc *StatisticsCollector) Collect() *Stat {
	stat := Stat{
		ByMethod:   sc.stat.ByMethod,
		ByConsumer: sc.stat.ByConsumer,
	}
	stat.Timestamp = time.Now().Unix()
	sc.reset()

	return &stat
}

type aclMethods [][]string

type aclAuth struct {
	acl map[string]aclMethods
}

func newAclAuth(aclData string) (*aclAuth, error) {
	acl := make(map[string][]string)
	if err := json.Unmarshal([]byte(aclData), &acl); err != nil {
		return nil, fmt.Errorf("failed to parse ACL data: %s", err)
	}

	auth := &aclAuth{
		acl: make(map[string]aclMethods, len(acl)),
	}

	for consumer, methods := range acl {
		auth.acl[consumer] = make(aclMethods, len(methods))
		for i, method := range methods {
			auth.acl[consumer][i] = strings.Split(method, "/")
		}
	}

	return auth, nil
}

func (aa *aclAuth) isAllowed(consumer string, method string) bool {
	methodParts := strings.Split(method, "/")

	if allowedMethods, ok := aa.acl[consumer]; ok {
	allowedLoop:
		for _, allowedMethod := range allowedMethods {
			if len(allowedMethod) > len(methodParts) {
				continue
			}

			for methodIndex, methodPart := range allowedMethod {
				if methodParts[methodIndex] == methodPart || methodPart == "*" {
					continue
				} else {
					break allowedLoop
				}
			}

			return true
		}
	}

	return false
}

func StartMyMicroservice(ctx context.Context, listenAddr string, aclData string) error {
	acl, err := newAclAuth(aclData)
	if err != nil {
		return err
	}
}
