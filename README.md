# tem - tiny go template cli renderer
```shell
echo '[{{ time.Now | time.DateTime }}] {{ "USER" | env.Or "anonymous" }}@{{ hostname }} {{ "hello templates!" | string.Upper }}' | tem
```

## Installation
```shell
go install -trimpath -ldflags "-s -w" github.com/JFAexe/tem/cmd/tem@latest
```
