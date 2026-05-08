# tem - tiny go template cli renderer
[![GitHub Release](https://img.shields.io/github/v/release/JFAexe/tem?style=for-the-badge&color=%2300ADD8)](https://github.com/JFAexe/tem/releases/latest)
[![License](https://img.shields.io/github/license/JFAexe/tem?style=for-the-badge&color=%2300ADD8)](LICENSE)

```shell
echo '
[[- $files := list.New -]]
[[- range $f := filepath.Walk "**/functions/[a-z]*.go" true | list.Reverse -]]
  [[- $files = $files | list.Concat ( map
    "name" ( $f.Name | string.TrimSuffix ( $f.Name | filepath.Ext ) )
    "path" $f.Path
  ) -]]
[[- end -]]
---
[[ map "files" $files | data.ToYAML -]]
' | tem -l '[[' -r ']]'
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
  TEM_VERSION="0.7.0"
  TEM_SYSTEM="linux"
  TEM_ARCH="amd64"
  TEM_ARCHIVE="$HOME/Downloads/tem_${TEM_VERSION}.tar.gz"
  TEM_PATH="$HOME/.local/bin/"

  wget "https://github.com/JFAexe/tem/releases/download/v${TEM_VERSION}/tem_${TEM_VERSION}_${TEM_SYSTEM}_${TEM_ARCH}.tar.gz" -O "$TEM_ARCHIVE"
  tar -xzf "$TEM_ARCHIVE" -C "$TEM_PATH" "tem"
)
```

### Prebuilt binaries for Windows via powershell
```powershell
$TEM_VERSION     = "0.7.0"
$TEM_SYSTEM      = "windows"
$TEM_ARCH        = "amd64"
$TEM_ARCHIVE     = "$env:USERPROFILE/Downloads/tem_${TEM_VERSION}.zip"
$TEM_INSTALL_DIR = "$env:LOCALAPPDATA/tem"

Invoke-WebRequest -Uri "https://github.com/JFAexe/tem/releases/download/v${TEM_VERSION}/tem_${TEM_VERSION}_${TEM_SYSTEM}_${TEM_ARCH}.zip" -OutFile "$TEM_ARCHIVE"
New-Item -ItemType Directory -Path "$TEM_INSTALL_DIR" -Force | Out-Null
Expand-Archive -Path "$TEM_ARCHIVE" -DestinationPath "$TEM_INSTALL_DIR" -Force

$ENV_PATH = [Environment]::GetEnvironmentVariable("Path", "User")

if ($ENV_PATH -notlike "*$TEM_INSTALL_DIR*") {
  [Environment]::SetEnvironmentVariable("Path", "$ENV_PATH;$TEM_INSTALL_DIR", "User")
}
```

## Functions
[`// TODO`](pkg/template/functions/)
