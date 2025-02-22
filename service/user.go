package service

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	database "todolist.go/db"
)

const userkey = "user"
const userNamekey = "username"

func NewUserForm(ctx *gin.Context) {
	ctx.HTML(http.StatusOK, "new_user_form.html", gin.H{"Title": "Register user"})
}

func UserPage(ctx *gin.Context) {
	userName := sessions.Default(ctx).Get(userNamekey)
	ctx.HTML(http.StatusOK, "user.html", gin.H{"Title": "User page", "Username": userName})
}

func RegisterUser(ctx *gin.Context) {
	// フォームデータの受け取り
	username := ctx.PostForm("username")
	password := ctx.PostForm("password")
	passwordConfirm := ctx.PostForm("password_confirm")
	switch {
	case username == "":
		ctx.HTML(http.StatusBadRequest, "new_user_form.html", gin.H{"Title": "Register user", "Error": "Usernane is not provided", "Username": username})
	case password == "":
		ctx.HTML(http.StatusBadRequest, "new_user_form.html", gin.H{"Title": "Register user", "Error": "Password is not provided", "Username": username, "Password": password})
	case passwordConfirm == "":
		ctx.HTML(http.StatusBadRequest, "new_user_form.html", gin.H{"Title": "Register user", "Error": "Re-Input of password is not provided", "Username": username, "Password": password})
	case password != passwordConfirm:
		ctx.HTML(http.StatusBadRequest, "new_user_form.html", gin.H{"Title": "Register user", "Error": "Password does not match", "Username": username, "Password": password})
	}

	if ok, msg := checkPasswordFormat(password); !ok {
		ctx.HTML(http.StatusBadRequest, "new_user_form.html", gin.H{"Title": "Register user", "Error": msg, "Username": username, "Password": password})
		return
	}

	// DB 接続
	db, err := database.GetConnection()
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}

	// 重複チェック
	var duplicate int
	err = db.Get(&duplicate, "SELECT COUNT(*) FROM users WHERE name=?", username)
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	if duplicate > 0 {
		ctx.HTML(http.StatusBadRequest, "new_user_form.html", gin.H{"Title": "Register user", "Error": "Username is already taken", "Username": username, "Password": password})
		return
	}
	// DB への保存
	result, err := db.Exec("INSERT INTO users(name, password) VALUES (?, ?)", username, hash(password))
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}

	// 保存状態の確認
	id, _ := result.LastInsertId()
	var user database.User
	err = db.Get(&user, "SELECT id, name, password FROM users WHERE id = ?", id)
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}

	session := sessions.Default(ctx)
	session.Set(userkey, id)
	session.Set(userNamekey, username)
	session.Save()

	ctx.Redirect(http.StatusFound, "/")
}

func EditUserForm(ctx *gin.Context) {
	ctx.HTML(http.StatusOK, "form_edit_user.html", gin.H{"Title": "Edit user"})
}

func EditUser(ctx *gin.Context) {
	// フォームデータの受け取り
	currentUserName := sessions.Default(ctx).Get(userNamekey).(string)
	newUsername := ctx.PostForm("new_username")
	password := ctx.PostForm("password")
	newPassword := ctx.PostForm("new_password")
	newPasswordConfirm := ctx.PostForm("new_password_confirm")
	switch {
	case newUsername == "" && newPassword == "":
		ctx.HTML(http.StatusBadRequest, "form_edit_user.html", gin.H{"Title": "Edit user", "Error": "No Changes", "NewUsername": newUsername, "NewPassword": newPassword})
		return
	case newUsername == currentUserName && newPassword == "":
		ctx.HTML(http.StatusBadRequest, "form_edit_user.html", gin.H{"Title": "Edit user", "Error": "Same with current name", "NewUsername": newUsername, "NewPassword": newPassword})
		return
	case password == "":
		ctx.HTML(http.StatusBadRequest, "form_edit_user.html", gin.H{"Title": "Edit user", "Error": "Password is not provided", "NewUsername": newUsername, "NewPassword": newPassword})
		return
	case newPassword != "" && newPassword != newPasswordConfirm:
		ctx.HTML(http.StatusBadRequest, "form_edit_user.html", gin.H{"Title": "Edit user", "Error": "Re-Input of new password is not provided", "NewUsername": newUsername, "NewPassword": newPassword})
		return
	}

	// DB 接続
	db, err := database.GetConnection()
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}

	// 重複チェック
	if newUsername != "" {
		var duplicate int
		err = db.Get(&duplicate, "SELECT COUNT(*) FROM users WHERE name=?", newUsername)
		if err != nil {
			Error(http.StatusInternalServerError, err.Error())(ctx)
			return
		}
		if duplicate > 0 {
			ctx.HTML(http.StatusBadRequest, "form_edit_user.html", gin.H{"Title": "Edit user", "Error": "Username is already taken", "NewUsername": newUsername, "NewPassword": newPassword})
			return
		}
	} else {
		newUsername = currentUserName
	}

	if newPassword != "" {
		if ok, msg := checkPasswordFormat(newPassword); !ok {
			ctx.HTML(http.StatusBadRequest, "form_edit_user.html", gin.H{"Title": "Edit user", "Error": msg, "NewUsername": newUsername, "NewPassword": newPassword})
			return
		}
	} else {
		newPassword = password
	}

	// ユーザの取得
	var user database.User
	err = db.Get(&user, "SELECT id, name, password, is_deleted FROM users WHERE id = ?", sessions.Default(ctx).Get(userkey))
	if err != nil {
		ctx.HTML(http.StatusBadRequest, "form_edit_user.html", gin.H{"Title": "Edit user", "Error": "You don't exist", "NewUsername": newUsername, "NewPassword": newPassword})
		return
	}

	// パスワードの照合
	if hex.EncodeToString(user.Password) != hex.EncodeToString(hash(password)) {
		ctx.HTML(http.StatusBadRequest, "form_edit_user.html", gin.H{"Title": "Edit user", "Error": "Incorrect password", "NewUsername": newUsername, "NewPassword": newPassword})
		return
	}

	_, err = db.Exec("UPDATE users SET name = ?, password = ? WHERE id = ?", newUsername, hash(newPassword), user.ID)
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}

	session := sessions.Default(ctx)
	session.Set(userNamekey, newUsername)
	session.Save()

	ctx.Redirect(http.StatusFound, "/user")
}

