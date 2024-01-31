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

---

# Example

* --emu=palladium --metrics --test
* --emu=palladium --metrics
* --emu=palladium --num=4 --test
* --emu=palladium --num=4
