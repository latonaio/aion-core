# mysql-kube
mysql-kube は、Kubernetes 上で MariaDB(MySQL) の Pod を立ち上げ稼働させるための マイクロサービス です。    
本リポジトリには、必要なマニフェストファイルが入っています。  
また、本リポジトリには、MySQLの初期設定と、Pod立ち上げ後のテーブルの作成に関する手順が含まれています。  
AIONでは、MySQLは主に、エッジアプリケーションで発生した静的なデータを保持・維持するために用いられます。  

# 動作環境

* OS: Linux OS  

* CPU: ARM/AMD/Intel  

* Kubernetes  


## mysql-kube を用いたエッジコンピューティングアーキテクチャ(OMOTE-Bakoアプリケーションの例)  
mysql-kube は、下記の黄色い枠の部分のリソースです。  
![mysql_omotebako](docs/omotebako_architecture.drowio.png)  

# MySQL の Initial Setup  
以下の手順でMySQLのPodを立ち上げます。  

[1] mysql_init内に初期データ挿入用のSQLファイルを配置してください  

```
$ mkdir mysql_init
$ cp 初期データ挿入用のSQLファイルパス mysql_init/
```

[2] 以下コマンドを実行してください  

```
$ make install-default PV_SIZE=1Gi USER_NAME=${MYSQL_USER} USER_PASSWORD=${MYSQL_PASSWORD}
PV_SIZE 任意のストレージサイズ
MYSQL_USER: 任意の「MySQLユーザ名」
MYSQL_PASSWORD: 任意の「MySQLパスワード」
```

[3] 以下コマンドでMySQLのPodが正常に起動している事を確認してください  

```
$ kubectl get po | grep mysql
```
# MySQL 立上げ・稼働 のための Kubernetes マニフェストファイル の設定
MySQL の Initial Setup により、以下の通りにマニフェストファイルが作成されます。

* ポート: 3306   
* コンテナイメージ: mariadb:10.6   
* volumeのマウント場所 
	* **persistentVolume**:
		* コンテナ: /var/lib/mysql
		* hostOS: /mnt/mysql_data
	* **initdb**:   
		* コンテナ: /docker-entrypoint-initdb.d
		* hostOS: /mnt/mysql_init
* タイムゾーン: Asia/Tokyo   

# MySQL における アプリケーション の コアテーブル の作成
MySQLデータベースに、アプリケーションのコアテーブルを作成します。  
例えば、OMOTE-Bakoアプリケーションのコアテーブル（＝主に ui-backend-for-omotebako の稼働に必要なコアテーブル）を作成する場合、以下のコマンドになります。  
```
$ kubectl exec -i <mysql-pods> -- /bin/sh -c "mysql -u <username> -p<password> --default-character-set=utf8 -D Omotebako" < ./sql/ui-backend-for-omotebako.sql
```
`<mysql-pods>`、`<username>`および`<password>`はセットアップ環境に合わせて変えること  

# MySQL における アプリケーション の 追加テーブル の作成    
MySQLデータベースに、アプリケーションの追加テーブルを作成します。  
例えば、calendar-module-kube の稼働に必要なカレンダーテーブルを追加する場合、以下のコマンドになります。

```
$ cd /path/to/calendar-module-kube-sql
$ kubectl exec -i <mysql-pods> -- /bin/sh -c "mysql -u <username> -p<password> --default-character-set=utf8" < ./calendar-module-kube-sql.sql
```
`<mysql-pods>`、`<username>`および`<password>`はセットアップ環境に合わせて変えること


# MariaDB について
エッジ環境はスペックの制限があるため、機能性とパフォーマンスのバランスに優れているMariaDB(MySQL)を採用しています。   
RDBMSにはSQLite、SQL ServerやPostgreSQLなどがあります。 

* SQLite: 軽量で手頃だが、大規模なシステムでは機能不十分  
* PostgreSQL: 高性能だが、処理コストが高い  

