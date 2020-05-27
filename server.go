package main

import (
	"context"
	"fmt"

	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/duration"
	"github.com/golang/protobuf/ptypes/empty"

	pb "github.com/slavaromanov/cyberbet-test-task/proto"
	"github.com/slavaromanov/cyberbet-test-task/storage"
)

var emp = &empty.Empty{}

type server struct {
	Storage *storage.Storage
}

func (s *server) Set(ctx context.Context, kv *pb.SetKeyValueRequest) (*empty.Empty, error) {
	if kv.Key == "" {
		return nil, fmt.Errorf("Key is empty!")
	}
	s.Storage.Put(kv.Key, kv.Value)
	return emp, nil
}

func (s *server) SetWithTTL(ctx context.Context, kvTTL *pb.SetKeyValueWithTTLRequest) (*empty.Empty, error) {
	dur, err := ptypes.Duration(kvTTL.Ttl)
	if err != nil {
		return nil, err
	}
	if kvTTL.Key == "" {
		return nil, fmt.Errorf("Key is empty!")
	}
	s.Storage.PutWithTTL(kvTTL.Key, kvTTL.Value, dur)
	return emp, nil
}

func (s *server) SetTTL(ctx context.Context, setTTL *pb.SetKeyTTLRequest) (*empty.Empty, error) {
	dur, err := ptypes.Duration(setTTL.Ttl)
	if err != nil {
		return emp, err
	}
	if setTTL.Key == "" {
		return nil, fmt.Errorf("Key is empty!")
	}
	s.Storage.SetTTL(setTTL.Key, dur)
	return emp, nil
}

func (s *server) GetValue(ctx context.Context, kReq *pb.ByKeyRequest) (*pb.ValueResponse, error) {
	val, err := s.Storage.Get(kReq.Key)
	if err != nil {
		return nil, err
	}
	return &pb.ValueResponse{Value: val}, nil
}

func (s *server) GetTTL(ctx context.Context, kReq *pb.ByKeyRequest) (*duration.Duration, error) {
	item, err := s.Storage.GetItem(kReq.Key)
	if err != nil {
		return nil, err
	}
	return ptypes.DurationProto(item.ExpiredAfter()), nil
}

func (s *server) GetValues(ctx context.Context, empt *empty.Empty) (*pb.GetValuesResponse, error) {
	return &pb.GetValuesResponse{Values: s.Storage.Values()}, nil
}

func (s *server) Delete(ctx context.Context, kReq *pb.ByKeyRequest) (*empty.Empty, error) {
	return emp, s.Storage.Delete(kReq.Key)
}
