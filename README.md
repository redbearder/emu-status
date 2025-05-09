# Argument
* --test
  * type: boolean
  * options: none
  * desc: get test data from local test file if set

* --emu=\<VALUE>
  * type: string
  * options: palladium / zebu
  * desc: get status data from set value

* --num=\<VALUE>
  * type: int
  * options: none
  * desc: request num from emu

* --metrics
  * type: boolean
  * options: none
  * desc: indicate prometheus metrics server is available

* --port=\<VALUE>
  * type: int
  * default: 27777
  * options: none
  * desc: indicate prometheus metrics exporter port
---

# Example

* --emu=palladium --metrics --test
* --emu=palladium --metrics
* --emu=palladium --metrics --port 9000
* --emu=palladium --num=4 --test
* --emu=palladium --num=4


# Build

```azure
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o emu-status .
```

# Run
```azure
nohup ./emu-status -emu palladium -metrics --port 9000 > emu-status.out 2>&1 &
```