func DeleteUser(ctx *gin.Context) {
	userID := sessions.Default(ctx).Get(userkey)

	// DB 接続
	db, err := database.GetConnection()
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}

	var user database.User
	err = db.Get(&user, "SELECT id, name, password, is_deleted FROM users WHERE id = ?", userID)
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}

	// 削除
	_, err = db.Exec("UPDATE users SET is_deleted = true WHERE id = ?", userID)
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}

	// セッションの削除
	Logout(ctx)
}

func LoginForm(ctx *gin.Context) {
	ctx.HTML(http.StatusOK, "login.html", gin.H{"Title": "Login"})
}

func Login(ctx *gin.Context) {
	username := ctx.PostForm("username")
	password := ctx.PostForm("password")

	db, err := database.GetConnection()
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}

	// ユーザの取得
	var user database.User
	err = db.Get(&user, "SELECT id, name, password, is_deleted FROM users WHERE name = ?", username)
	if err != nil {
		ctx.HTML(http.StatusBadRequest, "login.html", gin.H{"Title": "Login", "Username": username, "Error": "No such user"})
		return
	}

	// 削除済みユーザの場合
	if user.Deleted {
		ctx.HTML(http.StatusBadRequest, "login.html", gin.H{"Title": "Login", "Username": username, "Error": "No such user"})
		return
	}

	// パスワードの照合
	if hex.EncodeToString(user.Password) != hex.EncodeToString(hash(password)) {
		ctx.HTML(http.StatusBadRequest, "login.html", gin.H{"Title": "Login", "Username": username, "Error": "Incorrect password"})
		return
	}

	// セッションの保存
	session := sessions.Default(ctx)
	session.Set(userkey, user.ID)
	session.Set(userNamekey, user.Name)
	session.Save()

	ctx.Redirect(http.StatusFound, "/user")
}

func Logout(ctx *gin.Context) {
	session := sessions.Default(ctx)
	session.Clear()
	session.Options(sessions.Options{Path: "/", MaxAge: -1})
	session.Save()
	ctx.Redirect(http.StatusFound, "/")
}

func hash(pw string) []byte {
	const salt = "todolist.go#"
	h := sha256.New()
	h.Write([]byte(salt))
	h.Write([]byte(pw))
	return h.Sum(nil)
}

func checkPasswordFormat(pw string) (bool, string) {
	strlen := len(pw)
	if strlen < 8 {
		return false, "Password must be at least 8 characters"
	}
	var numUpper, numLower, numNumber int
	for _, c := range pw {
		switch {
		case 'A' <= c && c <= 'Z':
			numUpper++
		case 'a' <= c && c <= 'z':
			numLower++
		case '0' <= c && c <= '9':
			numNumber++
		}
	}
	if numNumber == strlen {
		return false, "Password only number is not allowed"
	}
	if numUpper+numLower == strlen {
		return false, "Password only alphabet is not allowed"
	}

	return true, ""
}
