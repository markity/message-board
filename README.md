# Message-Board

### 项目简介

这是一个作业，实现一个可评论可点赞的简易留言板功能, 亮点有JWT的实现, 实现事务级表锁, 以及利用事务保持数据一致性的操作，还有如何处理数据库错误(即重试策略)

### 功能支持概述

- 留言板消息的创建修改删除
- 留言板消息的获取支持分页
- 支持创建无限级的留言评论, 查询时返回一个树状结构, 支持留言评论的修改和删除
- 支持对留言板消息或任意级子消息的点赞
- 支持提交匿名消息或匿名评论
- 支持管理员权限, 管理员允许删除任何留言或评论
- 完备的wechat风格错误码设计

### 实现细节概述

- 管理员只允许删除特定消息, 不允许修改, 防止管理员乱改消息, 引发舆论
- JWT进行鉴权, 基于cookie实现JWT, 未登录的用户只能查看消息, 不能发送消息、评论或点赞
- JWT签发使用RSA非对称加密, 每次启动服务器都会新随机一个RSA密钥

### 接口概述(共13个接口)

- /user

  - POST 新建用户(interface1)
  - PUT 修改密码(interface2)或修改用户简介(interface3)
- /user/info/{string username}

  - GET 获得用户详细信息(就是用户的个性签名和创建时间)(interface4)
- /user/auth

  - POST 进行登录(interface5)
  - DELETE 进行注销(interface6)
- /message

  - GET 批量获得一定数目(10条)的消息, 及其各级子消息(interface7)
  - POST 发送顶级消息(interface8)
- /message/{int msgid}

  - GET 获取一条指定id消息的具体信息, 以及它的所有子消息(interface9)
  - POST 发送子消息, 其父消息的id为msgid(interface10)
  - PUT 修改消息(interface11), 或点赞(interface12)
  - DELETE 删除消息(interface13)

### 数据库的细节概述

- 数据库密码进行统一MD5加密, 不存储明文密码
- 不直接删除条目, 而是更新DELETE BIT
- 为了保证创建用户的原子性, 检测INSERT的错误码来判断用户是否存在, 而不使用查询后插入的策略
- 为了保证查询结果的一致性, 使用了事务查询
- 为了保证服务高可用, 使用重试策略, 当出现错误的时候重试一定次数, 如果都不成功才告知用户服务暂时不可用

### JWT的实现细节概述

**算法方面**

```plaintext
# 描述JWT元数据的JSON对象
header = {
   # 算法, RSA 256
   "alg":"RS256",
   # typ表示令牌的类型, JWT令牌
   "typ": "JWT"
}

----------------------------------------------------------------------

# 有效载荷部分
payload = {
# 到期时间时间戳
exp: 1669513868,
# jwt唯一ID, 为了保证每次生成的token都不一样而使用它
jti: 1234,
# 用户名, 指定一些接口操作的user对象
userid: 1,
# 是否为管理员, 管理员有权限删除所有消息
admin: false
}

----------------------------------------------------------------------

# 获得token(加密)
sign = HMACSHA256(base64UrlEncode(header) + "." + base64UrlEncode(payload),secret)
token = base64UrlEncode(header) + "." + base64UrlEncode(payload) + "." + base64UrlEncode(sign)

-----------------------------------------------------------------------

# 解密
# 解码payload可得到expire信息, 检查是否过期, 然后用公钥解密token, 对比header和payload是否相同
# 其实可以不用比对header和payload是否相同, 但为了更可靠, 选择了比对, 反正没啥损失
```

**实现方面**

使用中间件进行JWT鉴权, 每次开启服务器时调用系统随机数接口随机密钥. 鉴权时, 首先用私钥解密, 验证header和payload是否一致, 如果能解开, 说明JWT鉴权成功, 如果失败, 删除cookie, 返回响应错误码

### 数据表设计

**user**

