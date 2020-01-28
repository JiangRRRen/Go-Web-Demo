// This sample program demonstrates how to create race
// conditions in our programs. We don't want to do this.

package main
import (
	"database/sql"
	"fmt"
	_"github.com/Go-SQL-Driver/MySQL"
	"strings"
)
const(
	userName ="test"
	password = "asdfg12345"
	ip ="cdb-axt937vt.gz.tencentcdb.com"
	port="10059"
	dbName = "test"
)
var DB *sql.DB


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

type Orders struct{
	Order_num int `json:"order_num"`
	Order_date string `json:"order_date"`
	Cust_id string `json:"cust_id"`
}

func QueryOne(id int)(order Orders,err error){
	order = Orders{}
	row := DB.QueryRow("SELECT order_num, order_date,cust_id FROM orders WHERE order_num=?",id)
	err = row.Scan(&order.Order_num,&order.Order_date,&order.Cust_id)
	return
}

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

