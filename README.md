Aion core component
====

## リポジトリの概要
以下の機能を提供する
- AIONのメインコンポーネントの提供
- microserviceで利用するライブラリの提供(Python, Golang)
- kubernetesのデプロイメントConfigの提供

### ディレクトリ一覧
- cmd : aionコンポーネントのバイナリ実装
- config : プロジェクトファイル読み込み用モジュール
- internal : aion-core専用のモジュール
- pkg : 外部ライブラリとして呼び出すことが可能なモジュール
- proto : gRPCプロトコルバッファの定義ファイル
- python : マイクロサービスのPython用ライブラリ
- test : テスト用モジュール
- yaml : プロジェクトファイルのサンプルYaml

---
## AION core コンポーネント
### 一覧
- service-broker : マイクロサービスの管理および実行経路を制御
- kanban-server: マイクロサービス間で使用するカンバンのメッセージング処理を実施
- send-anything : リモートの端末間でファイルを転送

---

## Kubernetesの設定  
下記URLに記載されている「Statusが終了になっているPodが一定数以上になると削除する。」の項目を行う。  
https://latonaio.atlassian.net/wiki/spaces/TOYOB/pages/811859969/Kubernetes