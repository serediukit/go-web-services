package main

import "context"

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
}

func (s *AdminServerImpl) Logging(context.Context, *Nothing) (*Event, error) {
	return &Event{}, nil
}

func (s *AdminServerImpl) Statistics(context.Context, *StatInterval) (*Stat, error) {
	return &Stat{}, nil
}

func NewAdminServer() *AdminServerImpl {
	return &AdminServerImpl{}
}