MariaDB(MySQL)はSQL ServerやPostgreSQLの中間に位置し、高速で実用性が高いため、LatonaおよびAIONではエッジ環境で採用されています。   

以下、MariaDBの特徴です。   

## MariaDB とは
MariaDBはMySQLから派生したもので、MySQLと高い互換性があります。   

《MySQLとMariaDBの違い》

|    |MariaDB|MySQL|   
|:---|:---|:---|    
|ライセンス|オープンソース(GPL)|オープンソース(プロプライエタリ・ライセンス)|   
|管理|コミュニティによる管理|Oracle社によるベンダー管理|   
|シェア|Linuxディストリビューションでの採用など急速に伸びている|非常に高い|   
|セキュリティ（暗号化機能）|暗号化の対象が多い|暗号化は限られている|   
|パフォーマンス|高い|MariaDBには劣る|   
|堅牢性|高い|普通|   
|クラスター構成|対応|非対応|   


## リレーショナル・データベース
MariaDB(MySQL)は、リレーショナルデータベースです。
リレーショナルデータベースとは、データベース(DB)におけるデータを扱う方法の1つで、主に2つの特徴があります。   

1. データは2次元(行×列)の表(テーブル)形式で表現   
2. 「キー」を利用して、複数の表を結合(リレーション)して利用可能   

データを2次元の表に分割し、また複数の表を様々な手法で結合して使うことで、複雑なデータを柔軟に扱うことができます。   

## 高い拡張性・柔軟性・速度
MariaDB(MySQL)の利点は以下の通りです。 

* システム規模が大きくなっても対応できる拡張性   
* さまざまなテーブルタイプのデータを統合できる柔軟性   
* 大規模なデータにも耐えうるような高速動作   
* データを保護するためのセキュリティ機能（データベースにアクセスするためのアクセス制限、盗み見防止のデータ暗号機能、Webサイトなどを安全に接続するためのセキュリティ技術など）   

## トランザクションとロールバック
トランザクションとはDBシステムで実行される処理のまとまり、または作業単位のことです。   
トランザクションを使うと複数のクエリをまとめて１つの処理として扱うことができます。   

* **処理の途中でエラーになって処理を取り消したいような場合**：「ロールバック; roll back」をすることで、そのトランザクションによる痕跡を消去してデータベースを一貫した状態（そのトランザクションを開始する前の状態）にリストアできます。   
* **あるトランザクションの全操作が完了した場合**：そのトランザクションはシステムによって「コミット; commit」され、DBに加えられた更新内容が恒久的なものとなります。コミットされたトランザクションがロールバックされることはありません。    

MariaDB(MySQL)では、トランザクション処理を行うことで、MariaDB(MySQL)のテーブルにデータを保存などをする際に、他のユーザーからのアクセスを出来ないようにテーブルをロックしています。

## MySQL Workbench
MySQL Workbenchとは、MySQLの公式サイトにてMySQL Serverと共に無料で配布されている、
データ・モデリング、SQL 開発、およびサーバー設定、ユーザー管理、バックアップなどの包括的な管理ツールのことです。
コマンドラインではなくビジュアル操作（GUI）に対応しています。
MySQL Workbench は Windows、Linux、Mac OS X で利用可能です。

* **データベース設計**：新規にER図が作成できるほか、既存のデータベースからER図を逆に生成することも可能です。   
* **データベース開発**：SQLのクエリ作成や実行だけでなく、最適化を直感的な操作で行えるビジュアル表示に対応しています。さらにSQLエディタにはカラーハイライト表示や自動補完機能のほか、SQLの実行履歴表示やSQLステートメントの再利用、オブジェクトブラウザにも対応しており、SQLエディタとしてもとても優秀な開発ツールです。   
* **データベース管理**：ヴィジュアルコンソールによってデータベースの可視性が高められており、MySQLの管理をより容易にする工夫が凝らされています。さらにビジュアル・パフォーマンス・ダッシュボードの実装により、パフォーマンス指標を一目で確認できます。   