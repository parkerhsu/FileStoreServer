package handler

import (
	"net/http"
	"io/ioutil"
	"time"
	"fmt"

	"FileStoreServer/util"
	"FileStoreServer/db"
)

var pwd_salt = "$%#120"

func SignupHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		data, err := ioutil.ReadFile("./static/view/signup.html")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Write(data)
	} else {
		r.ParseForm()
		userName := r.Form.Get("username")
		passwd := r.Form.Get("password")

		if len(userName) <= 3 || len(passwd) <= 3 {
			w.Write([]byte("Invalid parameters"))
			return
		}

		encPasswd := util.Sha1([]byte(passwd+pwd_salt))
		suc := db.UserSignUp(userName, encPasswd)
		if suc {
			w.Write([]byte("SUCCESS"))
		} else {
			
			w.Write([]byte("Sign up failed!"))
		}
	}
}

func SignInHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		http.Redirect(w, r, "/static/view/signin.html", http.StatusFound)
		return
	}
	r.ParseForm()
	username := r.Form.Get("username")
	passwd := r.Form.Get("password")
	encPasswd := util.Sha1([]byte(passwd+pwd_salt))

	// 密码检测
	pwdChecked := db.UserSignin(username, encPasswd)
	if !pwdChecked {
		w.Write([]byte("FAILED"))
		return
	}

	// 生成token
	token := GetToken(username)
	suc := db.UpdateToken(username, token)
	if !suc {
		w.Write([]byte("FAILED"))
		return
	}

	// 重定向到用户首页
	resp := util.RespMsg{
		Code: 0,
		Msg:  "OK",
		Data: struct {
			Location string
			Username string
			Token    string
		}{
			Location: "http://" + r.Host + "/static/view/home.html",
			Username: username,
			Token:    token,
		},
	}
	w.Write(resp.JSONBytes())
}

func UserInfoHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	username := r.Form.Get("username")
	token := r.Form.Get("token")

	isValid := IsTokenValid(token)
	if !isValid {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	user, err := db.GetUserInfo(username)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	resp := util.RespMsg{
		Code: 0,
		Msg:  "OK",
		Data: user,
	}
	w.Write(resp.JSONBytes())
}

func GetToken(username string) string {
	ts := fmt.Sprintf("%x", time.Now().Unix())
	tokenPrefix := util.MD5([]byte(username + ts))
	return tokenPrefix + ts[:8]
}

func IsTokenValid(token string) bool {
	if len(token) != 40 {
		return false
	}
	// TODO: 判断token的时效性，是否过期
	// TODO: 从数据库表tbl_user_token查询username对应的token信息
	// TODO: 对比两个token是否一致
	return true
}