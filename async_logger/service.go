package main

import (
	"context"
	"sync"
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

func (s *AdminServerImpl) Logging(_ *Nothing, srv Admin_LoggingServer) (*Event, error) {
	return &Event{}, nil
}

func (s *AdminServerImpl) Statistics(_ *StatInterval, srv Admin_StatisticsServer) (*Stat, error) {
	return &Stat{}, nil
}

func NewAdminServer() *AdminServerImpl {
	return &AdminServerImpl{}
}

type EventSubs struct {
	id   int
	subs map[int]chan *Event
	mux  *sync.Mutex
}

func NewEventSubs() *EventSubs {
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
