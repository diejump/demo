package api

import (
	"gin-demo/api/middleware"
	"gin-demo/dao"
	"gin-demo/model"
	"gin-demo/utils"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/gomodule/redigo/redis"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"strconv"
	"time"
)

var rds redis.Conn

func register(c *gin.Context) {
	if err := c.ShouldBind(&model.User{}); err != nil {
		utils.RespSuccess(c, "verification failed")
		return
	}
	Account := c.PostForm("account")
	Password := c.PostForm("password")
	Username := c.PostForm("username")

	flag := dao.SelectUser(Account)
	if flag {
		utils.RespFail(c, "user already exists")
		return
	}

	dao.AddUser(Account, Username, dao.GetPassWord(Password)) //添加用户

	utils.RespSuccess(c, "add user successful")
}

func login(c *gin.Context) {
	if err := c.ShouldBind(&model.UserLogin{}); err != nil {
		utils.RespFail(c, "verification failed")
		return
	}
	account := c.PostForm("account")
	password := c.PostForm("password")

	flag := dao.SelectUser(account)
	if !flag {
		utils.RespFail(c, "user doesn't exists")
		return
	}

	selectPassword := dao.SelectPasswordFromAccount(account)

	err := bcrypt.CompareHashAndPassword(selectPassword, []byte(password))
	if err != nil {
		print(selectPassword)
		utils.RespFail(c, "wrong password")
		return
	}

	claim := model.MyClaims{
		Account: account, // 自定义字段
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 2).Unix(), // 过期时间
			Issuer:    "cxk",                                // 签发人
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claim)
	tokenString, _ := token.SignedString(middleware.Secret)
	c.JSON(http.StatusOK, gin.H{
		"message": "欢迎," + dao.FindUsernameFromAccount(account),
		"token":   tokenString,
	})
}

func UserName(c *gin.Context) {
	rds = dao.RedisPOllInit().Get()
	username, _ := redis.String(rds.Do("lrange", "username", 0, -1))
	if len(username) <= 0 { //列表中没有
		println("缓存中查询不到")
		name := dao.Username()
		for _, p := range name {
			if p != "" {
				_, err := rds.Do("lpush", "username", p)
				if err != nil {
					println("错误")
				}
				rds.Do("expire", "username", 10)
			}
		}
		c.JSON(http.StatusOK, gin.H{
			"status":  200,
			"message": name,
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"status":  200,
			"message": username,
		})
	}
}

func RaiseQuestion(c *gin.Context) {
	question := c.PostForm("question")
	if question == "" {
		c.String(http.StatusOK, "问题不合规")
		return
	}
	account, _ := c.Get("account")
	dao.SaveQuestion(question, account)
	c.String(http.StatusOK, "问题已发布！")
}

func ShowQuestions(c *gin.Context) {
	//account, _ := c.Get("account")
	question := dao.QueryQuestion()
	id := c.PostForm("id")
	if len(question) != 0 {
		if id == "" {
			c.JSON(http.StatusOK, "提示: 输入问题序号回答对应问题")
			c.String(http.StatusOK, "\n")
			for _, value := range question {
				c.JSON(http.StatusOK, gin.H{
					"问题序号": value.Id,
					"发布人":  dao.FindUsernameFromAccount(value.MasterAccount),
					"问题":   value.Question,
				})
			}
		} else {
			id, _ := strconv.Atoi(id)

			for _, value := range question {
				if value.Id == id {
					c.JSON(http.StatusOK, gin.H{
						"问题序号": value.Id,
						"发布人":  dao.FindUsernameFromAccount(value.MasterAccount),
						"问题":   value.Question,
					})
					c.String(http.StatusOK, "\n"+"评论区：")
					c.String(http.StatusOK, "\n")

					comments := dao.GetComments(id)
					if len(comments) != 0 {
						i := 1
						for _, value := range comments {
							c.String(http.StatusOK, "\n")
							c.String(http.StatusOK, "%d楼", i)
							c.String(http.StatusOK, "  "+dao.FindUsernameFromAccount(value.MasterAccount)+": ")
							c.String(http.StatusOK, value.Comment)
							i++
						}
					} else {
						c.String(http.StatusOK, "不要让楼主太寂寞~~")
					}
				}
			}
		}
	} else {
		c.String(http.StatusOK, "暂时没有问题哦！")
	}
}

func GetAccountFromToken(c *gin.Context) {
	account, _ := c.Get("account")
	utils.RespSuccess(c, account.(string))
}

func GiveComments(c *gin.Context) {
	account, _ := c.Get("account")

	id := c.PostForm("id")
	comment := c.PostForm("comment")
	if id == "" || comment == "" {
		c.JSON(http.StatusOK, "输入不合规")
		return
	}
	err := dao.SaveComments(id, comment, account)
	if err != nil {
		c.JSON(http.StatusOK, "评论失败！")
		return
	}
	c.JSON(http.StatusOK, "评论成功！")
	return
}

func ShowMyQuestion(c *gin.Context) {
	account, _ := c.Get("account")
	question := dao.QueryMyQuestion(account)
	if len(question) != 0 {
		for _, value := range question {
			//c.String(http.StatusOK, "\n")
			c.JSON(http.StatusOK, gin.H{
				"问题序号": value.Id,
				"发布人":  dao.FindUsernameFromAccount(account),
				"问题":   value.Question,
			})
			comment := dao.ShowMyQuestionAnswer(value.Id)
			if len(comment) == 0 {
				c.String(http.StatusOK, "不要让楼主寂寞太久~~")
			}
			i := 1
			for _, value2 := range comment {
				c.String(http.StatusOK, "\n")
				c.String(http.StatusOK, "%d楼", i)
				c.String(http.StatusOK, "  "+dao.FindUsernameFromAccount(value2.MasterAccount)+": ")
				c.String(http.StatusOK, value2.Comment)
				i++
			}
			c.String(http.StatusOK, "\n"+"\n"+"\n")
		}
	} else {
		c.String(http.StatusOK, "暂时还没有发布问题！")

	}
}

func DeleteMyComments(c *gin.Context) {
	account, _ := c.Get("account")
	comment := c.PostForm("comment")
	question_id := c.PostForm("id")
	if dao.DeleteComments(comment, question_id, account) != nil {
		c.String(http.StatusOK, "删除失败")
		return
	}
	c.String(http.StatusOK, "删除评论成功！")
}

func DeleteMyQuestion(c *gin.Context) {
	account, _ := c.Get("account")
	id := c.PostForm("id")
	if dao.DeleteMyQuestionAndComments(id, account) != nil {
		c.String(http.StatusOK, "删除问题失败！")
		return
	}
	c.String(http.StatusOK, "删除问题和评论成功！")
}

func UpdateMyQuestion(c *gin.Context) {
	account, _ := c.Get("account")
	id := c.PostForm("id")
	question := c.PostForm("question")
	if dao.UpdateQuestion(id, question, account) != nil {
		c.String(http.StatusOK, "修改问题失败")
		return
	}
	c.String(http.StatusOK, "修改问题成功")
}
