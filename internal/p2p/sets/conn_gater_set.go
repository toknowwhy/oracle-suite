""

package sets

import (
"github.com/libp2p/go-libp2p-core/connmgr"
"github.com/libp2p/go-libp2p-core/control"
"github.com/libp2p/go-libp2p-core/network"
"github.com/libp2p/go-libp2p-core/peer"
"github.com/multiformats/go-multiaddr"
)

// ConnGaterSet implements the connmgr.ConnectionGater and allow to aggregate
// multiple instances of this interface.
type ConnGaterSet struct {
connGaters []connmgr.ConnectionGater
}

// NewConnGaterSet creates new instance of the ConnGaterSet.
func NewConnGaterSet() *ConnGaterSet {
return &ConnGaterSet{}
}

// Add adds new connmgr.ConnectionGater to the set.
func (c *ConnGaterSet) Add(connGaters ...connmgr.ConnectionGater) {
c.connGaters = append(c.connGaters, connGaters...)
}

// InterceptAddrDial implements the connmgr.ConnectionGater interface.
func (c *ConnGaterSet) InterceptAddrDial(id peer.ID, addr multiaddr.Multiaddr) bool {
for _, connGater := range c.connGaters {
if !connGater.InterceptAddrDial(id, addr) {
return false
}
}
return true
}

// InterceptPeerDial implements the connmgr.ConnectionGater interface.
func (c *ConnGaterSet) InterceptPeerDial(id peer.ID) bool {
for _, connGater := range c.connGaters {
if !connGater.InterceptPeerDial(id) {
return false
}
}
return true
}

// InterceptAccept implements the connmgr.ConnectionGater interface.
func (c *ConnGaterSet) InterceptAccept(network network.ConnMultiaddrs) bool {
for _, connGater := range c.connGaters {
if !connGater.InterceptAccept(network) {
return false
}
}
return true
}

// InterceptSecured implements the connmgr.ConnectionGater interface.
func (c *ConnGaterSet) InterceptSecured(dir network.Direction, id peer.ID, network network.ConnMultiaddrs) bool {
for _, connGater := range c.connGaters {
if !connGater.InterceptSecured(dir, id, network) {
return false
}
}
return true
}

// InterceptUpgraded implements the connmgr.ConnectionGater interface.
func (c *ConnGaterSet) InterceptUpgraded(conn network.Conn) (bool, control.DisconnectReason) {
for _, connGater := range c.connGaters {
if allow, reason := connGater.InterceptUpgraded(conn); !allow {
return allow, reason
}
}
return true, 0
}

var _ connmgr.ConnectionGater = (*ConnGaterSet)(nil)
