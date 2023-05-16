# BiliBackup

自动下载 BiliBili 指定收藏夹内的视频，并上传到 OneDrive,Google Drive,WebDev,阿里云盘(WebDev).... 等可被 [Rclone](https://github.com/rclone/rclone) 挂载的云存储。

## 安装

1. 从 [Release](https://github.com/Privilege-privacy/BiliBackup/releases) 下载相应版本的二进制文件。


2. 安装 Rclone 和 FFmpeg (如果不需要将视频转换为 mp4 格式 FFmpeg 可以不用下载）
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
./BiliBackup -f 1235678 -pn 5 -remote onedrive:/bili
~~~
运行时将会创建 `bili.db` 以避免在下次运行时重复备份已有视频。
### 命令参数

> **-f** 收藏夹 ID
>> 例如： https://space.bilibili.com/486906719/favlist?fid=1201125119 这个收藏夹
> -f 命令所需要的收藏夹 ID 就是 fid 后面的数字
> **你指定的收藏夹 ID 必须为公开收藏夹，否则会找不到收藏夹.**

> **-pn** 所需要备份的收藏夹页数，默认将备份整个收藏夹
>> 可以指定你每次需要备份的页数，例如 -pn 1 将只备份最新一页

> **-remote** 命令指定所需要备份到的的云存储名称和路径，比如 **onedrive:/bili** 表示将备份到 OneDrive 下的 bili 目录下。
>> 你配置的 Rclone remote 挂载名和挂载类型，可以用 Rclone config 命令查看。

> **-convert** 是否转换视频格式为 mp4 (转换较为耗时，默认为 `false` 不进行转换且需要自行安装 FFmpeg)
>>当 Web 端接口解析视频下载链接失败时，会使用 Tv 端接口下载音视频，默认不执行音频合并。

> **-thread** 下载线程数 (默认为 4 )

## Docker 部署

1. 先在你机器上配置好 Rclone , 然后将 `~/.config/rclone/rclone.conf` 文件内的内容复制到 BiliBackup 目录里的 `rclone.conf` 文件里。
~~~
git clone https://github.com/privileges-privacy/BiliBackup.git
cd BiliBackup
cat ~/.config/rclone/rclone.conf > rclone.conf
~~~


2. 修改 BiliBackup 目录下的 `init.sh` 文件， 根据你的需要，修改第三行 `./BiliBackup` 的启动命令。


3. 构建 Docker 镜像并启动容器。
~~~
docker-compose up -d
~~~
##### 在 Crontab 中添加计划任务，就可以每天定时备份收藏夹了。
~~~
00 4 * * * docker start BiliBackup
~~~

## 感谢
#### 获取视频下载地址：  [FastestBilibiliDownloader](https://github.com/sodaling/FastestBilibiliDownloader)
