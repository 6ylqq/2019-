package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"github.com/garyburd/redigo/redis"
	"github.com/gorilla/mux"
	"gopkg.in/mgo.v2"
	"io"
	"math/rand"
	"net/http"
	"os"
	"strconv"
)


//定义RESTful的一系列操作
type Person struct {
	ID        string   `json:"id,omitemty"`
	Firstname string   `json:"firstname,omitempty"`
	Lastname  string   `json:"lastname,omitempty"`
	Address   *Address `json:"address,omitempty"`
}

type Address struct {
	City     string `json:"city,omitempty"`
	Province string `json:"province,omitempty"`
}

var people []Person

func GetPeople(w http.ResponseWriter,req *http.Request){
	json.NewEncoder(w).Encode(people)
}

func GetPersonbyID(w http.ResponseWriter,req *http.Request){
	params:=mux.Vars(req)
	for _,item:=range people{
		if item.ID==params["id"]{
			json.NewEncoder(w).Encode(item)
			return
		}
	}
	json.NewEncoder(w).Encode(people)
}

func PostPerson(w http.ResponseWriter, req *http.Request) {
	params := mux.Vars(req)
	var person Person
	_ = json.NewDecoder(req.Body).Decode(&person)//解码josn编码
	person.ID = params["id"]
	people = append(people, person)//切片添加
	json.NewEncoder(w).Encode(people)
}

func DeletePerson(w http.ResponseWriter, req *http.Request) {
	params := mux.Vars(req)
	for index, item := range people {
		if item.ID == params["id"] {
			people = append(people[:index], people[index+1:]...)
			break
		}
	}
	json.NewEncoder(w).Encode(people)//重新编码全部以应用变更
}

//主程序数据类型
type Value struct {
	Who string
	Value_amount string
	Where string
	Length []int
}

func main(){

	people = append(people, Person{ID: "1", Firstname: "yang", Lastname: "zedong", Address: &Address{City: "Chengdu", Province: "Sichuan"}})
	people = append(people, Person{ID: "2", Firstname: "liu", Lastname: "xinling", Address: &Address{City: "Qingdao", Province: "Shandong"}})
	people = append(people, Person{ID: "3", Firstname: "Xia", Lastname: "qiyang", Address: &Address{City: "Chengdu", Province: "Sichuan"}})
	people = append(people, Person{ID: "4", Firstname: "Lei", Lastname: "shiduo", Address: &Address{City: "Chengdu", Province: "Sichuan"}})
	people = append(people, Person{ID: "5", Firstname: "liu", Lastname: "liu", Address: &Address{City: "Chengdu", Province: "Sichuan"}})

	//对客户端进行初始化，将数据操作等放上去
	router:=mux.NewRouter()
	router.HandleFunc("/people",GetPeople).Methods("GET")
	router.HandleFunc("/people/{id}", GetPersonbyID).Methods("GET")

	// 将信息发送到客户端
	router.HandleFunc("/people/{id}", PostPerson).Methods("POST")

	// 根据ID删除信息，切片特性
	router.HandleFunc("/people/{id}", DeletePerson).Methods("DELETE")

	// 上传数据成功
	fmt.Println("The loadcoal host has been created!")
	fmt.Println("You can view some info on: http://localhost:9899/people or http://localhost:9899/people/{id}")

	http.ListenAndServe(":9899", router)//启动API响应操作，由于没整好验证，所以想要验证下面的功能要把这个注释掉
	//使用JWT Token/SSL HTTPS安全机制验证用户是否正确，确保可以进如系统

	//（时间原因，还在理解这两个的机制，大概意思应该是密码的解密加密问题，然后套上Redis作为缓存，验证成功后就可以进行下一步操作


	//导入csv文件
	fileName:="C:\\Users\\杨泽东\\go\\src\\bank\\info.csv"
	fs,err:=os.Open(fileName)
	if err!=nil{
		fmt.Println("can not open the file,err is %v",err)
	}
	defer fs.Close()

	//使用csv的函数创建一个NewReader
	r:=csv.NewReader(fs)

	//呼叫Redis
	conn,err:=redis.Dial("tcp","127.0.0.1:6379")
	if(err!=nil){
		fmt.Println("conn redis failed,",err)
		return
	}
	defer conn.Close()

	//呼叫MongoDB数据库
	session,err:=mgo.Dial("")
	if err!=nil{
		panic(err)
	}
	defer session.Close()
	db:=session.DB("test")
	c:=db.C("bank")

	dbsize:=0//记录数据库大小，由于找不到相关函数，只能出此下策
	var true_row [][]int//使用矩阵储存节点之间的路径长度，便于后面的最短路径算法
	for {
		row,err:=r.Read()
		if err!=nil && err!=io.EOF {
			fmt.Println("can not read,err is %v", err)
		}
		if err==io.EOF{
			break
		}
		//将个点位置距离转换为int类型，这里采用一个节点的路径对应为矩阵的一行
		//这里有个无名BUG，strconv的函数返回值是没错的，但在赋值的时候，直接跳到了数据库session的Panic函数
		for count:=3;count<14;count++{
			true_row[dbsize][count],_=strconv.Atoi(row[count])
		}
		//使用mongodb储存csv文件
		err=c.Insert(
			&Value{
				Who:          row[0],
				Value_amount: row[1],
				Where:   row[2],
				Length:   true_row[dbsize],
			})
		if err!=nil{
			panic(err)
		}
		dbsize++//记录数据库大小
	}

	var seletc_value []Value//存放随机抽取的数据
	c.Find(nil).All(&seletc_value)//访问并存储c这个集合的全部元素


	 rand.Seed(100)
	var amount,i int
	fmt.Println("Please enter the person amount you want to look up:")//6. 抽样百分比可输入，根据百分比随机抽样座位号
	fmt.Scanf("%d",amount)
	amount*=dbsize

	for ; amount>0;amount-- {
		i=rand.Intn(dbsize)//随机抽取数据库中的座位数据并缓存在Redis中，但目前依然存在会抽到重复的号码问题
		_, err = conn.Do("set", "Who", seletc_value[i].Who)
		if err != nil {
			fmt.Println("err falied!")
		}
		_, err = conn.Do("set", "Value_amount", seletc_value[i].Value_amount)
		if err != nil {
			fmt.Println("err falied!")
		}
		_, err = conn.Do("set", "Addression", seletc_value[i].Where)
		if err != nil {
			fmt.Println("err falied!")
		}
		_, err = conn.Do("set", "Length", seletc_value[i].Length)
		if err != nil {
			fmt.Println("err falied!")
		}
	}

//最短路径算法，目前 是打算用之前生成的二维数组进行，但需求中说明要在随机抽取的座位中抽取，也就是抽取后还需要建立一个全新的数组
//然后时间不怎么够，但可以根据上学期的C语言代码进行仿制，关键是抽取后新二维数组的建立，需要在原始数据上进行删改

//9,输出路径图片目前是打算仿照C语言的Floyd算法进行路径的输出，可视化图片输出可能就需要另外的包


//10,很遗憾时间没安排好，之后继续完善
}

//func NewGraph()[][]int{
//
//}