# What?

Prometheus Metrics Exporter by Shell

So It's Name shell_exporter

# talking is cheap, show the result

## config.yml configure file

```yaml
shells:
  - name: fruit_random_price
    const_labels:
      env: test
      mode: script
    labels_regexp: (?P<name>.+), (?P<type>.+), (?P<size>.+) (?P<value>[0-9.]+)
    help: "show fruit random number by shell."
    cmd: echo -n -e "apple, fruit, big $RANDOM\ntomato, vegetable, small $RANDOM"
    bin: /bin/bash
  - name: process_total
    const_labels:
      env: test
      mode: shell
      os: ubuntu
      name: ssh
    labels_regexp: (?P<hostname>.+)\n(?P<value>[0-9.]+)
    help: "show process count total."
    cmd: hostname; ps -ef | grep bash | grep -v 'grep' | wc -l
    bin: /bin/bash
  - name: ss_estab_total
    const_labels:
      env: product
      mode: script
      os: debian
    labels_regexp: (?P<hostname>.+)\n(?P<value>[0-9.]+)
    help: "show ssh estab count by shell."
    cmd: hostname; ss -antp | grep ssh | grep ESTAB | wc -l
    bin: /bin/bash
```

## Http Reqeust Metrics
```
# HELP process_total show process count total.
# TYPE process_total gauge
process_total{env="test",hostname="ozhiwei",mode="shell",name="ssh",os="ubuntu"} 8
# HELP ss_estab_total show ssh estab count by shell.
# TYPE ss_estab_total gauge
ss_estab_total{env="product",hostname="ozhiwei",mode="script",os="debian"} 5
# HELP fruit_random_price show fruit random number by shell.
# TYPE fruit_random_price gauge
fruit_random_price{env="test",mode="script",name="apple",size="big",type="fruit"} 24075
fruit_random_price{env="test",mode="script",name="tomato",size="small",type="vegetable"} 10732
```

## Build & Install

```
git clone github.com/ozhiwei/shell_exporter

cd shell_exporter

docker run -it --rm -v $(pwd):/data/ -w /data golang make
```

## Usage

copy the shell_exporter to any node and run it (don't forget configure config.yml)