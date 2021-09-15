# aion-core

aion-coreはマイクロサービスプラットフォームであるAIONを動作させるのに必要なオープンソースシステムです。

aion-coreはAIONプラットフォーム上でマイクロサービスを動作させるために、以下のリソースを提供しています。

* AIONのメインコンポーネント
* マイクロサービスで利用するライブラリ
* kubernetesへのデプロイに必要な設定ファイル

また、aion-coreは単体のマシンで動作するシングルモードと、複数のマシン間でクラスタ構成をとるクラスタモードの２通りでの動作が可能です。

**目次**

* [動作環境](#動作環境)
    * [前提条件](#前提条件)
* [AIONで利用しているミドルウェア群](#AIONで利用しているミドルウェア群)
    * [Redis](#redis)
    * [Mongo DB](#mongo-db)
    * [mysql](#mysql)
* [マイクロサービス構成の例](#マイクロサービス構成の例)
* [シングルモードとクラスタモード](#シングルモードとクラスタモード)
    * [シングルモード](#シングルモード)
    * [クラスタモード](#クラスタモード)
* [セットアップ](#セットアップ)
    * [hostnameの設定](#hostnameの設定)
    * [ディレクトリの作成](#ディレクトリの作成)
    * [kubernetesのインストール](#1.kubernetesのインストール)
    * [AIONのセットアップ](#AIONのセットアップ)
    * [aion-core-manifestsの配置](#aion-core-manifestsの配置)
    * [project.ymlの設定](#project.ymlの設定)
* [Master nodeの構築(シングルモード/クラスタモードのMaster)](#Master-nodeの構築)
    * [1.Kubeadmでセットアップ](#1kubeadmでセットアップ)
    * [2.Flannelのコンテナをデプロイする](#2flannelのコンテナをデプロイする)
    * [3.Master Nodeの隔離を無効にする](#3master-nodeの隔離を無効にする)
    * [4.Master Nodeがクラスターに参加していることを確認する](#4master-nodeがクラスターに参加していることを確認する)
    * [5.(クラスタモードのみ)aionctlのインストール](#aionctlのインストール)
* [Worker nodeの構築(クラスタモードのWorke)](#Worker-nodeの構築)
    * [1.ノードをワーカーノードとしてclusterに参加させる](#1ノードをワーカーノードとしてclusterに参加させる)
    * [2.secret情報をconfigに書き込む](#2secret情報をconfigに書き込む)
    * [3.参加したクラスターにaion-coreをdeploy](#3参加したクラスターにaion-coreをdeploy)
* [シングルモードでのAIONの起動と停止](#シングルモードでのAIONの起動と停止)
    * [defaultネームスペース](#defaultネームスペース)
    * [prjネームスペース](#prjネームスペース)
    * [AIONの起動](#aionの起動)
* [クラスタモードでのAIONの起動と停止](#クラスタモードでのAIONの起動と停止)
    * [aion-core-manifestのビルド](#aion-core-manifestのビルド)
    * [各manifestファイルを修正](#各manifestファイルを修正)
    * [起動](#起動)
    * [停止](#停止)
* [動作確認](#動作確認)

## 動作環境

### 前提条件

動作には以下の環境であることを前提とします。
また、aion-coreの動作にはKubernetesのインストールが必要です。

* OS: Linux
* CPU: Intel64/AMD64/ARM64

## AIONで利用しているミドルウェア群

AIONでは以下のデータベースを採用しております。 aion-coreと同時にkubernetes上に展開されます。

- Redis
- Mongo DB
- MySQL

### Redis

Redisは高速で永続化可能なインメモリデータベースです。AIONでは、主に以下の用途でRedisを利用しています。

* AIONプラットフォーム上で動作するマイクロサービス間のKanbanデータの受け渡し

* 各マイクロサービスで利用できるデータキャッシュサーバ

* フロントエンドで発生した動的データを保持

### Mongo DB

MongoDBはNoSQLの一種でドキュメント指向データベースと言われるDBです。スキーマレスでデータを保存し、永続化をサポートしています。 AIONでは、各マイクロサービスのLogをKanban
Replicatorを通して保存する役割を担っています。

### mysql

AIONでは、主にフロントエンドで発生した静的データが保持されます。mysqlを立ち上げる場合は[こちら](https://github.com/latonaio/mysql-kube) を参照してください。

### WebRTC

AIONでは、ブラウザで利用可能な API として、ビデオ、音声、および一般的なデータをリアルタイムにやり取りすることができます。

### gRPC

AIONでは、あるマイクロサービスからのリクエストに対して応答し、別のマイクロサービスへ送信することで、双方のマイクロサービスが通信をできるようにします。

### RabbitMQ

AIONでは、メッセージングアーキテクチャの一構成例として、RabbitMQを用いてキューを用いた非同期処理を行います。詳しくは[こちら](https://github.com/latonaio/rabbitmq-for-kubernetes)を参照してください。

## AION を用いたシステム構成の例

### AION の基本的な構成

AION がマイクロサービスの起動と通信を管理します。
Send Anything からリクエストを送り、他のデバイスと接続することで拡張を行なうことができます。
![マイクロサービス構成の例0](https://raw.githubusercontent.com/latonaio/aion-core/main/documents/aion-core-architecture.png)

### AION のメッセージングアーキテクチャの一例（RabbitMQ）

AION がマイクロサービスの起動を行い、マイクロサービス間の通信を RabbitMQ で管理します。
RabbitMQ での通信により長時間安定したシステムが実現されます。
さらに柔軟性の高さからシステムの拡張を容易に行うことができます。
(例えば、gRPCのような、より重厚なメッセージングアーキテクチャを採用する場合、もしくは、gRPCとRabbitMQを組み合わせる場合の方が適切なときもあります)
![マイクロサービス構成の例1](documents/aion-core-example1.drawio.png)

### AION のアーキテクチャの一例（WebRTC）

AION のフロントエンドにWebRTCを実装して、フロントエンド／ブラウザからバックエンドサービス等へ、ビデオ・音声など、任意のデータ入力を、リアルタイムに送信することができます。   
![マイクロサービス構成の例2](documents/aion-core-example2.png)

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

Kubernentsクラスターを構築するツールであるKubeadmを用いてセットアップを行います。

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

#### 5. os再起動

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
docker tag latonaio/envoy:latest localhost:31112/envoy:latest
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
### project.ymlの設定

aion-coreでは、マイクロサービスをデプロイするために、YAML形式の定義ファイルを作成する必要があります。


#### 配置

シングルモードで利用する場合は、以下のディレクトリに project.ymlを配置します。

```
project.ymlは/var/lib/aion/(namespace)/configの中に配置する。
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


## Master nodeの構築(シングルモード/クラスタモードのMaster)

### 1.Kubeadmでセットアップ

KubernentsのMaster Nodeのセットアップを行いますが、ホスト側のIPアドレスがKubernentesの設定ファイルに書き込まれるため、静的IPアドレスを設定しておくことをおすすめします。

```
sudo kubeadm init --pod-network-cidr=10.244.10.0/16
mkdir $HOME/.kube/
sudo cp /etc/kubernetes/admin.conf $HOME/.kube/config
sudo chown $(id -u):$(id -g) $HOME/.kube/config
```

※「apiserver-cert-extra-sans」オプションは外部サーバからkubectlで接続したい場合、接続元のIPアドレスを入力する項目になります（不要であればオプションごと削除して構いません）

### 2.Flannelのコンテナをデプロイする

ポッド間の通信を行うためのコンテナがクラスター上にデプロイされていないためCNIのネットワークアドオンをデプロイする。

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

## worker nodeの構築(クラスタモードのWorker)

### 1.ノードをワーカーノードとしてclusterに参加させる

```shell
# master nodeで下記のコマンドを実行
kubeadm token create --print-join-command
# 実行すると下記のコマンドが出るのでworker側で実行する
kubeadm join {マスターノードのIP}:6443 --token {token} --discovery-token-ca-cert-hash sha256:{hash値} 
```

※ worker node側でkubeadm initですでにclusterを立ち上げている場合は`sudo kubeadm reset`でclusterをリセットする

### 2.secret情報をconfigに書き込む

master nodeの`/etc/kubernetes/admin.conf`内の設定ファイルを、worker nodeの`~/.kube/config`にコピー

### 3.Master Nodeと共にnodeがクラスターに参加していることを確認する

下記のコマンドを実行し、master nodeと自分のnodeが表示され、StatusがREADYになっていれば完了です。

```
kubectl get node
```

## シングルモードでのAIONの起動と停止

各種起動/停止用のスクリプトは、aion-core-manifestの中にあります。

### defaultネームスペース

#### 起動

```shell
$ sh kubectl-apply.sh
```

#### 停止

```shell
$ sh kubectl-delete.sh
```

### prjネームスペース

#### 起動

```
$ kubectl apply -f generated/prj.yml
```

##### 停止

```
$ kubectl delete -f generated/prj.yml
```

## クラスタモードでのAIONの起動と停止

###  aion-core-manifestのビルド

```
$ make build-master HOST={masterのHOSTNAME}
$ make build-worker HOST={workerのHOSTNAME}
```

### 各manifestファイルを修正

#### project.ymlの各microserviceに対して、targetNodeパラメータを追加

```yaml
startup: no
ports: hoge
...
targetNode: {workerのHOSTNAME}
```

### mysql-kubeのdeployment.ymlに対してnamespaceとnodeSelectorを追加

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

### aion-core外部で実行している各サービスのmanifestに対してnamespace, nodeSelectorを追加

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

### 起動

```
$ make apply-master
$ make apply-worker HOST={workerのHOSTNAME}
```

### 停止

```
$ make delete-worker HOST={workersのHOSTNAME}
$ make delete-master
```

## 動作確認
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

