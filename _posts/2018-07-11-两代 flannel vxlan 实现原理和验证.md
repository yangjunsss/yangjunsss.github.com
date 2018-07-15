---
description: "
SDN 改变了传统的网络世界，在 Underlay 之上构建 Overlay 灵活性带来了巨大的红利，如[fastly成本案例](https://www.fastly.com/blog/building-and-scaling-fastly-network-part-1-fighting-fib)。在容器生态中，[flannel](https://github.com/coreos/flannel/)为容器集群构建 Overlay 网络。本文介绍 Flannel 介绍第一代和第二代 Flannel VXLAN （2017年）的实现方式。
"
---

### 摘要

  SDN 改变了传统的网络世界，在 Underlay 之上构建 Overlay 灵活性带来了巨大的红利，如[fastly成本案例](https://www.fastly.com/blog/building-and-scaling-fastly-network-part-1-fighting-fib)。在容器生态中，[flannel](https://github.com/coreos/flannel/)为容器集群构建 Overlay 网络。本文介绍 Flannel 介绍第一代和第二代 Flannel VXLAN （2017年）的实现方式。

### flannel vxlan 核心设计和历史

关于 VXLAN 的认识可见 [这里](yangjunsss.github.com/_posts/2016-07-12-初识 vxlan.md)，简单来讲就是在 Underlay 之上使用 UDP 封装一层二层数据包，从而实现跨 Underlay 三层的网络实现一种 Overlay 的逻辑二层网络，逻辑网络与物理网络解耦，从而实现灵活的组网需求。在一个二层数据包发送和接收的时候核心主要是 RIB 路由表、FDB 转发表、 ARP 路由表，即 VXLAN 要解决 VM MAC 地址寻址、跨三层 VM IP 地址寻址的问题，并实现全网路由分发和同步，就各家有各家的细节方案，这里讨论容器生态中 Flannel 的实现方案。

在最新的 Flannel VXLAN 代码 [vxlan.go](https://github.com/coreos/flannel/blob/master/backend/vxlan/vxlan.go) 官方有一段注释说明如下：

```text

// Some design notes and history:
// VXLAN encapsulates L2 packets (though flannel is L3 only so don't expect to be able to send L2 packets across hosts)
// The first versions of vxlan for flannel registered the flannel daemon as a handler for both "L2" and "L3" misses
// - When a container sends a packet to a new IP address on the flannel network (but on a different host) this generates
//   an L2 miss (i.e. an ARP lookup)
// - The flannel daemon knows which flannel host the packet is destined for so it can supply the VTEP MAC to use.
//   This is stored in the ARP table (with a timeout) to avoid constantly looking it up.
// - The packet can then be encapsulated but the host needs to know where to send it. This creates another callout from
//   the kernal vxlan code to the flannel daemon to get the public IP that should be used for that VTEP (this gets called
//   an L3 miss). The L2/L3 miss hooks are registered when the vxlan device is created. At the same time a device route
//   is created to the whole flannel network so that non-local traffic is sent over the vxlan device.
//
// In this scheme the scaling of table entries (per host) is:
//  - 1 route (for the configured network out the vxlan device)
//  - One arp entry for each remote container that this host has recently contacted
//  - One FDB entry for each remote host
//
// The second version of flannel vxlan removed the need for the L3MISS callout. When a new remote host is found (either
// during startup or when it's created), flannel simply adds the required entries so that no further lookup/callout is required.
//
//
// The latest version of the vxlan backend  removes the need for the L2MISS too, which means that the flannel deamon is not
// listening for any netlink messages anymore. This improves reliability (no problems with timeouts if
// flannel crashes or restarts) and simplifies upgrades.
//
// How it works:
// Create the vxlan device but don't register for any L2MISS or L3MISS messages
// Then, as each remote host is discovered (either on startup or when they are added), do the following
// 1) create routing table entry for the remote subnet. It goes via the vxlan device but also specifies a next hop (of the remote flannel host).
// 2) Create a static ARP entry for the remote flannel host IP address (and the VTEP MAC)
// 3) Create an FDB entry with the VTEP MAC and the public IP of the remote flannel daemon.
//
// In this scheme the scaling of table entries is linear to the number of remote hosts - 1 route, 1 arp entry and 1 FDB entry per host
```

大致意思是：
  1. 第一代 VXLAN 我们使用 L2&L3 Miss 事件监听来实现 Container ARP 和 FDB 路由表的更新，但可靠性和效率不高
  2. 第二带 VXLAN 实现采用 RIB+ARP+FDB 的新方式实现，路由记录与物理机器成线性关系，提高可靠性和性能

### L2&L3 Miss VXLAN 实现方案

#### 理论基础

组网：

![img](/images/l2l3miss_flannel_impl.png)

流程：
  1. 当 VM0 第一次发送一个数据包的时候，发现目的地址 `10.20.1.3` 在同一个子网，转为二层转发，查询本地 VM ARP 表，无记录发送 ARP 请求；
  2. vxlan 开启了的 proxy、l3miss 功能，数据包通过 vtep0， ARP 请求不对外广播，转为本地代答，查询 Host ARP 表，无记录，触发 L2Miss 事件，ARP 表是用于三层 IP 进行二层 MAC 转发的映射表，存储着 IP-MAC-NIC 记录，在二层转发过程中往往需要根据 IP 地址查询对应的 MAC 地址从而通过数据链路转发到目的接口中；
  3. L2Miss 事件被 Flannel Daemon 捕捉到，Daemon 根据自身的 Etcd 存储的路由数据库返回对应 `10.20.1.3` 的 MAC 地址 `e6:4b:f9:ce:d7:7b` 并存储 Host ARP 表；
  4. vtep0 命中 ARP 记录回复 ARP Reply；
  5. VM0 收到 ARP Reply 后存 VM ARP 表，开始发送数据，携带目的 `e6:4b:f9:ce:d7:7b` 地址；
  6. vtep0 查询 bridge 的 FDB（Forwarding Database entry） 转发表，询问 where `e6:4b:f9:ce:d7:7b` send to? 这个时候发生 L3miss 事件，FDB 表为 2 层交换机的转发表，FDB 存储这 MAC - PORT 的映射关系，用于 MAC数据包从哪个接口出，在 vxlan 中还可以通过`dst` 指定目的 tunnel vtep 的地址；
  7. Flannel Daemon 捕捉 L3miss 事件，并向 FDB 表中加入目的 `e6:4b:f9:ce:d7:7b` 的数据包发送给对端 Host `192.168.100.3` 这台机器；
  8. 此时 vtep0 对数据包进行 UDP 封装，并填充 VNI 号为 1，并与对端 `192.168.100.3` 建立隧道，对端收到 vxlan 包进行拆分，根据 VNI 分发到对应的 vtep 上，拆分后重新转回 Bridge，bridge 根据 dst mac 地址转发到对应的 veth 接口上，此时就完成了整个数据包的转发。这里不仅仅支持同 VNI 二层转发，也支持跨 VNI 三层转发。

Flannel 有能力知道容器的 IP 地址和 MAC 地址，同时在每一个 host 上部署的对应的 Daemon 实例，所以只要利用某种监听机制就能动态添加路由并实现转发面的连通，而监听 ARP miss 的能力依赖 Kernel 提供的 VXLAN DOVE 机制。

add DOVE extensions for VXLAN
This patch provides extensions to VXLAN for supporting Distributed Overlay Virtual Ethernet (DOVE) networks. The patch includes:

	+ a dove flag per VXLAN device to enable DOVE extensions
	+ ARP reduction, whereby a bridge-connected VXLAN tunnel endpoint answers ARP requests from the local bridge on behalf of remote DOVE clients
	+ route short-circuiting (aka L3 switching). Known destination IP addresses use the corresponding destination MAC address for switching rather than going to a (possibly remote) router first.
	+ netlink notification messages for forwarding table and L3 switching misses

```c
// L2Miss - find dest from MAC in FDB
+	f = vxlan_find_mac(vxlan, eth->h_dest);
+	if (f == NULL) {
+		did_rsc = false;
+		dst = vxlan->gaddr;
+		if (!dst && (vxlan->flags & VXLAN_F_L2MISS) &&
+		    !is_multicast_ether_addr(eth->h_dest))
+			vxlan_fdb_miss(vxlan, eth->h_dest);
+	}
```

```c
// L3Miss - find MAC from IP in ARP Table
+	n = neigh_lookup(&arp_tbl, &tip, dev);
+
+	if (n) {...}
+ else if (vxlan->flags & VXLAN_F_L3MISS)
+		vxlan_ip_miss(dev, tip);
```

可以看到内核在查询 `vxlan_find_mac` FDB 转发时未命中则发送 l2miss netlink 通知，在查询 `neigh_lookup` ARP 表时未命中则发送 l3miss netlink 通知，以便有机会给用户态的 client 补充路由，这就是第一代 flannel vxlan 的实现基础。

#### 模拟验证

环境：
  1. 2 台 Centos7.x 机器，2张网卡
  2. 2 个 Bridge，2 张 vtep 网卡
  3. 2 个 Namespace，2 对 veth 接口

步骤：
  1. 创建 Namespace 网络隔离空间模拟容器或虚拟
  2. 创建 veth 接口
  3. 创建 Bridge，并与 veth 相连
  4. 创建 vtep
  5. vm0 ping vm1，在 veth 上观察 l2miss 和 l3miss 事件
  6. 添加 ARP 和 FDB 路由，验证 vm0 ping vm1 的连通性

配置后的网口如下：

```sh
[root@i-7dlclo08 ~]# ip a
1: lo: <LOOPBACK,UP,LOWER_UP> mtu 65536 qdisc noqueue state UNKNOWN qlen 1
    link/loopback 00:00:00:00:00:00 brd 00:00:00:00:00:00
    inet 127.0.0.1/8 scope host lo
       valid_lft forever preferred_lft forever
    inet6 ::1/128 scope host
       valid_lft forever preferred_lft forever
2: eth0: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc pfifo_fast state UP qlen 1000
    link/ether 52:54:ca:9d:db:ff brd ff:ff:ff:ff:ff:ff
    inet 192.168.100.2/24 brd 192.168.100.255 scope global dynamic eth0
       valid_lft 34300sec preferred_lft 34300sec
    inet6 fe80::76ef:824d:95ef:18a3/64 scope link
       valid_lft forever preferred_lft forever
4: veth0@if3: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc noqueue master br0 state UP qlen 1000
    link/ether 4a:1a:2b:4a:5f:48 brd ff:ff:ff:ff:ff:ff link-netnsid 0
    inet6 fe80::481a:2bff:fe4a:5f48/64 scope link
       valid_lft forever preferred_lft forever
5: br0: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1450 qdisc noqueue state UP qlen 1000
    link/ether 4a:1a:2b:4a:5f:48 brd ff:ff:ff:ff:ff:ff
    inet6 fe80::481a:2bff:fe4a:5f48/64 scope link
       valid_lft forever preferred_lft forever
7: vtep0: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1450 qdisc noqueue master br0 state UNKNOWN qlen 1000
    link/ether 4e:94:01:25:a2:fa brd ff:ff:ff:ff:ff:ff
```

测试连通性如下：

同 VNI,`10.20.1.4 - 10.20.1.3`
![img](/images/ping_same_vni.png)

不同 VNI, `10.20.1.4 - 10.20.2.2`
![img](/images/ping_other_vni.png)


#### l2miss l3miss 方案的缺陷

1. 每一台 Host 需要配置所有互通 Container 的 ARP、FDB 记录，导致路由记录较多，不适合大型组网
2. 通过 netlink 通知的效率不高
3. Flannel Daemon 异常后，无法更新 ARP 和 FDB 表影响 Container 之间互通


### 路由网关方案 VXLAN 实现方案

#### 理论基础

组网：
![img](/images/dvr_flannel_impl.png)

在最新的 flannel vxlan 实现上，flannel 把 L2MISS & L3MISS 已经移除了，flannel deamon 不再监听 netlink 通知，而是给 vtep 分配一个三层地址，本地主机配置一条远端的子网路由，nexthop 指向 vtep 分配的三层 IP 地址，并配置好 FDB 转发表和 ARP 表，当数据包到达远端 HOST 后，再通过 Bridge 的网关送达目的 veth ，flannel 本质使用了网关路由的方式实现了一种分层路由，这样的好处就是 HOST 不需要配置所有的 Container FDB 和 ARP 地址路由，使路由信息与物理机器线性相关，基本做到每一台主机 1 route，1 arp entry and 1 FDB entry。

#### 模拟验证

环境：
  1. 2 台 Centos7.x 机器，2张网卡
  2. 2 个 Bridge，2 张 vtep 网卡
  3. 3 个 Namespace，3 对 veth 接口

步骤：
  1. 创建 Namespace 网络隔离空间模拟容器或虚拟
  2. 创建 veth 接口
  3. 创建 Bridge，并与 veth 相连
  4. 创建 vtep，配置 IP，添加网关路由
  5. vm0 ping vm1，vm0 ping vm2

```sh
# HOST0
[root@i-7dlclo08 ~]# ip -d a
1: lo: <LOOPBACK,UP,LOWER_UP> mtu 65536 qdisc noqueue state UNKNOWN qlen 1
    link/loopback 00:00:00:00:00:00 brd 00:00:00:00:00:00 promiscuity 0
    inet 127.0.0.1/8 scope host lo
       valid_lft forever preferred_lft forever
    inet6 ::1/128 scope host
       valid_lft forever preferred_lft forever
2: eth0: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc pfifo_fast state UP qlen 1000
    link/ether 52:54:ca:9d:db:ff brd ff:ff:ff:ff:ff:ff promiscuity 0
    inet 192.168.100.2/24 brd 192.168.100.255 scope global dynamic eth0
       valid_lft 26787sec preferred_lft 26787sec
    inet6 fe80::76ef:824d:95ef:18a3/64 scope link
       valid_lft forever preferred_lft forever
40: br0: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc noqueue state UP qlen 1000
    link/ether 5a:5f:4f:3c:4d:a6 brd ff:ff:ff:ff:ff:ff promiscuity 0
    bridge forward_delay 1500 hello_time 200 max_age 2000
    inet 10.20.1.1/24 scope global br0
       valid_lft forever preferred_lft forever
    inet6 fe80::585f:4fff:fe3c:4da6/64 scope link
       valid_lft forever preferred_lft forever
42: veth0@if41: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc noqueue master br0 state UP qlen 1000
    link/ether ee:d9:bf:29:7a:96 brd ff:ff:ff:ff:ff:ff link-netnsid 1 promiscuity 1
    veth
    bridge_slave
    inet6 fe80::ecd9:bfff:fe29:7a96/64 scope link
       valid_lft forever preferred_lft forever
43: vtep0: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1450 qdisc noqueue state UNKNOWN qlen 1000
    link/ether 52:2d:1f:cb:13:55 brd ff:ff:ff:ff:ff:ff promiscuity 0
    vxlan id 1 dev eth0 srcport 0 0 dstport 4789 nolearning proxy l2miss l3miss ageing 300
    inet 10.20.1.0/32 scope global vtep0
       valid_lft forever preferred_lft forever

# HOST1

[root@i-hh5ai710 ~]# ip -d a
1: lo: <LOOPBACK,UP,LOWER_UP> mtu 65536 qdisc noqueue state UNKNOWN qlen 1
    link/loopback 00:00:00:00:00:00 brd 00:00:00:00:00:00 promiscuity 0
    inet 127.0.0.1/8 scope host lo
       valid_lft forever preferred_lft forever
    inet6 ::1/128 scope host
       valid_lft forever preferred_lft forever
2: eth0: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc pfifo_fast state UP qlen 1000
    link/ether 52:54:d5:9b:94:4c brd ff:ff:ff:ff:ff:ff promiscuity 0
    inet 192.168.100.3/24 brd 192.168.100.255 scope global dynamic eth0
       valid_lft 27216sec preferred_lft 27216sec
    inet6 fe80::baef:a34c:3194:d36e/64 scope link
       valid_lft forever preferred_lft forever
77: br0: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc noqueue state UP qlen 1000
    link/ether d6:ca:34:af:d7:fd brd ff:ff:ff:ff:ff:ff promiscuity 0
    bridge forward_delay 1500 hello_time 200 max_age 2000
    inet 10.20.2.1/24 scope global br0
       valid_lft forever preferred_lft forever
    inet6 fe80::d4ca:34ff:feaf:d7fd/64 scope link
       valid_lft forever preferred_lft forever
79: veth0@if78: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc noqueue master br0 state UP qlen 1000
    link/ether 0e:45:51:9b:1b:23 brd ff:ff:ff:ff:ff:ff link-netnsid 0 promiscuity 1
    veth
    bridge_slave
    inet6 fe80::c45:51ff:fe9b:1b23/64 scope link
       valid_lft forever preferred_lft forever
80: vtep0: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1450 qdisc noqueue state UNKNOWN qlen 1000
    link/ether 52:2f:17:7c:bc:0f brd ff:ff:ff:ff:ff:ff promiscuity 0
    vxlan id 2 dev eth0 srcport 0 0 dstport 4789 ageing 300
    inet 10.20.2.0/32 scope global vtep0
       valid_lft forever preferred_lft forever
82: veth1@if81: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc noqueue master br0 state UP qlen 1000
    link/ether e2:df:ec:93:e7:8b brd ff:ff:ff:ff:ff:ff link-netnsid 1 promiscuity 1
    veth
    bridge_slave
    inet6 fe80::e0df:ecff:fe93:e78b/64 scope link
       valid_lft forever preferred_lft forever

```

测试从 10.20.1.4 与 10.20.2.4、10.20.2.5 的连通性

![imp](/images/vxlan_ping_2th.png)

#### 清理环境

```sh
ip link delete [dev]
ip link delete [dev]
ip link delete [dev]
ip netns delete [name]
```

### 总结
以上就是对 Flannel 2代 vxlan 实现基本原理的解析和验证，可以看到 SDN 的 Overlay 配置很灵活也很巧妙，基本 Overlay 的数据包通过 vxlan 这种技术能穿透 Underlay 网络，同时主机中迭代着两层网络配置也带来了一定的复杂性，在整个过程中数据包也是实际从 Container 中封装穿透的，安全性对于 Overlay 网络是一个重要的问题。

##### 附录

[Linux 上实现 vxlan 网络](http://cizixs.com/2017/09/28/linux-vxlan)
[Kernel Map](http://www.makelinux.net/kernel_map/)
[Network Stack](http://www.cs.dartmouth.edu/~sergey/io/netreads/path-of-packet/Network_stack.pdf)
[Packet Flow](https://www.ccnahub.com/ip-fundamentals/understanding-packet-flow-across-the-network-part1/)
[TCP/IP Network Stack](https://www.cubrid.org/blog/understanding-tcp-ip-network-stack)
