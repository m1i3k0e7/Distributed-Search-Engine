package trie

import (
	"fmt"
	"strings"

	cmap "github.com/orcaman/concurrent-map/v2"
)

type Node struct {
	IsEnd          bool                              `json:"is_end"`
	Children       cmap.ConcurrentMap[string, *Node] `json:"children"`
	ChildrenRecall map[string]*Node                  `json:"children_recall"`
	Parent 	       *Node                             `json:"parent"`
	Word           string                            `json:"word"`
}

func NewTrieNode(word string, parent *Node) *Node {
	m := cmap.New[*Node]()
	return &Node{
		IsEnd:    false,
		Children: m,
		Parent:   parent,
		Word:    word,
	}
}

type Trie struct {
	Root *Node `json:"root"`
}

func NewTrie() *Trie {
	return &Trie{Root: NewTrieNode("", nil)}
}

func (trie *Trie) Insert(title string) {
	splitedTitle := strings.Split(title, " ")
	words := make([]string, 0)
	for _, word := range splitedTitle {
		if len(word) > 0 {
			words = append(words, strings.ToLower(word))
		}
	}

	node := trie.Root
	for i := 0; i < len(words); i++ {
		word := words[i]
		if _, ok := node.Children.Get(word); !ok {
			node.Children.Set(word, NewTrieNode(word, nil))
		}
		node, _ = node.Children.Get(word)
	}
	node.IsEnd = true
}

func (trie *Trie) Search(query string) bool {
	splitedQuery := strings.Split(query, " ")
	words := make([]string, 0)
	for _, word := range splitedQuery {
		if len(word) > 0 {
			words = append(words, strings.ToLower(word))
		}
	}

	node := trie.Root
	for i := 0; i < len(words); i++ {
		word := words[i]
		if _, ok := node.Children.Get(word); !ok {
			return false
		}
		node, _ = node.Children.Get(word)
	}

	return node.IsEnd
}

func (trie *Trie) StartsWith(prefix string) bool {
	splitedPrefix := strings.Split(prefix, " ")
	prefixes := make([]string, 0)
	for _, word := range splitedPrefix {
		if len(word) > 0 {
			prefixes = append(prefixes, strings.ToLower(word))
		}
	}

	node := trie.Root
	for i := 0; i < len(prefixes); i++ {
		c := prefixes[i]
		if _, ok := node.Children.Get(c); !ok {
			return false
		}
		node, _ = node.Children.Get(c)
	}

	return true
}

func (trie *Trie) FindAllByPrefix(prefix string) []string {
	splitedPrefix := strings.Split(prefix, " ")
	prefixes := make([]string, 0)
	for _, word := range splitedPrefix {
		if len(word) > 0 {
			prefixes = append(prefixes, strings.ToLower(word))
		}
	}

	node := trie.Root
	for i := 0; i < len(prefixes); i++ {
		c := prefixes[i]
		if _, ok := node.Children.Get(c); !ok {
			return nil
		}
		node, _ = node.Children.Get(c)
	}
	words := make([]string, 0)
	trie.dfs(node, prefix, &words)
	
	return words
}

func (trie *Trie) dfs(node *Node, word string, words *[]string) {
	if node.IsEnd {
		*words = append(*words, word)
	}

	for c, child := range node.Children.Items() {
		trie.dfs(child, word+c, words)
	}
}

func (trie *Trie) Merge(other *Trie) {
	if other == nil {
		return
	}

	var mergeNodes func(n1, n2 *Node)
	mergeNodes = func(n1, n2 *Node) {
		for c, child := range n2.Children.Items() {
			if v, ok := n1.Children.Get(c); ok {
				mergeNodes(v, child)
			} else {
				n1.Children.Set(c, child)
			}
		}
		n1.IsEnd = n1.IsEnd || n2.IsEnd
	}

	mergeNodes(trie.Root, other.Root)
}

func traverse(node *Node, prefix string) {
	if node.IsEnd {
		fmt.Println(prefix)
	}

	for c, child := range node.Children.Items() {
		traverse(child, prefix+c)
	}
}

func (trie *Trie) Traverse() {
	traverse(trie.Root, "")
}