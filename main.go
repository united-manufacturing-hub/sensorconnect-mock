package main

import (
	"github.com/gin-contrib/gzip"
	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"github.com/united-manufacturing-hub/umh-utils/logger"
	"go.uber.org/zap"
	"math/rand"
	"time"
)

func main() {
	// Initialize zap logging
	log := logger.New("LOGGING_LEVEL")
	defer func(logger *zap.SugaredLogger) {
		err := logger.Sync()
		if err != nil {
			panic(err)
		}
	}(log)

	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	// Add a ginzap middleware, which:
	//   - Logs all requests, like a combined access and error log.
	//   - Logs to stdout.
	//   - RFC3339 with UTC time format.
	router.Use(ginzap.Ginzap(zap.L(), time.RFC3339, true))

	// Logs all panic to error log
	//   - stack means whether output the stack info.
	router.Use(ginzap.RecoveryWithZap(zap.L(), true))

	// Use gzip for all requests
	router.Use(gzip.Gzip(gzip.DefaultCompression))

	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, "online")
	})
	router.POST("/", Handler)

	// Listen and serve on 1337
	zap.S().Infof("Starting server on port 1337")
	err := router.Run(":1337")
	if err != nil {
		zap.S().Fatalf("Error starting server: %s", err.Error())
	}
}

type RequestBody struct {
	Code string      `json:"code"`
	Cid  int         `json:"cid"`
	Adr  string      `json:"adr"`
	Data RequestData `json:"data"`
}

type RequestData struct {
	DataToSend []string `json:"datatosend"`
}

type ResponseBody struct {
	Data map[string]ResponseData `json:"data"`
	Cid  int                     `json:"cid"`
	Code int                     `json:"code"`
}

type ResponseData struct {
	Data interface{} `json:"data,omitempty"`
	Code int         `json:"code"`
}

func Handler(c *gin.Context) {
	var requestBody RequestBody
	var responseBody ResponseBody
	responseBody.Data = make(map[string]ResponseData)

	err := c.ShouldBindJSON(&requestBody)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		zap.S().Errorf("Error parsing request body: %s", err.Error())
		return
	}

	zap.S().Debugf("Request body: %+v", requestBody)

	if requestBody.Adr == "/getdatamulti" {
		if stringInSlice("/deviceinfo/serialnumber/", requestBody.Data.DataToSend) {
			responseBody.Data["/deviceinfo/serialnumber/"] = ResponseData{
				Data: "000201610192",
				Code: 200,
			}
		}
		if stringInSlice("/deviceinfo/productcode/", requestBody.Data.DataToSend) {
			responseBody.Data["/deviceinfo/productcode/"] = ResponseData{
				Data: "AL1350",
				Code: 200,
			}
		}
		if stringInSlice("/iolinkmaster/port[1]/iolinkdevice/deviceid", requestBody.Data.DataToSend) {
			responseBody.Data["1_deviceid"] = ResponseData{
				Code: 200,
				Data: 278531,
			}
		}
		if stringInSlice("/iolinkmaster/port[1]/iolinkdevice/vendorid", requestBody.Data.DataToSend) {
			responseBody.Data["1_vendorid"] = ResponseData{
				Code: 200,
				Data: 42,
			}
		}
		if stringInSlice("/iolinkmaster/port[1]/mode", requestBody.Data.DataToSend) {
			responseBody.Data["1_mode"] = ResponseData{
				Code: 200,
				Data: 0x0003,
			}
		}
		if stringInSlice("/iolinkmaster/port[1]/iolinkdevice/pdin", requestBody.Data.DataToSend) {
			responseBody.Data = map[string]ResponseData{
				"/iolinkmaster/port[1]/iolinkdevice/pdin": {
					Code: 200,
					Data: randString(4),
				},
			}
		}
		responseBody.Code = 200
		responseBody.Cid = requestBody.Cid
	}

	zap.S().Debugf("Response body: %+v", responseBody)
	c.JSON(200, responseBody)
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

const bits = "01"

func randString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = bits[rand.Intn(len(bits))]
	}
	return string(b)
}
