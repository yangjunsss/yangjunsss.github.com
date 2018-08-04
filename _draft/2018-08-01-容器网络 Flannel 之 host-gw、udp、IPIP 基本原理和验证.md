---
description: "
之前一篇博文分析了 Flannel 的 vxlan 的基本原理和验证，对于 Flannel 来说后端的网络实现还有 host-gw、UDP、IPIP 等，索性在这里都分析和验证了。
"
---
### 摘要

之前一篇博文分析了 Flannel 的 vxlan 的基本原理和验证，对于 Flannel 来说后端的网络实现还有 host-gw、UDP、IPIP 等，索性在这里都分析和验证了。

### Host-gw 模式

Host-gw 的基本原理是直接在 host 主机上配置 Overlay 的 subnet 对端 host 的路由信息，整个后端核心代码量在 50 行左右，如下：

```golang
package hostgw

func (be *HostgwBackend) RegisterNetwork(ctx context.Context, wg sync.WaitGroup, config *subnet.Config) (backend.Network, error) {
	n := &backend.RouteNetwork{
		SimpleNetwork: backend.SimpleNetwork{
			ExtIface: be.extIface,
		},
		SM:          be.sm,
		BackendType: "host-gw",
		Mtu:         be.extIface.Iface.MTU,
		LinkIndex:   be.extIface.Iface.Index,
	}
	n.GetRoute = func(lease *subnet.Lease) *netlink.Route {
		return &netlink.Route{
			Dst:       lease.Subnet.ToIPNet(), // 目的地址路由，如 10.20.1.0/24
			Gw:        lease.Attrs.PublicIP.ToIP(), // Underlay Dst Host IP 192.168.100.3
			LinkIndex: n.LinkIndex,
		}
	}

	attrs := subnet.LeaseAttrs{
		PublicIP:    ip.FromIP(be.extIface.ExtAddr),
		BackendType: "host-gw",
	}

	l, err := be.sm.AcquireLease(ctx, &attrs)
	switch err {
	case nil:
		n.SubnetLease = l

	case context.Canceled, context.DeadlineExceeded:
		return nil, err

	default:
		return nil, fmt.Errorf("failed to acquire lease: %v", err)
	}

	return n, nil
}
```

模拟组网：

