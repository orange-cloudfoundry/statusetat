package plugin

import (
	"fmt"
	"reflect"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/orange-cloudfoundry/statusetat/v2/common"
	"github.com/orange-cloudfoundry/statusetat/v2/models"
	"github.com/orange-cloudfoundry/statusetat/v2/notifiers/plugin/proto"
)

func IncidentToProto(incident models.Incident) *proto.Incident {
	return &proto.Incident{
		Guid:           incident.GUID,
		CreatedAt:      timeToLocalizedTime(incident.CreatedAt),
		UpdatedAt:      timeToLocalizedTime(incident.UpdatedAt),
		State:          proto.Incident_State(incident.State),
		ComponentState: proto.Incident_ComponentState(incident.ComponentState),
		Components:     componentsToProto(incident.Components),
		Messages:       messagesToProto(incident.Messages),
		Metadata:       metadataToMap(incident.Metadata),
		IsScheduled:    incident.IsScheduled,
		ScheduledEnd:   timeToLocalizedTime(incident.ScheduledEnd),
		Origin:         incident.Origin,
	}
}

func ProtoToIncident(incident *proto.Incident) models.Incident {
	return models.Incident{
		GUID:           incident.GetGuid(),
		CreatedAt:      localizedTimeToTime(incident.GetCreatedAt()),
		UpdatedAt:      localizedTimeToTime(incident.GetUpdatedAt()),
		State:          models.IncidentState(incident.GetState()),
		ComponentState: models.ComponentState(incident.GetComponentState()),
		Components:     protoToComponents(incident.GetComponents()),
		Messages:       protoToMessages(incident.GetMessages()),
		Metadata:       mapToMetadata(incident.GetMetadata()),
		IsScheduled:    incident.GetIsScheduled(),
		ScheduledEnd:   localizedTimeToTime(incident.GetScheduledEnd()),
		Origin:         incident.GetOrigin(),
	}
}

func NotifyRequestToProto(notifyRequest *models.NotifyRequest) *proto.NotifyRequest {
	return &proto.NotifyRequest{
		Incident:      IncidentToProto(notifyRequest.Incident),
		TriggerByUser: notifyRequest.TriggerByUser,
		Subscribers:   notifyRequest.Subscribers,
	}
}

func ProtoToNotifyRequest(notifyRequest *proto.NotifyRequest) *models.NotifyRequest {
	return &models.NotifyRequest{
		Incident:      ProtoToIncident(notifyRequest.Incident),
		TriggerByUser: notifyRequest.GetTriggerByUser(),
		Subscribers:   notifyRequest.GetSubscribers(),
	}
}

func timeToLocalizedTime(t time.Time) *proto.LocalizedTime {
	return &proto.LocalizedTime{
		Time:     timestamppb.New(t),
		Location: t.Location().String(),
	}
}

func localizedTimeToTime(t *proto.LocalizedTime) time.Time {
	newTime := time.Unix(t.GetTime().GetSeconds(), int64(t.GetTime().GetNanos()))
	if t.GetLocation() != "" {
		loc, err := time.LoadLocation(t.GetLocation())
		if err == nil {
			newTime = newTime.In(loc)
		}

	}
	return newTime
}

func metadataToMap(metadata []models.Metadata) map[string]string {
	return common.MetadataToMap(metadata)
}

func mapToMetadata(metadata map[string]string) []models.Metadata {
	metadataList := make([]models.Metadata, len(metadata))
	i := 0
	for k, v := range metadata {
		metadataList[i] = models.Metadata{
			Key:   k,
			Value: v,
		}
		i++
	}

	return metadataList
}

func messagesToProto(messages []models.Message) []*proto.Message {
	protoMessages := make([]*proto.Message, len(messages))
	for i, message := range messages {
		protoMessages[i] = messageToProto(message)
	}
	return protoMessages
}

func protoToMessages(protoMessages []*proto.Message) []models.Message {
	messages := make([]models.Message, len(protoMessages))
	for i, protoMessage := range protoMessages {
		messages[i] = protoToMessage(protoMessage)
	}
	return messages
}

func messageToProto(message models.Message) *proto.Message {
	return &proto.Message{
		Guid:      message.GUID,
		CreatedAt: timeToLocalizedTime(message.CreatedAt),
		Title:     message.Title,
		Content:   message.Content,
	}
}

func protoToMessage(message *proto.Message) models.Message {
	return models.Message{
		GUID:         message.GetGuid(),
		IncidentGUID: "",
		CreatedAt:    localizedTimeToTime(message.GetCreatedAt()),
		Title:        message.GetTitle(),
		Content:      message.GetContent(),
	}
}

func componentsToProto(components *models.Components) []*proto.Component {
	if components == nil {
		return make([]*proto.Component, 0)
	}
	protoComponents := make([]*proto.Component, len(*components))
	for i, component := range *components {
		protoComponents[i] = componentToProto(component)
	}
	return protoComponents
}

func protoToComponents(protoComponents []*proto.Component) *models.Components {
	components := make(models.Components, len(protoComponents))
	for i, protoComponent := range protoComponents {
		components[i] = protoToComponent(protoComponent)
	}
	return &components
}

func componentToProto(component models.Component) *proto.Component {
	return &proto.Component{
		Name:  component.Name,
		Group: component.Group,
	}
}

func protoToComponent(component *proto.Component) models.Component {
	return models.Component{
		Name:  component.GetName(),
		Group: component.GetGroup(),
	}
}

func listInterfaceToListString(val interface{}) []string {
	valOf := reflect.ValueOf(val)
	listString := make([]string, valOf.Len())
	for i := 0; i < valOf.Len(); i++ {
		listString[i] = fmt.Sprint(valOf.Index(i).Interface())
	}
	return listString
}

func mapInterfaceToListString(val interface{}) map[string]interface{} {
	valOf := reflect.ValueOf(val)
	m := make(map[string]interface{})
	for _, k := range valOf.MapKeys() {
		m[fmt.Sprint(k.Interface())] = valOf.MapIndex(k).Interface()
	}
	return m
}
