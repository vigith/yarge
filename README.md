Yet Another Range
=================

A rewrite of [range](https://github.com/square/libcrange) to suit the ephemeral nature of cloud architecture. This is not a replacement
for the old range, since this understands only a subset of features and also have added few features which are very valuable in cloud environment.

 * a new querying language 
 * dynamic store
 * reverse lookup are faster
 * options for fast querying (ie, return the first result)
 * written in go

## Range Expression

### Simple Operations
  * `%range1` == expand range1
  * `%range1:KEYS` == show all the KEYS in range1
  * `%range1:FOO`  == show the value for key FOO in range1
  * `%RANGE`   == toplevel (here RANGE is a keyword)
  * `%%RANGE`  == second level (can go any level down with more `%` till you hit leaf node)
  * `%%range1` == second level w.r.t `range1` (you can go any level down with `%`)
  * `*hostname`  == get cluster where this hostname is present
  * `*value;KEY` == get the cluster where KEY=value
  * `*value;KEY:HINT` == get the cluster where KEY=value, HINT is to scope within a toplevel  

### Set Operatons:
  * `%range1 , %range2` == union (space is optional)
  * `%range1 ,- %range2` == set difference
  * `%range1 ,& %range2` == intersection
  * `%range1,(%range1 ,& %range2)` == set operations with grouping using brackets

### Advanced Operations

  * `%*value`     == cluster operation on reverse lookup
  * `%*value;KEY` == cluster operation on reverse lookup with KEY where KEY=value
  * `%*value;KEY:HINT` == cluster operation on reverse lookup where KEY=value, HINT is to scope within a toplevel
  * `%(%range1:FOO,(%range1:BAR ,& %range2:MOO))` == set operations with grouping using brackets
  * `%(*value1;KEY1:HINT1 ,& *value2;KEY2:HINT2)` == cluster lookup the result of a set operation done on reverse lookups

I would suggest you to read `expr.peg` to understand all the possbile query combinations. The AST evaluator evaluates from Right to Left.

## Deployment

If you are planning to use *etcdstore* as the store for the range then we need to setup etcd cluster.

### Etcd

*WIP*

## Development

### Requirements
#### PEG
#### Etcd
#### Yaml Parser


## Miscellaneous

### Bugs
Probably many, please open issues. Patches are most welcome.

### Author
[Vigith Maurice](https://github.com/vigith)
