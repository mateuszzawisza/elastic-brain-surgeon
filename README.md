#Elastic-brain-surgeon

This is initial version of a simple tool that will tell you if your ES cluster is suffering from split brain at the moment.

##Usage

```bash
# if there is no split brain
./brain  -elasticsearch-list 127.0.0.1:9200,127.0.0.1:9201,127.0.0.1:9202 --print
Everything is ok
master: elasticsearch-box-1
  node 0: elasticsearch-box-1
  node 1: elasticsearch-box-2
  node 2: elasticsearch-box-3
```


```bash
# if there is split brain
./brain  -elasticsearch-list 127.0.0.1:9200,127.0.0.1:9201,127.0.0.1:9202 --print
The brian is split!
master: elasticsearch-box-1
  node 0: elasticsearch-box-1
  node 1: elasticsearch-box-2
master: elasticsearch-box-3
  node 0: elasticsearch-box-3
```
