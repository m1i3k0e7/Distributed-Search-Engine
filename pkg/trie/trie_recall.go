package trie

import (
	"encoding/json"
	// "log"
	"slices"
	"strings"

	"github.com/pkg/errors"
)

func (trie *Trie) FindAllByPrefixForRecall(prefix string) []string {
	splitedPrefix := strings.Split(prefix, " ")
	prefixes := make([]string, 0)
	for _, word := range splitedPrefix {
		if len(word) > 0 {
			prefixes = append(prefixes, strings.ToLower(word))
		}
	}

	words := make([]string, 0)
	node := trie.Root
	cuttedPrefix := ""
	lastIncluded := true
	for i := 0; i < len(prefixes); i++ {
		c := prefixes[i]
		if _, ok := node.ChildrenRecall[c]; !ok && i < len(prefixes)-1 {
			return nil
		} else if i == len(prefixes)-1 {
			lastIncluded = false
			for word := range node.ChildrenRecall {
				if strings.HasPrefix(word, c) {
					trie.dfsForRecall(node.ChildrenRecall[word], cuttedPrefix + " " + word, &words)
				}
			}
			break
		}

		node = node.ChildrenRecall[c]
		cuttedPrefix = cuttedPrefix + " " + c
	}
	
	if lastIncluded {
		trie.dfsForRecall(node, prefix, &words)
	}

	return words
}

func (trie *Trie) dfsForRecall(node *Node, word string, words *[]string) {
	if len(*words) >= 10 {
		return
	}

	if node.IsEnd && !slices.Contains(*words, word) {
		*words = append(*words, word)
	}

	for c, child := range node.ChildrenRecall {
		trie.dfsForRecall(child, word + " " + c, words)
	}
}

func (trie *Trie) SearchForRecall(word string) bool {
	words := []rune(word)
	node := trie.Root
	for i := 0; i < len(words); i++ {
		c := string(words[i])
		if _, ok := node.ChildrenRecall[c]; !ok {
			return false
		}
		node = node.ChildrenRecall[c] // nolint:golint,gosimple
	}

	return node.IsEnd
}

func ParseTrieNode(str string) (*Node, error) {
	var data map[string]interface{}
	err := json.Unmarshal([]byte(str), &data)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal data")
	}

	node := &Node{
		IsEnd:          false,
		ChildrenRecall: make(map[string]*Node),
	}

	for key, value := range data {
		childData, ok := value.(map[string]interface{})
		if !ok {
			return nil, errors.Wrap(errors.Errorf("invalid child data for key: %s", key), "failed to assert type")
		}

		childNode, err := parseTrieNodeChild(childData)
		if err != nil {
			return nil, errors.WithMessage(err, "parseTrieNodeChild error")
		}

		node.ChildrenRecall[key] = childNode
	}

	return node, nil
}

func parseTrieNodeChild(data map[string]interface{}) (*Node, error) {
	node := &Node{
		IsEnd:          false,
		ChildrenRecall: make(map[string]*Node),
	}

	isEnd, ok := data["is_end"].(bool)
	if ok {
		node.IsEnd = isEnd
	}

	childrenData, ok := data["children_recall"].(map[string]interface{})
	if !ok {
		return nil, errors.Wrap(errors.New("invalid children data"), "failed to assert type")
	}

	for key, value := range childrenData {
		childData, ok := value.(map[string]interface{})
		if !ok {
			return nil, errors.Wrap(errors.Errorf("invalid child data for key: %s", key), "failed to assert type")
		}

		childNode, err := parseTrieNodeChild(childData)
		if err != nil {
			return nil, errors.WithMessage(err, "parseTrieNodeChild error")
		}

		node.ChildrenRecall[key] = childNode
	}

	return node, nil
}

func (trie *Trie) TraverseForRecall() {
	traverseForRecall(trie.Root, "")
}

func traverseForRecall(node *Node, prefix string) {
	if node.IsEnd {
		return
	}

	for c, child := range node.ChildrenRecall {
		traverseForRecall(child, prefix+c)
	}
}