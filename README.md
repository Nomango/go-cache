# Cache缓存器

适用于内存缓存的简单工具类

### Usage

缓存对象并取出
```golang
c := cache.New()

// 保存两个对象
c.Set("num", 123)
c.Set("str", "test")

// 取出对象
if value, ok := c.Get("num"); ok {
    num := value.(int)
    println(num)
}
if value, ok := c.Get("str"); ok {
    str := value.(string)
    println(str)
}

// 遍历对象
c.Range(func(key string, value interface{})bool{
    if key == "num" {
        num := value.(int)
        println(num)
    }
    return true  // 返回true表示继续遍历
})

// 移除一个缓存对象
c.Delete("num")
// 获取缓存数量
count := c.Len()
// 清空缓存
c.Flush()
```

全局缓存
```golang
// 初始化全局缓存
options := &cache.Options{}
cache.Init(options)

// 保存一个对象
cache.Set("num", 123)

// 取出对象
if value, ok := cache.Get("num"); ok {
    num := value.(int)
    println(num)
}

// 移除一个缓存对象
cache.Delete("num")

// 获取全局缓存
global := cache.Global()
```

保存一个对象并设置过期时间
```golang
c := cache.New()

// 保存一个3秒过期的对象
expiration := time.Second * 3
c.SetWithExpiration("num", 123, expiration)
```

设置默认过期时间，和自动清理时长
```golang
options := &cache.Options{
    DefaultExpiration: time.Second * 3,  // 默认3秒过期
    CleanInterval:     time.Minute,      // 每分钟清理一次
}

c := cache.NewWithOptions(options)
```

同时支持LRU缓存机制
```golang
options := &cache.Options{
    Capacity: 2,  // Capacity不为0时自动开启LRU，LRU缓存的容量为2
}

c := cache.NewWithOptions(options)
```

### 已知问题

使用LRU Cache时，如果设置了自动清理（`options.CleanInterval`不为0），可能有潜在的性能问题

Cache的自动清理会调用 `ClearExpired` 方法，该方法执行时会打开写锁，直到清理完成

跑性能测试结果如下

|待删除的对象数量|一次CleanUp耗时|
|-------------|--------------|
|1k           |697.929µs     |
|1w           |4.749823ms    |
|10w          |53.20541ms    |
|100w         |806.008161ms  |

待清除的对象数量超过1w时，一次CleanUp阻塞可能会对性能有较大影响，所以在代码中做了限制，一次最多清理1w个
