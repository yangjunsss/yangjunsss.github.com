---
layout: post
title: 一个写线程，一个读线程，crash了
categories: [ios]
tags: [Thread, NSMutableArray]
fullview: false
---

实际上我偷了个懒，就crash了，读写线程一般发生crash情况有：
1.访问一个被释放的内存，比如delete或replace操作；
2.访问越界

比如：

{% highlight objective-c %}
- (void) testOneReplaceThreadAndMultiThreadReadBug2
{
    NSMutableArray *array = [NSMutableArray new];
    VideoItem *emptyI = [VideoItem new];
    NSInteger i = 0;
    [array addObject:emptyI];
    while (true) {
        if (i++>10000000) {
            break;
        }
        [array addObject:emptyI];
        array[[array count] - 1] = [VideoItem new];
        dispatch_async(dispatch_get_global_queue(DISPATCH_QUEUE_PRIORITY_DEFAULT, 0), ^{
            VideoItem *e = array[[array count] - 1];
        });
    }
}
{% endhighlight %}

就会发生EXC_BAD_ACCESS