| 字段名 | id                          | username         | password_crypto | created_at                | personal_signature | admin                         | deleted                   |
| ------ | --------------------------- | ---------------- | --------------- | ------------------------- | ------------------ | ----------------------------- | ------------------------- |
| 类型   | INT                         | VARCHAR(32)      | TINYBLOB        | DATETIME                  | VARCHAR(200)       | TINYINT                       | TINYINT                   |
| 约束   | PRIMARY KEY, AUTO_INCREMENT | NOT NULL, UNIQUE | NOT NULL        | NOT NULL                  | NULL               | NOT NULL, DEFAULT 0           | NOT NULL, DEFAULT 0       |
| 说明   | 主键                        | 用户名, 唯一     | 加密后的密码    | 如果该条目被删除, 被置为1 | 用户的创建日期     | 如果该条目为1, 那么他为管理员 | 如果该条目被删除, 被置为1 |

一些说明:

- deleted字段在当前项目没有用到, 此项目不允许删除用户信息, 保留是为了表的一致性
- 创建用户的时候会利用到UNIQUE的特性, 不采用查询后插入的方法, 而是直接插入, 如果错误码为1062(duplicate entry)那么就代表已经存在, 直接告知用户该用户名已被注册
- 如果personal_signature为null, 那么代表此人没有个性签名

**message**

| 字段名 | id                          | content      | sender_user_id | parent_message_id                                                          | thumbs_up            | anonymous               | created_at       | deleted                                                     |
| ------ | --------------------------- | ------------ | -------------- | -------------------------------------------------------------------------- | -------------------- | ----------------------- | ---------------- | ----------------------------------------------------------- |
| 类型   | INT                         | VARCHAR(500) | INT            | INT                                                                        | INT                  | TINYINT                 | DATETIME         | TINYINT                                                     |
| 约束   | PRIMARY KEY, AUTO_INCREMENT | NOT NULL     | NOT NULL       | DEFAULT NULL                                                               | NOT NULL, DEFAULT 0  | NOT NULL, DEFAULT 0     | NOT NULL         | NOT NULL, DEFAULT 0                                         |
| 说明   | 主键                        | 消息的内容   | 发送者的用户id | 表示父消息, 如果没有父消息, 那么<br />它就是顶级消息(也就是留言, 而非评论) | 表示该消息的点赞数目 | 该消息是否匿名, 1则匿名 | 该消息的创建日期 | 如果该条目被删除, 被置为1, 此时<br />不再展示这条以及子条目 |

一些说明:

- deleted字段用于表示该条目是否被删除, 当一条消息被删除时, 不更新其子条目的deleted bit, 但认为子条目也被删除. 为了效率, 不逐个更新子条目的deleted bit了

**thumb_message_user**

这张表是用来维系点赞用户与消息的一对多关系的表, 用于实现一个用户仅能点赞一个评论一次的需求

| 字段名 | id                          | user_id  | message_id |
| ------ | --------------------------- | -------- | ---------- |
| 类型   | INT                         | INT      | INT        |
| 约束   | PRIMARY KEY, AUTO_INCREMENT | NOT NULL | NOT NULL   |
| 说明   | 主键                        | 用户方id | 消息的id   |

一点细节:

点赞的过程包括查询此表以及更新此表的过程, 这是一个查询后插入操作, 需要考虑并发下的数据不一致问题, 因此不给user_id和message_id加索引. 这样查询该表的之后事务会锁住这张表, 使之并发安全

**distributed_lock**

| 字段名 | id                          | tbname           |
| ------ | --------------------------- | ---------------- |
| 类型   | INT                         | VARCHAR(32)      |
| 约束   | PRIMARY KEY, AUTO_INCREMENT | NOT NULL, UNIQUE |
| 说明   | 主键                        | 表明             |

为了实现事务级表锁, 需要锁表的操作都先从这个表里面拿锁

### 接口详述

**统述**

- 所有接口的status code都是200, 具体的错误码取决于返回的error_code字段以及错误信息msg
- 为了简化实现, 这些接口只从入参中提取所需的参数, 不检查多余的参数
- 有的接口会需要JWT鉴权, 如果鉴权失败, 会清空cookie, 返回一定的错误码. 对不需要鉴权的接口, 会在鉴权中间件中清空cookie
- 关于错误码风格: 200xx代表的是通用错误(比如服务暂时不可用, JWT验证失败等), 而201xx或202xx等是给某个具体接口设计的错误码, 见下面通用错误码汇总

