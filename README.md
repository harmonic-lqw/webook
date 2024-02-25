# webook
a book not for read

# assignment

## week2

### edit

+ ![image-20231129221700201](README.assets/image-20231129221700201.png)
+ ![image-20231129221721112](README.assets/image-20231129221721112.png)

### profile

+ ![image-20231129221740351](README.assets/image-20231129221740351.png)

## week3

+ ![image-20231206202302666](README.assets/image-20231206202302666.png)
+ ![image-20231206202637718](README.assets/image-20231206202637718.png)
+ ![image-20231206202651890](README.assets/image-20231206202651890.png)

## week6

+ ![image-20240122230607206](README.assets/image-20240122230607206.png)

  

### 思路

+ 创建一个为handler打印log的中间件

  ![image-20240122230711838](README.assets/image-20240122230711838.png)

+ 在初始化中添加中间件

  ![image-20240122230728116](README.assets/image-20240122230728116.png)

+ 每次 `err != nil`时，在上下文的err列表中添加err， 最后中间件会自动检测err并打印日志

  ![image-20240122230922617](README.assets/image-20240122230922617.png)

## week10

### 具体修改的文件

+ 将 `internal/events/article` 下的 `producer` 和 `consumer` 集成 `vector`，并修改 `ProduceReadEvent` 和 `Consume` 方法的代码
+ [webook/internal/events/article/producer.go at main · harmonic-lqw/webook (github.com)](https://github.com/harmonic-lqw/webook/blob/main/internal/events/article/producer.go)
+ [webook/internal/events/article/consumer.go at main · harmonic-lqw/webook (github.com)](https://github.com/harmonic-lqw/webook/blob/main/internal/events/article/consumer.go)

### 统计指标

+ `summary` ，并以 `topic` 区分业务，因为比较关心生产一条消息和消费一条消息的时间

### 告警思路

+ 在生产耗时和消费耗时的差距设置告警，如果消费耗时比生产耗时高出某一个阈值，考虑可能会出现消息积压问题，此时需要采取批量消费和异步消费等措施

### 效果

+ ![image-20240225220554969](README.assets/image-20240225220554969.png)
+ ![image-20240225220841729](README.assets/image-20240225220841729.png)

