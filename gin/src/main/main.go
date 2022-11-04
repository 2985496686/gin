package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/gorilla/sessions"
	"log"
	"net/http"
	"reflect"
	"time"
)

type Login struct {
	Name     string `json:"username" form:"username" binding:"NotNilAndNotAdmin"`
	Password string `json:"password" form:"password" binding:"required"`
}

// NotNullAndAdmin 自定义结构体验证
func NotNullAndAdmin(v *validator.Validate, topStruct reflect.Value, currentStructOrField reflect.Value,
	field reflect.Value, fieldType reflect.Type, fieldKind reflect.Kind, param string) bool {
	value := field.Interface().(string)
	return value != "" && value != "admin"
}

// 请求路径: /form   POST
func handleForm(context *gin.Context) {
	name := context.PostForm("username")
	password := context.PostForm("userpassword")
	context.JSON(http.StatusOK, gin.H{})
	context.String(http.StatusOK, name+"  "+password)
}

// 请求路径: /:name/:action  GET
func handleHelloWorld(context *gin.Context) {
	//获取请求路径中的参数
	name := context.Param("name")
	action := context.Param("action")
	//获取url后面带的参数 例localhost:8080?name=zhang
	queryName := context.Query("name")
	context.String(http.StatusOK, name+action+queryName)
}

// 请求路径:/hello
func handleRedirect(context *gin.Context) {
	//重定向
	context.Redirect(http.StatusMovedPermanently, "/form")
}

// MiddleWare 定义中间件
// 中间件类似于拦截器或过滤器，会在路由方法执行前和执行后进行前置处理和后置处理
func MiddleWare() gin.HandlerFunc {
	return func(context *gin.Context) {
		//前置处理
		context.Set("Request", "MiddleWare")
		t1 := time.Now()
		//后置处理
		context.Next()
		since := time.Since(t1)
		fmt.Printf("方法执行耗时:%s\n", since)
	}
}

// 通过密匙获取存储session的store
var store = sessions.NewCookieStore([]byte("session-secret"))

func main() {

	//将自定义的结构体校验方法绑定到validator中
	//validate := binding.Validator.Engine().(*validator.Validate)
	//validate.RegisterValidation("NotNilAndNotAdmin", NotNullAndAdmin,)
	//创建路由
	r := gin.Default() //默认会使用两个中间件，Logger(), Recovery()
	//使用中间件
	r.Use(MiddleWare())
	//加载模板
	r.LoadHTMLGlob("src/resources/html/*")
	r.GET("/index", func(context *gin.Context) {
		context.HTML(http.StatusOK, "index.tmpl", gin.H{
			"title": "hello",
			"text":  "你好，世界！",
		})
	})

	//分组路由
	group1 := r.Group("/v1")

	{
		group1.GET("/b1", func(context *gin.Context) {
			context.String(http.StatusOK, "hello world!1")
		})

		group1.GET("/b2", func(context *gin.Context) {
			context.String(http.StatusOK, "hello world! %d", 1)
		})
	}

	//单个文件上传
	r.POST("/upload", func(context *gin.Context) {
		//设置最大上传大小
		r.MaxMultipartMemory = 8 << 20
		//从form中获取文件
		file, err := context.FormFile("file")
		if err != nil {
			log.Println("文件上传失败！")
			context.Error(errors.New("文件太大"))
		}
		//保存图片
		err = context.SaveUploadedFile(file, "D:/图片/"+file.Filename)
		if err != nil {
			log.Println("文件保存失败！")
		}
		context.String(http.StatusOK, "文件名为:%s", file.Filename)
	})

	//上传多个文件
	r.POST("/upload/files", func(context *gin.Context) {
		//获取表单中的所有数据
		form, _ := context.MultipartForm()
		//获取上传的多个文件
		files := form.File["files"]
		for _, file := range files {
			if err := context.SaveUploadedFile(file, "D:/图片/"+file.Filename); err != nil {
				context.String(http.StatusBadRequest, "文件上传失败！")
				return
			}
		}
		context.String(http.StatusOK, "文件上传失败！")
	})
	//设置404页面
	r.NoRoute(func(context *gin.Context) {
		context.HTML(http.StatusNotFound, "404.tmpl", "")
	})

	r.POST("/login", MiddleWare(), func(context *gin.Context) {
		var user Login
		//将表单数据转换为json
		if err := context.ShouldBindJSON(&user); err != nil {
			context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		//将类序列为json串。注意:这里只有大写的字段才能被序列化为json串
		userJson, _ := json.Marshal(user)
		context.String(http.StatusOK, string(userJson))
	})

	r.GET("/middleware", func(context *gin.Context) {
		get, _ := context.Get("Request")
		context.String(http.StatusOK, "%s", get)
	})

	r.GET("/cookie", func(context *gin.Context) {
		//获取cookie
		_, err := context.Cookie("goSessionId")
		if err != nil {
			//cookie不存在，创建cookie
			myUuid, _ := uuid.NewRandom()
			context.SetCookie("key_cookie", myUuid.String(), 60, "/", "localhost", false, true)
		}
	})

	r.GET("/saveSession", func(context *gin.Context) {
		//获取session，为获取到创建一个新的session
		session, err := store.Get(context.Request, "session1")
		if err != nil {
			context.String(http.StatusOK, "session获取错误！")
		}
		session.Values["name"] = "张三"
		session.Values["sex"] = "男"
		//保存更改,会在客户端创建一个 name=session1 的cookie
		session.Save(context.Request, context.Writer)
	})

	r.GET("/getSession", func(context *gin.Context) {
		session, err := store.Get(context.Request, "session1")
		if err != nil {
			context.String(http.StatusOK, "session获取错误！")
		}
		name := session.Values["name"]
		sex := session.Values["sex"]
		context.String(http.StatusOK, "name:%s  sex:%s", name, sex)
	})

	//设置监听窗口，默认为8080
	err := r.Run(":8080")
	if err != nil {
		log.Println("连接错误！")
	}
}
