# aion-core

aion-coreはAIONのプラットフォームにあるマイクロサービスを動作させるのに必要なオープンソースシステムです。

AIONのメインコンポーネント、マイクロサービスで利用するライブラリ、kubernetesのデプロイメントに必要なConfigなどを提供しております。

aion-coreは単体nodeでのdeployと、worker nodとしてのdeployの2通りのdeployが可能です。

**目次**

* [aion-core](#aion-core)
    * [マイクロサービス構成の例](#マイクロサービス構成の例)
    * [cluster構成](#cluster構成)
    * [動作環境](#動作環境)
        * [前提条件](#前提条件)
    * [OS側の事前準備](#os側の事前準備)
        * [hostnameの設定](#hostnameの設定)
    * [Databaseについて](#databaseについて)
        * [Redis](#redis)
        * [Mongo DB](#mongo-db)
        * [mysql](#mysql)
    * [セットアップ(master/worker共通)](#セットアップmasterworker共通)
        * [ディレクトリ](#ディレクトリ)
        * [1.kubernetes](#1kubernetes)
            * [a.Dockerをインストール&amp;有効化](#adockerをインストール有効化)
            * [b.kubeadm、kubelet、kubectlをインストール](#bkubeadmkubeletkubectlをインストール)
        * [2.AION](#2aion)
            * [a.DOCKER_BUILDKITの環境変数を設定](#adocker_buildkitの環境変数を設定)
            * [b.daemon.jsonの内容を変更](#bdaemonjsonの内容を変更)
            * [c.os再起動](#cos再起動)
            * [d.aion-coreのbuild](#daion-coreのbuild)
        * [3. pyhon-base-imagesのセットアップ](#3-pyhon-base-imagesのセットアップ)
        * [4.project.ymlの設定](#4projectymlの設定)
            * [配置](#配置)
            * [項目定義](#項目定義)
        * [5.envoyのdocker imageを準備](#5envoyのdocker-imageを準備)
        * [6.aion-core-manifestsの配置](#6aion-core-manifestsの配置)
    * [master nodeのdeploy](#master-nodeのdeploy)
        * [1.Kubeadmでセットアップ](#1kubeadmでセットアップ)
        * [2.Flannelのコンテナをデプロイする](#2flannelのコンテナをデプロイする)
        * [3.Master Nodeの隔離を無効にする](#3master-nodeの隔離を無効にする)
        * [4.Master Nodeがクラスターに参加していることを確認する](#4master-nodeがクラスターに参加していることを確認する)
    * [worker nodeのdeploy(複数node構成にしない場合は飛ばして可)](#worker-nodeのdeploy複数node構成にしない場合は飛ばして可)
        * [1.ノードをワーカーノードとしてclusterに参加させる](#1ノードをワーカーノードとしてclusterに参加させる)
        * [2.secret情報をconfigに書き込む](#2secret情報をconfigに書き込む)
        * [3.各manifestファイルを修正](#3各manifestファイルを修正)
            * [project.ymlの各microserviceに対して、targetNodeパラメータを追加](#projectymlの各microserviceに対してtargetnodeパラメータを追加)
            * [Aion-coreのmanifestに対してnodeSelectorを追加](#aion-coreのmanifestに対してnodeselectorを追加)
            * [mysql-kubeのdeployment.ymlに対してnodeSelectorを追加](#mysql-kubeのdeploymentymlに対してnodeselectorを追加)
            * [aion-core外部で実行している各サービスのmanifestに対してnodeSelectorを追加](#aion-core外部で実行している各サービスのmanifestに対してnodeselectorを追加)
        * [4.参加したクラスターにaion-coreをdeploy](#4参加したクラスターにaion-coreをdeploy)
    * [AIONの起動と停止(master/worker)](#aionの起動と停止masterworker)
        * [defaultネームスペース](#defaultネームスペース)
            * [起動](#起動)
            * [停止](#停止)
                * [aion-coreのみを停止](#aion-coreのみを停止)
                * [aion全体を停止](#aion全体を停止)
        * [prjネームスペース](#prjネームスペース)
            * [起動](#起動-1)
                * [aion-coreのみを停止](#aion-coreのみを停止-1)
                * [aionを停止](#aionを停止)
        * [AIONの起動](#aionの起動)

## マイクロサービス構成の例

![マイクロサービス構成の例](https://raw.githubusercontent.com/latonaio/aion-core/main/documents/aion-core-architecture.png)

## cluster構成

単体nodeでclusterを運用する場合、aion-coreと各microserviceはmaster nodeに対してdeployされることになります。

一方で、aion-coreをdeployした複数のエッジ端末でclusterを構成するようなケースの場合、master nodeとしてkubernetesを構築したエッジ端末に対して、その他のエッジ端末をworker
nodeとして紐づけて、単一のclusterで管理することが可能です。

master nodeとworker nodeの構築手順は一部異なりますが、OS,端末サイドのセットアップ手順や必要なリソースは基本的に共通です。

## 動作環境

### 前提条件

動作には以下の環境であることを前提とします。

* OS: Linux
* CPU: Intel64/AMD64/ARM64

## OS側の事前準備

* ネットワークIPアドレスの固定
* hostnameの設定

### hostnameの設定

AIONではLinuxの端末名を頼りに端末間通信を行うので、端末名を一台ごとに異なるものに変えておく必要があります。端末名を変更する場合は以下のコマンドを実行してください。

```
hostnamectl set-hostname [new device name]
```

その後一度ターミナルを閉じ、開き直し、 以下のコマンドを実行して端末名が変更されていることを確認します。

```
hostnamectl
```

## Databaseについて

AIONでは以下のデータベースを採用しております。 aion-coreと同時にkubernetes上に展開されます。

- Redis
- Mongo DB

### Redis

Redisは高速で永続化可能なインメモリデータベースです。AIONでは、主に以下の用途でRedisを利用しています。

* AIONプラットフォーム上で動作するマイクロサービス間のKanbanデータの受け渡し

* 各マイクロサービスで利用できるデータキャッシュサーバ

### Mongo DB

MongoDBはNoSQLの一種でドキュメント指向データベースと言われるDBです。スキーマレスでデータを保存し、永続化をサポートしています。 AIONでは、各マイクロサービスのLogをKanban
Replicatorを通して保存する役割を担っています。

### mysql

AIONではmysqlデータベースを使用する場合は別途デプロイが必要です。 mysqlを立ち上げる場合は[こちら](https://github.com/latonaio/mysql-kube) を参照してください。

## セットアップ(master/worker共通)

### ディレクトリ

はじめに作業ファイル等を配置するディレクトリを作成します。

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

### 1.kubernetes

#### a.Dockerをインストール&有効化

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

#### b.kubeadm、kubelet、kubectlをインストール

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

### 2.AION

#### a.DOCKER_BUILDKITの環境変数を設定

```
echo 'export DOCKER_BUILDKIT=1' >> ~/.bashrc
```

#### b.daemon.jsonの内容を変更

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

#### c.os再起動

```shell
source ~/.bashrc
reboot 
```

#### d.aion-coreのbuild

```shell
cd $(hostname)/AionCore
git clone https://github.com/latonaio/aion-core.git
cd aion-core
docker login
make docker-build
cd ..
```

### 3. pyhon-base-imagesのセットアップ

一部のマイクロサービスのDockerイメージには、以下のベースイメージが必要となります。

- latonaio/l4t
- latonaio/pylib-lite

pyhon-base-imagesのREADMEを参照し、これらのベースイメージを準備してください。

### 4.project.ymlの設定

#### 配置

project.ymlを配置します。

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

### 5.envoyのdocker imageを準備

```
docker login
docker pull envoyproxy/envoy:v1.16-latest
docker tag latonaio/envoy:latest localhost:31112/envoy:latest
```

### 6.aion-core-manifestsの配置

```
cd ~/$(hostname)/AionCore
git clone https://github.com/latonaio/aion-core-manifests.git
cd aion-core-manifests
```

## 単体nodeのdeploy

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

## worker nodeのdeploy(複数node構成にしない場合は飛ばして可)

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

### 3.各manifestファイルを修正

#### project.ymlの各microserviceに対して、targetNodeパラメータを追加

```yaml
startup: no
ports: hoge
...
targetNode: ＄{NODE_NAME}
```

#### Aion-coreのmanifestに対してnodeSelectorを追加

```shell
cd ~/$(hostname)/AionCore/aion-core-manifest
vi generated/deafult.yaml
```

```yaml
template:
  metadata:
    labels:
      app: hoge
  spec:
    containers:
    ...
    spec.template.spec.nodeSelect:
      kubernetes.io/hostname: ${NODE_NAME}
```

#### mysql-kubeのdeployment.ymlに対してnodeSelectorを追加

```yaml
template:
  metadata:
    labels:
      app: hoge
  spec:
    containers:
    ...
    spec.template.spec.nodeSelect:
      kubernetes.io/hostname: ${NODE_NAME}
```  

#### aion-core外部で実行している各サービスのmanifestに対してnodeSelectorを追加

```yaml
template:
  metadata:
    labels:
      app: hoge
  spec:
    containers:
    ...
    spec.template.spec.nodeSelect:
      kubernetes.io/hostname: ${NODE_NAME}
```

### 4.Master Nodeと共にnodeがクラスターに参加していることを確認する

下記のコマンドを実行し、master nodeと自分のnodeが表示され、StatusがREADYになっていれば完了です。

```
kubectl get node
```

## AIONの起動と停止(master/worker)

### defaultネームスペース

#### 起動

```shell
# NODE_NAMEはdeployするnodeの名前を指定（他のnodeの指定も可）
make apply-node NODE-NAME=${NODE_NAME}
```

#### 停止

##### aion-coreのみを停止

defaultネームスペースで起動しているaion-coreのみ停止する（サーバは停止しない）

```shell
$ bash kubectl-delete only-aion.sh
```

##### aion全体を停止

```shell
# NODE_NAMEはdeployするnodeの名前を指定（他のnodeの指定も可）
make delete-node NODE-NAME=${NODE_NAME}
```

### prjネームスペース

#### 起動

prjネームスペースでaionを起動する

```
$ kubectl apply -f generated/prj.yml
```

##### aion-coreのみを停止

prjネームスペースで起動しているaion-coreのみ停止する（サーバは停止しない）

```
$ bash kubectl-delete only-aion-prj.sh
```

##### aionを停止

prjネームスペースで起動しているaionを停止する（prjネームスペースごと）

```
$ kubectl delete -f generated/prj.yml
```

### AIONの起動

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

