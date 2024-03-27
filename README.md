# FS-Shadow

`fs-shadow` is a robust and efficient file system watcher that recursively watches all files and folders for changes. Built with compatibility in mind, it functions seamlessly across Windows, Linux, and MacOS. Think of it like `fs-notify`, but with recursive watching capabilities.

## Features

- 🔍 **Recursive Watching**: Monitor entire directories, including their sub-directories and files.
- 💡 **Cross-Platform**: Supports Windows, Linux, and MacOS.
- 🚀 **Real-time Events**: Get instant feedback with event-driven mechanisms.
- 🛠 **Flexible**: Easily integrate into existing projects.

## Installation

To use `fs-shadow` in your project:

```bash
go get github.com/ayhanozemre/fs-shadow
```

## Usage
Here is a basic example to monitor changes in the `/tmp/fs-shadow`` directory on Linux:

```go
package main

import (
	"fmt"
	"github.com/ayhanozemre/fs-shadow/watcher"
	log "github.com/sirupsen/logrus"
)

func main() {
	log.SetLevel(log.DebugLevel)
	tw, _, err := watcher.NewFSWatcher("/tmp/fs-shadow")

	if err == nil {
		go func() {
			ch := tw.GetEvents()
			err := tw.GetErrors()
			for {
				select {
				case p := <-ch:
					log.Debug(fmt.Sprintf("Event-> Name:%s UUID:%s", p.Name, p.UUID))
				case e := <-err:
					log.Debug("Error->", e)
				}
				tw.PrintTree("TREE")
			}
		}()

		done := make(chan bool)
		<-done
		tw.Stop()
	} else {
		log.Error(err)
	}
}
```


## API
### `NewFSWatcher(path string) (*FSWatcher, error)`
Creates a new file system watcher for the given path.

### `GetEvents() <-chan *Event`
Returns a channel of events, notifying about file or folder changes.

### `GetErrors() <-chan error`
Returns a channel of errors, if any occur during the watching process.

### `Stop()`
Stops the file system watcher.

## Contributing
We welcome contributions! If you'd like to help improve fs-shadow, please fork the repository and submit a pull request.

## License
This project is licensed under the MIT License.