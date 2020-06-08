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
    Version: "v2",
    Domain: endpoints.Domain{
        Local: "http://localhost:8000",
        LocalDev: "https://local-dev.hoge.com",
        Dev: "https://v2.dev.hoge.com",
        Prod: "https://v2.hoge.com",
    }
})

// エンドポイントの追加
// Versionsが指定されていない場合、全てのバージョンに含まれる
ew.GET("/users", userHandler.GetUsers, endpoints.Desc{
    Name: "userIndex",
    Query: "page=2&itemsPerPage=50",
    Desc: "ユーザ一覧を取得する"
})

// このエンドポイントはv2にのみ含まれる
ew.GET("/messages", messageHandler.GetMessages, endpoints.Desc{
    Name: "messageIndex",
    Query: "page=2&itemsPerPage=50",
    Desc: "メッセージ一覧を取得する",
    Versions: endpoints.Versions{"v2"},
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
