---
layout: post
title: Notification的坑
categories: [ios]
tags: [Notification]
fullview: false
keywords:

---

在事件驱动的消息处理中，Notification用起来很方便。

坑一：NSNotificationQueue的addObserver方式是`[NSNotificationCenter defaultCenter] addObserver`，而不是`[NSNotificationQueue defaultQueue] addObserver`

Notification提供异步post方式NSNotificationQueue，通过enqueueNotification的接口把Notification入队列，并且提供NSPostASAP, NSPostWhenIdle, and NSPostNow这3种时刻来执行，意思分别为在runloop结束时，在线程idle时和立刻。比如：

{% highlight objective-c %}
...
[[NSNotificationQueue defaultQueue] addObserver:self selector:@selector(doIdleTask) name:@"idleTask" object:nil]; // 错误，但编译正确不报错- -
[[NSNotificationQueue defaultQueue] enqueueNotification:[NSNotification notificationWithName:@"idleTask" object:self] postingStyle:NSPostWhenIdle];
...

- (void)doIdleTask
{
    NSLog(@"do task in idle");
    [[NSNotificationQueue defaultQueue] enqueueNotification:[NSNotification notificationWithName:@"idleTask" object:self] postingStyle:NSPostWhenIdle];
}

{% endhighlight %}

可以重复利用当前线程Idle来碎片化执行任务。

在addObserver的时候很容易写成：`[[NSNotificationQueue defaultQueue] addObserver:self selector:@selector(doIdleTask) name:@"idleTask" object:nil];`，而且这样写并不报错，因为[NSNotificationQueue defaultQueue]返回的是id,在编译期它能是任意的对象，所以能关联上任意子类的方法，从而都没有编译错误，但事实上NSNotificationQueue并没有addObserver方法，所有Notification的addObserver都使用[NSNotificationCenter defaultCenter]。正确的写法是：

{% highlight objective-c %}
[[NSNotificationCenter defaultCenter] addObserver:self selector:@selector(doIdleTask) name:@"idleTask" object:nil];

{% endhighlight %}

