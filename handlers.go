package main

import "github.com/gin-gonic/gin"

// ------ Service ------
func handlePing(c *gin.Context) {
	c.String(200, "pong")
}

// ------ Test ------
func handlePanic(c *gin.Context) {
	panic("test panic for Sentry")
}
