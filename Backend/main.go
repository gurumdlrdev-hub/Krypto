package main

import (
	"log"

	"github.com/gin-gonic/gin"
)

func main() {

	InitDB()

	r := gin.Default()

	r.Static("/static", "../Frontend")

	// Page Views (Both Clean URL and .html support)
	r.GET("/", func(c *gin.Context) {
		c.File("../Frontend/home.html")
	})
	r.GET("/login", func(c *gin.Context) {
		c.File("../Frontend/login.html")
	})
	r.GET("/signup", func(c *gin.Context) {
		c.File("../Frontend/signup.html")
	})
	r.GET("/encrypt", func(c *gin.Context) {
		c.File("../Frontend/encrypt.html")
	})
	r.GET("/decrypt", func(c *gin.Context) {
		c.File("../Frontend/decrypt.html")
	})

	// Image Cipher Routes
	r.GET("/encrypt/image", func(c *gin.Context) {
		c.File("../Frontend/encryptimage.html")
	})
	r.GET("/encryptimage.html", func(c *gin.Context) {
		c.File("../Frontend/encryptimage.html")
	})
	r.POST("/encrypt/image", gin.WrapF(encryptImage))

	r.GET("/decrypt/image", func(c *gin.Context) {
		c.File("../Frontend/decryptimage.html")
	})
	r.GET("/decryptimage.html", func(c *gin.Context) {
		c.File("../Frontend/decryptimage.html")
	})
	r.POST("/decrypt/image", gin.WrapF(decryptImage))

	// Audio Cipher Routes
	r.GET("/encrypt/audio", func(c *gin.Context) {
		c.File("../Frontend/encryptaudio.html")
	})
	r.GET("/encryptaudio.html", func(c *gin.Context) {
		c.File("../Frontend/encryptaudio.html")
	})
	r.POST("/encrypt/audio", gin.WrapF(encryptAudio))

	r.GET("/decrypt/audio", func(c *gin.Context) {
		c.File("../Frontend/decryptaudio.html")
	})
	r.GET("/decryptaudio.html", func(c *gin.Context) {
		c.File("../Frontend/decryptaudio.html")
	})
	r.POST("/decrypt/audio", gin.WrapF(decryptAudio))

	// Video Cipher Routes
	r.GET("/encrypt/video", func(c *gin.Context) {
		c.File("../Frontend/encryptvideo.html")
	})
	r.GET("/encryptvideo.html", func(c *gin.Context) {
		c.File("../Frontend/encryptvideo.html")
	})
	r.POST("/encrypt/video", gin.WrapF(encryptVideoPipeline))

	r.GET("/decrypt/video", func(c *gin.Context) {
		c.File("../Frontend/decryptvideo.html")
	})
	r.GET("/decryptvideo.html", func(c *gin.Context) {
		c.File("../Frontend/decryptvideo.html")
	})
	r.POST("/decrypt/video", gin.WrapF(decryptVideoPipeline))

	r.POST("/signup", gin.WrapF(signupHandler))
	r.POST("/login", gin.WrapF(loginHandler))

	log.Println("KRYPTO Server with Gin Metrics running at: http://localhost:2409")
	if err := r.Run(":2409"); err != nil {
		log.Fatalf("Server Starting Error: %v", err)
	}
}
