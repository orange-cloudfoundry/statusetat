package plugin

import (
	"context"
	"fmt"
	"github.com/orange-cloudfoundry/statusetat/common"
	"github.com/orange-cloudfoundry/statusetat/config"
	"github.com/orange-cloudfoundry/statusetat/models"
	"github.com/orange-cloudfoundry/statusetat/notifiers/plugin/proto"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/structpb"
)

type GRPCClient struct {
	client proto.NotifierClient
}

func (m *GRPCClient) Init(baseInfo config.BaseInfo, params map[string]interface{}) error {
	paramsStruct, err := structpb.NewStruct(common.CleanupMap(params))
	if err != nil {
		return err
	}
	_, err = m.client.Init(context.Background(), &proto.InitRequest{
		BaseInfo: &proto.BaseInfo{
			BaseUrl:  baseInfo.BaseURL,
			Support:  baseInfo.Support,
			Contact:  baseInfo.Contact,
			Title:    baseInfo.Title,
			TimeZone: baseInfo.TimeZone,
		},
		Params: paramsStruct,
	})
	return err
}

func (m *GRPCClient) Name() (string, error) {
	resp, err := m.client.Name(context.Background(), &emptypb.Empty{})
	if err != nil {
		return "", err
	}
	return resp.GetName(), nil
}

func (m *GRPCClient) Id() (string, error) {
	resp, err := m.client.Id(context.Background(), &emptypb.Empty{})
	if err != nil {
		return "", err
	}
	return resp.GetId(), nil
}

func (m *GRPCClient) Notify(incident models.Incident) error {
	resp, err := m.client.Notify(context.Background(), &proto.NotifyRequest{
		Incident: IncidentToProto(incident),
	})
	if err != nil {
		return err
	}
	if resp.GetError() != nil {
		return fmt.Errorf(resp.GetError().GetDetail())
	}
	return nil
}

func (m *GRPCClient) NotifySubscriber(incident models.Incident, subscribers []string) error {
	resp, err := m.client.NotifySubscriber(context.Background(), &proto.NotifySubscriberRequest{
		Incident:    IncidentToProto(incident),
		Subscribers: subscribers,
	})
	if err != nil {
		return err
	}
	if resp.GetError() != nil {
		return fmt.Errorf(resp.GetError().GetDetail())
	}
	return nil
}

func (m *GRPCClient) MetadataFields() ([]models.MetadataField, error) {

	resp, err := m.client.MetadataFields(context.Background(), &emptypb.Empty{})
	if err != nil {
		return nil, err
	}

	fields := make([]models.MetadataField, len(resp.GetFields()))
	for i, field := range resp.GetFields() {
		m := models.MetadataField{
			Name:         field.GetName(),
			Id:           field.GetId(),
			Info:         field.GetInfo(),
			InputType:    models.InputTypeMetadata(field.GetInputType()),
			ForScheduled: field.GetForScheduled(),
			DefaultOpt:   field.GetDefaultOptKey(),
		}
		if field.GetRadioOpts() != nil {
			m.Opts = field.GetRadioOpts().Opts
		}
		if field.GetSelectOpts() != nil {
			m.Opts = field.GetSelectOpts().AsMap()
		}
		fields[i] = m
	}
	return fields, err

}
