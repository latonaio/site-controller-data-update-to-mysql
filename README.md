# site-controller-data-update-to-mysql
site-controller-data-update-to-mysqlは、エッジ端末からWindows端末のディレクトリを監視し、そのディレクトリに新しいサイトコントローラー連携ファイル（CSV）が作成された場合に、その情報をkanbanに渡す、エッジ端末上で動くマイクロサービスです。

## サイトコントローラーとは
サイトコントローラーとは、主に宿泊業向けに、予約情報をまとめて管理運用するサービスです。   
サイトコントローラーは、サービスオーナー各社により提供されております。   

## 概要
site-controller-data-update-to-mysqlは、 一定時間おきにWindows端末ディレクトリを走査し、新しいファイルが作成されたら、以下の処理を行います。   
監視するディレクトリは`services.yml`の`volumeMountPathList`で、走査間隔は`POLLING_INTERVAL`で指定します。   
   
1. 監視しているディレクトリに新しく作成されたファイルを、`/var/lib/aion/defalt/Data/`配下にコピーする。   
2. MySQLの`csv_upload_transaction`テーブルに、タイムスタンプ、1.でコピーしたファイルパス、status="before"を挿入する。 監視しているディレクトリに新しく作成されたファイルを、`/var/lib/aion/defalt/Data/`配下にコピーする。   
3. kanbanに、`omotebako-sc`のconnectionKeyで、タイムスタンプと1.でコピーしたファイルパスを送る。 MySQLの`csv_upload_transaction`テーブルに、タイムスタンプ、1.でコピーしたファイルパス、status="before"を挿入する。


## 動作環境
site-controller-data-update-to-mysqlは、aion-coreのプラットフォーム上での動作を前提としています。    
使用する際は、事前に下記の通りAIONの動作環境を用意してください。

* OS: Linux OS   
  
* CPU: ARM/AMD/Intel   
  
* Kubernetes   
  
* AION のリソース   

## セットアップ
### 1. エッジ端末でWindowsサーバの共有フォルダをマウント

#### 1.1. エッジ端末でWindows共有フォルダをマウント
1.1.1 Windows側で共有フォルダを作成
- 任意のディレクトリに共有フォルダ用のフォルダを新規作成する
- 作成したフォルダを右クリック→「プロパティ」→「共有タブ」を開く
- 「共有」ボタンを押し、プルダウンから「Everyone」を選択し、追加
- 「共有」を押し、「終了」を押す
- コマンドプロンプトを開き、`ipconfig`と入力
- IPv4 アドレスからwindowsサーバのIPアドレスを確認する（後で使う）

※ 参考URL   

https://bsj-k.com/win10-common-file/   

1.2.2 エッジ端末でWindowsの共有フォルダをマウント   

smbnetfsを使用します。マウントする際、ルートユーザーでないと挙動がおかしくなるため、ルートユーザーでマウント実行することを推奨します。   

```
# install smbnetfs
sudo apt install smbnetfs

# make folder
mkdir ~/.smb
cd ~/.smb

# copy default config files
cp /etc/samba/smb.conf .
cp /etc/smbnetfs.conf .

# make config files
echo "host {windowsIPアドレス} WORKGROUP visible=true" > smbnetfs.host
echo "auth {windowsユーザー名} {windowsユーザーパスワード}" > smbnetfs.auth

# drop file permission
chmod 600 smbnetfs.host smbnetfs.auth

# mount
mkdir /mnt/windows
sudo -i
smbnetfs /mnt/windows
```
マウントした後、共有フォルダがあることを確認します。デスクトップに共有フォルダを置いた場合、`/{windowsIPアドレス}/Users/{ユーザー名}/Desktop`にあると思われます。

※ マウントを解除したい場合   
```
# unmount for root user
umount /mnt/windows

# unmount for non-root users
fusermount -u test
```

※ 参考URL   

https://ruco.la/memo/tag/fuse   

### 2. Dockerイメージの作成   
```
cd site-controller-data-update-to-mysql
make docker-build
```

## 起動方法   
### デプロイ on AION   
`services.yml`に設定を記載し、AionCore経由でコンテナを起動します。   

例）   
```
  site-controller-data-update-to-mysql:
    scale: 1
    startup: yes
    always: yes
    env:
      KANBAN_ADDR: aion-statuskanban:10000
      MYSQL_USER: MYSQL_USER_XXX
      MYSQL_HOST: mysql
      MYSQL_PASSWORD: MYSQL_PASSWORD_XXX
      POLLING_INTERVAL: 5
      SITE_CONTOROLLER_NAME: XXX
      MOUNT_PATH: /mnt/windows/{共有フォルダへのパス}
    nextService:
      sc_csv:
        - name: omotebako-sc
    volumeMountPathList:
      - /mnt/windows:/mnt/windows/{共有フォルダへのパス}
```
{共有フォルダへのパス}の例（デスクトップにある場合）：`{windowsIPアドレス}/Users/{ユーザー名}/Desktop/{フォルダ名}`


## I/O   
kanbanのメタデータから下記の情報を入出力します。   

### input   
kanbanのメタデータ   

### output   
* outputPath: `/var/lib/aion/defalt/Data/`配下にコピーしたcsvのファイルパス
* timestamp:

## エラー対処法   
1. ホストのディレクトリがマウントできないと言われた時
```
Error: failed to start container "site-controller-data-update-to-mysql-001": Error response from daemon: error while creating mount source path '/mnt/windows': mkdir /mnt/windows: file exists
```
→ mac側の共有フォルダから削除   
→ 再度共有フォルダに追加・linuxでマウント   

## ローカルからデバッグする場合   
MYSQLの情報をコマンドライン引数で渡すことによって、別のローカル端末からエッジ端末のデータベースにアクセスすることができます。   

