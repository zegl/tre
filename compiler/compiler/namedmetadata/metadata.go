package namedmetadata

import "github.com/llir/llvm/ir/types"

type Specialized struct {
	// Metadata name.
	Name string
	// Associated metadata.
	//Metadata []*Metadata
	Data map[string]string
}

func (s Specialized) Ident() string {
	return "yolo"
}

func (s Specialized) MetadataNode() {

}

func (s Specialized) Type() types.Type {
	return &types.MetadataType{}
}

// Def returns the LLVM syntax representation of the definition of the named
// metadata.
// func (md *Named) Def() string {
// 	buf := &bytes.Buffer{}
// 	buf.WriteString("!{")
// 	for i, metadata := range md.Metadata {
// 		if i != 0 {
// 			buf.WriteString(", ")
// 		}
// 		buf.WriteString(metadata.Ident())
// 	}
// 	buf.WriteString("}")
// 	return buf.String()
// }
//
