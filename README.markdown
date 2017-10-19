The [bakery algorithm][1] is a way for software processes to achieve mutual
exclusion for the usage of a shared resource.

## Running

```bash

$ go get github.com/yati-sagade/bakery
$ bakery -nodes 10 -iters 100000

```

[1]: http://lamport.azurewebsites.net/pubs/bakery.pdf
