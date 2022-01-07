package apistruct

import (
	"context"

	"github.com/filecoin-project/go-jsonrpc/auth"
	"github.com/google/uuid"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	metrics "github.com/libp2p/go-libp2p-metrics"
	protocol "github.com/libp2p/go-libp2p-protocol"

	"github.com/EpiK-Protocol/go-epik-gateway/epik/api"
)

type CommonStruct struct {
	Internal struct {
		AuthVerify func(ctx context.Context, token string) ([]auth.Permission, error) `perm:"read"`
		AuthNew    func(ctx context.Context, perms []auth.Permission) ([]byte, error) `perm:"admin"`

		NetConnectedness            func(context.Context, peer.ID) (network.Connectedness, error)    `perm:"read"`
		NetPeers                    func(context.Context) ([]peer.AddrInfo, error)                   `perm:"read"`
		NetConnect                  func(context.Context, peer.AddrInfo) error                       `perm:"write"`
		NetAddrsListen              func(context.Context) (peer.AddrInfo, error)                     `perm:"read"`
		NetDisconnect               func(context.Context, peer.ID) error                             `perm:"write"`
		NetFindPeer                 func(context.Context, peer.ID) (peer.AddrInfo, error)            `perm:"read"`
		NetPubsubScores             func(context.Context) ([]api.PubsubScore, error)                 `perm:"read"`
		NetAutoNatStatus            func(context.Context) (api.NatInfo, error)                       `perm:"read"`
		NetBandwidthStats           func(ctx context.Context) (metrics.Stats, error)                 `perm:"read"`
		NetBandwidthStatsByPeer     func(ctx context.Context) (map[string]metrics.Stats, error)      `perm:"read"`
		NetBandwidthStatsByProtocol func(ctx context.Context) (map[protocol.ID]metrics.Stats, error) `perm:"read"`
		NetAgentVersion             func(ctx context.Context, p peer.ID) (string, error)             `perm:"read"`
		NetPeerInfo                 func(context.Context, peer.ID) (*api.ExtendedPeerInfo, error)    `perm:"read"`
		NetBlockAdd                 func(ctx context.Context, acl api.NetBlockList) error            `perm:"admin"`
		NetBlockRemove              func(ctx context.Context, acl api.NetBlockList) error            `perm:"admin"`
		NetBlockList                func(ctx context.Context) (api.NetBlockList, error)              `perm:"read"`

		ID      func(context.Context) (peer.ID, error)        `perm:"read"`
		Version func(context.Context) (api.APIVersion, error) `perm:"read"`

		LogList     func(context.Context) ([]string, error)     `perm:"write"`
		LogSetLevel func(context.Context, string, string) error `perm:"write"`

		GC func(context.Context) error `perm:"admin"`

		Shutdown func(context.Context) error                    `perm:"admin"`
		Session  func(context.Context) (uuid.UUID, error)       `perm:"read"`
		Closing  func(context.Context) (<-chan struct{}, error) `perm:"read"`
	}
}

// CommonStruct

func (c *CommonStruct) AuthVerify(ctx context.Context, token string) ([]auth.Permission, error) {
	return c.Internal.AuthVerify(ctx, token)
}

func (c *CommonStruct) AuthNew(ctx context.Context, perms []auth.Permission) ([]byte, error) {
	return c.Internal.AuthNew(ctx, perms)
}

func (c *CommonStruct) NetPubsubScores(ctx context.Context) ([]api.PubsubScore, error) {
	return c.Internal.NetPubsubScores(ctx)
}

func (c *CommonStruct) NetConnectedness(ctx context.Context, pid peer.ID) (network.Connectedness, error) {
	return c.Internal.NetConnectedness(ctx, pid)
}

func (c *CommonStruct) NetPeers(ctx context.Context) ([]peer.AddrInfo, error) {
	return c.Internal.NetPeers(ctx)
}

func (c *CommonStruct) NetConnect(ctx context.Context, p peer.AddrInfo) error {
	return c.Internal.NetConnect(ctx, p)
}

func (c *CommonStruct) NetAddrsListen(ctx context.Context) (peer.AddrInfo, error) {
	return c.Internal.NetAddrsListen(ctx)
}

func (c *CommonStruct) NetDisconnect(ctx context.Context, p peer.ID) error {
	return c.Internal.NetDisconnect(ctx, p)
}

func (c *CommonStruct) NetFindPeer(ctx context.Context, p peer.ID) (peer.AddrInfo, error) {
	return c.Internal.NetFindPeer(ctx, p)
}

func (c *CommonStruct) NetAutoNatStatus(ctx context.Context) (api.NatInfo, error) {
	return c.Internal.NetAutoNatStatus(ctx)
}

func (c *CommonStruct) NetBandwidthStats(ctx context.Context) (metrics.Stats, error) {
	return c.Internal.NetBandwidthStats(ctx)
}

func (c *CommonStruct) NetBandwidthStatsByPeer(ctx context.Context) (map[string]metrics.Stats, error) {
	return c.Internal.NetBandwidthStatsByPeer(ctx)
}

func (c *CommonStruct) NetBandwidthStatsByProtocol(ctx context.Context) (map[protocol.ID]metrics.Stats, error) {
	return c.Internal.NetBandwidthStatsByProtocol(ctx)
}

func (c *CommonStruct) NetBlockAdd(ctx context.Context, acl api.NetBlockList) error {
	return c.Internal.NetBlockAdd(ctx, acl)
}

func (c *CommonStruct) NetBlockRemove(ctx context.Context, acl api.NetBlockList) error {
	return c.Internal.NetBlockRemove(ctx, acl)
}

func (c *CommonStruct) NetBlockList(ctx context.Context) (api.NetBlockList, error) {
	return c.Internal.NetBlockList(ctx)
}

func (c *CommonStruct) NetAgentVersion(ctx context.Context, p peer.ID) (string, error) {
	return c.Internal.NetAgentVersion(ctx, p)
}

func (c *CommonStruct) NetPeerInfo(ctx context.Context, p peer.ID) (*api.ExtendedPeerInfo, error) {
	return c.Internal.NetPeerInfo(ctx, p)
}

// ID implements API.ID
func (c *CommonStruct) ID(ctx context.Context) (peer.ID, error) {
	return c.Internal.ID(ctx)
}

// Version implements API.Version
func (c *CommonStruct) Version(ctx context.Context) (api.APIVersion, error) {
	return c.Internal.Version(ctx)
}

func (c *CommonStruct) LogList(ctx context.Context) ([]string, error) {
	return c.Internal.LogList(ctx)
}

func (c *CommonStruct) LogSetLevel(ctx context.Context, group, level string) error {
	return c.Internal.LogSetLevel(ctx, group, level)
}

func (c *CommonStruct) GC(ctx context.Context) error {
	return c.Internal.GC(ctx)
}

func (c *CommonStruct) Shutdown(ctx context.Context) error {
	return c.Internal.Shutdown(ctx)
}

func (c *CommonStruct) Session(ctx context.Context) (uuid.UUID, error) {
	return c.Internal.Session(ctx)
}

func (c *CommonStruct) Closing(ctx context.Context) (<-chan struct{}, error) {
	return c.Internal.Closing(ctx)
}
