# M3u8

M3u8 - a mini m3u8 downloader written in Golang for downloading and merging TS(Transport Stream) files.

You only need to specify the flags(`u`, `o`, `c`) to run, downloader will automatically download the m3u8 and parse it, 
then download and merge all TS files.


## Features

- Download and parse m3u8（VOD）
- Support Master playlist
- Support encrypted TS
- Support merge TS

## Usage

**installation**

```
go get -u github.com/oopsguy/m3u8
```

**build**

```
go build
```

**run**

Linux & MacOS

```
./m3u8 -u=http://example.com/index.m3u8 -o=/data/example
```

Windows PowerShell

```
.\m3u8 -u="http://example.com/index.m3u8" -o="D:\data\example"
```

**help**

```bash
m3u8 -h
```

## Download

[Binary packages and source code](https://github.com/oopsguy/m3u8/releases)

## Screenshots

![Demo](./screenshots/demo.gif)

## References

- [HTTP Live Streaming draft-pantos-http-live-streaming-23](https://tools.ietf.org/html/draft-pantos-http-live-streaming-23#section-4.3.4.2)
- [MPEG transport stream - Wikipedia](https://en.wikipedia.org/wiki/MPEG_transport_stream)


## License

[MIT License](./LICENSE)