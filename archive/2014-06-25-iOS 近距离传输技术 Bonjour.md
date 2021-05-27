---
description: "
一直想看看近距离传输，特别是“快牙”火了之后，本来想看看古老的蓝牙传输，但这种技术显然落后了，耗电，慢和不稳定。[bonjour](https://developer.apple.com/bonjour/index.html)是apple提供的一种基于TCP/IP，无配置的通信服务，可用于近距离传输，遵循Apache 2.0 license，支持wi-fi和蓝牙。名字也有点意思，“你好”，很神似。
"
---

一直想看看近距离传输，特别是“快牙”火了之后，本来想看看古老的蓝牙传输，但这种技术显然落后了，耗电，慢和不稳定。[bonjour](https://developer.apple.com/bonjour/index.html)是apple提供的一种基于TCP/IP，无配置的通信服务，可用于近距离传输，遵循Apache 2.0 license，支持wi-fi和蓝牙。名字也有点意思，“你好”，很神似。Apple从7.0后提供了一个对bonjour的high level的Framework，`MultipeerConnectivity.framework`.官方有个[sample code](https://developer.apple.com/library/ios/samplecode/MultipeerGroupChat/Introduction/Intro.html#//apple_ref/doc/uid/DTS40013691)（我在2台ios7的pad上已实验正常）就是使用MultipeerConnectivity.framework来做的，但是对于需要支持ios7以下的App来说就歇菜了，不然自己用NSNetService或者CFNetService，或者考虑给这个framework做兼容。

##### 流程步骤

1. Publishing a Network Service

  发布一个网络服务，并schedule default model runloop来lisntening远端请求，采用服务的命名、类型、域名来替换传统的IP地址和端口。命名具有可读性。如：Zealous Lizard's Tune Studio._music._tcp.local，整个过程如图：

  ![tu1](/images/屏幕快照 2014-06-26 12.00.35 AM.png)


* Browsing for and connecting to a Network Service

  询问服务和询问域名，bonjour的询问方式更加明确，比如用“What print services are available?”替换掉“What services are you running?”，整个过程如图：

![tu2](/images/屏幕快照 2014-06-26 12.02.14 AM.png)

* Resolution

  service name 和 地址端口的相互解析，用service name重新广播得到最新的IP地址和端口。

![tu3](/images/屏幕快照 2014-06-26 12.05.06 AM.png)

**API layers：**

Top

^

|  NSNetService NSNetServiceBrowser

|  CFNetServices

|  DNS　Service Discovery API

Low


publish：

```objective-c
NSNetService *service = [[NSNetService alloc] initWithDomain:@""  // default using the compute name,@"local" prevent My Mac or wide-area
                                                      type:@"_music._tcp" // publish TCP/IP music service
                                                      name:@""];
        [service scheduleInRunLoop:[NSRunLoop currentRunLoop] forMode:NSRunLoopCommonModes];
        [service setDelegate:self];
        [service publish]; // if u want to stop,use [service stop]

// you need to get the state of the service , impl the delegate protocal
// netServiceWillPublish:
// netServiceDidPublish:
// netService:didNotPublish:
// netServiceDidStop:
```

browsing：

```objective-c
NSNetServiceBrowser *browser = [NSNetServiceBrowser new];
[browser setDelegate:self];
[browser searchForServicesOfType:@"_music._tcp"
                 inDomain:@""]; // can pass @""(limited wide-area),@"local"(local LAN), or custom
// as the same, impl the delegate
// netServiceBrowserWillSearch:
// netServiceBrowserDidStopSearch:
// netServiceBrowser:didNotSearch:
// the moreComing = NO doesn't indicate that browsing has finished. Dirty here I think
// netServiceBrowser:didFindService:moreComing:
// netServiceBrowser:didRemoveService:moreComing:
```


官方有个Browsing的[sample](https://developer.apple.com/library/ios/samplecode/BonjourWeb/Introduction/Intro.html)


Connecting:
Three ways:
1,Using stream
2,Connecting by hostname
3,Connecting by IP
