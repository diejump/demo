package dao

import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gomodule/redigo/redis"
	"golang.org/x/crypto/bcrypt"
	"log"
	"time"
)

type User struct {
	Username string
	Account  string
	Password string
}

type Questions struct {
	Id            int
	MasterAccount string
	Question      string
}

type Comments struct {
	MasterAccount string
	Comment       string
}

var db *sql.DB

var Rds redis.Conn

func RedisPOllInit() *redis.Pool {
	return &redis.Pool{
		MaxIdle:     5,
		MaxActive:   0,
		Wait:        true,
		IdleTimeout: time.Duration(1) * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", "127.0.0.1:6379")
			if err != nil {
				fmt.Println(err)
				return nil, err
			}
			println("连接成功")
			redis.DialDatabase(0)
			return c, err
		},
	}
}

func RedisClose() {
	Rds.Close()
}

func InitDB() {
	var err error

	dsn := "root:123@tcp(127.0.0.1:3306)/school?charset=utf8mb4&parseTime=True&loc=Local"

	db, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalln(err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("DB connect success")
	return
}

func AddUser(Account, Username string, Password []byte) {
	sqlstr := "insert into user (username,account,password) values (?,?,?)"
	_, err := db.Exec(sqlstr, Username, Account, Password)
	if err != nil {
		fmt.Printf("insert failed, err:%v\n", err)
		return
	}
	log.Println("insert success")
}

func SelectUser(Account string) bool {
	sqlStr := "select password from user where account =?"
	var password string
	db.QueryRow(sqlStr, Account).Scan(&password)
	if password != "" {
		return true
	}
	return false
}

func SelectPasswordFromAccount(Account string) []byte {
	sqlstr := "select password from user where account=?"
	var password []byte
	db.QueryRow(sqlstr, Account).Scan(&password)
	return password
}

func Username() [20]string {
	sqlstr := "select username from user"
	var username [20]string
	rows, err := db.Query(sqlstr)
	if err != nil {
		fmt.Printf("query failed, err:%v\n", err)
	}
	defer rows.Close()
	i := 0
	for rows.Next() {
		err := rows.Scan(&username[i])
		i++
		if err != nil {
			fmt.Printf("scan failed, err:%v\n", err)
		}
	}
	return username
}

func SaveQuestion(question string, account any) {
	sqlstr := "insert into question (master_account,question) values (?,?)"
	_, err := db.Exec(sqlstr, account, question)
	if err != nil {
		fmt.Printf("insert failed, err:%v\n", err)
		return
	}
	log.Println("insert success")
	_, _ = db.Exec("alter table `question` drop `id`")
	_, _ = db.Exec("alter table `question` add `id` int not null first")
	_, _ = db.Exec("alter table `question` modify column `id` int not null AUTO_INCREMENT,ADD PRIMARY KEY(id)")
}

func QueryQuestion() []Questions {
	sqlStr := "select id,master_account,question from question"
	rows, err := db.Query(sqlStr)
	if err != nil {
		fmt.Printf("query failed, err:%v\n", err)
		return nil
	}
	defer rows.Close()

	question := make([]Questions, 0)

	for rows.Next() {
		var u Questions
		err := rows.Scan(&u.Id, &u.MasterAccount, &u.Question)
		if err != nil {
			fmt.Printf("scan failed, err:%v\n", err)
			return nil
		}
		question = append(question, u)
	}
	return question
}

func FindUsernameFromAccount(account any) string {
	sqlstr := "select username from user where account = ?"
	var username string
	db.QueryRow(sqlstr, account).Scan(&username)
	return username
}

func SaveComments(id, comment string, account any) error {
	sqlstr := "insert into comments (question_id,master_account,comment) values (?,?,?)"
	_, err := db.Exec(sqlstr, id, account, comment)
	if err != nil {
		fmt.Printf("评论插入失败, err:%v\n", err)
		return err
	}
	log.Println("评论插入成功")
	return nil
}

func GetComments(id int) []Comments {
	sqlStr := "select master_account,comment from comments where question_id=?"
	rows, err := db.Query(sqlStr, id)
	if err != nil {
		fmt.Printf("query failed, err:%v\n", err)
		return nil
	}

	defer rows.Close()

	comment := make([]Comments, 0)

	for rows.Next() {
		var u Comments
		err := rows.Scan(&u.MasterAccount, &u.Comment)
		if err != nil {
			fmt.Printf("scan failed, err:%v\n", err)
			return nil
		}
		comment = append(comment, u)
	}
	return comment
}

func QueryMyQuestion(account any) []Questions {
	sqlStr := "select id,question from question where master_account=?"
	rows, err := db.Query(sqlStr, account)
	if err != nil {
		fmt.Printf("query failed, err:%v\n", err)
		return nil
	}
	defer rows.Close()

	question := make([]Questions, 0)

	for rows.Next() {
		var u Questions
		err := rows.Scan(&u.Id, &u.Question)
		if err != nil {
			fmt.Printf("scan failed, err:%v\n", err)
			return nil
		}
		question = append(question, u)
	}
	return question
}

func ShowMyQuestionAnswer(question_id int) []Comments {
	sqlStr := "select master_account,comment from comments where question_id=?"
	rows, err := db.Query(sqlStr, question_id)

	if err != nil {
		fmt.Printf("query failed, err:%v\n", err)
		return nil
	}
	defer rows.Close()

	comment := make([]Comments, 0)

	for rows.Next() {
		var u Comments
		err := rows.Scan(&u.MasterAccount, &u.Comment)
		if err != nil {
			fmt.Printf("scan failed, err:%v\n", err)
			return nil
		}
		comment = append(comment, u)
	}
	return comment
}

func DeleteComments(comment, question_id string, master_account any) error {
	var c string
	db.QueryRow("select comment from comments where master_account=? and question_id=?", master_account, question_id).Scan(&c)
	if c == "" {
		return errors.New("")
	}

	sqlstr := "delete from comments where master_account=? and comment=? and question_id=?"
	_, err := db.Exec(sqlstr, master_account, comment, question_id)
	if err != nil {
		fmt.Printf("删除评论失败, err:%v\n", err)
		return err
	}
	log.Println("删除成功")
	return nil
}

func DeleteMyQuestionAndComments(id string, master_account any) error {
	var c string
	db.QueryRow("select question from question where id=? and master_account=?", id, master_account).Scan(&c)
	if c == "" {
		return errors.New("")
	}

	sqlstr1 := "delete from question where master_account=? and id=?"
	//sqlstr2 := "delete from comments where question_id=?"
	_, err := db.Exec(sqlstr1, master_account, id)
	//_, err2 := db.Exec(sqlstr2, id)
	if err != nil {
		fmt.Printf("删除问题失败, err:%v\n", err)
		return err
	}
	log.Println("删除成功")

	/*questions1 := QueryQuestion()
	question_id1 := make([]int, 0)
	for _, value := range questions1 {
		question_id1 = append(question_id1, value.Id)
	} //删除后、重新排序前各问题id
	//_, _ = db.Exec("update comments set former_question_id= question_id")*/

	_, _ = db.Exec("alter table `question` drop `id`")
	_, _ = db.Exec("alter table `question` add `id` int not null first")
	_, _ = db.Exec("alter table `question` modify column `id` int not null AUTO_INCREMENT,ADD PRIMARY KEY(id)")

	/*questions2 := QueryQuestion()
	question_id2 := make([]int, 0)
	for _, value := range questions2 {
		question_id2 = append(question_id2, value.Id)
	} //重新排序后id

	for i := 0; i < len(questions1); i++ {
		sqlstr := "update comments set question_id=? where question_id=?"
		db.Exec(sqlstr, question_id2[i], question_id1[i])
	}*/
	return nil
}

func UpdateQuestion(id, question string, master_account any) error {
	var c string
	db.QueryRow("select question from question where id=? and master_account=?", id, master_account).Scan(&c)
	if c == "" {
		return errors.New("")
	}

	sqlstr := "update question set question=? where id=? and master_account=?"
	_, err := db.Exec(sqlstr, question, id, master_account)
	if err != nil {
		fmt.Printf("update failed, err:%v\n", err)
		return err
	}
	log.Println("update success")
	return nil
}

func GetPassWord(password string) []byte { //密码加密
	password2, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return password2
}
