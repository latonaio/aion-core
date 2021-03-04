## microservice client ライブラリ 

microservice と kanban 間の通信を担うクライアントライブラリを提供します。

### 使用方法

`pkg/msclient` をインポートしてください。

#### microservice から kanban へデータを送信

microservice から kanban へのデータ送信を行う手順は下記の通りです。

1. microservice client object を生成します。
2. 送信するデータ郡を用意します。
3. データ郡を元に output request object を構築します。(これが kanban 側に送信するデータの実体となります)
4. output request object を kanban 側に送信します。

下記サンプルコードは、`map[string]interface{}` 型で用意されたデータを kanban に送信するコードのサンプルです。

```
func writeKanban(data map[string]interface{}) error {
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    // microservice client object を生成
    client, err := msclient.NewKanbanClient(ctx)

    // 送信するデータ群を用意
    var options []msclient.Option
    options = append(options, msclient.SetResult(true))
    options = append(options, msclient.SetDataPath("/foo"))
    options = append(options, msclient.SetConnectionKey("default"))
    options = append(options, msclient.SetFileList([]string{"file 1", "file 2"}))
    options = append(options, msclient.SetMetadata(data))
    options = append(options, msclient.SetDeviceName("sample service"))

    // output request object を構築
    req, err := msclient.NewOutputData(options...)
    if err != nil {
	    return fmt.Errorf("failed to construct output request: %v", err)
    }

    // kanban へ output request object を送信
    if err := client.OutputKanban(req); err != nil {
	    return fmt.Errorf("failed to output to kanban: %v", err)
    }
    return nil
}
```

#### kanban からのデータを microservice で受信

kanban からの通信を microservice で受信する手順は下記の通りです。
1. microservice client object を生成します。
2. kanban channel 生成します。
3. kanban channel を監視し、受信を検知します。

下記のサンプルコードは kanban からの通信をポーリングし、受信データの内容をログに出力するサンプルです。 

``` 
func main() {
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    // microservice client object を生成
    client, err := msclient.NewKanbanClient(ctx)
    if err != nil {
	    log.Fatalf("failde to construct kanban client: %v", err)
    }

    msName := "sample service"

    // kanban channel を生成
    kanbanCh, err := client.GetKanbanCh(msName, client.GetProcessNumber())
    if err != nil {
	    log.Fatalf("failed to get kanban channel: %v", err)
    }

    // kanban channel を監視
    for {
	    select {
	    case k := <-kanbanCh:
		    metadata, err := k.GetMetadataByMap()
		    if err != nil {
			    log.Printf("failed to get metadatas: %v", err)
		    }
		    log.Print(metadata)
	    }
    }
}
```