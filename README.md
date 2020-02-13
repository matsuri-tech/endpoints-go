# endpoints-go

- .endpoints.json ファイルを自動生成するための Echo ラッパー

## Usage

```go
e := echo.New()

ew := endpoints.NewEchoWrapper(e)

// ドメインの設定
ew.AddEnv(endpoints.Env{
    Version: "v1",
    Domain: endpoints.Domain{
        Local: "http://localhost:8000",
        LocalDev: "https://local-dev.hoge.com",
        Dev: "https://dev.hoge.com",
        Prod: "https://hoge.com",
    },
})

// エンドポイントの追加
ew.GET("/user", userHandler.GetUsers, endpoints.Desc{
    Name: "userIndex",
    Query: "page=2&itemsPerPage=50",
    Desc: "ユーザ一覧を取得する"
})

// グループの作成とエンドポイントの追加
articles := ew.Group("/articles")
articles.POST("/", articleHandler.CreateArticle, endpoints.Desc{
    Name: "createArticle",
    Query: "",
    Desc: "記事を新規作成する",
})

// .endpoints.jsonファイルの出力
if err := ew.Generate(".endpoints.json"); err != nil {
    log.Printf("failed to generate endpoints file: %v", err)
}
```
