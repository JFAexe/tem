# tem - tiny go template cli renderer
[![GitHub Release](https://img.shields.io/github/v/release/JFAexe/tem?style=for-the-badge&color=%2300ADD8)](https://github.com/JFAexe/tem/releases/latest)
[![License](https://img.shields.io/github/license/JFAexe/tem?style=for-the-badge&color=%2300ADD8)](LICENSE)

> Just because you can, doesn't mean you should.

```shell
echo '
[[- $files := list.New -]]
[[- range $f := filepath.Walk "pkg/**/*.go" true -]]
  [[- $files = ( map
    "path" $f.RelPath
    "data" ( $f.AbsPath | file | to.Bytes )
  ) | list.Append $files -]]
[[- end -]]
---
[[ $files | list.SortBy "path" | data.ToYAML ]]
' | tem
```

## Installation
> **DO NOT run any shell commands unless you understand them**

### Building via `go install`
```shell
go install -trimpath -ldflags "-s -w" github.com/JFAexe/tem/cmd/tem@latest
```

### Prebuilt binaries for Linux/Darwin via shell
```shell
(
  TEM_SYSTEM="darwin"
  TEM_ARCH="arm64"
  TEM_DOWNLOAD_PATH="$HOME/Downloads"
  TEM_INSTALL_PATH="$HOME/.local/bin/"

  TEM_URL=$(curl -sL https://api.github.com/repos/JFAexe/tem/releases/latest | grep -o "https://[^\"]*${TEM_SYSTEM}_${TEM_ARCH}[^\"]*")
  TEM_ARCHIVE="$TEM_DOWNLOAD_PATH/${TEM_URL##*/}"

  curl -sL "$TEM_URL" -o "$TEM_ARCHIVE" && tar -xzf "$TEM_ARCHIVE" -C "$TEM_INSTALL_PATH" "tem"
)
```

### Prebuilt binaries for Windows via powershell
```powershell
$TEM_SYSTEM        = "windows"
$TEM_ARCH          = "amd64"
$TEM_DOWNLOAD_PATH = "$env:USERPROFILE\Downloads"
$TEM_INSTALL_PATH  = "$env:LOCALAPPDATA\tem"

$RELEASE     = Invoke-RestMethod -Uri "https://api.github.com/repos/JFAexe/tem/releases/latest"
$TEM_URL     = $RELEASE.assets.browser_download_url | Where-Object { $_ -match "${TEM_SYSTEM}_${TEM_ARCH}" }
$TEM_ARCHIVE = "$TEM_DOWNLOAD_PATH\$($TEM_URL.Split('/')[-1])"

Invoke-WebRequest -Uri $TEM_URL -OutFile $TEM_ARCHIVE
New-Item -ItemType Directory -Path $TEM_INSTALL_DIR -Force | Out-Null
Expand-Archive -Path $TEM_ARCHIVE -DestinationPath $TEM_INSTALL_DIR -Force

$ENV_PATH = [Environment]::GetEnvironmentVariable("Path", "User")

if ($ENV_PATH -notlike "*$TEM_INSTALL_DIR*") {
  [Environment]::SetEnvironmentVariable("Path", "$ENV_PATH;$TEM_INSTALL_DIR", "User")
}
```

## Functions
[`// TODO`](pkg/template/functions/)
