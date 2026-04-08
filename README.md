# tem - go template cli renderer
```shell
echo '[{{ timeNow | timeFormatDateTime }}] {{ env "USER" }}@{{ hostname }}' | tem
```

## Installation
```shell
go install -trimpath -ldflags "-s -w" github.com/JFAexe/tem/cmd/tem@latest
```
