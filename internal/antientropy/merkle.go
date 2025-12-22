package antientropy

import (
	"crypto/sha256"
	"encoding/hex"
	"sort"
)

type MerkleNode struct {
	Hash     string
	Left     *MerkleNode
	Right    *MerkleNode
	StartKey string
	EndKey   string
	IsLeaf   bool
	Keys     []string
}

type MerkleTree struct {
	root     *MerkleNode
	depth    int
	keyCount int
}

func Build(keyHashses map[string]string, depth int) *MerkleTree {
	if len(keyHashses) == 0 {
		return &MerkleTree{
			root:  nil,
			depth: depth,
		}
	}

	// sort keys
	keys := make([]string, 0, len(keyHashses))
	for k := range keyHashses {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	root := buildNode(keys, keyHashses, depth, 0, "", "\xff")

	return &MerkleTree{
		root:     root,
		depth:    depth,
		keyCount: len(keys),
	}
}

func buildNode(keys []string, keyHashes map[string]string, maxDepth, currDepth int, startKey, endKey string) *MerkleNode {
	if len(keys) == 0 {
		return &MerkleNode{
			Hash:     hashEmpty(),
			StartKey: startKey,
			EndKey:   endKey,
			IsLeaf:   true,
		}
	}

	if currDepth >= maxDepth || len(keys) <= 4 {
		return createLeaf(keys, keyHashes, startKey, endKey)
	}

	mid := len(keys) / 2
	midKey := keys[mid]

	leftKeys := keys[:mid]
	rightKeys := keys[mid:]

	left := buildNode(leftKeys, keyHashes, maxDepth, currDepth+1, startKey, midKey)
	right := buildNode(rightKeys, keyHashes, maxDepth, currDepth+1, midKey, endKey)

	combinedHash := hashStrings(left.Hash, right.Hash)

	return &MerkleNode{
		Hash:     combinedHash,
		Left:     left,
		Right:    right,
		StartKey: startKey,
		EndKey:   endKey,
		IsLeaf:   false,
	}
}

func createLeaf(keys []string, keyHashes map[string]string, startKey, endKey string) *MerkleNode {
	var hashes []string
	for _, k := range keys {
		hashes = append(hashes, keyHashes[k])
	}

	return &MerkleNode{
		Hash:     hashStrings(hashes...),
		StartKey: startKey,
		EndKey:   endKey,
		IsLeaf:   true,
		Keys:     keys,
	}
}

type KeyRange struct {
	Start string
	End   string
}

func Compare(local, remote *MerkleTree) []KeyRange {
	if local.root == nil && remote.root == nil {
		return nil
	}

	if local.root == nil {
		return []KeyRange{
			{
				Start: "",
				End:   "\xff",
			},
		}
	}

	if remote.root == nil {
		return []KeyRange{
			{
				Start: "",
				End:   "\xff",
			},
		}
	}

	return compareNodes(local.root, remote.root)
}

func compareNodes(local, remote *MerkleNode) []KeyRange {
	if local.Hash == remote.Hash {
		return nil
	}

	if local.IsLeaf || remote.IsLeaf {
		return []KeyRange{
			{
				Start: local.StartKey,
				End:   local.EndKey,
			},
		}
	}

	var ranges []KeyRange
	ranges = append(ranges, compareNodes(local.Left, remote.Left)...)
	ranges = append(ranges, compareNodes(local.Right, remote.Right)...)

	return ranges
}

func hashEmpty() string {
	h := sha256.Sum256([]byte{})
	return hex.EncodeToString(h[:])
}

func hashStrings(strs ...string) string {
	h := sha256.New()
	for _, s := range strs {
		h.Write([]byte(s))
	}

	return hex.EncodeToString(h.Sum(nil))
}

func (mt *MerkleTree) GetRootHash() string {
	if mt.root == nil {
		return hashEmpty()
	}

	return hashStrings(mt.root.Hash)
}
