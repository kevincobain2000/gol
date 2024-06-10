<p align="center">
  <a href="https://github.com/kevincobain2000/gol">
    <img alt="gol" src="https://imgur.com/sktoYPP.png" width="120">
  </a>
</p>
<h1 align="center">
  Logs Viewer
</h1>

<p align="center">
  View realtime logs in browser<br>
  Advanced regex search<br>
  Zero dependency
</p>

**Quick Setup:** One command to install and run.

**Hassle Free:** Doesn't require elastic search or other shebang.

**Platform:** Supports (arm64, arch64, Mac, Mac M1, Ubuntu and Windows).

**Flexible:** Works with multiple logs file, with massive size support.

**Supports** Log rotation or tar and gz compressed.

**Intelligent** Judges log level based on non patented algorithm.

**Search** Fast search with regex.

**Realtime** Tail logs in real time in browser.



### Install using [go](https://github.com/kevincobain2000/gobrew)

```bash
go install github.com/kevincobain2000/gol@latest
```

### Install using curl

Use this method if go is not installed on your server

```bash
curl -sL https://raw.githubusercontent.com/kevincobain2000/gol/master/install.sh | sh
mv gol /usr/local/bin/
```

## Examples

```sh
# run in current directory
gol

gol -f=/var/log/*.log
gol -f=/var/log/*.log.tar.gz
gol -f=/var/log/*.log*
```

**All done!**

## CHANGE LOG

- **v1.0.0** - Initial release