| 错误码 | 20000 | 20001                                                                              | 20002                                                                                                                         | 20003                                                                                              |
| ------ | ----- | ---------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------- |
| 说明   | 成功  | 服务暂时不可用, 这种错误出现在查找数据<br />库三次尝试均失败之后, 要求用户稍后再试 | 鉴权失败, Token不合法, 对于需要JWT鉴权<br />的接口, 中间件首先检查用户的JWT是否有效<br />如果有效, 才进入下一级中间件继续处理 | 不合法的参数, 可能发生在用户的<br />输入与预期不符的情况下, 比如用户<br />确定密码与原密码不同等等 |
| msg    | OK    | service not available temporarily                                                 | invalid identity token                                                                                                        | invalid parameters                                                                                 |

**1. 新建用户**

接口:

| 接口地址 | HTTP方法 | 是否JWT鉴权 |
| -------- | -------- | ----------- |
| /user    | POST     | 否          |

入参(form-data):

| 键名     | username                                                      | password                                                    | personal_signature         |
| -------- | ------------------------------------------------------------- | ----------------------------------------------------------- | -------------------------- |
| 格式要求 | 长度限制为[4, 20]个字符, 只允许<br />大小写字符, 数字和下划线 | 要求[6, 25]的长度限制, 只允<br />许大小写字符, 数字和下划线 | [0, 150]长度限制的utf8字符 |
| 是否必填 | 是                                                            | 是                                                          | 否                         |

特别说明:

- personal_signature如果为""或压根没有这个字段, 那么数据库保存的personal_signature则为NULL

出参(json)

| 键名     | error_code | msg      |
| -------- | ---------- | -------- |
| 字段类型 | int number | string   |
| 是否必填 | 是         | 是       |
| 说明     | 错误码     | 错误信息 |

可能的错误码:

| 错误码 | 20000 | 20003                | 20101                  | 20001                              |
| ------ | ----- | -------------------- | ---------------------- | ---------------------------------- |
| 说明   | 成功  | 失败, 错误的入参参数 | 失败, 该用户名已被占用 | 服务暂时不可用                     |
| msg    | OK    | invalid parameters   | occupied user name     | service not available temporarily |

**2. 修改密码**

接口:

| 接口地址 | HTTP方法 | 是否JWT鉴权 |
| -------- | -------- | ----------- |
| /user    | PUT      | 是          |

入参(form-data):

| 键名     | password                                                    | password_verify                  | put_type            | old_password |
| -------- | ----------------------------------------------------------- | -------------------------------- | ------------------- | ------------ |
| 格式要求 | 要求[6, 25]的长度限制, 只允<br />许大小写字符, 数字和下划线 | 同password, 且要求和password相同 | 填写change_password | 原密码       |
| 是否必填 | 是                                                          | 是                               | 是                  | 是           |

出参(json):

| 键名     | error_code | msg      |
| -------- | ---------- | -------- |
| 类型     | int number | string   |
| 是否必填 | 是         | 是       |
| 说明     | 错误码     | 错误信息 |

增加说明:

- 使用此接口修改密钥, 要求用户有合法的JWT
- 使用此接口修改成功密钥后, 讲删除用户方jwt cookie, 之后需用户重新登录

可能的错误代码:

| 错误码 | 20000 | 20001                              | 20002                  | 20003                                                                                                                                    | 20201                 |
| ------ | ----- | ---------------------------------- | ---------------------- | ---------------------------------------------------------------------------------------------------------------------------------------- | --------------------- |
| 说明   | 成功  | 服务暂时不可用                     | 鉴权失败, Token不合法  | 失败, 错误的入参参数, 在此接口中,<br />password不合法, 两次密码不同, <br />old_password格式不合法,<br /> put_type未知都被归结于这个错误 | 错误的原密码          |
| msg    | OK    | service not available temporarily | invalid identity token | invalid parameters                                                                                                                       | wrong former password |

**3. 修改个性签名**

