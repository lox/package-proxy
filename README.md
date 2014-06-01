Package Manager Cache (pmc) 
===========================

Pmc is a caching reverse proxy designed for the purpose of caching package managers. 

Currently pmc supports Ubuntu/Apt. 

Running
========

```bash
go build pmc.go
./pmc
```

Todo
====

 * Disk-based cache (presently in memory)
 * Expiration (currently requires a restart)