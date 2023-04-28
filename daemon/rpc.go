// Package daemon provides gRPC interface for management of vpn on the device and various related functionalities,
// such as communication with the backend api and configuration management.
package daemon

import (
	"time"

	"github.com/NordSecurity/nordvpn-linux/auth"
	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/NordSecurity/nordvpn-linux/daemon/dns"
	"github.com/NordSecurity/nordvpn-linux/daemon/firewall"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/meshnet"
	"github.com/NordSecurity/nordvpn-linux/nc"
	"github.com/NordSecurity/nordvpn-linux/network"
	"github.com/NordSecurity/nordvpn-linux/networker"
	"github.com/NordSecurity/nordvpn-linux/request"

	"github.com/go-co-op/gocron"
)

// RPC is a gRPC server.
type RPC struct {
	environment     internal.Environment
	ac              auth.Checker
	cm              config.Manager
	dm              *DataManager
	api             *core.DefaultAPI
	serversAPI      core.ServersAPI
	credentialsAPI  core.CredentialsAPI
	cdn             core.CDN
	repo            *RepoAPI
	authentication  core.Authentication
	lastServer      core.Server
	version         string
	systemInfoFunc  func(string) string
	networkInfoFunc func() string
	httpClient      *request.HTTPClient
	events          *Events
	// factory picks which VPN implementation to use
	factory          FactoryFunc
	endpointResolver network.EndpointResolver
	endpoint         network.Endpoint
	scheduler        *gocron.Scheduler
	netw             networker.Networker
	publisher        events.Publisher[string]
	nameservers      dns.Getter
	ncClient         nc.NotificationClient
	supportChecker   SupportChecker
	analytics        events.Analytics
	fileshare        meshnet.Fileshare
	pb.UnimplementedDaemonServer
}

func NewRPC(
	environment internal.Environment,
	ac auth.Checker,
	cm config.Manager,
	dm *DataManager,
	api *core.DefaultAPI,
	serversAPI core.ServersAPI,
	credentialsAPI core.CredentialsAPI,
	cdn core.CDN,
	repo *RepoAPI,
	authentication core.Authentication,
	version string,
	fw firewall.Service,
	httpClient *request.HTTPClient,
	events *Events,
	factory FactoryFunc,
	endpointResolver network.EndpointResolver,
	netw networker.Networker,
	publisher events.Publisher[string],
	nameservers dns.Getter,
	ncClient nc.NotificationClient,
	supportChecker SupportChecker,
	analytics events.Analytics,
	fileshare meshnet.Fileshare,
) *RPC {
	return &RPC{
		environment:      environment,
		ac:               ac,
		cm:               cm,
		dm:               dm,
		api:              api,
		serversAPI:       serversAPI,
		credentialsAPI:   credentialsAPI,
		cdn:              cdn,
		repo:             repo,
		authentication:   authentication,
		version:          version,
		systemInfoFunc:   getSystemInfo,
		networkInfoFunc:  getNetworkInfo,
		httpClient:       httpClient,
		factory:          factory,
		events:           events,
		endpointResolver: endpointResolver,
		scheduler:        gocron.NewScheduler(time.UTC),
		netw:             netw,
		publisher:        publisher,
		nameservers:      nameservers,
		ncClient:         ncClient,
		supportChecker:   supportChecker,
		analytics:        analytics,
		fileshare:        fileshare,
	}
}