# tem - go template cli renderer

```shell
echo '[{{ timeNow | timeFormatDateTime }}] {{ env "USER" }}@{{ hostname }}' | ./tem
```
