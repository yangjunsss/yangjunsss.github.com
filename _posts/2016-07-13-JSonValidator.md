---
layout: post
title: 写的另一个Xcode插件YJsonValidator
categories: [工具]
tags: [工具]
fullview: false
keywords: YJsonValidator

---

新的App架构可以不用升级App就能通过配置化的文件对在线PaaS资源进行安装升级了，我们这一套配置化文件都是用的json格式，这两天正在写插件，就写了一个校验json的插件，而之前的做法是去[JSon Formatter](https://jsonformatter.curiousconcept.com/)去复制粘贴，而且网站有时又比较慢，用起来不方便。

Code和Readme看这里[YJsonValidator](https://github.com/yangjunsss/YJsonValidator)

没找到合适的插件，就自己写了，写起来很简单，利用NSJSONSerialization的error message，它的错误信息了包含了出错json的位置信息，只需要做个正规匹配把位置找出来，然后设置下当前Editor光标位置就可以了，有了这个对我写配置文件就很方便了。

#### Demo

![pimg](http://yangjunsss.github.io/assets/media/3dec26cdfb1f83eecf44e21a7b70b70e.gif)

#### References:

 * [Xcode Headers](https://github.com/luisobo/Xcode-RuntimeHeaders)


###### Copyright © 2016 yangjunsss. All rights reserved.