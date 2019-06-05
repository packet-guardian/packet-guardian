# User Agent Parsing Library

This library parses a browser user agent and extracts information about the
client system such as OS name, version, and edition name.

## Supported OSes

- Windows (Version and VersionName)
- macOS (Version and VersionName)
- Linux (Distro)
- ChromeOS
- Android (Version)
- iOS (Version and Distro (iPod, iPad, iPhone))
- Kindle

## Example

```go
package main

import (
    "fmt"

    "github.com/packet-guardian/useragent"
)

func main() {
    agent := "Mozilla/6.0 (Windows NT 6.2; WOW64; rv:16.0.1) Gecko/20121011 Firefox/16.0.1"
    osInfo := useragent.ParseUserAgent(agent)

    fmt.Println(osInfo.OS)
    fmt.Println(osInfo.Distor)
    fmt.Println(osInfo.Version)
    fmt.Println(osInfo.VersionName)
}
```
