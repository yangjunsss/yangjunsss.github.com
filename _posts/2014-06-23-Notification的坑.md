---
layout: post
title: Notification的坑
categories: [ios]
tags: [Notification]
fullview: false
keywords:

---

在事件驱动的消息处理中，Notification用起来很方便。
**坑一：NSNotificationQueue的addObserver方式是`[NSNotificationCenter defaultCenter] addObserver`，而不是`[NSNotificationQueue defaultQueue] addObserver`**
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

**坑二：Notification原则上是不支持多线程的，post在哪个线程，observer就在哪个线程接收，所以别以为网络线程要刷新UI的时候去post，然后执行UI操作，这个坑会导致crash。**

Apple提供了一种比较粗糙的实现方式来支持多线程，实质上是进行了传递，“传递者”实现如下：

{% highlight objective-c %}
#import <Foundation/Foundation.h>

@interface NotificationTransfer : NSObject <NSMachPortDelegate>
@property (nonatomic) NSMutableArray *notifications;
@property (nonatomic) NSThread *thread;
@property (nonatomic) NSLock *lock;
@property (nonatomic) NSMachPort *port;

- (id) init;
- (void) setUpThreadingSupport;
- (void) handleMachMessage:(void *)msg;
- (void) processNotification:(NSNotification *)notification;
@end


#import "NotificationTransfer.h"

@implementation NotificationTransfer

- (id) init
{
    if (self = [super init]) {
        
    }
    return self;
}
- (void) setUpThreadingSupport
{
    if (self.notifications) {
        return;
    }
    self.notifications = [NSMutableArray new];
    self.lock = [NSLock new];
    self.thread = [NSThread currentThread];
    self.port = [NSMachPort new];
    [self.port setDelegate:self];
    [[NSRunLoop currentRunLoop] addPort:self.port forMode:(__bridge NSString *) kCFRunLoopCommonModes];
}

- (void) handleMachMessage:(void *)msg
{
    [self.lock lock];
    
    while ([self.notifications count]) {
        NSNotification *notification = [self.notifications objectAtIndex:0];
        [self.notifications removeObjectAtIndex:0];
        [self.lock unlock];
        [self processNotification:notification];
        [self.lock lock];
    }
    [self.lock unlock];
}

- (void) processNotification:(NSNotification *)notification
{
    NSThread *ct = [NSThread currentThread];
    if (ct != _thread) {
        [self.lock lock];
        [self.notifications addObject:notification];
        [self.lock unlock];
        [self.port sendBeforeDate:[NSDate date] components:nil from:nil reserved:0];
    }else{
        NSLog(@"process notification %@,is main %zd",[NSThread currentThread],[NSThread isMainThread]);
    }
}
@end
{% endhighlight %}

processNotification会接收post子线程的notification，然后发现不是当前注册的thread就通过schedule一个port并缓存Notification，最后会执行handleMachMessage从而调用逻辑函数，所以这样做需要注册一个NSThread对象，需要统一定义好事件的接收selector。

调用的code：

{% highlight objective-c %}
- (void)viewDidLoad
{
    [super viewDidLoad];
    _transfer = [NotificationTransfer new];
    [_transfer setUpThreadingSupport]; // 在主线程定义了_transfer
    [[NSNotificationCenter defaultCenter] addObserver:_transfer selector:@selector(processNotification:) name:@"notifi" object:nil]; // 添加了selector
    dispatch_queue_t queue = dispatch_queue_create("queue1", DISPATCH_QUEUE_CONCURRENT);
    dispatch_async(queue, ^{
        [[NSNotificationCenter defaultCenter] postNotificationName:@"notifi" object:nil]; // 子线程post，最终Notification会转发到主线程执行processNotification
    });
}
{% endhighlight %}



