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
