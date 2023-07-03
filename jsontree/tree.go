package jsontree

import (
	"github.com/go-openapi/jsonpointer"
)

//
//type Node interface {
//	PrimitiveNode | ArrayNode | ObjectNode
//}
//
//type PrimitiveNode struct {
//	Name string
//}
//
//type ArrayNode struct {
//	Name     string
//	Children []interface{}
//}
//
//type ObjectNode struct {
//	Name     string
//	Children map[string]interface{}
//}
//
//func NewRootNode() ObjectNode {
//	return ObjectNode{
//		Children: make(map[string]interface{}),
//	}
//}

type Node struct {
	Name     string
	Children map[string]Node
}

func NewNode(name string) Node {
	return Node{
		Name:     name,
		Children: make(map[string]Node),
	}
}

// ParseJsonPtr does not have context about the type of the created node, it will be created to ObjectNode.
func ParseJsonPtr(root *Node, ptr jsonpointer.Pointer) (Node, error) {
	tks := ptr.DecodedTokens()
	cur := *root
	for _, tk := range tks {
		if _, ok := cur.Children[tk]; !ok {
			cur.Children[tk] = NewNode(tk)
		}
		cur = cur.Children[tk]
	}
	return *root, nil
}
