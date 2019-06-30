# M3U8

M3U8 是一个使用了 Go 语言编写的迷你 M3U8 下载工具。你只需指定必要的 flag (`u`、`o`、`c`) 来运行, 工具就会自动帮你解析 M3U8 文件，并将 TS 片段下载下来合并成一个文件。


## 功能

- 下载和解析 M3U8（仅限 VOD 类型）
- 下载 TS 失败重试
- 解析 Master playlist
- 解密 TS
- 合并 TS 片段

## 用法

### 源码方式

```bash
go run main.go -u=http://example.com/index.m3u8 -o=/data/example
```

### 二进制方式:

Linux 和 MacOS

```
./m3u8 -u=http://example.com/index.m3u8 -o=/data/example
```

Windows PowerShell

```
.\m3u8.exe -u="http://example.com/index.m3u8" -o="D:\data\example"
```

参数说明：

```
- u M3U8 地址
- o 文件保存目录
- c 下载协程并发数，默认 25
```

部分链接可能限制请求频率，可根据实际情况调整 `c` 参数的值。

## 下载

[二进制文件](https://github.com/oopsguy/m3u8/releases)

## 截屏

![Demo](./screenshots/demo.gif)

## 参考资料

- [TS科普 2 包头](https://blog.csdn.net/cabbage2008/article/details/49281729)
- [HTTP Live Streaming draft-pantos-http-live-streaming-23](https://tools.ietf.org/html/draft-pantos-http-live-streaming-23#section-4.3.4.2)
- [MPEG transport stream - Wikipedia](https://en.wikipedia.org/wiki/MPEG_transport_stream)


## License

[MIT License](./LICENSE)