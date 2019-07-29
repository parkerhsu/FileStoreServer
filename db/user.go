package db

import (
	"FileStoreServer/db/mysql"
	_ "database/sql"
	"log"
	"fmt"
)

// User : 用户表model
type User struct {
	Username     string
	Email        string
	Phone        string
	SignupAt     string
	LastActiveAt string
	Status       int
}

func UserSignUp(userName string, passwd string) bool {
	stmt, err := mysql.DBConn().Prepare("insert into tbl_user(user_name, user_pwd) " +
		"values (?,?)")
	if err != nil {
		log.Printf("Failed to prepare: %v\n", err)
		return false
	}
	defer stmt.Close()

	res, err := stmt.Exec(userName, passwd)
	if err != nil {
		log.Printf("Failed to insert: %v\n", err)
		return false
	}
	if newRow, err := res.RowsAffected(); err == nil && newRow > 0 {
		return true
	}
	return false
}


// 判断密码是否正确
func UserSignin(username string, encPasswd string) bool {
	stmt, err := mysql.DBConn().Prepare("select * from tbl_user where user_name=?")
	if err != nil {
		log.Println(err)
		return false
	}
	defer stmt.Close()

	rows, err := stmt.Query(username)
	if err != nil {
		log.Println(err)
		return false
	} else if rows == nil {
		log.Printf("username not found: %s\n", username)
		return false
	}

	pRows := mysql.ParseRows(rows)
	if len(pRows) > 0 && string(pRows[0]["user_pwd"].([]byte)) == encPasswd {
		return true
	}
	return false
}

func UpdateToken(username string, token string) bool {
	stmt, err := mysql.DBConn().Prepare(
	"replace into tbl_user_token(user_name, user_token) values(?,?)")
	if err != nil {
		log.Println(err)
		return false
	}
	defer stmt.Close()

	_, err = stmt.Exec(username, token)
	if err != nil {
		log.Println(err)
		return false
	}
	return true
}


// GetUserInfo : 查询用户信息
func GetUserInfo(username string) (User, error) {
	user := User{}

	stmt, err := mysql.DBConn().Prepare(
		"select user_name,signup_at from tbl_user where user_name=? limit 1")
	if err != nil {
		fmt.Println(err.Error())
		return user, err
	}
	defer stmt.Close()

	// 执行查询的操作
	err = stmt.QueryRow(username).Scan(&user.Username, &user.SignupAt)
	if err != nil {
		return user, err
	}
	return user, nil
}