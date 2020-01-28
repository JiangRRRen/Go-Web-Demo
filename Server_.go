package main

import (
	"encoding/json"
	"net/http"
	"path"
	"strconv"
)


func main(){
	server := http.Server{
		Addr:"127.0.0.1:8080",
	}
	http.HandleFunc("/order/",handleRequest)
	server.ListenAndServe()
}

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


