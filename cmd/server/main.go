package main

import (
	"flag"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/m1i3k0e7/distributed-search-engine/internal/config"
	"github.com/m1i3k0e7/distributed-search-engine/internal/handler"
	"github.com/m1i3k0e7/distributed-search-engine/internal/indexing/kvdb"
	"github.com/rs/cors"
)

var (
	mode         = flag.Int("mode", 1, "1-standalone web server, 2-grpc index server, 3-distributed web server")
	rebuildIndex = flag.Bool("index", false, "rebuild index from csv file when server starting")
	port         = flag.Int("port", 0, "port for web server or grpc index server")
	dbPath       = flag.String("dbPath", "", "path to the local kvdb database")
	totalWorkers = flag.Int("totalWorkers", 0, "total number of index workers in the distributed system")
	workerIndex  = flag.Int("workerIndex", 0, "index worker id in the distributed system")
	trieDBPath   = "../../internal/indexing/trie/storage/trie_bolt" // Path to the trie database file
)

var (
	dbType      = kvdb.BOLT
	csvFilesDir     = config.RootPath + "/../data/test"
	etcdServers = []string{"127.0.0.1:2379"}
)

func StartGin() {
	engine := gin.Default()
	gin.SetMode(gin.ReleaseMode)

	engine.Use(handler.GetUserInfo)

	engine.POST("/search", handler.SearchAll)
	engine.POST("/associate", handler.AssociateQuery)

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173", "http://127.0.0.1:5173"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowCredentials: true,
	})
	handler := c.Handler(engine)

	addr := "127.0.0.1:" + strconv.Itoa(*port)
	log.Printf("Starting server on %s", addr)
	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatalf("ListenAndServe: %v", err)
	}
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