package schema

import (
	_ "embed"
	"fmt"
	"github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/node/bindnode"
	"github.com/ipld/go-ipld-prime/schema"
)

var (

	// ProllyNodePrototype represents the IPLD node prototype of Metadata.
	// See: bindnode.Prototype.
	ProllyNodePrototype schema.TypedPrototype

	ChunkConfigPrototype schema.TypedPrototype

	//go:embed schema.ipldsch
	schemaBytes []byte
)

func init() {
	typeSystem, err := ipld.LoadSchemaBytes(schemaBytes)
	if err != nil {
		panic(fmt.Errorf("failed to load schema: %w", err))
	}
	ProllyNodePrototype = bindnode.Prototype(&ProllyNode{}, typeSystem.TypeByName("ProllyNode"))

	ChunkConfigPrototype = bindnode.Prototype(&ChunkConfig{}, typeSystem.TypeByName("ChunkConfig"))
}
