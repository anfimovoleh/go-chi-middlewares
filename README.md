# go-chi-middlewares
Common middlewares used in projects with go-chi. Just import necessary middlewares without reimplementation of them each time in a new project. 

## Getting started

Add go-chi-middlewares as a dependency to your project

```shell
go get github.com/anfimovoleh/go-chi-middlewares
```  
The code above fetches the go-chi-middlewares as your project dependency. 

## Current list of middlewares
* Logger(zap);
* Basic Auth;
* Verify remote address is private.

We are planning to add next middlewares in the near future: 
* JWT Auth;
* OAuth2
* Rate limiter;