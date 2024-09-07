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
  Advanced regex search<br>
  Low Mem Footprint<br>
  Single binary
</p>

<h3 align="center">
  Supports
</h3>

<p align="center">
  Docker Container logs from path<br>
  Docker Container logs<br>
  SSH remote logs<br>
  STDIN logs<br>
  Local logs<br>
  Tar logs<br>
</p>

- **Quick Setup:** One command to install and run.

- **Hassle Free:** Doesn't require elastic search or other shebang.

- **Platform:** Supports (arm64, arch64, Mac, Mac M1, Ubuntu and Windows).

- **Flexible:** View docker logs, remote logs over ssh, files on disk and piped inputs in browser.

- **Intelligent** Smartly judges log level, and dates.

- **Search** Fast search with regex.

- **Realtime** Tail logs in real time in browser.

- **Log Rotation** Supports log rotation and watch for new log files.

- **Embed in GO** Easily embed in your existing Go app.

<h1 align="center">
  View in Browser
</h1>

<p align="center">
 Intuitive UI to view logs in browser
</p>

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
![go-build-cli-time](https://coveritup.app/badge?org=kevincobain2000&repo=gol&branch=master&type=go-build-cli-time)
![go-build-all-time](https://coveritup.app/badge?org=kevincobain2000&repo=gol&branch=master&type=go-build-all-time)

![go-test-run-time](https://coveritup.app/badge?org=kevincobain2000&repo=gol&branch=master&type=go-test-run-time)
![coverage](https://coveritup.app/badge?org=kevincobain2000&repo=gol&branch=master&type=coverage)
![go-binary-size](https://coveritup.app/badge?org=kevincobain2000&repo=gol&branch=master&type=go-binary-size)
![go-mod-dependencies](https://coveritup.app/badge?org=kevincobain2000&repo=gol&branch=master&type=go-mod-dependencies)

![npm-install-time](https://coveritup.app/chart?org=kevincobain2000&repo=gol&branch=master&type=npm-install-time&theme=light&line=fill&width=150&height=150&output=svg)
![npm-build-time](https://coveritup.app/chart?org=kevincobain2000&repo=gol&branch=master&type=npm-build-time&theme=light&line=fill&width=150&height=150&output=svg)
![go-build-cli-time](https://coveritup.app/chart?org=kevincobain2000&repo=gol&branch=master&type=go-build-cli-time&theme=light&line=fill&width=150&height=150&output=svg)
![go-build-all-time](https://coveritup.app/chart?org=kevincobain2000&repo=gol&branch=master&type=go-build-all-time&theme=light&line=fill&width=150&height=150&output=svg)
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

### CLI - Basic Example

```sh
# run in current directory for pattern
gol "*log" "access/*log.tar.gz"
```

### CLI - Advanced Examples

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

### Embed in GO

If you don't want to use CLI to have seperate port and want to integrate within your existing Go app.


```go
import (
	"fmt"
	"net/http"

	"github.com/kevincobain2000/gol"
)

func main() {
    // init with options of file path you want to watch
    g := gol.NewGol(func(o *gol.GolOptions) error {
        o.FilePaths = []string{"*.log"}
        return nil
    })

    // register following two routes
    http.HandleFunc("/gol/api", g.Adapter(g.NewAPIHandler().Get))
    http.HandleFunc("/gol", g.Adapter(g.NewAssetsHandler().Get))

    // start server as usual
    http.ListenAndServe("localhost:8080", nil)
}
```

## CHANGE LOG

- **v1.0.0** - Initial release.
- **v1.0.3** - Multiple file patterns, and pipe input support.
- **v1.0.4** - Support os.Args for quick view.
- **v1.0.5** - Support ssh logs.
- **v1.0.6** - UI shows grouped output.
- **v1.0.7** - Support docker logs.
- **v1.0.14** - Sleak UI changes and support dates.
- **v1.0.17** - Support both ignore and include patterns.
- **v1.0.21** - Better logging.
- **v1.0.22** - Support UA.
- **v1.0.24** - Dropdown on files.
- **v1.0.25** - Searchable files.
- **v1.1.0** - Embed in GO, buggy.
- **v1.1.1** - Embed in GO, stable.
- **v1.1.2** - Go VUP
- **v1.1.3** - Node VUP and debounce for better performance.

## Limitations

- **Docker Logs:** Only supports logs from containers running on the same machine.
- **fmt, stdout:** For embedded use, fmt and stdout logs are not intercepted.

  **Tip:** If you want to capture, then run your app by piping output as `./app >> logs.log`.


## Development Notes

```sh
# Get some fake logs
mkdir -p testdata
while true; do date >> testdata/test.log; sleep 1; done

# Start the API
cd frontend
go run main.go --cors=4321 --open=false -f="../testdata/*log"
# API development on http://localhost:3003/api

# Start the frontend
npm install
npm run dev
# Frontend development on http://localhost:4321/
```