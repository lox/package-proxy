Package Manager Cache (pmc) 
===========================

Pmc is a caching reverse proxy designed for the purpose of caching package managers. 

Currently pmc supports Ubuntu/Apt. 

Running
========

```bash
go get github.com/lox/pmc
pmc
```

Todo
====

 * Disk-based cache (presently in memory)
 * Expiration (currently requires a restart)