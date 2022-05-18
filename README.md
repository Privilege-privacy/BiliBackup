# BiliBackup

自动下载 BiliBili 指定收藏夹内的视频，并上传到 OneDrive,Google Drive,WebDev,阿里云盘(WebDev).... 等可被 [Rclone](https://github.com/rclone/rclone) 挂载的云存储。

## 安装

----
1. 从 [Release](https://github.com/privileges-privacy/BiliBackup/releases) 下载相应版本的的压缩包，解压后进入目录。  
  

2. 安装 Rclone 和 FFmpeg 
~~~
sudo apt-get update && sudo apt-get install -y rclone ffmpeg
~~~
3. Rclone 挂载你需要上传到的云存储，Google 上教程很多。
~~~
rclone config
~~~
### 运行
#### 启动 Rclone Rcd 后端
~~~
nohup rclone rcd --rc-no-auth >/dev/null 2>&1 &
~~~
#### 启动 BiliBackup  
示例：
~~~
./main -f 1235678 -pn 5 -remote onedrive:/bili
~~~
### 命令参数

---
> **-f** 收藏夹 ID
>> 例如： https://space.bilibili.com/486906719/favlist?fid=1201125119 这个收藏夹  
> -f 命令所需要的收藏夹 ID 就是 fid 后面的数字  
> **你指定的收藏夹 ID 必须为公开收藏夹，否则会找不到收藏夹.**

> **-pn** 所需要备份的收藏夹页数，默认每次将备份最新的一页
>> 如果想备份所有视频就随便填写一个很大的数字，比如 100000，当返回的 Json 数据 has_more 为 false 时，将会自动停止。

> **-remote** 命令指定所需要备份到的的云存储名称和路径，比如 **onedrive:/bili** 表示将备份到 OneDrive 下的 bili 目录下。
>> 你配置的 Rclone remote 挂载名和挂载类型，可以用 Rclone config 命令查看。

## Docker 部署

---
1. 先在你机器上配置好 Rclone , 然后将 `~/.config/rclone/rclone.conf` 文件内的内容复制到 BiliBackup 目录里的 `rclone.conf` 文件里。 
~~~
git clone https://github.com/privileges-privacy/BiliBackup.git
cd BiliBackup
cat ~/.config/rclone/rclone.conf > rclone.conf
~~~  
  

2. 修改 BiliBackup 目录下的 `init.sh` 文件， 根据你的需要，修改第三行 `./main` 的启动命令。
  

3. 构建 Docker 镜像并启动容器。 
~~~
docker build --rm -t bilibackup .
docker run -it --name bili bilibackup
~~~
##### 在 Crontab 中添加计划任务，就可以每天定时备份收藏夹了。
~~~
00 4 * * * docker start bili
~~~








