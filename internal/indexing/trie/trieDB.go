package storage

import (
	"bytes"
	"log"

	"github.com/m1i3k0e7/distributed-search-engine/internal/indexing/kvdb"
	"github.com/m1i3k0e7/distributed-search-engine/pkg/trie"
)


type TrieDB struct {
	db kvdb.IKeyValueDB
	path string
}

func NewTrieDB(dbPath string) (*TrieDB, error) {
	newDB, err := kvdb.GetKvDb(kvdb.BOLT, dbPath)
	if err != nil {
		return nil, err
	}
	trieDB := &TrieDB{
		db:   newDB,
		path: dbPath,
	}

	err = trieDB.db.SetBucket("trie")
	if err != nil {
		trieDB.Close()
		return nil, err
	}

	return trieDB, nil
}

func (t *TrieDB) StoreTrie(trie *trie.Trie) error {
	if trie == nil {
		return nil
	}

	trieJson, err := trie.MarshalJSON()
	if err != nil {
		return err
	}

	err = t.db.Set([]byte("trie"), trieJson)

	return err
}

func (t *TrieDB) AssociateQuery(query string) ([]string, error) {
	value, err := t.db.Get([]byte("trie"))
	if err != nil {
		log.Printf("Error retrieving trie from database: %v", err)
		return nil, err
	}

	if value == nil {
		log.Println("No trie found in the database.")
		return nil, nil // No trie stored
	}

	trieInstance := &trie.Trie{}
	replacedValue := bytes.Replace(value, []byte("children"), []byte("children_recall"), -1)
	node, err := trie.ParseTrieNode(string(replacedValue))
	if err != nil {
		log.Printf("Error parsing trie node: %v", err)
		return nil, err
	}

	rootNode := trie.NewTrieNode("", nil)
	rootNode.ChildrenRecall = node.ChildrenRecall["root"].ChildrenRecall
	trieInstance.Root = rootNode

	associatedQueries := trieInstance.FindAllByPrefixForRecall(query)

	return associatedQueries, nil
}

func (t *TrieDB) IterTrie() error {
	value, err := t.db.Get([]byte("trie"))
	if err != nil {
		log.Printf("Error retrieving trie from database: %v", err)
		return err
	}

	if value == nil {
		log.Println("No trie found in the database.")
		return nil // No trie stored
	}

	trieInstance := &trie.Trie{}
	replacedValue := bytes.Replace(value, []byte("children"), []byte("children_recall"), -1)
	node, err := trie.ParseTrieNode(string(replacedValue))
	if err != nil {
		log.Printf("Error parsing trie node: %v", err)
		return err
	}

	rootNode := trie.NewTrieNode("root", nil)
	rootNode.ChildrenRecall = node.ChildrenRecall["root"].ChildrenRecall
	trieInstance.Root = rootNode
	for key, _ := range trieInstance.Root.ChildrenRecall {
		log.Printf("Trie word: %s", key)
		log.Printf("Trie node: %v", trieInstance.Root.ChildrenRecall[key].Word)
	}

	return nil
}

func (t *TrieDB) Close() error {
	return t.db.Close()
}