接口:

| 接口地址 | HTTP方法 | 是否JWT鉴权 |
| -------- | -------- | ----------- |
| /user    | PUT      | 是          |

入参(form-data):

| 键名     | personal_signature                                           | put_type               |
| -------- | ------------------------------------------------------------ | ---------------------- |
| 格式要求 | 要求[0, 150]的长度限制, 只允<br />许大小写字符, 数字和下划线 | 填写personal_signature |
| 是否必填 | 否                                                           | 是                     |

特别说明:

- 如果入参personal_signature的值为"", 那么数据库会将personal_signature字段设置为NULL
- 如果put_type为personal_signatrue, 且没有personal_signature字段, 那么数据库会将personal_signature字段设置为NULL

出参(json):

| 键名     | error_code | msg      |
| -------- | ---------- | -------- |
| 类型     | int number | string   |
| 是否必填 | 是         | 是       |
| 说明     | 错误码     | 错误信息 |

增加说明:

- 使用此接口修改密钥, 要求用户有合法的JWT, 通过JWT信息判断用户username

可能的错误代码:

| 错误码 | 20000 | 20001                              | 20002                  | 20003                |
| ------ | ----- | ---------------------------------- | ---------------------- | -------------------- |
| 说明   | 成功  | 服务暂时不可用                     | 鉴权失败, Token不合法  | 失败, 错误的入参参数 |
| msg    | OK    | service not available temporarily | invalid identity token | invalid parameters   |

**4. 获得用户信息:**

接口:

| 接口地址                     | HTTP方法 | 是否JWT鉴权 |
| ---------------------------- | -------- | ----------- |
| /user/info/{string username} | GET      | 否          |

入参(接口地址入参):

| 键名     | username                               |
| -------- | -------------------------------------- |
| 类型     | string                                 |
| 是否必填 | 是                                     |
| 格式要求 | [4,20]长度, 由大小写字符以及下划线组成 |
| 说明     | 需要查询用户的用户名                   |

出参(json):

| 键名     | created_at                                     | username                                 | personal_signature                             | error_code | msg      |
| -------- | ---------------------------------------------- | ---------------------------------------- | ---------------------------------------------- | ---------- | -------- |
| 类型     | string                                         | string                                   | string                                         | int number | string   |
| 是否必填 | 否                                             | 否                                       | 否                                             | 是         | 是       |
| 说明     | 用户创建日期, 如果用<br />户不存在, 则无该字段 | 用户名, 如果用户不<br />存在, 则无该字段 | 用户个性签名, 如果用<br />户不存在, 则无该字段 | 错误码     | 错误信息 |

可能的错误码:

| 错误码 | 20000    | 20001                              | 20003                | 20401            |
| ------ | -------- | ---------------------------------- | -------------------- | ---------------- |
| 说明   | 查询成功 | 服务暂时不可用                     | 失败, 错误的入参参数 | 失败, 没有该用户 |
| msg    | OK       | service not available temporarily | invalid parameters   | no such username |

**5. 登录**

接口:

| 接口地址   | HTTP方法 | 是否JWT鉴权 |
| ---------- | -------- | ----------- |
| /user/auth | POST     | 否          |

入参(form-data):

| 键名     | username                                                | password                                                |
| -------- | ------------------------------------------------------- | ------------------------------------------------------- |
| 类型     | string                                                  | string                                                  |
| 是否必填 | 是                                                      | 是                                                      |
| 格式要求 | 要求长度[4, 20], 只允许<br />大小写字符, 数字以及下划线 | 要求长度[6, 25], 只允许<br />大小写字符, 数字以及下划线 |
| 说明     | 用户名                                                  | 密码                                                    |

出参(json):

| 键名     | error_code | msg      |
| -------- | ---------- | -------- |
| 类型     | int number | string   |
| 是否必填 | 是         | 是       |
| 说明     | 错误码     | 错误信息 |

可能的错误码:

