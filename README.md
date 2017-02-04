# 使用Golang快速构建JSON API服务

***代码仅用于介绍nex包的用法, 存在多处不严谨***

越来越多的WEB应用采用前后端完全分离, WEB前端使用Angular/Vue之类的库, 用JSON通过RESTful与服务器通讯,
同样的服务器接口不但可以用于WEB前端, 同时能用于动应用前端.

这个例子通过一个失物招领的信息管理系统来介绍如何使用[nex](https://github.com/chrislonng/nex)包快速
构建JSON API服务, 限于作者水平有限, 如果发现错误, 欢迎指正.

## REPO名字的由来

名字`Yue`来源于典故`拾金不昧`中的主人公, 穷秀才`何岳`两次将捡到的金子物归原主, 正好和这里例子`失物招领`
系统名字吻合. 这个系统就是个基本的CRUD系统, 主要操作`Clue(线索)`, 对`Clue`的增删改查, 希望可以通过这个
简单的例子, 能让读者明白`nex`如何使用, 为了保存示例简单, 减少依赖, 数据没使用数据库, 直接使用一个数组来
存储所有的`Clue`信息.

完整代码: https://github.com/chrislonng/yue

需要下载依赖
```
go get github.com/gorilla/mux
go get github.com/chrislonng/nex
```

这个代码不包括WEB前端, 只有后端的JSON API服务, 读者可以使用PostMan测试, Repo中包含PostMan的配置
可以导入PostMan. 如果之前没有使用过PostMan的同学, 可以从https://www.getpostman.com/获取.

## 示例

提供JSON API服务的RESTful接口, 使用了`mux.Router`进行多路复用
```
r.Handle("/clues", nex.Handler(createClue)).Methods("POST")          //创建
r.Handle("/clues", nex.Handler(clueList)).Methods("GET")             //获取列表

r.Handle("/clues/{id}", nex.Handler(clueInfo)).Methods("GET")        //获取Clue信息
r.Handle("/clues/{id}", nex.Handler(updateClue)).Methods("PUT")      //更新
r.Handle("/clues/{id}", nex.Handler(deleteClue)).Methods("DELETE")   //删除

r.Handle("/blob", nex.Handler(uploadFile)).Methods("POST")           //上传
```

### 使用中间件
```
// logMiddleware: 为每个请求打印日志
// startTimeMiddleware: 记录请求开始时间
// endTimeMiddleware: 记录请求结束时间, 用于统计系统性能, 或将数据放入Promethus
nex.Before(logMiddleware, startTimeMiddleware)
nex.After(endTimeMiddleware)
```

### 对请求与响应的自动序列化和反序列化
```
func createClue(c *ClueInfo) (*StringMessage, error) {
	title := strings.TrimSpace(c.Title)
	number := strings.TrimSpace(c.Number)

	if title == "" || number == "" {
		return nil, errors.New("title and number can not empty")
	}

	db.clues = append(db.clues, *c)
	return SuccessResponse, nil
}
```

上面的代码片段中, 返回参数必须是两个, 一个是正常逻辑返回到客户端的数据, 另一个是发生错误时, 返回给客户端的
数据, `nex`提供了一个对`error`的默认的Encode函数, 后面会介绍如何自定义编码函数, `c *ClueInfo`是将客
户端数据反序列化后的结构体

### 使用查询参数
```
func clueList(query nex.Form) (*ClueListResponse, error) {
	s := query.Get("start")
	c := query.Get("count")

	var start, count int
	var err error

	if s == "" {
		start = 0
	} else {
		start, err = strconv.Atoi(s)
		if err != nil {
			return nil, err
		}
	}

	if c == "" {
		count = len(db.clues)
	} else {
		count, err = strconv.Atoi(c)
		if err != nil {
			return nil, err
		}
	}

	return &ClueListResponse{Data: db.clues[start : start+count]}, nil
}
```

### 使用查询参数辅助函数
```
// 与clueList函数功能相同, 使用query辅助函数
func clueList2(query nex.Form) (*ClueListResponse, error) {
	start := query.IntOrDefault("start", 0)
	count := query.IntOrDefault("count", len(db.clues))

	return &ClueListResponse{Data: db.clues[start: start+count]}, nil
}
```

在参数列表中使用`nex.Form`或者`*nex.Form`, 可以自动获取查询参数, 具体用法和原生`http.Request`中的`Form`
一样, 同时也可以使用`nex.PostForm`或者`*nex.PostForm`, 获取`Post`参数, 是对`http.Request`中的`PostForm`
的封装

### 获取原始的`Request`
```
func clueInfo(r *http.Request) (*ClueInfoResponse, error) {
	id, err := parseID(r)
	if err != nil {
		return nil, err
	}

	return &ClueInfoResponse{Data:&db.clues[id-1]}, nil
}
```

可以在函数签名中直接使用`*http.Request`获取原始的Request


### 组合使用不同的参数
```
func updateClue(r *http.Request, c *ClueInfo) (*StringMessage, error) {
	id, err := parseID(r)
	if err != nil {
		return nil, err
	}

	title := strings.TrimSpace(c.Title)
	number := strings.TrimSpace(c.Number)

	if title == "" || number == "" {
		return nil, errors.New("title and number can not empty")
	}

	db.clues[id] = *c

	return SuccessResponse, nil
}
```
所有`nex`支持的类型, 都可以在函数签名中使用, 没有顺序要求, `nex`支持的类型, 详见[nex](https://github.com/chrislonng/nex)

### 在参数中使用`http.Header`
```
func deleteClue(h http.Header, r *http.Request) (*StringMessage, error) {
	t := h.Get("Authorization")
	if t != token {
		return nil, errors.New("permission denied")
	}

	id, err := parseID(r)
	if err != nil {
		return nil, err
	}

	db.clues = append(db.clues[:id], db.clues[id:]...)
	return SuccessResponse, nil
}
```

这里为了演示如何在`nex`的函数中使用`http.Header`, 在删除`Clue`是需要客户端在`Header`中加入`Authorization`字段
值等于服务器的`token`时, 才能删除, 这里仅用于严实才这样写的

### 上传文件
```
func uploadFile(form *multipart.Form) (*BlobResponse, error) {
	uploaded, ok := form.File["uploadfile"]
	if !ok {
		return nil, errors.New("can not found `uploadfile` field")
	}

	localName := func(filename string) string {
		ext := filepath.Ext(filename)
		id := time.Now().Format("20060102150405.999999999")
		return id + ext
	}

	var fds []io.Closer
	defer func() {
		for _, fd := range fds {
			fd.Close()
		}
	}()

	files := make(map[string]string)
	for _, fh := range uploaded {
		fileName := localName(fh.Filename)
		files[fh.Filename] = fileName

		// upload file
		uf, err := fh.Open()
		fds = append(fds, uf)
		if err != nil {
			return nil, err
		}

		// local file
		lf, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY, 0660)
		fds = append(fds, lf)
		if err != nil {
			return nil, err
		}

		_, err = io.Copy(lf, uf)
		if err != nil {
			return nil, err
		}
	}

	return &BlobResponse{Data: files}, nil
}
```

使用`*multipart.Form`类型, 可以获取`http.Request`中的`MultipartForm`字段, 这里上传的文件, 简单的存在本地
并返回文件在服务器的文件名

### 自定义错误编码函数
```
nex.SetErrorEncoder(func(err error) interface{} {
    return &ErrorMessage{
        Code:  -1000,
        Error: err.Error(),
    }
})
```
上面的代码通过自定义错误编码函数, 将所有的错误信息`Code`设为-1000, 实际开发中可能会根据不同的错误生成不同的错误码,
以及返回相应的错误信息, 包括过滤一部分服务器的敏感信息, 通常可以在开发过程中, 通过golang的`+build`来设置不同的`tags`
最终在`release`和`develop`版本包含不同级别的错误信息.

## 总结

`nex`主要用于将一个符合`nex`签名的函数转换成符合`http.Handler`接口的结构, 并在请求到达时, 自动进行依赖注入,
相对于`HandleFunc`更加便于写单元测试, 并且减少在各个接口中序列化反序列化中的大量冗余代码, 我在使用[go-kit](https://github.com/go-kit/kit)
的过程中就存在这个问题.

相关功能逻辑单元如果需要新的依赖, 只需要在函数签名中新加一个参数即可, 在实际使用中还是比较方便, `nex`的函数必须
包含两个返回值, 一个返回值代表正常返回数据, 另一个返回值代表错误信息

---

欢迎任何关于`nex`的建议及意见, e-mail: chris@lonng.org, 欢迎Star, [nex传送门](https://github.com/chrislonng/nex)

## License
Copyright (c) <2016> <chris@lonng.org>


Permission is hereby granted, free of charge, to any person obtaining a copy of 
this software and associated documentation files (the "Software"), to deal in 
the Software without restriction, including without limitation the rights to use, 
copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the 
Software, and to permit persons to whom the Software is furnished to do so, subject 
to the following conditions:

The above copyright notice and this permission notice shall be included in all 
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, 
INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A 
PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT 
HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION 
OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE 
SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.