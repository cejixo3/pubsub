# A simple pub-sub

### Requirements
- Go 1.13. It must to work on a lower version but was tested only on 1.13

### Installation
The recommended way to get started using the *pubsub* library is by using go modules to install the dependency in your project.
```shell script
go get github.com/cejixo3/pubsub.git
```

### Usage
This a rough example how to use this library together with http server ([gin](https://github.com/gin-gonic/gin) in example).
Refer to documentation for more details.
```go
package main

import (
	"github.com/cejixo3/pubsub.git"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
)

func main() {
	ps := pubsub.New()
	r := gin.Default()
	apiv1 := r.Group("/api/v1")

	apiv1.PUT("/subscribe/tn/:topic-name/sn/:subscription-name", func(c *gin.Context) {
		// some validation here
		ps.Subscribe(c.Param("topic-name"), c.Param("subscription-name"))
		c.AbortWithStatus(http.StatusNoContent)
	})

	apiv1.DELETE("/unsubscribe/tn/:topic-name/sn/:subscription-name", func(c *gin.Context) {
		// some validation here
		ps.Unsubscribe(c.Param("topic-name"), c.Param("subscription-name"))
		c.AbortWithStatus(http.StatusNoContent)
	})

	apiv1.POST("/publish/tn/:topic-name", func(c *gin.Context) {
		// some validation here
		b, err := ioutil.ReadAll(c.Request.Body)
		if err != nil {
			// process error
		}
		ps.Publish(c.Param("topic-name"), b)
		c.AbortWithStatus(http.StatusAccepted)
	})

	apiv1.GET("/unsubscribe/tn/:topic-name/sn/:subscription-name", func(c *gin.Context) {
		// some validation here
		if msg, err := ps.Poll(c.Param("topic-name"), c.Param("subscription-name")); err == pubsub.ErrNoSubscriptions {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		} else {
			if msg == nil {
				c.AbortWithStatus(http.StatusNotFound)
			} else {
				c.Data(http.StatusOK, "application/json", msg)
			}
		}
	})
	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}

``` 

### Testing
```shell script
make test
```

### Todo
- [ ] Prettify tests
- [ ] Implement removing keys from p.hm[tn] when subscription list is empty
- [ ] Add ability to use custom storage (for testing another data structures for example)
- [ ] Add benchmarks
- [ ] Optimize map key sizes (use hashing for example)
- [ ] Play with garbage collector for reducing gc count (in application, not library)
- [ ] Improve naming in code 
- [ ] Maybe separate ```pubsub.go``` into several parts (```errors.go```,```storage.go```, etc)
- [ ] Improve publish method performance (run in several goroutines for example, or something else) 