| 错误码 | 20000    | 20001                              | 20003                | 20501                                                |
| ------ | -------- | ---------------------------------- | -------------------- | ---------------------------------------------------- |
| 说明   | 登录成功 | 服务暂时不可用                     | 失败, 错误的入参参数 | 失败, 没有该用户或密码错误                           |
| msg    | OK       | service not available temporarily | invalid parameters   | login failed, check<br />password and username again |

额外说明:

- 此接口会给用户签发jwt用于鉴权, 放在cookie之中

**6. 注销**

接口:

| 接口地址   | HTTP方法 | 是否JWT鉴权 |
| ---------- | -------- | ----------- |
| /user/auth | DELETE   | 是          |

入参: 无

出参(json):

| 键名     | error_code | msg      |
| -------- | ---------- | -------- |
| 类型     | int number | string   |
| 是否必填 | 是         | 是       |
| 说明     | 错误码     | 错误信息 |

可能的错误码:

| 错误码 | 20000    | 20001                              | 20002                  |
| ------ | -------- | ---------------------------------- | ---------------------- |
| 说明   | 注销成功 | 服务暂时不可用                     | 鉴权失败, Token不合法  |
| msg    | OK       | service not available temporarily | invalid identity token |

**7. 批量获得一定数目的消息及子消息**

接口:

| 接口地址 | HTTP方法 | 是否JWT鉴权 |
| -------- | -------- | ----------- |
| /message | GET      | 否          |

入参(form-data):

| 键名     | entry_num                                    | page_num                                                      |
| -------- | -------------------------------------------- | ------------------------------------------------------------- |
| 类型     | int number                                   | int number                                                    |
| 是否必填 | 否                                           | 否                                                            |
| 参数要求 | 取值范围[1, 50]                              | 正整数                                                        |
| 说明     | 指定一页的长度, 如果<br />没有这一条, 默认10 | 1代表第一页, 最新的消息在前面,<br />如果未填写此参数, 默认为1 |

出参(json):

| 键名     | error_code | msg      | messages                                                                                 |
| -------- | ---------- | -------- | ---------------------------------------------------------------------------------------- |
| 类型     | int number | string   | array                                                                                    |
| 是否必填 | 是         | 是       | 否                                                                                       |
| 说明     | 错误码     | 错误信息 | 如果error_code不为20000, 就没有这个字段<br />, array的每一个子项都是树状的, 见下面的例子 |

messages例子:

```plaintext
[
  {
    message_id: 1,
    message_content: "你好, 这是一条留言",
    sender_user_name: "Markity",
    created_at: "2004-01-19 12:31:43",
    thumbs_up: 0,
    anonymous: false,
    son_messages: [
      {
        message_id: 2,
        message_content: "你好, 这又是一条留言",
        created_at: "2004-01-19 12:31:43",
        thumbs_up: 3,
        # 当anonymous字段为true的时候, 该消息没有sender_user_name字段
        anonymous: true,
        son_messages: nil
      }
    ]
  }, ...更多内容
]
```

说明:

- 当一条都没有的时候, 返回一个空的数组, 即为 `[]`

可能的错误码:

| 错误码 | 20000    | 20001                              | 20002                  | 20003                |
| ------ | -------- | ---------------------------------- | ---------------------- | -------------------- |
| 说明   | 注销成功 | 服务暂时不可用                     | 鉴权失败, Token不合法  | 失败, 错误的入参参数 |
| msg    | OK       | service not available temporarily | invalid identity token | invalid parameters   |

**8. 发送顶级消息**

接口:

| 接口地址 | HTTP方法 | 是否JWT鉴权 |
| -------- | -------- | ----------- |
| /message | POST     | 是          |

入参(form-data):

| 键名     | content          | anonymous                        |
| -------- | ---------------- | -------------------------------- |
| 类型     | string           | bool                             |
| 是否必填 | 是               | 否                               |
| 参数要求 | 长度范围[5, 300] | true或false                      |
| 说明     | 留言内容         | 如果没有该字段<br />默认为不匿名 |

出参(json):

| 键名     | error_code | msg      | message_id                              |
| -------- | ---------- | -------- | --------------------------------------- |
| 类型     | int number | string   | int number                              |
| 是否必填 | 是         | 是       | 否                                      |
| 说明     | 错误码     | 错误信息 | 如果error_code不为20000, 就没有这个字段 |

