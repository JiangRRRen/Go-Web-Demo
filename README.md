---
title: GoWeb编程7-Web实战1
category:
  - go
tags:
 - 计算机网络
 - go
mathjax: true
date: 2020-01-27 15:29:27
---

本节是针对前面学到的数据库操作和Web编程的综合应用。

# 1. 数据库操作

用mysql进行测试，测试用数据源如下，其中`order_num`作为主键。

<img src="https://uk-1259555870.cos.eu-frankfurt.myqcloud.com/20200127153214.png"  style="zoom:75%;display: block; margin: 0px auto; vertical-align: middle;">

## 1.1 连接初始化

连接初始化我们需要4块代码：

- 连接信息
- 全局变量：指向数据库的指针
- 初始化函数
- 一个结构变量，作为数据载体

首先是连接信息：

```go
const(
	userName ="test"
	password = "asdfg12345"
	ip ="cdb-axt937vt.gz.tencentcdb.com"
	port="10059"
	dbName = "test"
)
```

关键词`const`加上括号，有两个作用：

- 变量值自增，类似于C++的枚举（这里用不到）
- 全部设置为常量类型，比较方便

---

之后我们需要声明一个全局变量`DB`，作为数据库的指针。

```go
var DB *sql.DB
```

---

然后重点就是初始化连接函数：

```go
func init(){
	connectInfo:=[]string{userName,":",password,"@tcp(",ip,":",port,")/", dbName, "?charset=utf8"}
	path := strings.Join(connectInfo,"")
	DB,_=sql.Open("mysql",path)

	if err:=DB.Ping(); err !=nil{
		fmt.Println("open database fail!")
		return
	}
	fmt.Println("connect success")
}
```

必须声明为`init()`，这个函数就会在主函数之前调用。这个函数首先**将零散的连接信息拼接成一个字符切片。**`strings.Join`方法将切片的字符串元素拼接起来，比如：

```go
string [] tmpStr={abc,def,ghi};
string jn = string.Join(tmpStr，"-");
//jn="abc-def-ghi"
```

我们之前已经通过如下语句引入了Mysql驱动：

```go
import _"github.com/Go-SQL-Driver/MySQL"
```

所以`DB,_=sql.Open("mysql",path)`可以直接连接mysql。链结构，还要利用`ping`来检测连接是否成功。

---

之后，我们需要针对特定的数据指定相应的结构体来承载数据：

```go
type Orders struct{
	Order_num int `json:"order_num"`
	Order_date string `json:"order_date"`
	Cust_id int `json:"cust_id"`
}
```

## 1.2 查询

查询我们分为两类：

- 单行查询，指定主键，查询内容
- 多行查询，返回所有内容

单行查询非常简单，思路是：**传入查询的主键，返回查询结果（order结构）和错误**。具体步骤是：

1. 新建一个order结构，**不要使用`:=`**。
2. 调用查询语句
3. 将查询结果扫描进order结构，同时赋值错误，**不要使用`：=`**

```go
func QueryOne(id int)(order Orders,err error){
	order = Orders{}
	row := DB.QueryRow("SELECT order_num, order_date,cust_id FROM orders WHERE order_num=?",id)
	err = row.Scan(&order.Order_num,&order.Order_date,&order.Cust_id)
	return
}
```

主要需要关注查询语句，可以通过`order_num=?`向其中填值。

---

多行查询比较特殊，我们调用查询语句后返回的一个特殊的数据结构`database/sql.Rows`

<img src="https://uk-1259555870.cos.eu-frankfurt.myqcloud.com/20200127155739.png"  style="zoom:75%;display: block; margin: 0px auto; vertical-align: middle;">

结构体的`Next()`函数能够判断，是否存在下一行，因此我们需要循环读取，**每次扫描都会清除一行**：

```go
func QueryMulti()(orders[] Orders,err error){
	orders =[]Orders{}
	rows,err:=DB.Query("SELECT * FROM orders")
	if err!=nil{
		fmt.Printf("Query failed,err:%v\n", err)
		return
	}
	i:=0
	for rows.Next(){
		orders=append(orders, Orders{})
		newerr := rows.Scan(&orders[i].Order_num,&orders[i].Order_date,&orders[i].Cust_id)
		if newerr != nil {
			fmt.Printf("Scan failed,err:%v\n", err)
		}
		i++
	}
	return
}
```

另一个重点是**改用切片的形式返回数据**，在next循环中切片需要用`append`添加！

## 1.2 修改数据

改动数据有三种方式：增加、删除、更新，这三种方法其实大同小异。这里我们需要**使用事务来保证原子性**，防止并发时的资源竞争。

一般查询使用的是db对象的方法，事务则是使用另外一个对象。sql.Tx对象。使用db的Begin方法可以创建tx对象。

创建Tx对象的时候，会从连接池中取出连接。事务的连接生命周期从Beigin函数调用起，直到Commit和Rollback函数的调用结束。事务也提供了prepare语句的使用方式，但是需要使用Tx.Stmt方法创建。**所以在事务开启过程中，是不能使用DB的方法的！**

```go
tx, err := db.Begin()
db.Exec(query1) //wrong!
tx.Exec(query2)
tx.commit()
```

