---
layout: post
title: ios8 is coming
categories: [ios]
tags: [WWDC2014, ios]
fullview: false
keywords: WWDC2014,ios8,touchID api,Swift language
description: WWDC2014视频在这里，当看到开放`Touch ID API`时激动了，这意味着指纹识别技术正式进入大规模应用阶段，但如何监管却是个大问题。

---

[WWDC2014视频在这里](http://v.youku.com/v_show/id_XNzIwODkzMTM2.html)，逐步开放`Touch ID API`意味着指纹识别技术正式进入大规模应用阶段，`Passcode`要退休了；）
当看到Apple放开了沙盒的限制，`extension`，这意味这App可以以widget的形式嵌入到其他应用中，会上演示了在iPhoto中应用第三方的美化工具进行编辑照片，Safari中应用了翻译工具，值得一提的是在Notification Center中，App可以获得更多的交互；）

这次iOS更新开放了4000+个API，涉及到Extenstion,TouchID,iCloudKit,PhotoKit,
Camera APIs,HealthKit,HomeKit,Metal.

在游戏方面，引入了Metal

纵观整个keynote，Apple在交互性上做了很多的推进，不仅仅是应用之间，同时也在设备之间。

最后引入New Programming Language：`Swift`，Dev举起屌，叫了！演示了Sort和加密的performance，动态类型，小哥当场制作的个小游戏，亮点是同样的code可以直接从2D转为3D。Swift被设计成***Modern、Interactive、Safe、Fast and Powerful***

Swift兼容oc的code，并被编译成优化的Native code用来保证Performance，这点可以保持怀疑；），仍然是ARC来做内存管理。

最有趣的应该是Interactive Playgrounds了，看上去这个东西可以让你边写代码边看到运行结果，甚至出performance的数据，这应该是最新鲜和有趣的feature了

Swift官方介绍[详见](https://developer.apple.com/swift/)

第一个Hello World的program是这样：

```
println("Hello,World")
```

它被编译在全局区域，不需要main函数。

翻了下Swift的官方文档，600+页，除了语言的基本内容：变量定义，类和结构定义，属性和方法定义，还有闭包、Function、subscript、ARC、Optional Chaining、Extensions等特性。

编程语言变得越来越高级，我们变得越来越傻吗？



