package plugin

import (
	"context"
	"fmt"

	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/orange-cloudfoundry/statusetat/config"
	"github.com/orange-cloudfoundry/statusetat/models"
	"github.com/orange-cloudfoundry/statusetat/notifiers/plugin/proto"
)

type GRPCServer struct {
	proto.UnsafeNotifierServer
	// This is the real implementation
	Impl Notifier
}

func (s *GRPCServer) Init(ctx context.Context, request *proto.InitRequest) (*emptypb.Empty, error) {
	err := s.Impl.Init(
		config.BaseInfo{
			BaseURL:  request.GetBaseInfo().GetBaseUrl(),
			Support:  request.GetBaseInfo().GetSupport(),
			Contact:  request.GetBaseInfo().GetContact(),
			Title:    request.GetBaseInfo().GetTitle(),
			TimeZone: request.GetBaseInfo().GetTimeZone(),
		},
		request.GetParams().AsMap(),
	)
	if err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (s *GRPCServer) Name(ctx context.Context, request *emptypb.Empty) (*proto.NameResponse, error) {
	name, err := s.Impl.Name()
	if err != nil {
		return nil, err
	}
	return &proto.NameResponse{Name: name}, nil
}

func (s *GRPCServer) Description(ctx context.Context, empty *emptypb.Empty) (*proto.DescriptionResponse, error) {
	desc, err := s.Impl.Description()
	if err != nil {
		return nil, err
	}
	return &proto.DescriptionResponse{Description: desc}, nil
}

func (s *GRPCServer) Id(ctx context.Context, request *emptypb.Empty) (*proto.IdResponse, error) {
	id, err := s.Impl.Id()
	if err != nil {
		return nil, err
	}
	return &proto.IdResponse{Id: id}, nil
}

func (s *GRPCServer) Notify(ctx context.Context, request *proto.NotifyRequest) (*proto.ErrorResponse, error) {
	err := s.Impl.Notify(ProtoToNotifyRequest(request))
	if err != nil {
		return &proto.ErrorResponse{
			Error: &proto.Error{
				Detail: err.Error(),
			},
		}, nil
	}
	return &proto.ErrorResponse{}, nil
}

func (s *GRPCServer) MetadataFields(ctx context.Context, request *emptypb.Empty) (*proto.ListMetadataField, error) {
	fields, err := s.Impl.MetadataFields()
	if err != nil {
		return nil, err
	}
	protoFields := make([]*proto.MetadataField, len(fields))
	for i, field := range fields {
		err := field.Validate()
		if err != nil {
			return nil, err
		}
		m := &proto.MetadataField{
			Name:         field.Name,
			Id:           field.Id,
			Info:         field.Info,
			InputType:    proto.MetadataField_InputType(field.InputType),
			ForScheduled: field.ForScheduled,
		}
		if field.InputType == models.Radio {
			m.Opts = &proto.MetadataField_RadioOpts{
				RadioOpts: &proto.RadioOpts{
					Opts: listInterfaceToListString(field.Opts),
				},
			}
		}
		if field.InputType == models.Select {
			st, err := structpb.NewStruct(mapInterfaceToListString(field.Opts))
			if err != nil {
				return nil, err
			}
			m.Opts = &proto.MetadataField_SelectOpts{
				SelectOpts: st,
			}
		}
		if field.DefaultOpt != nil {
			m.DefaultOpt = &proto.MetadataField_DefaultOptKey{
				DefaultOptKey: fmt.Sprint(field.DefaultOpt),
			}
		}
		protoFields[i] = m
	}
	return &proto.ListMetadataField{Fields: protoFields}, nil
}

func (s *GRPCServer) PreCheck(ctx context.Context, request *proto.NotifyRequest) (*proto.ErrorResponse, error) {
	protoToIncident := ProtoToIncident(request.Incident)
	err := s.Impl.PreCheck(&protoToIncident)
	if err != nil {
		return &proto.ErrorResponse{
			Error: &proto.Error{
				Detail: err.Error(),
			},
		}, nil
	}
	return &proto.ErrorResponse{}, nil
}
