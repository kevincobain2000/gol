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
  Single binary
</p>

**Quick Setup:** One command to install and run.

**Hassle Free:** Doesn't require elastic search or other shebang.

**Platform:** Supports (arm64, arch64, Mac, Mac M1, Ubuntu and Windows).

**Flexible:** Works with multiple logs file, with massive size support.

**Supports** Plain text, ansii outputs, tar and gz compressed.

**Intelligent** Smartly judges log level.

**Search** Fast search with regex.

**Realtime** Tail logs in real time in browser.

**Watch Changes** Supports log rotation and watch for new log files.

<h1 align="center">
  View in Browser
</h1>

<p align="center">
  <a href="https://github.com/kevincobain2000/gol">
    <img alt="gol" src="https://imgur.com/MJmCAsw.png">
  </a>
</p>

## Reports from [coveritup](https://coveritup.app/readme?org=kevincobain2000&repo=gol&branch=master)

<p align="center">
  <a href="https://coveritup.app/readme?org=kevincobain2000&repo=gol&branch=master">
    <img alt="gol" src="https://coveritup.app/progress?org=kevincobain2000&repo=gol&branch=master&type=coverage&theme=dark&style=bar" width="150">
  </a>
</p>

![npm-install-time](https://coveritup.app/badge?org=kevincobain2000&repo=gol&branch=master&type=npm-install-time)
![npm-build-time](https://coveritup.app/badge?org=kevincobain2000&repo=gol&branch=master&type=npm-build-time)
![go-build-time](https://coveritup.app/badge?org=kevincobain2000&repo=gol&branch=master&type=go-build-time)
![go-lint-errors](https://coveritup.app/badge?org=kevincobain2000&repo=gol&branch=master&type=go-lint-errors)

![go-test-run-time](https://coveritup.app/badge?org=kevincobain2000&repo=gol&branch=master&type=go-test-run-time)
![coverage](https://coveritup.app/badge?org=kevincobain2000&repo=gol&branch=master&type=coverage)
![go-binary-size](https://coveritup.app/badge?org=kevincobain2000&repo=gol&branch=master&type=go-binary-size)
![go-mod-dependencies](https://coveritup.app/badge?org=kevincobain2000&repo=gol&branch=master&type=go-mod-dependencies)

![npm-install-time](https://coveritup.app/chart?org=kevincobain2000&repo=gol&branch=master&type=npm-install-time&theme=light&line=fill&width=150&height=150&output=svg)
![npm-build-time](https://coveritup.app/chart?org=kevincobain2000&repo=gol&branch=master&type=npm-build-time&theme=light&line=fill&width=150&height=150&output=svg)
![go-build-time](https://coveritup.app/chart?org=kevincobain2000&repo=gol&branch=master&type=go-build-time&theme=light&line=fill&width=150&height=150&output=svg)
![go-lint-errors](https://coveritup.app/chart?org=kevincobain2000&repo=gol&branch=master&type=go-lint-errors&theme=light&line=fill&width=150&height=150&output=svg)
![go-test-run-time](https://coveritup.app/chart?org=kevincobain2000&repo=gol&branch=master&type=go-test-run-time&theme=light&line=fill&width=150&height=150&output=svg)
![coverage](https://coveritup.app/chart?org=kevincobain2000&repo=gol&branch=master&type=coverage&theme=light&line=fill&width=150&height=150&output=svg)
![go-binary-size](https://coveritup.app/chart?org=kevincobain2000&repo=gol&branch=master&type=go-binary-size&theme=light&line=fill&width=150&height=150&output=svg)
![go-mod-dependencies](https://coveritup.app/chart?org=kevincobain2000&repo=gol&branch=master&type=go-mod-dependencies&theme=light&line=fill&width=150&height=150&output=svg)


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
```

## Advanced Examples

```sh
gol -f="/var/log/*.log"
gol -f="/var/log/*.log.tar.gz"
gol -f="/var/log/*.log*"
```

## CHANGE LOG

- **v1.0.0** - Initial release