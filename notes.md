2024-3-23:
0. 设计一个分布式缓存系统：
   0.1 内存不够时，采用什么样的淘汰策略
   0.2 并发写冲突
1. groupcache 是go语言版的 memcached， geecache的特性：
    1.1 单机缓存和基于HTTP的分布式缓存
    1.2 LRU(Least Recently Used, 最近最少访问)缓存策略
    1.3 GO锁机制，防止缓存击穿
    1.4 一致性哈希选择节点，实现负载均衡
    1.5 节点之间使用protobuf通信（优化通信）
2. 三种缓存淘汰策略（FIFO/LFU/LRU）：
   2.1 FIFO，淘汰最早添加的记录。| 用一个队列，新增记录就添加到队尾，内存不够时淘汰队首
   2.2 LFU，淘汰最不常用的记录。 | 用一个队列，维护记录被访问次数（次数从高到低排序）  
                                 受历史数据影响较大，有可能某个数据某个时间被大量访问，但后面几乎没人访问，却不能被淘汰
   2.3 LRU，如果数据最近被访问了，那么将来被访问的概率会很高。| 维护一个队列，某个记录被访问就移动到队尾，每次依然从队首开始淘汰
        核心：2个数据结构：一个map和一个双向链表


tips:
1. 引用其他包的方法和变量，要带上包名。
2. import时是 module/packagename
3. 工程目录结构: 最好不要在根目录下放多个go文件，因为子包引用根目录下的东西时还不知道怎么引用......可能Go工程里应该多建立子包，根目录下放些运行的main函数就够了。 
4. 根目录下的package也最好叫main吧

2024-4-14：
1. 通过为cache增加sync.Mutex来支持并发读写
2. 抽象一个只读数据结构 ByteView，用来表示缓存值
3. 为用户提供一个Getter接口（callback函数），用于在缓存中查不到时去数据源查询获取数据并添加到缓存

tips: 延迟初始化（lazy initialization）。一个对象的延迟初始化意味着该对象的创建将会延迟至第一次使用该对象时。主要用于提高性能，并减少程序内存要求


2024-4-21：
1. 节点间使用http通信
2. 单机节点搭建 http Server, 返回缓存值。
HTTPPool是http通信的核心数据结构，并为其实现通信核心的ServeHTTP()方法：校验访问路径的前缀，通过groupName查找到group，调用group的Get()方法获得只读缓存


2024-5-2：
1. 实现一致性哈希（hash函数定义、虚拟节点生成、虚拟节点与真实节点映射）
   一致性哈希算法将 key 映射到 2^32 的空间中，将这个数字首尾相连，形成一个环。
   为了解决数据倾斜问题：
   第一步，计算虚拟节点的 Hash 值，放置在环上。
   第二步，计算 key 的 Hash 值，在环上顺时针寻找到应选取的虚拟节点，例如是 peer2-1，那么就对应真实节点 peer2。
2. 实现http客户端，访问远程节点的url，获取返回值
   抽象接口 PeerPicker（PickPeer方法选择节点）和 PeerGetter(Get方法从对应group中查找缓存值)
3. 为httpPool添加节点选择的功能，在HTTPPool中添加peers（一致性哈希的类型）和 httpGetters(映射key和节点的url)
   实例化一致性哈希
   为每个节点创建一个 http客户端 httpGetter

2024-6-23:
1. 实现singleflight防止缓存击穿（g.load调用时，g.loader.Do()会调用多次）。加了dup属性保证匿名函数fn()只被调用一次。

2024-6-30：
添加到github仓库：
在github上创建一个新的repository，注意默认分支最好写master。
（因为本地 init之后叫main分支，直接提交会自动提交一个新分支main）

1. 本地工程 git init
2. git add .
3. git commit -m "first commit"
4. git remote add origin 远程仓库url地址
5. git pull origin master 
6. git push -u origin master 