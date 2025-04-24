package plugin

import (
	"context"

	pluginhc "github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"

	"github.com/orange-cloudfoundry/statusetat/v2/config"
	"github.com/orange-cloudfoundry/statusetat/v2/models"
	"github.com/orange-cloudfoundry/statusetat/v2/notifiers/plugin/proto"
)

// Handshake is a common handshake that is shared by plugin and host.
var Handshake = pluginhc.HandshakeConfig{
	// This isn't required when using VersionedPlugins
	ProtocolVersion:  1,
	MagicCookieKey:   "BASIC_PLUGIN",
	MagicCookieValue: "hello",
}

type Base struct {
	BaseInfo config.BaseInfo
	Params   map[string]interface{}
}

type Notifier interface {
	Init(baseInfo config.BaseInfo, params map[string]interface{}) error
	Name() (string, error)
	Description() (string, error)
	Id() (string, error)
	MetadataFields() ([]models.MetadataField, error)
	Notify(notifyReq *models.NotifyRequest) error
	PreCheck(incident *models.Incident) error
}

type NotifierGRPCPlugin struct {
	// GRPCPlugin must still implement the Plugin interface
	pluginhc.Plugin
	// Concrete implementation, written in Go. This is only used for plugins
	// that are written in Go.
	Impl Notifier
}

func (p *NotifierGRPCPlugin) GRPCServer(broker *pluginhc.GRPCBroker, s *grpc.Server) error {
	proto.RegisterNotifierServer(s, &GRPCServer{Impl: p.Impl})
	return nil
}

func (p *NotifierGRPCPlugin) GRPCClient(ctx context.Context, broker *pluginhc.GRPCBroker, c *grpc.ClientConn) (interface{}, error) {
	return &GRPCClient{client: proto.NewNotifierClient(c)}, nil
}