![img](http://yangjunsss.github.io/images/flannel/host_gw.png)

模拟验证：

```sh
# HOST0
[root@i-7dlclo08 ~]# ip r
default via 192.168.100.1 dev eth0  proto static  metric 100
10.20.1.0/24 dev br0  proto kernel  scope link  src 10.20.1.1
10.20.2.0/24 via 192.168.100.3 dev eth0
192.168.100.0/24 dev eth0  proto kernel  scope link  src 192.168.100.2  metric 100

# HOST1
[root@i-hh5ai710 ~]# ip r
default via 192.168.100.1 dev eth0  proto static  metric 100
10.20.1.0/24 via 192.168.100.2 dev eth0
10.20.2.0/24 dev br0  proto kernel  scope link  src 10.20.2.1
192.168.100.0/24 dev eth0  proto kernel  scope link  src 192.168.100.3  metric 100

# Guest0 ping Guest1
[root@i-7dlclo08 ~]# ip netns exec ns0 ping 10.20.2.2
PING 10.20.2.2 (10.20.2.2) 56(84) bytes of data.
64 bytes from 10.20.2.2: icmp_seq=1 ttl=62 time=1.29 ms
64 bytes from 10.20.2.2: icmp_seq=2 ttl=62 time=0.868 ms
```

可以看到实现非常简单，双方配置对端路由就行了，和 vxlan 不同，不需要拆包和封包，IP 层完全暴露在网络中，数据包到达对端时也是完全裸露的，直接根据路由表进行了转发，也意味着 Underlay 的安全策略需要和 Overlay 一致。

### UDP 模式
UDP 模式和 vxlan 类似是一种隧道实现，即为 host 创建一个 tun 的设备，tun 设备是一个虚拟网路设备，通常一端连接着 kernel 网络协议栈，而另一端就取决于网络设备驱动的实现，一般连接着应用进程，网络数据包发送到这个 tun 设备上后从管道的出口到达应用程序，这个时候应用程序可以根据需求对数据包进行拆包和解包并传送给 eth0 或其他网络设备，从而到达隧道的另一端。
而 flannel 的 udp 模式 tun 设备连接的另一端是 flanneld 进程，这里不单独做验证。

模拟组网：

![img](http://yangjunsss.github.io/images/flannel/udp.png)

可以看到数据包传送给 flanneld 进程，然后由 flanneld 进行 UDP 拆包和封包移交给目的设备。

核心代码：

```golang

func configureIface(ifname string, ipn ip.IP4Net, mtu int) error {
	iface, err := netlink.LinkByName(ifname)
	if err != nil {
		return fmt.Errorf("failed to lookup interface %v", ifname)
	}

	// Ensure that the device has a /32 address so that no broadcast routes are created.
	// This IP is just used as a source address for host to workload traffic (so
	// the return path for the traffic has an address on the flannel network to use as the destination)
	ipnLocal := ipn
	ipnLocal.PrefixLen = 32

	err = netlink.AddrAdd(iface, &netlink.Addr{IPNet: ipnLocal.ToIPNet(), Label: ""})
	if err != nil {
		return fmt.Errorf("failed to add IP address %v to %v: %v", ipnLocal.String(), ifname, err)
	}

	err = netlink.LinkSetMTU(iface, mtu)
	if err != nil {
		return fmt.Errorf("failed to set MTU for %v: %v", ifname, err)
	}

	err = netlink.LinkSetUp(iface)
	if err != nil {
		return fmt.Errorf("failed to set interface %v to UP state: %v", ifname, err)
	}

	// explicitly add a route since there might be a route for a subnet already
	// installed by Docker and then it won't get auto added
	err = netlink.RouteAdd(&netlink.Route{
		LinkIndex: iface.Attrs().Index,
		Scope:     netlink.SCOPE_UNIVERSE,
		Dst:       ipn.Network().ToIPNet(),
	})
	if err != nil && err != syscall.EEXIST {
		return fmt.Errorf("failed to add route (%v -> %v): %v", ipn.Network().String(), ifname, err)
	}

	return nil
}

func (n *network) processSubnetEvents(batch []subnet.Event) {
	for _, evt := range batch {
		switch evt.Type {
		case subnet.EventAdded:
			log.Info("Subnet added: ", evt.Lease.Subnet)

			setRoute(n.ctl, evt.Lease.Subnet, evt.Lease.Attrs.PublicIP, n.port)

		case subnet.EventRemoved:
			log.Info("Subnet removed: ", evt.Lease.Subnet)

			removeRoute(n.ctl, evt.Lease.Subnet)

		default:
			log.Error("Internal error: unknown event type: ", int(evt.Type))
		}
	}
}
```

```sh
ip link del veth0
ip link del br0
ip netns del ns0
ip link add br0 type bridge
ip addr add 10.20.1.1/24 dev br0
# ip link set br0 address 52:2d:1f:cb:13:55
ip link set br0 up
ip netns add ns0
ip link add veth0 type veth peer name veth0-0
ip link set dev veth0 up
ip link set dev veth0 master br0
ip link set dev veth0-0 netns ns0
ip netns exec ns0 ip link set lo up
ip netns exec ns0 ip link set veth0-0 name eth0
ip netns exec ns0 ip addr add 10.20.1.2/24 dev eth0
ip netns exec ns0 ip link set eth0 up
ip netns exec ns0 ip route add default via 10.20.1.1 dev eth0
ip r add 10.20.2.0/24 via 192.168.100.3 dev eth0

ip netns exec ns0 ping 10.20.2.2



ip r add 10.20.1.0/24 via 192.168.100.2 dev eth0