可能的错误码:

| 错误码 | 20000    | 20001                              | 20002                  | 20003                |
| ------ | -------- | ---------------------------------- | ---------------------- | -------------------- |
| 说明   | 留言成功 | 服务暂时不可用                     | 鉴权失败, Token不合法  | 失败, 错误的入参参数 |
| msg    | OK       | service not available temporarily | invalid identity token | invalid parameters   |

**9. 获得指定id消息的信息以及子消息的信息**

接口:

| 接口地址             | HTTP方法 | 是否JWT鉴权 |
| -------------------- | -------- | ----------- |
| /message/{int msgid} | GET      | 否          |

入参(接口地址参数):

| 键名     | msgid      |
| -------- | ---------- |
| 类型     | int number |
| 是否必填 | 是         |

出参(json):

| 键名     | error_code | msg      | message                                                                         |
| -------- | ---------- | -------- | ------------------------------------------------------------------------------- |
| 类型     | int number | string   | dict                                                                            |
| 是否必填 | 是         | 是       | 否                                                                              |
| 说明     | 错误码     | 错误信息 | 如果error_code不为20000, 就没有这<br />个字段, 这是一个树状的结构, 见下面的例子 |

message示例:

```plaintext
{
  message_id: 1,
  message_content: "你好, 这是一条留言",
  sender_user_name: "Markity",
  created_at: "2004-01-19 12:31:43",
  thumbs_up: 0,
  anonymous: false,
  son_messages: [
    {
      message_id: 2,
      message_content: "你好, 这又是一条留言",
      created_at: "2004-01-19 12:31:43",
      thumbs_up: 3,
      # 当anonymous字段为true的时候, 该消息没有sender_user_name字段
      anonymous: true,
      son_messages: nil
    }
  ]
}
```

可能的错误码:

| 错误码 | 20000    | 20001                              | 20002                  | 20901              |
| ------ | -------- | ---------------------------------- | ---------------------- | ------------------ |
| 说明   | 留言成功 | 服务暂时不可用                     | 鉴权失败, Token不合法  | 失败, 没有该条留言 |
| msg    | OK       | service not available temporarily | invalid identity token | no such comment    |

**10. 发送子消息**

接口:

| 接口地址             | HTTP方法 | 是否JWT鉴权 |
| -------------------- | -------- | ----------- |
| /message/{int msgid} | POST     | 是          |

实现细节:

- 如果msgid转化为int失败, 那么

入参1(接口地址传参):

| 键名 | msgid      |
| ---- | ---------- |
| 类型 | int number |
| 说明 | 父消息的id |

入参2(form-data):

| 键名     | content                | anonymous              |
| -------- | ---------------------- | ---------------------- |
| 类型     | string                 | bool                   |
| 是否必填 | 是                     | 否                     |
| 格式要求 | 要求[5, 300]个utf8字符 | true或false            |
| 说明     | 内容                   | 是否匿名, 默认为非匿名 |

出参(json):

| 键名     | message_id                                                    | error_code | msg      |
| -------- | ------------------------------------------------------------- | ---------- | -------- |
| 类型     | int                                                           | string     | string   |
| 是否必填 | 否                                                            | 是         | 是       |
| 说明     | 发送的消息的id, 如果error_code<br />不为20000, 那么没有该字段 | 错误码     | 错误消息 |

可能的错误码:

| 错误码 | 20000    | 20001                              | 20002                  | 21001              |
| ------ | -------- | ---------------------------------- | ---------------------- | ------------------ |
| 说明   | 评论成功 | 服务暂时不可用                     | 鉴权失败, Token不合法  | 错误, 没有这个消息 |
| msg    | OK       | service not available temporarily | invalid identity token | no such message    |

**11. 修改消息**

接口:

| 接口地址             | HTTP方法 | 是否JWT鉴权 |
| -------------------- | -------- | ----------- |
| /message/{int msgid} | PUT      | 是          |

入参1(接口地址传参):

