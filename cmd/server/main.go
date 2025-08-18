package main

import (
	"flag"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/m1i3k0e7/distributed-search-engine/internal/indexing/kvdb"
	"github.com/m1i3k0e7/distributed-search-engine/internal/config"
	"github.com/m1i3k0e7/distributed-search-engine/internal/handler"
)

var (
	mode         = flag.Int("mode", 1, "1-standalone web server, 2-grpc index server, 3-distributed web server")
	rebuildIndex = flag.Bool("index", false, "rebuild index from csv file when server starting")
	port         = flag.Int("port", 0, "port for web server or grpc index server")
	dbPath       = flag.String("dbPath", "", "path to the local kvdb database")
	totalWorkers = flag.Int("totalWorkers", 0, "total number of index workers in the distributed system")
	workerIndex  = flag.Int("workerIndex", 0, "index worker id in the distributed system")
)

var (
	dbType      = kvdb.BOLT
	csvFile     = config.RootPath + "/../data/bili_video.csv"
	etcdServers = []string{"127.0.0.1:2379"}
)

func StartGin() {
	engine := gin.Default()
	gin.SetMode(gin.ReleaseMode)

	engine.Use(handler.GetUserInfo)
	// classes := [...]string{"Home, Kitchen, Pets", "Grocery & Gourmet Foods", "Men's Shoes", "Kids's Fashion", "Women's Shoes", "Accessories", "Bags & Luggage", "Industrial Supplies", "Stores", "Men's Clothing", "Women's Clothing", "TV, Audio & Cameras", "Beauty & Health", "Home & Kitchen", "Pet Supplies", "Music", "Toys & Baby Products", "Sports & Fitness", "Car & Motorbike", "Appliances"}
	// engine.GET("/", func(ctx *gin.Context) {
	// 	ctx.HTML(http.StatusOK, "search.html", classes)
	// })
	// engine.GET("/up", func(ctx *gin.Context) {
	// 	ctx.HTML(http.StatusOK, "up_search.html", classes)
	// })

	engine.POST("/search", handler.SearchAll)
	engine.Run("127.0.0.1:" + strconv.Itoa(*port))
}

func main() {
	flag.Parse()

	switch *mode {
	case 1, 3:
		WebServerMain(*mode) //1. standalone mode 3：distributed mode
		StartGin()
	case 2:
		GrpcIndexerMain() // 2: start grpc index server
	}
}

// go run ./example/main -mode=1 -index=true -port=5678 -dbPath=data/local_db/video_bolt
// go run ./example/main -mode=2 -index=true -port=5600 -dbPath=data/local_db/video_bolt -totalWorkers=2 -workerIndex=0
// go run ./example/main -mode=2 -index=true -port=5601 -dbPath=data/local_db/video_bolt -totalWorkers=2 -workerIndex=1
// go run ./example/main -mode=3 -index=false -port=5678
