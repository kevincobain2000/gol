<p align="center">
  <a href="https://github.com/kevincobain2000/gol">
    <img alt="gol" src="https://imgur.com/sktoYPP.png" width="120">
  </a>
</p>

<h1 align="center">
  Logs Viewer
</h1>

<p align="center">
  View realtime logs in your fav browser<br>
  Local, Docker, Remote, Pipes<br>
  Advanced regex search<br>
  Low Mem Footprint<br>
  Single binary
</p>

**Quick Setup:** One command to install and run.

**Hassle Free:** Doesn't require elastic search or other shebang.

**Platform:** Supports (arm64, arch64, Mac, Mac M1, Ubuntu and Windows).

**Flexible:** View docker logs, remote logs over ssh, files on disk and piped inputs in browser.

**Intelligent** Smartly judges log level, and dates.

**Search** Fast search with regex.

**Realtime** Tail logs in real time in browser.

**Watch Changes** Supports log rotation and watch for new log files.

<h1 align="center">
  View in Browser
</h1>

<p align="center">
  <a href="https://github.com/kevincobain2000/gol">
    <img alt="gol" src="https://imgur.com/fBK0hGa.png">
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
```

## Examples

```sh
# run in current directory for *log & ./*/*log
gol
```

```sh
# run in current directory for pattern
gol "storage/*log" "access/*log.tar.gz"
```

## Advanced Examples

All patterns work in combination with each other.

```sh
# search using pipe and file patterns
demsg | gol -f="/var/log/*.log"

# over ssh
# port optional (default 22), password optional (default ''), private_key optional (default $HOME/.ssh/id_rsa)
gol -s="user@host[:port] [password=/path/to/password] [private_key=/path/to/key] /app/*logs"

# Docker all container logs
gol -d=""

# Docker specific container logs
gol -d="container-id"

# Docker specific path on a container
gol -d="container-id /app/logs.log"

# All patterns combined
gol -d="container-id" \
    -d="container-id /app/logs.log" \
    -s="user@host[:port] [password=/path/to/password] [private_key=/path/to/key] /app/*logs" \
    -f="/var/log/*.log"
```

## CHANGE LOG

- **v1.0.0** - Initial release.
- **v1.0.3** - Multiple file patterns, and pipe input support.
- **v1.0.4** - Support os.Args for quick view.
- **v1.0.5** - Support ssh logs.
- **v1.0.6** - UI shows grouped output.
- **v1.0.7** - Support docker logs.
- **v1.0.14** - Sleak UI changes and support dates.

## Limitations

- **Docker Logs:** Only supports logs from containers running on the same machine.