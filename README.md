# AION-CoreのREADME作成

**目次**
- [概要](#概要)
- [動作環境](#動作環境)
    - [前提条件](#1.前提条件)
    - [事前準備](#2.事前準備)
- [インストール](#インストール)
    - [ディレクトリを作成](#ディレクトリを作成)
    - [Golang](#Golang)
    - [kubernetes](#kubernetes)
        - [MasterNode](#MasterNode)
            - [1.JetsonのSwap機能をOFFにする](#1.JetsonのSwap機能をOFFにする)
            - [2.Dockerをインストール&有効化する](#2.Dockerをインストール&有効化する)
            - [3.kubeadm、kubelet、kubectlをインストールする](#3.kubeadm、kubelet、kubectlをインストールする)
            - [4.Kubeadmでセットアップを行う](#4.Kubeadmでセットアップを行う)
            - [5.Flannelのコンテナをデプロイする](#5.Flannelのコンテナをデプロイする)
    - [aion-core](#aion-core)
    - [envoy](#envoy)
    - [project.ymlの配置](#project.ymlの配置)
    - [AIONの起動、停止](#AIONの起動、停止)

## 概要
aion-coreはAIONTMのプラットフォームにあるマイクロサービスを動作させるのに必要なオープンソースシステムです。
AIONTMのメインコンポーネント、マイクロサービスで利用するライブラリ、kubernetesのデプロイメントに必要なConfigなどを提供しております。

## 動作環境
### 1.前提条件
動作には以下の環境であることを前提とします。
* Ubuntu OS
* ARM CPU搭載のデバイス

### 2.事前準備
実行環境に以下のソフトウェアがインストールされている事を前提とします。
* kubernetesのインストール
* envoyのインストール
* project-yamlsのインストール
* aion-core-manifestsのインストール

https://github.com/microsoftgraph/ruby-connect-rest-sample/blob/master/README-Localized/README-ja-jp.md#%E5%89%8D%E6%8F%90%E6%9D%A1%E4%BB%B6

```
AIONではLinuxの端末名を頼りに端末間通信を行う
端末名を一台ごとに異なるものに変えておく必要がある
端末名を変更する場合は以下のコマンドを実行
hostnamectl set-hostname [new device name]
その後一度ターミナルを閉じ、開き直す
以下のコマンドを実行し、変更できていることを確認する
hostnamectl
```

## インストール
### ディレクトリを作成
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

##### 2.Dockerをインストール&有効化する
```
sudo apt install docker.io
sudo systemctl start docker
sudo systemctl enable docker
```

```
ログインユーザにDockerコマンドの実行権限を付与する必要がある
権限を付与するには以下のコマンドを実行する
sudo gpasswd -a $USER docker
sudo systemctl restart docker
```

##### 3.kubeadm、kubelet、kubectlをインストールする
今回は、Kubernentsクラスターを構築するツールであるKubeadmを用いてセットアップを行う。
```
sudo apt update && sudo apt install -y apt-transport-https curl
curl -s https://packages.cloud.google.com/apt/doc/apt-key.gpg | sudo apt-key add -
cat <<EOF | sudo tee /etc/apt/sources.list.d/kubernetes.list
deb https://apt.kubernetes.io/ kubernetes-xenial main
EOF
sudo apt update && sudo apt install -y kubelet kubeadm kubectl
sudo apt show kubelet kubeadm kubectl
```

##### 4.Kubeadmでセットアップを行う
ここでKubernentsのMaster Nodeのセットアップを行うが、ここでホスト側のIPアドレスがKubernentesの設定ファイルに書き込まれるため、静的IPアドレスを設定しておくことをおすすめします。
```
sudo kubeadm init --pod-network-cidr=10.244.10.0/16
mkdir $HOME/.kube/
sudo cp /etc/kubernetes/admin.conf $HOME/.kube/config
sudo chown $(id -u):$(id -g) $HOME/.kube/config
```

※「apiserver-cert-extra-sans」オプションは外部サーバからkubectlで接続したい場合、接続元のIPアドレスを入力する項目になります（不要であればオプションごと削除して構いません）

##### 5.Flannelのコンテナをデプロイする
ポッド間の通信を行うための、コンテナがクラスター上にデプロイされていないためCNIのネットワークアドオンをデプロイする。
```
kubectl apply -f https://raw.githubusercontent.com/coreos/flannel/2140ac876ef134e0ed5af15c65e414cf26827915/Documentation/kube-flannel.yml
```

##### 6.Master Nodeの隔離を無効にする
```
kubectl taint nodes --all node-role.kubernetes.io/master-
```

※デフォルトでは、マスターノードに対してSystem系以外のPodが配置されないよう設定されているため

##### 7.Master Nodeがクラスターに参加していることを確認する
下記のコマンドを実行し、NodeのStatusがReadyになっていればセットアップが完了
```
kubectl get node
```

### aion-core
```
echo 'export DOCKER_BUILDKIT=1' >> ~/.bashrc
sudo vi /etc/docker/daemon.json
以下の内容に変更する　（※i押す前ならddで１行ずつ消せる）
=====
{
    "default-runtime": "nvidia",
    "runtimes": {
        "nvidia": {
            "path": "/usr/bin/nvidia-container-runtime",
            "runtimeArgs": []
        }
    },
    "features": { "buildkit": true }
}
=====
source ~/.bashrc
reboot
⇒Ubuntuを再起動する
```
```
cd $(hostname)/AionCore
git clone https://github.com/latonaio/aion-core.git
cd aion-core
docker login
make docker-build

cd ..
git clone https://github.com/latonaio/python-base-images.git
cd python-base-images
make docker-build-pylib-lite
make docker-build-l4t
```

### envoy
```
docker login
docker pull latonaio/envoy:latest
docker tag latonaio/envoy:latest localhost:31112/envoy:latest
```

### project.ymlの配置
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
microservices.[service-name].withoutKanban：
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
```

project.ymlの配置
```
project.ymlは/var/lib/aion/(namespace)/configの中に配置する。
```

## AIONの起動、停止
```
aion-core-manifestsリポジトリをクローンする
cd ~/$(hostname)/AionCore
git clone https://github.com/latonaio/aion-core-manifests.git
cd aion-core-manifests
```

```
用途に応じていずれかのコマンドを実行する
defaultネームスペースでaionを起動する
$ bash kubectl-apply.sh
defaultネームスペースで起動しているaion-coreのみ停止する（サーバは停止しない）
$ bash kubectl-delete only-aion.sh
defaultネームスペースで起動しているaionを停止する
$ bash kubectl-delete.sh
 
prjネームスペースでaionを起動する
$ kubectl apply -f generated/prj.yml
prjネームスペースで起動しているaion-coreのみ停止する（サーバは停止しない）
$ bash kubectl-delete only-aion-prj.sh
prjネームスペースで起動しているaionを停止する（prjネームスペースごと）
$ kubectl delete -f generated/prj.yml
```

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