package main

import (
	"crypto/tls"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/penglongli/gin-metrics/ginmetrics"
)

func responseMsg(c *gin.Context) string {
	name, err := os.Hostname()
	if err != nil {
		panic(err)
	}
	response := ""
	for k, v := range c.Request.Header {
		response += k + ": " + v[0] + "\n"
	}

	response += "\nHostname: " + name + "\n"
	response += "Path: " + c.Request.URL.Path + "\n"
	response += "L3 IP: " + strings.Split(c.Request.RemoteAddr, ":")[0] + "\n"
	if c.GetHeader("Cf-Connecting-Ip") != "" {
		response += "L7 IP: " + c.GetHeader("Cf-Connecting-Ip") + "\n"
	} else {
		response += "L7 IP: " + c.ClientIP() + "\n"
	}

	backend := os.Getenv("BACKEND")
	if backend != "" {
		tr := &http.Transport{ //解决x509: certificate signed by unknown authority
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client := &http.Client{
			Transport: tr, //解决x509: certificate signed by unknown authority
		}
		res, err := client.Get(backend)
		if err != nil {
			response += "Error: " + err.Error() + "\n"
		} else {
			defer res.Body.Close()
			sitemap, err := ioutil.ReadAll(res.Body)
			if err != nil {
				log.Fatal(err)
			}
			response += "\nFrom Backend:\n" + string(sitemap)
		}
	}
	return response
}

func main() {
	r := gin.Default()

	m := ginmetrics.GetMonitor()

	// +optional set metric path, default /debug/metrics
	m.SetMetricPath("/metrics")
	// +optional set slow time, default 5s
	m.SetSlowTime(10)
	// +optional set request duration, default {0.1, 0.3, 1.2, 5, 10}
	// used to p95, p99
	m.SetDuration([]float64{0.1, 0.3, 1.2, 5, 10})

	// set middleware for gin
	m.Use(r)

	r.GET("/frontend", func(c *gin.Context) {
		c.String(200, responseMsg(c))
	})
	r.GET("/", func(c *gin.Context) {
		c.String(200, responseMsg(c))
	})
	if gin.Mode() != gin.ReleaseMode {
		r.Run("127.0.0.1:8080")
	} else {
		r.Run(":8080")
	}
}
