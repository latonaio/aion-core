# aion-core
aion-coreはAIONのプラットフォームにあるマイクロサービスを動作させるのに必要なオープンソースシステムです。
AIONのメインコンポーネント、マイクロサービスで利用するライブラリ、kubernetesのデプロイメントに必要なConfigなどを提供しております。

**目次**
- [マイクロサービス構成の例](#マイクロサービス構成の例)
- [動作環境](#動作環境)
    - [前提条件](#前提条件)
- [事前準備](#事前準備)
- [hostnameの設定](#hostnameの設定)
- [セットアップ](#セットアップ)
    - [ディレクトリ](#ディレクトリ)
    - [kubernetes](#kubernetes)
    - [aion-core](#aion-core)
    - [envoy](#envoy)
    - [project.yml](#project.yml)
        - [配置](#配置)        
        - [項目定義](#項目定義)  
- [AIONの起動と停止](#AIONの起動と停止)
    - [aion-core-manifests](#aion-core-manifests)
    - [default](#default)
        - [起動](#起動)
        - [停止](#停止)
            - [aion-coreのみを停止](#aion-coreのみを停止)
            - [aionを停止](#aionを停止)
    - [prjネームスペース](#prjネームスペース)
        - [起動](#起動)
            - [aion-coreのみを停止](#aion-coreのみを停止)
            - [aionを停止](#aionを停止)
- [AIONの起動確認](#AIONの起動確認)
            
## マイクロサービス構成の例
<a target="_blank" href="https://github.com/latonaio/aion-core/blob/main/documents/aion-core-architecture.png">
<img src="https://raw.githubusercontent.com/latonaio/aion-core/main/documents/aion-core-architecture.png" width="300">
</a>

## 動作環境
### 前提条件
動作には以下の環境であることを前提とします。
* Ubuntu OS
* ARM CPU搭載のデバイス

## 事前準備
* ネットワークIPアドレスの固定
* hostnameの設定

### hostnameの設定
AIONではLinuxの端末名を頼りに端末間通信を行うので
端末名を一台ごとに異なるものに変えておく必要があります。
端末名を変更する場合は以下のコマンドを実行。
```
hostnamectl set-hostname [new device name]
```
その後一度ターミナルを閉じ、開き直し、
以下のコマンドを実行したら変更できていることを確認します。
```
hostnamectl
```

## セットアップ
### ディレクトリ
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

### kubernetes
#### 1.Dockerをインストール&有効化
```
sudo apt install docker.io
sudo systemctl start docker
sudo systemctl enable docker
```

ログインユーザにDockerコマンドの実行権限を付与する必要があります。
権限を付与するには以下のコマンドを実行する。
```
sudo gpasswd -a $USER docker
sudo systemctl restart docker
```

#### 2.kubeadm、kubelet、kubectlをインストール
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

#### 3.Kubeadmでセットアップ
KubernentsのMaster Nodeのセットアップを行いますが、ホスト側のIPアドレスがKubernentesの設定ファイルに書き込まれるため、静的IPアドレスを設定しておくことをおすすめします。
```
sudo kubeadm init --pod-network-cidr=10.244.10.0/16
mkdir $HOME/.kube/
sudo cp /etc/kubernetes/admin.conf $HOME/.kube/config
sudo chown $(id -u):$(id -g) $HOME/.kube/config
```

※「apiserver-cert-extra-sans」オプションは外部サーバからkubectlで接続したい場合、接続元のIPアドレスを入力する項目になります（不要であればオプションごと削除して構いません）

#### 4.Flannelのコンテナをデプロイする
ポッド間の通信を行うためのコンテナがクラスター上にデプロイされていないためCNIのネットワークアドオンをデプロイする。
```
kubectl apply -f https://raw.githubusercontent.com/coreos/flannel/2140ac876ef134e0ed5af15c65e414cf26827915/Documentation/kube-flannel.yml
```

#### 5.Master Nodeの隔離を無効にする
```
kubectl taint nodes --all node-role.kubernetes.io/master-
```

※ デフォルトでは、マスターノードに対してSystem系以外のPodが配置されないよう設定されているため

#### 6.Master Nodeがクラスターに参加していることを確認する
下記のコマンドを実行し、NodeのStatusがReadyになっていればセットアップが完了です。
```
kubectl get node
```

### aion-core
```
echo 'export DOCKER_BUILDKIT=1' >> ~/.bashrc
```

daemon.jsonの内容を変更します。
```
sudo vi /etc/docker/daemon.json
// 以下の内容に変更する
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
reboot // Ubuntuを再起動する
```
```
cd $(hostname)/AionCore
git clone https://github.com/latonaio/aion-core.git
cd aion-core
docker login
make docker-build

```

#### pyhon-base-imagesのセットアップ

```
cd ..
```

pyhon-base-imagesのセットアップは<a href="https://github.com/latonaio/python-base-images">こちら</a>のREADMEを参照してください。

### envoy
```
docker login
docker pull envoyproxy/envoy:v1.16-latest
docker tag latonaio/envoy:latest localhost:31112/envoy:latest
```
※ v1.16以降はdockerのARCのタイプがlinux/arm64に対応している

### project.yml
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

## AIONの起動と停止
### aion-core-manifests
aion-core-manifestsリポジトリをGit Cloneします。
```
cd ~/$(hostname)/AionCore
git clone https://github.com/latonaio/aion-core-manifests.git
cd aion-core-manifests
```

### default
#### 起動
```
defaultネームスペースでaionを起動する
$ bash kubectl-apply.sh
```

#### 停止
##### aion-coreのみを停止
defaultネームスペースで起動しているaion-coreのみ停止する（サーバは停止しない）
```
$ bash kubectl-delete only-aion.sh
```
##### aionを停止
defaultネームスペースで起動しているaionを停止する
```
$ bash kubectl-delete.sh
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