| 键名     | msgid        |
| -------- | ------------ |
| 类型     | int number   |
| 是否必填 | 是           |
| 说明     | 修改消息的id |

入参2(form-data):

| 键名     | content                | put_type   | anonymous                                                             |
| -------- | ---------------------- | ---------- | --------------------------------------------------------------------- |
| 类型     | string                 | int number | bool                                                                  |
| 是否必填 | 是                     | 是         | 否                                                                    |
| 格式要求 | 要求[5, 300]个utf8字符 | 填edit     | true或false                                                           |
| 说明     | 新的内容               | 表示修改   | 如果不存在该字段则不<br />改变原消息的匿名与否,<br />即与原来保持一致 |

出参(json):

| 键名     | error_code | msg      |
| -------- | ---------- | -------- |
| 类型     | string     | string   |
| 是否必填 | 是         | 是       |
| 说明     | 错误码     | 错误消息 |

可能的错误码:

| 错误码 | 20000    | 20001                              | 20002                  | 21101              | 20003                | 21102                    |
| ------ | -------- | ---------------------------------- | ---------------------- | ------------------ | -------------------- | ------------------------ |
| 说明   | 修改成功 | 服务暂时不可用                     | 鉴权失败, Token不合法  | 错误, 没有这个消息 | 失败, 错误的入参参数 | 错误, 没有修改权限       |
| msg    | OK       | service not available temporarily | invalid identity token | no such message    | invalid parameters   | no permission to edit it |

**12. 点赞消息**

接口:

| 接口地址             | HTTP方法 | 是否JWT鉴权 |
| -------------------- | -------- | ----------- |
| /message/{int msgid} | PUT      | 是          |

入参1(接口地址传参):

| 键名     | msgid                |
| -------- | -------------------- |
| 类型     | int number           |
| 是否必填 | 是                   |
| 说明     | 需要点赞消息的消息id |

入参2(form-data):

| 键名     | put_type       |
| -------- | -------------- |
| 类型     | int number     |
| 是否必填 | 是             |
| 格式要求 | 填thumb_ub     |
| 说明     | 表示点赞该消息 |

出参(json):

| 键名     | error_code | msg      |
| -------- | ---------- | -------- |
| 类型     | string     | string   |
| 是否必填 | 是         | 是       |
| 说明     | 错误码     | 错误消息 |

可能的错误码:

| 错误码 | 20000    | 20001                              | 20002                  | 21201              | 20003                | 21202                |
| ------ | -------- | ---------------------------------- | ---------------------- | ------------------ | -------------------- | -------------------- |
| 说明   | 修改成功 | 服务暂时不可用                     | 鉴权失败, Token不合法  | 错误, 没有这个消息 | 失败, 错误的入参参数 | 错误, 你已经点赞了它 |
| msg    | OK       | service not available temporarily | invalid identity token | no such message    | invalid parameters   | You already liked it |

**13. 删除消息**

接口:

| 接口地址             | HTTP方法 | 是否JWT鉴权 |
| -------------------- | -------- | ----------- |
| /message/{int msgid} | DELETE   | 是          |

入参1(接口地址传参):

| 键名     | msgid                                      |
| -------- | ------------------------------------------ |
| 类型     | int number                                 |
| 是否必填 | 是                                         |
| 说明     | 需要删除的消息id, 需要本人是该消息的创建者 |

出参(json):

| 键名     | error_code | msg      |
| -------- | ---------- | -------- |
| 类型     | string     | string   |
| 是否必填 | 是         | 是       |
| 说明     | 错误码     | 错误消息 |

可能的错误码:

| 误码 | 20000    | 20001                              | 20002                  | 21301              | 20003                | 21302                                |
| ---- | -------- | ---------------------------------- | ---------------------- | ------------------ | -------------------- | ------------------------------------ |
| 说明 | 删除成功 | 服务暂时不可用                     | 鉴权失败, Token不合法  | 错误, 没有这个消息 | 失败, 错误的入参参数 | 错误, 不是这条消息的创建者, 没有权限 |
| msg  | OK       | service not available temporarily | invalid identity token | no such message    | invalid parameters   | no permission to delete it           |
