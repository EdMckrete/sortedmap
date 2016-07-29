package sortedmap

import "fmt"

// BPlusTree interface declares the available methods available for a B+Tree
type BPlusTree interface {
	SortedMap
	Flush(andPurge bool) (rootObjectNumber uint64, rootObjectOffset uint64, rootObjectLength uint64, err error)
	Purge() (err error)
}

// BPlusTreeCallbacks specifies the interface to a set of callbacks provided by the client
type BPlusTreeCallbacks interface {
	GetNode(objectNumber uint64, objectOffset uint64, objectLength uint64) (nodeByteSlice []byte, err error)
	PutNode(nodeByteSlice []byte) (objectNumber uint64, objectOffset uint64, err error)
	PackKey(key Key) (packedKey []byte, err error)
	UnpackKey(payloadData []byte) (key Key, bytesConsumed uint64, err error)
	PackValue(value Value) (packedValue []byte, err error)
	UnpackValue(payloadData []byte) (value Value, bytesConsumed uint64, err error)
}

// New is used to construct an in-memory B+Tree supporting the Tree interface
//
// Note that if the B+Tree will reside only in memory, callback argument may be nil.
// That said, there is no advantage to using a B+Tree for an in-memory collection over
// the llrb-provided collection implementing the same APIs.
func NewBPlusTree(maxKeysPerNode uint64, compare Compare, callbacks BPlusTreeCallbacks) (tree BPlusTree) {
	minKeysPerNode := maxKeysPerNode >> 1
	if (0 == maxKeysPerNode) != ((2 * minKeysPerNode) != maxKeysPerNode) {
		err := fmt.Errorf("maxKeysPerNode (%v) invalid - must be a positive number that is a multiple of 2", maxKeysPerNode)
		panic(err)
	}

	rootNode := &btreeNodeStruct{
		objectNumber:        0, //              To be filled in once root node is posted
		objectOffset:        0, //              To be filled in once root node is posted
		objectLength:        0, //              To be filled in once root node is posted
		items:               0,
		loaded:              true, //           Special case in that objectNumber == 0 means it has no onDisk copy
		dirty:               true,
		root:                true,
		leaf:                true,
		tree:                nil, //            To be set just below
		parentNode:          nil,
		kvLLRB:              NewLLRBTree(compare),
		nonLeafLeftChild:    nil,
		rootPrefixSumChild:  nil,
		prefixSumItems:      0,   //            Not applicable to root node
		prefixSumParent:     nil, //            Not applicable to root node
		prefixSumLeftChild:  nil, //            Not applicable to root node
		prefixSumRightChild: nil, //            Not applicable to root node
	}

	treePtr := &btreeTreeStruct{
		minKeysPerNode:     minKeysPerNode,
		maxKeysPerNode:     maxKeysPerNode,
		Compare:            compare,
		BPlusTreeCallbacks: callbacks,
		root:               rootNode,
	}

	rootNode.tree = treePtr

	tree = treePtr

	return
}

// OldBPlusTree is used to re-construct a B+Tree previously persisted
func OldBPlusTree(rootObjectNumber uint64, rootObjectOffset uint64, rootObjectLength uint64, compare Compare, callbacks BPlusTreeCallbacks) (tree BPlusTree) {
	rootNode := &btreeNodeStruct{
		objectNumber:        rootObjectNumber,
		objectOffset:        rootObjectOffset,
		objectLength:        rootObjectLength,
		items:               0, //                 To be filled in once root node is loaded
		loaded:              false,
		dirty:               false,
		root:                true,
		leaf:                true, //              To be updated once root node is loaded
		tree:                nil,  //              To be set just below
		parentNode:          nil,
		kvLLRB:              nil, //               To be filled in once root node is loaded
		nonLeafLeftChild:    nil, //               To be filled in once root node is loaded
		rootPrefixSumChild:  nil, //               To be filled in once root node is loaded (nil if root is also leaf)
		prefixSumItems:      0,   //               Not applicable to root node
		prefixSumParent:     nil, //               Not applicable to root node
		prefixSumLeftChild:  nil, //               Not applicable to root node
		prefixSumRightChild: nil, //               Not applicable to root node
	}

	treePtr := &btreeTreeStruct{
		minKeysPerNode:     0, //                  To be filled in once root node is loaded
		maxKeysPerNode:     0, //                  To be filled in once root node is loaded
		Compare:            compare,
		BPlusTreeCallbacks: callbacks,
		root:               rootNode,
	}

	rootNode.tree = treePtr

	treePtr.loadNode(rootNode)

	tree = treePtr

	return
}