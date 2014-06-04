---
layout: post
title: ios8 is coming
categories: [ios]
tags: [WWDC2014, ios]
fullview: false
keywords: WWDC2014,ios8,touchID api,Swift language
description: WWDC2014视频在这里，当看到开放`Touch ID API`时激动了，这意味着指纹识别技术正式进入大规模应用阶段，`Passcode`要退休了；）

---

[WWDC2014视频在这里](http://v.youku.com/v_show/id_XNzIwODkzMTM2.html)，当看到开放`Touch ID API`时激动了，这意味着指纹识别技术正式进入大规模应用阶段，`Passcode`要退休了；）
当看到Apple放开了沙盒的限制，很多人该忙起来了，苹果称之为`extension`，这意味这App可以以widget的形式嵌入到其他应用中，会上演示了在iPhoto中应用第三方的美化工具进行编辑照片，safari中应用了翻译工具，值得一提的是在Notification Center中，App可以获得更多的交互，Taobao，JD等O2O应用该忙起来了。实际上也带来了另一个潜在的商业契机；）

当看到开放iCloud Kit，就笑了，为创业公司加油！这次ios更新开放了4000+个API，涉及到Extenstion,TouchID,iCloudKit,PhotoKit,Camera APIs,HealthKit,HomeKit,Metal.

在游戏方面，引入了Metal，不是很懂- -

纵观整个keynote，Apple在交互性上做了很多的推进，不仅仅是应用之间，同时也在设备之间。

最后引入New Programming Language：`Swift`，Dev举起屌，叫了！演示了Sort和加密的performance，动态类型，小哥当场制作的个小游戏，亮点是同样的code可以直接从2D转为3D。Swift被设计成`Modern、Interactive、Safe、Fast and Powerful`，oc本身是很动态的语言，因为基于message forwarding的机制，底层msg_send接管了所有函数调用的this，函数名，参数和返回值，可以想象Swift会是很动态的语言。比如函数调用不会异常，数组越界不会异常，包装函数返回值，动态加载函数和变量等等都不难实现。

Swift兼容oc的code，并被编译成优化的Native code用来保证Performance，这点可以保持怀疑；），仍然是ARC来做内存管理。

最有趣的应该是Interactive Playgrounds了，看上去这个东西可以让你边写代码边看到运行结果，甚至出performance的数据，这应该是最新鲜和有趣的feature了

Swift官方介绍[详见](https://developer.apple.com/swift/)

第一个Hello World的program是这样：
{% highlight c++ %}
println("Hello,World")
{% endhighlight %}
它被编译在全局区域，不需要main函数，甚至连结束符；都不需要。翻了下Swift的官方文档，600+页，除了语言的基本内容：变量定义，类和结构定义，属性和方法定义，还有闭包、Function、subscript、ARC、Optional Chaining、Extensions等特性。

![p1](/assets/media/屏幕快照 2014-06-04 11.23.08 PM.png)

编程语言变得越来越高级，越来越傻瓜化，我一直在意淫以后的编程技术应该只要程序员用嘴巴说或者手势就能制作出一个程式了- -



