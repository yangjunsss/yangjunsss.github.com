---
layout: post
title: 写了一个Xcode插件XXAlignOnSave
categories: [工具]
tags: [工具]
fullview: false
keywords: XXAlignOnSave

---

博客好像2年没更新了，也忙会了2年，似乎也没停过，折腾来折腾去，要说得太多，不说了，写点code吧，这两天用了下XAlign的插件，这个插件用作对选中的code进行简单的align，但人都懒，如果能在保存的时候就自动align就好了，于是就自己动手改造了这个，算是写得第一个插件，这2年一直也忙于项目，只是偶尔会在github上面提几个bug，希望以后多写点open source的东西在自己github上面。

Code和Readme看这里[XXAlignOnSave](https://github.com/yangjunsss/XXAlignOnSave)，对原来XAlign的code没有侵入性，不用改XAlign的code，只通过Catogery和Swizzle来做的AOP，这也是做plugin的核心思想，因为Apple并没有直接提供plugin文档和接口，必须拦截IDE的方法调用。

主要做的几个事：

* 植入NSMenuItem到XAlign菜单
* Swizzle XCode的ide_saveDocument方法，用于执行align操作
* 在align之前保存当前光标的位置，replace后要恢复
* 监听IDEEditorDocumentDidChangeNotification事件，判断当前文件是否有更改，没有则无需align
* 获取当前文件的type，目前只格式化:

```
@[@"public.c-header",@"public.c-plus-plus-header",@"public.c-source",@"public.objective-c-source",@"public.c-plus-plus-source",@"public.objective-c-plus-plus-source"
```


References:

 * [Xcode Headers](https://github.com/luisobo/Xcode-RuntimeHeaders)