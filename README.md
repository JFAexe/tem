# tem - tiny go template cli renderer
```shell
echo '[{{ time.Now | time.DateTime }}] {{ "USER" | env.Or "anonymous" }}@{{ hostname }}' | tem
```

## Installation
Via `go install`:
```shell
go install -trimpath -ldflags "-s -w" github.com/JFAexe/tem/cmd/tem@latest
```

Prebuilt binaries:
```shell
(
  TEM_VERSION="0.4.0"
  wget "https://github.com/JFAexe/tem/releases/download/v${TEM_VERSION}/tem_${TEM_VERSION}_linux_amd64.tar.gz" -O "tem_${TEM_VERSION}.tar.gz"
  tar -xzf "tem_${TEM_VERSION}.tar.gz" -C "$HOME/.local/bin/" "tem"
)
```