```go
func Insert(order Orders)(err error){
	tx,err := DB.Begin()
	if err != nil{
		fmt.Println("tx fail")
		return
	}
	stmt,err := tx.Prepare("INSERT INTO orders (`order_num`,`order_date`,`cust_id`) VALUES(?,?,?)")
	if err != nil{
		fmt.Println("Prepare fail")
		return
	}
	res,err:=stmt.Exec(order.Order_num,order.Order_date,order.Cust_id)
	if err != nil{
		fmt.Println("Exec fail")
		return
	}
	tx.Commit()
	fmt.Println(res.LastInsertId())
	return
}

func Delete(id int) (err error){
	tx,err := DB.Begin()
	if err != nil{
		fmt.Println("tx fail")
		return
	}
	stmt,err := tx.Prepare("DELETE FROM orders WHERE order_num = ?")
	if err != nil{
		fmt.Println("Prepare fail")
		return
	}
	res,err:=stmt.Exec(id)
	if err != nil{
		fmt.Println("Exec fail")
		return
	}
	tx.Commit()
	fmt.Println(res.LastInsertId())
	return
}
```

# 2. 服务端操作

服务端操作我们遵从REST原理：设计那些**通过标准的几个动作来操纵资源**。假设我们需要三个操作：

- 获取单行数据GET
- 获取全部数据MULTIGET
- 插入数据INSERT
- 删除数据DELETE

首先是主程序：

```go
func main(){
	server := http.Server{
		Addr:"127.0.0.1:8080",
	}
	http.HandleFunc("/order/",handleRequest)
	server.ListenAndServe()
}
```

这里使用了**多路复用器handleRequest**，多路复用器会根据请求使用的HTTP方法，把请求转发给相应的CRUD(create, read, update, delete)处理器函数。

```go
func handleRequest(w http.ResponseWriter, r *http.Request){
	var err error
	switch r.Method{
	case "GET":
		err = handleGet(w,r)
	case "MULTIGET":
		err = handleMultiGet(w,r)
	case "INSERT":
		err = handleInsert(w,r)
    case "DELETE":
        err = handleDelete(w,r)
	}
	if err!=nil{
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
```

下面就是多路复用器CRUD函数的实现：

## 2.1 查询数据

这一类操作，我们只需要发送相应的主键ID，即可操作数据。由于查询结果需要通过响应包返回，所以我们还需要将其转化为JSON格式，完整步骤如下：

1. 从URL中解析出想要查询的主键ID
2. 调用数据库查询函数获取查询结果（order结构体）
3. 将结果转化为JSON格式
4. 封装响应包（头，主体，状态码）

```go
func handleGet(w http.ResponseWriter, r *http.Request) (err error){
	id, err :=strconv.Atoi(path.Base(r.URL.Path))
	if err!=nil{
		return
	}
	order, err := QueryOne(id)
	if err!=nil{
		return
	}

	output, err := json.MarshalIndent(&order, "", "\t\t")
	if err != nil {
		return
	}
	w.Header().Set("Content-Type","application/json")
	w.Write(output)
	w.WriteHeader(200)
	return
}
```

在命令行输入：

```
curl -i -X GET http://127.0.0.1:8080/order/20011
```

即可查询主键为20011的数据

----

集体查询不需要解析URL，

```go
func handleMultiGet(w http.ResponseWriter, r *http.Request)(err error){
	//_, err =strconv.Atoi(path.Base(r.URL.Path))
	if err!=nil{
		return
	}
	orders, err := QueryMulti()
	if err!=nil{
		return
	}
	output, err := json.MarshalIndent(&orders, "", "\t\t")
	if err != nil {
		return
	}
	w.Header().Set("Content-Type","application/json")
	w.Write(output)
	w.WriteHeader(200)
	return
}
```

在命令行输入：

```
curl -i -X MULTIGET http://127.0.0.1:8080/order/
```

即可查询所有数据。

## 2.2 删除数据

删除数据不需要返回内容，所以更简单：

```go
func handleDelete(w http.ResponseWriter, r *http.Request)(err error) {
	id,err:=strconv.Atoi(path.Base(r.URL.Path))
	if err!=nil{
		return
	}
	err = Delete(id)
	if err != nil {
		return
	}
	w.WriteHeader(200)
	return
}
```

使用方法集体查询一样，只是将指令MULTIGET换成DELETE。

## 2.3 插入数据

插入数据有点像查询数据的逆过程：

1. 解析URL，读取JSON数据
2. 解析JSON数据，形成order结构体
3. 调用数据库插入函数

````go
func handleInsert(w http.ResponseWriter, r *http.Request)(err error){
	len := r.ContentLength
	body := make([]byte,len)
	r.Body.Read(body)
	var order Orders
	json.Unmarshal(body,&order)
	err = Insert(order)
	if err!=nil{
		return
	}
	w.WriteHeader(200)
	return
}
````

请求包的主体BODY是一个二进制数据，所以我们需要创建一个二进制切片来存储。

在命令行输入：

```
curl -i -X INSERT -H "Content-Type: application/json"  
    -d '{"order_date":"2020-01-01 00:00:00","cust_id":"10086"}' http://127.0.0.1:8080/order/
```

# Go-Web-Demo
