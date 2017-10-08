package gin

import (
    "github.com/gin-gonic/gin"
    "io/ioutil"
    "github.com/wonktnodi/lollipop/pkg/log"
)

// DebugHandler creates a dummy handler function, useful for quick integration tests
func DebugHandler(logger log.Logger) gin.HandlerFunc {
    return func(c *gin.Context) {
        param := c.Params.ByName("param")
        logger.Debug("Method:", c.Request.Method)
        logger.Debug("URL:", c.Request.RequestURI)
        logger.Debug("Query:", c.Request.URL.Query())
        logger.Debug("Params:", c.Params)
        logger.Debug("Headers:", c.Request.Header)
        body, _ := ioutil.ReadAll(c.Request.Body)
        c.Request.Body.Close()
        logger.Debug("Body:", string(body))
        c.JSON(200, gin.H{
            "message": "pong",
            "uri": param,
        })
    }
}

