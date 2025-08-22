package indexing

import (
	"encoding/csv"
	"io"
	"log"
	"os"
	"strconv"
	"strings"

	// "time"
	uuid "github.com/google/uuid"
	farmhash "github.com/leemcloughlin/gofarmhash"
	search_proto "github.com/m1i3k0e7/distributed-search-engine/api/proto/search"
	"github.com/m1i3k0e7/distributed-search-engine/internal/search/common"
	"github.com/m1i3k0e7/distributed-search-engine/pkg/logger"
	"github.com/m1i3k0e7/distributed-search-engine/pkg/preprocessing"
	"github.com/m1i3k0e7/distributed-search-engine/pkg/trie"
	proto "google.golang.org/protobuf/proto"
	"github.com/m1i3k0e7/distributed-search-engine/internal/indexing/trie"
)

func BuildIndexFromDir(csvFilesDir string, indexer IIndexer, totalWorkers, workerIndex int) {
	files, err := os.ReadDir(csvFilesDir)
	if err != nil {
		log.Printf("read dir %s failed: %s", csvFilesDir, err)
		return
	}

	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".csv") {
			continue
		}
		csvFile := csvFilesDir + "/" + file.Name()
		log.Printf("start to build index from file: %s", csvFile)
		BuildIndexFromFile(csvFile, indexer, totalWorkers, workerIndex)
	}
}

// Write all documents in csvFile to indexer
//
// totalWorkers: total number of workers; workerIndex: index of this worker, set to 0 if only one worker
func BuildIndexFromFile(csvFile string, indexer IIndexer, totalWorkers, workerIndex int) {
	file, err := os.Open(csvFile)
	if err != nil {
		log.Printf("open file %s failed: %s", csvFile, err)
		return
	}
	defer file.Close()

	queryTrie := trie.NewTrie();
	reader := csv.NewReader(file)
	progress := 0
	for {
		record, err := reader.Read()
		if err != nil {
			if err != io.EOF {
				log.Printf("read record failed: %s", err)
			}
			break
		}

		if len(record) < 9 {
			continue
		}

		docId := uuid.New().String()
		
		if totalWorkers > 0 && int(farmhash.Hash32WithSeed([]byte(docId), 0)) % totalWorkers != workerIndex {
			log.Printf("skip document %s for worker %d", docId, workerIndex)
			continue
		}
		
		product := &search_proto.Product{
			Id:     docId,
			Name:  record[0],
			Category: record[1],
			Image: record[3],
		}
		
		n, _ := strconv.ParseFloat(record[5], 64)
		product.Ratings = float64(n)
		m, _ := strconv.Atoi(record[6])
		product.NoRatings = int32(m)
		
		n, _ = strconv.ParseFloat(record[7], 64)
		product.DiscountPrice = float64(n)
		n, _ = strconv.ParseFloat(record[8], 64)
		product.ActualPrice = float64(n)
		
		queryTrie.Insert(record[0]);
	
		keywords := preprocessing.PreprocessForLargeDataset(record[0])
		if len(keywords) > 0 {
			for _, word := range keywords {
				word = strings.TrimSpace(word)
				if len(word) > 0 {
					product.Keywords = append(product.Keywords, strings.ToLower(word))
				}
			}
		}
		AddProduct2Index(product, indexer)
		progress++
		if progress % 100 == 0 {
			logger.Log.Printf("processed %d documents", progress)
		}
	}

	err = storeTrieToDB(queryTrie)
	if err != nil {
		panic(err)
	}

	logger.Log.Printf("add %d documents to index totally", progress)
}

func AddProduct2Index(product *search_proto.Product, indexer IIndexer) {
	doc := search_proto.Document{Id: product.Id}
	bs, err := proto.Marshal(product)
	if err == nil {
		doc.Bytes = bs
	} else {
		log.Printf("serielize video failed: %s", err)
		return
	}

	keywords := make([]*search_proto.Keyword, 0, len(product.Keywords))
	for _, word := range product.Keywords {
		keywords = append(keywords, &search_proto.Keyword{Field: "content", Word: strings.ToLower(word)})
	}
	
	doc.Keywords = keywords
	doc.BitsFeature = common.GetClassBits([]string{product.Category}) | common.GetClassBits(product.Keywords)

	indexer.AddDoc(doc)
}

func storeTrieToDB(trie *trie.Trie) error {
	trieDB, err := storage.NewTrieDB("../../internal/indexing/trie/storage/trie_bolt" )
	if err != nil {
		return err
	}

	err = trieDB.StoreTrie(trie)
	if err != nil {
		return err
	}

	err = trieDB.Close()
	if err != nil {
		return err
	}

	return nil
}
