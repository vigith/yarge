rangeexpr
=========

Range queries are parsed using [PEG](http://en.wikipedia.org/wiki/Parsing_expression_grammar) (parsing expression grammar).


## Few Valid Queries

### Simple Operations
  * %range1 == expand range1
  * %range1:KEYS == show all the KEYS in range1
  * %range1:FOO == show the value for key FOO in range1

### Set Operatons:
  * %range1 , %range2 == union (space is optional)
  * %range1 ^ %range2 == set difference
  * %range1 & %range2 == intersection
  * %range1,(%range1 & %range2) == set operations with grouping using brackets

### Advanced Operations
  * %RANGE   == toplevel (here RANGE is a keyword)
  * %%RANGE  == second level (here RANGE is a keyword)
  * %%range1 == second level w.r.t range1
  * *hostname == get cluster where this hostname is present
  * *value;KEY:HINT == get the cluster where KEY=value, HINT is to scope within a toplevel

I would suggest you to read `rangeexpr.peg` to understand all the possbile query combinations.

### Author
[Vigith Maurice](https://github.com/vigith)
