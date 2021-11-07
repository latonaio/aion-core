# aion-core

aion-core は、主にエッジコンピューティング向けの マイクロサービスアーキテクチャ・プラットフォームである AION を動作させるのに必要な オープンソースレポジトリ です。

aion-core は、以下のリソースを提供しています。

* AIONのメインコンポーネント
* AIONの関連ライブラリ等
* Kubernetesの初期設定・デプロイに必要な設定ファイル

aion-core の動作方法として、単体のマシンで動作するシングルモードと、複数のマシンでクラスタ構成をとるクラスタモードの、2通りの動作方法があります。  
シングルモードでは、エッジコンピューティング環境の単体のマシンに Kubernetes の Master Node のみが構成され動作します。  
クラスタモードでは、主にエッジコンピューティング環境の複数のマシンでにわたって、 Kubernetes の Master Node と Wokrer Nodes が構成され動作します。

**目次**

* [動作環境](#動作環境)
    * [前提条件](#前提条件)
* [AIONの概要](#AIONの概要)
* [AIONのアーキテクチャ](#AIONのアーキテクチャ)
* [AIONの主要構成](#AIONの主要構成)
    * [Service Broker](#Service-Broker)
    * [Status Kanban および Kanban Replicator](#Status-Kanban-および-Kanban-Replicator)
    * [Send Anything](#Send-Anything)
    * [その他](#その他)
* [AIONにおけるミドルウェアとフレームワーク](#AIONにおけるミドルウェアとフレームワーク)
    * [RabbitMQ](#RabbitMQ) 
    * [Fluentd](#Fluentd)  
    * [Redis](#redis)
    * [Envoy](#Envoy)  
    * [MongoDB](#mongodb)
    * [MySQL](#MySQL)
    * [WebRTC](#WebRTC)
    * [gRPC](#gRPC)
    * [ReactJS](#ReactJS)
* [AIONを用いたシステム構成の例](#AIONを用いたシステム構成の例)
    * [AION のメッセージングアーキテクチャ（RabbitMQ）](#AIONのメッセージングアーキテクチャ（RabbitMQ）)
    * [AION のアーキテクチャの一例（WebRTC）](#AIONのアーキテクチャの一例（WebRTC）)
* [AIONのランタイム環境](#AIONのランタイム環境)
* [シングルモードとクラスタモード](#シングルモードとクラスタモード)
    * [シングルモード](#シングルモード)
    * [クラスタモード](#クラスタモード)
* [セットアップ(シングルモード/クラスタモード共通)](#セットアップ(シングルモード/クラスタモード共通))
    * [hostnameの設定](#hostnameの設定)
    * [ディレクトリの作成](#ディレクトリの作成)
    * [kubernetesのインストール](#1.kubernetesのインストール)
    * [AIONのセットアップ](#AIONのセットアップ)
    * [aion-core-manifestsの配置](#aion-core-manifestsの配置)
    * [services.ymlの設定](#services.ymlの設定)
    * [aion-core-manifestsのビルド・修正(シングルモード/クラスタモードで異なります)](#aion-core-manifestsのビルド・修正(シングルモード/クラスタモードで異なります))
* [Master Nodeの構築(シングルモード/クラスタモードのMaster)](#Master-nodeの構築)
    * [1.Kubeadmでセットアップ](#1kubeadmでセットアップ)
    * [2.Flannelをデプロイする](#2flannelをデプロイする)
    * [3.Master Nodeの隔離を無効にする](#3master-nodeの隔離を無効にする)
    * [4.Master Nodeがクラスターに参加していることを確認する](#4master-nodeがクラスターに参加していることを確認する)
    * [5.(クラスタモードのみ)aionctlのインストール](#aionctlのインストール)
* [Worker Nodeの構築(クラスタモードのWorker)](#Worker-nodeの構築)
    * [1.ノードをワーカーノードとしてclusterに参加させる](#1ノードをワーカーノードとしてclusterに参加させる)
    * [2.secret情報をconfigに書き込む](#2secret情報をconfigに書き込む)
    * [3.参加したクラスターにaion-coreをdeploy](#3参加したクラスターにaion-coreをdeploy)
* [AIONの起動と停止（シングルモード/クラスタモード共通）](#AIONの起動と停止（シングルモード/クラスタモード共通）)
    * [起動](#起動)
    * [停止](#停止)
* [aion-core の起動と停止（シングルモード/クラスタモード共通）](#aion-core-の起動と停止（シングルモード/クラスタモード共通）)
    * [起動](#起動)
    * [停止](#停止)

## 動作環境
aion-coreならびに関連リソースならびにエッジアプリケーションやマイクロサービス等を、安定的に動作させるには、以下の環境であることを前提とします。  

* OS: Linux
* CPU: ARM/AMD/Intel  
* Memory: 8GB 以上推奨  
* Storage: 64GB 以上推奨 (OS領域とは別に主にコンテナイメージ実装・稼働のために必要です。通常のエッジ端末で64GBを確保するには、外付けMicroSDやSSDが必要です）   

## AIONの概要
AIONは、100% Linux のオープンソース環境をベースとして構築された、主にエッジコンピューティングのための、マイクロサービス志向のコンピューティング・プラットフォーム環境です。   
ほぼ全てのマイクロサービス、ミドルウェアがエッジ端末上でコンテナ化されており、エッジ端末上に構築された コンテナオーケストレーションシステムのKubernetesによって制御・監視されています。

## AIONのアーキテクチャ

![マイクロサービス構成の例0](documents/aion-core-architecture.png)

## AIONの主要構成  

AIONでは、主要構成として以下があります。 
Service Broker、Status Kanban および Kanban Replicator、Send Anything は、aion-core に含まれます。  

- Service Broker
- Status Kanban および Kanban Replicator
- Send Anything
- その他

### Service Broker
 
Service Brokerは、AION™のコア機能で、主にエッジコンテナオーケストレーション環境でのマイクロサービスの実行に関する統括制御をつかさどるモジュールです。  
AIONでは、Service Broker はそれ自体がマイクロサービスとして機能します。  
 
### Status Kanban および Kanban Replicator
 
Status Kanban および Kanban Replicatorは、それぞれAION™のコア機能の1つで、マイクロサービス間のかんばんのやりとりを制御します。AION™　にはカンバンロジックがあらかじめ含まれているため、コンピューティングリソースとストレージリソースが制限されたエッジで、1、10、または100ミリ秒のタイムサイクルでエンドポイントの高性能処理を実行できます。マイクロサービス（マイクロサービスA>マイクロサービスB>マイクロサービスCなど）の各連続処理に割り当てられたAまたは一部のカンバンカードは、エッジでのIoTおよびAI処理における大量の同時注文の一貫性とモデレーションを厳密に維持します。    
AIONでは、Status Kanban および Kanban Replicator は各々それ自体がマイクロサービスとして機能します。  

### Send Anything
 
Send Anythingは、エッジのAION™プラットフォームでソフトウェアのコアスタック専用に機能する統合カンバンネイティブデータ処理システムを提供します。  
Send Anything によるクロスデバイスかんばん処理システムは、AION™サービスブローカーによってオーケストレーションされ、多数のネットワークノード全体で、マイクロサービス指向アーキテクチャのデータ処理/インターフェースとアプリケーションのランタイムの柔軟なパターンを可能にします。  
AIONでは、Send Anything はそれ自体がマイクロサービスとして機能します。  

### その他
 
Data Sweeperは、マイクロサービスが生成した不要なファイルを定期的に削除する機能を提供します。これにより、ストレージリソースをクリーンアップして、エッジアプリケーションの実行時環境を安定かつ適度に保つことが可能になります。また、Data Sweeperはセキュリティブローカーとしても機能し、デバイス上の個人情報を自動的に消去することで、非常に安全なエッジ環境を確保し、個人のデータが外部に漏洩しないようにします。data-sweeper-kubeを立ち上げる場合は[こちら](https://github.com/latonaio/data-sweeper-kube)を参照してください。   
AIONでは、Data Sweeper はそれ自体がマイクロサービスとして機能します。  

## AIONにおけるミドルウェアとフレームワーク

AIONでは以下のミドルウェアとフレームワークを採用しております。 

- [RabbitMQ](https://github.com/latonaio/rabbitmq-on-kubernetes)
- [Fluentd](https://github.com/latonaio/fluentd-for-containers-mongodb-kube)    
- [Redis](https://github.com/latonaio/redis-cluster-kube)
- [Envoy](https://github.com/latonaio/envoy)
- [MongoDB](https://github.com/latonaio/mongodb-kube)
- [MySQL](https://github.com/latonaio/mysql-kube)
- [WebRTC](https://github.com/latonaio/webrtc)
- [gRPC](https://github.com/latonaio/grpc-io)
- [ReactJS](https://github.com/latonaio/react-js)

### RabbitMQ

AIONでは、AION がカンバンシステムと呼んでいる、マイクロサービス間のメッセージングアーキテクチャのコアアーキテクチャとして、RabbitMQ を採用しています。    
AION のカンバンシステムは、コンピューティングリソースとストレージリソースが制限されたエッジ環境で、1/10/100ミリ秒のタイムサイクルでエンドポイントにおけるマイクロサービス間の効率的・安定的処理をつかさどる、軽量なメッセージングアーキテクチャです。  
RabbitMQ について、詳しくは[こちら](https://github.com/latonaio/rabbitmq-on-kubernetes)を参照してください。  
AIONでは、RabbitMQ はマイクロサービスとして機能します。 

### Fluentd  

Fluentdは大量のログファイルを収集、解析し、ストレージに集約、保存を行うことができるオープンソースのデータコレクタです。  
AIONでは、Fluentdを用いてマイクロサービス単位で対象Podのログを監視し、必要なログをデータベースに保存します。  
AIONでは、Fluentd はマイクロサービスとして機能します。

### Redis

Redisは高速で永続化可能なインメモリデータベースです。AIONでは、主に以下の用途でRedisを利用しています。

* マイクロサービス間のメッセージデータの受け渡し

* マイクロサービスで常時利用可能なデータキャッシュ

* フロントエンドUIで発生した動的データを保持

AIONでは、Redis（RedisCluster）はマイクロサービスとして機能します。  

### Envoy

Envoy はマイクロサービス間のネットワーク制御をライブラリとしてではなく、ネットワークプロキシとして提供します。
AION ではネットワーク制御プロキシ、及びネットワークの負荷軽減を目的とするロードバランサーとして採用されています。

AIONでは、Envoy はマイクロサービスとして機能します。   

### MongoDB  

MongoDBはNoSQLの一種でドキュメント指向データベースと言われるDBです。スキーマレスでデータを保存し、永続化をサポートしています。 AIONでは、各マイクロサービスのLogをKanban
Replicatorを通して保存する役割を担っています。  
AIONでは、MongoDB はマイクロサービスとして機能します。  

### MySQL

AIONでは、主にフロントエンドUIで発生した静的データが保持されます。mysqlを立ち上げる場合は[こちら](https://github.com/latonaio/mysql-kube) を参照してください。  
AIONでは、MySQL はマイクロサービスとして機能します。  

### WebRTC

AIONでは、ブラウザで利用可能な API として、ビデオ、音声、および一般的なデータをリアルタイムにやり取りすることができます。

### gRPC

AIONでは、あるマイクロサービスからのリクエストに対して応答し、別のマイクロサービスへ送信することで、双方のマイクロサービスが通信をできるようにします。

### ReactJS

ReactJSは、ユーザインタフェース構築のためのJavaScriptライブラリです。   
AIONからのアウトプットをフロントエンドUIに表示したり、フロントエンドUIからの指示をバックエンド経由でAIONに伝えたりする役割を果たします。
ReactJSはコンポーネントベースで、大規模なJavaScriptコードを部品化させることで保守性を高めたり、既存のReactコンポーネントを再利用したりできるため、マイクロサービスアーキテクチャに適しています。

## AIONを用いたシステム構成の例

### AIONのメッセージングアーキテクチャ（RabbitMQ）

AION がマイクロサービスの起動を行い、マイクロサービス間の通信を RabbitMQ で管理します。
RabbitMQ での通信により長時間安定したシステムが実現されます。
さらに柔軟性の高さからシステムの拡張を容易に行うことができます。
(例えば、gRPCのような、より重厚なメッセージングアーキテクチャを採用する場合、もしくは、gRPCとRabbitMQを組み合わせる場合の方が適切なときもあります)
![マイクロサービス構成の例1](documents/aion-core-example1.drawio.png)

### AIONのアーキテクチャ（WebRTC）

AION のフロントエンドにWebRTCを実装して、フロントエンド／ブラウザからバックエンドサービス等へ、ビデオ・音声など、任意のデータ入力を、リアルタイムに送信することができます。   
![マイクロサービス構成の例2](documents/aion-core-example2.png)

## AIONのランタイム環境  

* AION-Core および data-sweeper-kube のランタイム環境は、[Golang](https://github.com/golang/go) で開発実装されています。  
* AION の 個別マイクロサービス等のランタイム環境は、[Golang](https://github.com/golang/go)、[Node.js](https://github.com/nodejs)、[Python](https://github.com/python)で開発実装されています。  
* AIONプラットフォームにおける 個別マイクロサービス等のランタイム環境として、上記以外の(または上記に加えて)任意のランタイム環境(例：[Rust](https://github.com/rust-lang)、C++、[Vue.js](https://github.com/vuejs))を選択肢として開発実装することができます。    
* AION では、例えば1つのエッジデバイス内などの、エッジコンピューティング環境等の制約されたリソース環境において、個別マイクロサービス等の要求仕様等に応じたプログラムの特性に合わせた、様々なランタイム環境を組み合わせて選択して開発実装することができます。  

## シングルモードとクラスタモード

### シングルモード

シングルモードでは、aion-coreはKubernetesのMaster node上に各種リソースおよびマイクロサービスがデプロイされます。   

シングルモードの特徴として、1つの端末上にマイクロサービスを展開し、aion-coreはそれらのサービスの起動または再起動、通信等を自動的に実行します。これにより、複数のマイクロサービスで構成されるシステムが実現できます。

### クラスタモード

クラスタモードでは、aion-coreはKubernetesのMaster node上にmaster-aionがデプロイされ、
Worker node上にworker-aionおよび各マイクロサービスがデプロイされます。

クラスタモードの特徴として、デプロイするマイクロサービスをWorker node単位で指定することができます。
デプロイ先の指示はmaster-aionから各worker-aionに対して振り分けられ、master-aion上でデプロイの状況などを見ることもできます。

## セットアップ(シングルモード/クラスタモード共通)

### hostnameの設定

AIONではLinuxの端末名を頼りに端末間通信を行うため、端末名を一台ごとに異なるものに変えておく必要があります。端末名を変更する場合は以下のコマンドを実行してください。

```
hostnamectl set-hostname [new device name]
```

その後一度ターミナルを閉じ、開き直し、 以下のコマンドを実行して端末名が変更されていることを確認します。

```
hostnamectl
```

### ディレクトリの作成

作業ファイル等を配置するディレクトリを作成します。

```
mkdir ~/$(hostname)
mkdir ~/$(hostname)/AionCore
mkdir ~/$(hostname)/DataSweeper
mkdir ~/$(hostname)/BackendService
mkdir ~/$(hostname)/Runtime
mkdir ~/$(hostname)/MysqlKube
mkdir ~/$(hostname)/UI
sudo mkdir -p /var/lib/aion
sudo mkdir -p /var/lib/aion/default/config
sudo mkdir -p /var/lib/aion/prj/config
sudo mkdir -p /var/lib/aion/Data
sudo mkdir -p /var/lib/aion/Data/deployment
sudo mkdir -p /var/lib/aion/prj/Data
```

### kubernetesのインストール

#### 1. Dockerをインストール&有効化

```
sudo apt install docker.io
sudo systemctl start docker
sudo systemctl enable docker
```

ログインユーザにDockerコマンドの実行権限を付与する必要があります。 権限を付与するには以下のコマンドを実行してください。

```
sudo gpasswd -a $USER docker
sudo systemctl restart docker
```

#### 2. kubeadm、kubelet、kubectlをインストール

Kubernetesクラスターを構築するツールであるKubeadmを用いてセットアップを行います。

```
sudo apt update && sudo apt install -y apt-transport-https curl
curl -s https://packages.cloud.google.com/apt/doc/apt-key.gpg | sudo apt-key add -
cat <<EOF | sudo tee /etc/apt/sources.list.d/kubernetes.list
deb https://apt.kubernetes.io/ kubernetes-xenial main
EOF
sudo apt update && sudo apt install -y kubelet kubeadm kubectl
sudo apt show kubelet kubeadm kubectl
```

#### 3. DOCKER_BUILDKITの環境変数を設定

```
echo 'export DOCKER_BUILDKIT=1' >> ~/.bashrc
```

#### 4. daemon.jsonの内容を変更

```shell
sudo vi /etc/docker/daemon.json
```

以下の内容に書き換え

```json
{
  "default-runtime": "nvidia",
  "runtimes": {
    "nvidia": {
      "path": "/usr/bin/nvidia-container-runtime",
      "runtimeArgs": []
    }
  },
  "features": {
    "buildkit": true
  }
}
```

#### 5. OS再起動

```shell
source ~/.bashrc
reboot 
```

### AIONのセットアップ

#### 1. aion-coreのbuild

```shell
cd $(hostname)/AionCore
git clone https://github.com/latonaio/aion-core.git
cd aion-core
docker login
make docker-build
cd ..
```

#### 2. pyhon-base-imagesのbuild

一部のマイクロサービスのDockerイメージには、以下のベースイメージが必要となります。

- latonaio/l4t
- latonaio/pylib-lite

pyhon-base-imagesのREADMEを参照し、これらのベースイメージを準備してください。

#### 3. envoyのdocker imageの用意

```
docker login
docker pull envoyproxy/envoy:v1.16-latest
```
#### 4. 各種マイクロサービスのbuild

AION上で動作させるためのマイクロサービスのDocker Imageを作成します。

クラスタモードで動作させる場合、デプロイ先のWorker Node上でそれぞれ個別にDocker Imageを作成する必要があります。


### aion-core-manifestsの配置

aion-coreをデプロイするためのマニフェストファイル群です。
クラスタモードで利用する場合は、master nodeのあるマシン上に配備してください。

```
cd ~/$(hostname)/AionCore
git clone https://github.com/latonaio/aion-core-manifests.git
cd aion-core-manifests
```
### services.ymlの設定

aion-coreでは、マイクロサービスをデプロイするために、YAML形式の定義ファイルを作成する必要があります。


#### 配置

シングルモードで利用する場合は、以下のディレクトリに services.ymlを配置します。

```
services.ymlは/var/lib/aion/(namespace)/configの中に配置する。
```

#### 項目定義

```
deviceName：自身のデバイス名。ここの値をdevices配下のマイクロサービスで環境変数DEVICE_NAMEとして参照できる

devices：通信相手となる端末の情報を記述する
devices.[device-name]：端末名。この下の階層にその端末の設定を記載する
devices.[device-name].addr：IPなど、その端末を参照できるアドレス
devices.[device-name].aionHome：AIONのホームディレクトリ

microservices：この端末で動かすマイクロサービスの情報を記述する
microservices.[service-name]：マイクロサービス名。この下の階層にそのサービスの設定を記載する

microservices.[service-name].startup：AIONの起動が完了したら、すぐ起動する（デフォルト:no）
microservices.[service-name].always：podが停止していたら自動で再起動する（デフォルト:no）
microservices.[service-name].env：この下の階層にKEY: VALUEで環境変数を定義することができる
microservices.[service-name].nextService：この下の階層に次のサービス一覧を記述する
microservices.[service-name].scale：同時起動数（デフォルト:1）
microservices.[service-name].privileged：Dockerの特権モードで動作させる
microservices.[service-name].serviceAccount：Kubernetesのサービスアカウントを付与する
microservices.[service-name].volumeMountPathList：この下の階層に、追加でマウントする一覧を記載する
microservices.[service-name].withoutKanban：カンバンを使用するかどうか
microservices.[service-name].targetNode：nodeをworker nodeとして運用する際のnode名
```

例）

```
  kube-etcd-sentinel:
    startup: yes
    always: yes
    withoutKanban: yes
    serviceAccount: controller-serviceaccount
    env:
      MY_NODE_NAME: YOUR_DEVICE_NAME
    targetNode: YOUR_NODE_NAME
```

##  aion-core-manifest のビルド・修正（シングルモード/クラスタモードで異なります）

###  aion-core-manifestのビルド（シングルモード）

```
make build
```

###  aion-core-manifestのビルド（クラスタモード）

```
$ make build-master HOST={masterのHOSTNAME}
$ make build-worker HOST={workerのHOSTNAME}
```

### 各manifestファイルを修正（クラスタモード）

#### project.ymlの各microserviceに対して、targetNodeパラメータを追加

```yaml
startup: no
ports: hoge
...
targetNode: {workerのHOSTNAME}
```

#### mysql-kubeのdeployment.ymlに対してnamespaceとnodeSelectorを追加

```yaml
metadata:
  namespace: {workerのHOSTNAME}

template:
  metadata:
    labels:
      app: hoge
  spec:
    containers:
    ...
    spec.template.spec.nodeSelect:
      kubernetes.io/hostname: {workerのHOSTNAME}
```  

#### aion-core外部で実行している各サービスのmanifestに対してnamespace, nodeSelectorを追加

volume mountで、ディレクトリパスなどの変更が必要な場合は、合わせて修正する

```yaml
metadata:
  namespace: {workerのHOSTNAME}

template:
  metadata:
    labels:
      app: hoge
  spec:
    containers:
    ...
    spec.template.spec.nodeSelect:
      kubernetes.io/hostname: {workerのHOSTNAME}
```

#### 起動

```
$ make apply-master
$ make apply-worker HOST={workerのHOSTNAME}
```

#### 停止

```
$ make delete-worker HOST={workersのHOSTNAME}
$ make delete-master
```

#### 動作確認
```
aion-coreが正常に動作しているか確認するには、以下のコマンドを実行する
$ kubectl get pod
または
$ kubectl get pod -n prj

以下の名前を含むpodが起動すればaion-coreは動作している
aion-servicebroker : マイクロサービスの呼び出しを管理する
aion-statuskanban : マイクロサービス間のデータ(看板)受け渡しを管理する
aion-sendanything : 端末間のデータ(看板)受け渡しを管理する
aion-kanban-replicator : 処理が終わった看板をログとしてmongodbに保管する
mongo : MongoDBサーバ
redis-cluster : Redisサーバ

その後、任意のマイクロサービスが起動しているかを確認する
```

## Master Nodeの構築（シングルモード/クラスタモードのMaster）

### 1.Kubeadmでセットアップ

KubernetesのMaster Nodeのセットアップを行いますが、ホスト側のIPアドレスがKubernentesの設定ファイルに書き込まれるため、静的IPアドレスを設定しておくことをおすすめします。

```
sudo kubeadm init --pod-network-cidr=10.244.10.0/16
mkdir $HOME/.kube/
sudo cp /etc/kubernetes/admin.conf $HOME/.kube/config
sudo chown $(id -u):$(id -g) $HOME/.kube/config
```

※「apiserver-cert-extra-sans」オプションは外部サーバからkubectlで接続したい場合、接続元のIPアドレスを入力する項目になります（不要であればオプションごと削除して構いません）

### 2.Flannelをデプロイする

Kubernetesにレイヤー3通信を実装するために、また、ポッド間の通信を行うために、Flannelをデプロイします。  

```
kubectl apply -f https://raw.githubusercontent.com/coreos/flannel/2140ac876ef134e0ed5af15c65e414cf26827915/Documentation/kube-flannel.yml
```

### 3.Master Nodeの隔離を無効にする

```
kubectl taint nodes --all node-role.kubernetes.io/master-
```

※ デフォルトでは、マスターノードに対してSystem系以外のPodが配置されないよう設定されているため

### 4.Master Nodeがクラスターに参加していることを確認する

下記のコマンドを実行し、NodeのStatusがReadyになっていればセットアップが完了です。

```
kubectl get node
```

### 5.(クラスタモードのみ)aionctlのインストール

```
cd /path/to/aion-core/
go install cmd/aionctl/main.go

```

## Worker Nodeの構築（クラスタモードのWorker）

### 1.ノードをワーカーノードとしてclusterに参加させる

```shell
# master nodeで下記のコマンドを実行
kubeadm token create --print-join-command
# 実行すると下記のコマンドが出るのでworker側で実行する
kubeadm join {マスターノードのIP}:6443 --token {token} --discovery-token-ca-cert-hash sha256:{hash値} 
```

※ Worker Node側でkubeadm initですでにclusterを立ち上げている場合は`sudo kubeadm reset`でclusterをリセットする

### 2.secret情報をconfigに書き込む

master nodeの`/etc/kubernetes/admin.conf`内の設定ファイルを、worker nodeの`~/.kube/config`にコピー

### 3.Master Nodeと共にnodeがクラスターに参加していることを確認する

下記のコマンドを実行し、master nodeと自分のnodeが表示され、StatusがREADYになっていれば完了です。

```
kubectl get node
```

## AIONの起動と停止（シングルモード/クラスタモード共通）

aion-core およびAION 稼働に必要なリソースをまとめて起動、停止します。

aion-core には、Service Broker, Kanban Server, Kanban Replicator, Send Anything が含まれます。
AION 稼働に必要なリソースには、Envoy, Redis, MongoDB などが含まれます。

以下の各Shellスクリプトは、aion-core-manifests の中にあります。


#### 起動

```shell
$ sh aion-start.sh
```

#### 停止

```shell
$ sh aion-stop.sh
```

## aion-core の起動と停止（シングルモード/クラスタモード共通）

aion-core を単体で起動、停止します。


#### 起動
```
$ sh aion-core-start.sh
```
#### 停止
```
$ sh aion-core-stop.sh
```
