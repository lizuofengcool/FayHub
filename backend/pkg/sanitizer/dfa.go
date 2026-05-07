package sanitizer

import (
	"strings"
	"sync"
	"unicode/utf8"
)

type dfaNode struct {
	children map[rune]*dfaNode
	isEnd    bool
}

type DFAMatcher struct {
	root *dfaNode
	mu   sync.RWMutex
}

var defaultMatcher = &DFAMatcher{
	root: &dfaNode{children: make(map[rune]*dfaNode)},
}

func GetMatcher() *DFAMatcher {
	return defaultMatcher
}

func (m *DFAMatcher) AddWord(word string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	node := m.root
	for _, ch := range word {
		if node.children[ch] == nil {
			node.children[ch] = &dfaNode{children: make(map[rune]*dfaNode)}
		}
		node = node.children[ch]
	}
	node.isEnd = true
}

func (m *DFAMatcher) RemoveWord(word string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.removeWord(m.root, []rune(word), 0)
}

func (m *DFAMatcher) removeWord(node *dfaNode, runes []rune, idx int) bool {
	if idx == len(runes) {
		if node.isEnd {
			node.isEnd = false
			return len(node.children) == 0
		}
		return false
	}

	ch := runes[idx]
	child, ok := node.children[ch]
	if !ok {
		return false
	}

	shouldDelete := m.removeWord(child, runes, idx+1)
	if shouldDelete {
		delete(node.children, ch)
		return len(node.children) == 0 && !node.isEnd
	}
	return false
}

func (m *DFAMatcher) BuildFromWords(words []string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.root = &dfaNode{children: make(map[rune]*dfaNode)}
	for _, word := range words {
		node := m.root
		for _, ch := range word {
			if node.children[ch] == nil {
				node.children[ch] = &dfaNode{children: make(map[rune]*dfaNode)}
			}
			node = node.children[ch]
		}
		node.isEnd = true
	}
}

func (m *DFAMatcher) Contains(text string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	runes := []rune(text)
	for i := 0; i < len(runes); i++ {
		node := m.root
		for j := i; j < len(runes); j++ {
			child, ok := node.children[runes[j]]
			if !ok {
				break
			}
			node = child
			if node.isEnd {
				return true
			}
		}
	}
	return false
}

func (m *DFAMatcher) FindAll(text string) []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []string
	runes := []rune(text)
	for i := 0; i < len(runes); i++ {
		node := m.root
		for j := i; j < len(runes); j++ {
			child, ok := node.children[runes[j]]
			if !ok {
				break
			}
			node = child
			if node.isEnd {
				result = append(result, string(runes[i:j+1]))
			}
		}
	}
	return result
}

func (m *DFAMatcher) Replace(text string, replaceChar rune) string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	runes := []rune(text)
	marked := make([]bool, len(runes))

	for i := 0; i < len(runes); i++ {
		node := m.root
		for j := i; j < len(runes); j++ {
			child, ok := node.children[runes[j]]
			if !ok {
				break
			}
			node = child
			if node.isEnd {
				for k := i; k <= j; k++ {
					marked[k] = true
				}
			}
		}
	}

	var builder strings.Builder
	builder.Grow(len(runes))
	for i, r := range runes {
		if marked[i] {
			builder.WriteRune(replaceChar)
		} else {
			builder.WriteRune(r)
		}
	}
	return builder.String()
}

func SanitizeText(input string) string {
	s := input
	s = SanitizeHTML(s)
	s = defaultMatcher.Replace(s, '*')
	return s
}

func HasSensitiveWord(input string) bool {
	return defaultMatcher.Contains(input)
}

func FindSensitiveWords(input string) []string {
	return defaultMatcher.FindAll(input)
}

func ValidateText(input string, minLen, maxLen int) (string, bool) {
	length := utf8.RuneCountInString(input)
	if length < minLen || length > maxLen {
		return "", false
	}

	sanitized := SanitizeText(input)
	if sanitized != input {
		return sanitized, true
	}

	return input, true
}

func SanitizeTextPreserve(input string) string {
	s := SanitizeHTML(input)
	if defaultMatcher.Contains(s) {
		return defaultMatcher.Replace(s, '*')
	}
	return s
}
