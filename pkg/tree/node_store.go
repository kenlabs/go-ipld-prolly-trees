package tree

import (
	"context"
	lru "github.com/hashicorp/golang-lru"
	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-datastore"
	blockstore "github.com/ipfs/go-ipfs-blockstore"
	"github.com/ipld/go-ipld-prime"
	_ "github.com/ipld/go-ipld-prime/codec/dagcbor"
	"github.com/ipld/go-ipld-prime/linking"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	"github.com/kenlabs/go-ipld-prolly-trees/pkg/tree/linksystem"
)

type StoreConfig struct {
	CacheSize int
}

var _ NodeStore = &BlockNodeStore{}

type BlockNodeStore struct {
	bs    blockstore.Blockstore
	lsys  *ipld.LinkSystem
	cache *lru.Cache
}

func (ns *BlockNodeStore) WriteProof(ctx context.Context, prf Proof, prefix *cid.Prefix) (cid.Cid, error) {
	var linkProto cidlink.LinkPrototype
	if prefix == nil {
		// default linkproto
		linkProto = DefaultLinkProto
	} else {
		linkProto = cidlink.LinkPrototype{Prefix: *prefix}
	}
	ipldNode, err := prf.ToNode()
	if err != nil {
		return cid.Undef, err
	}
	lnk, err := ns.lsys.Store(ipld.LinkContext{Ctx: ctx}, linkProto, ipldNode)
	if err != nil {
		return cid.Undef, err
	}
	c := lnk.(cidlink.Link).Cid

	go func() {
		if ns.cache != nil {
			ns.cache.Add(c, prf)
		}
	}()

	return c, nil
}

func (ns *BlockNodeStore) ReadProof(ctx context.Context, c cid.Cid) (Proof, error) {
	var inCache bool
	if ns.cache != nil {
		var res interface{}
		res, inCache = ns.cache.Get(c)
		if inCache {
			return res.(Proof), nil
		}
	}
	nd, err := ns.lsys.Load(ipld.LinkContext{Ctx: ctx}, cidlink.Link{Cid: c}, ProofPrototype.Representation())
	if err != nil {
		return nil, err
	}

	prf, err := UnwrapProof(nd)
	if err != nil {
		return nil, err
	}

	return *prf, nil
}

func (ns *BlockNodeStore) LinkSystem() *ipld.LinkSystem {
	return ns.lsys
}

func NewBlockNodeStore(bs blockstore.Blockstore, cfg *StoreConfig) (*BlockNodeStore, error) {
	lsys := linksystem.MkLinkSystem(bs)
	ns := &BlockNodeStore{
		bs:   bs,
		lsys: &lsys,
	}
	if cfg == nil {
		cfg = &StoreConfig{}
	}
	if cfg.CacheSize != 0 {
		var err error
		ns.cache, err = lru.New(cfg.CacheSize)
		if err != nil {
			return nil, err
		}
	}
	return ns, nil
}

func (ns *BlockNodeStore) WriteNode(ctx context.Context, nd *ProllyNode, prefix *cid.Prefix) (cid.Cid, error) {
	var linkProto cidlink.LinkPrototype
	if prefix == nil {
		// default linkproto
		linkProto = DefaultLinkProto
	} else {
		linkProto = cidlink.LinkPrototype{Prefix: *prefix}
	}
	ipldNode, err := nd.ToNode()
	if err != nil {
		return cid.Undef, err
	}
	lnk, err := ns.lsys.Store(ipld.LinkContext{Ctx: ctx}, linkProto, ipldNode)
	if err != nil {
		return cid.Undef, err
	}
	c := lnk.(cidlink.Link).Cid

	go func() {
		if ns.cache != nil {
			ns.cache.Add(c, nd)
		}
	}()

	return c, nil
}

func (ns *BlockNodeStore) ReadNode(ctx context.Context, c cid.Cid) (*ProllyNode, error) {
	var inCache bool
	if ns.cache != nil {
		var res interface{}
		res, inCache = ns.cache.Get(c)
		if inCache {
			return res.(*ProllyNode), nil
		}
	}
	nd, err := ns.lsys.Load(ipld.LinkContext{Ctx: ctx}, cidlink.Link{Cid: c}, ProllyNodePrototype.Representation())
	if err != nil {
		return nil, err
	}

	inode, err := UnwrapProllyNode(nd)
	if err != nil {
		return nil, err
	}

	return inode, nil
}

func (ns *BlockNodeStore) WriteTree(ctx context.Context, root *ProllyTree, prefix *cid.Prefix) (cid.Cid, error) {
	var linkProto cidlink.LinkPrototype
	if prefix == nil {
		// default linkproto
		linkProto = DefaultLinkProto
	} else {
		linkProto = cidlink.LinkPrototype{Prefix: *prefix}
	}
	ipldNode, err := root.ToNode()
	if err != nil {
		return cid.Undef, err
	}
	lnk, err := ns.lsys.Store(ipld.LinkContext{Ctx: ctx}, linkProto, ipldNode)
	if err != nil {
		return cid.Undef, err
	}
	c := lnk.(cidlink.Link).Cid

	go func() {
		if ns.cache != nil {
			ns.cache.Add(c, *root)
		}
	}()

	return c, nil
}

func (ns *BlockNodeStore) ReadTree(ctx context.Context, c cid.Cid) (*ProllyTree, error) {
	var inCache bool
	if ns.cache != nil {
		var res interface{}
		res, inCache = ns.cache.Get(c)
		if inCache {
			tree := res.(ProllyTree)
			return &tree, nil
		}
	}
	nd, err := ns.lsys.Load(ipld.LinkContext{Ctx: ctx}, cidlink.Link{Cid: c}, ProllyTreePrototype.Representation())
	if err != nil {
		return nil, err
	}

	root, err := UnwrapProllyTree(nd)
	if err != nil {
		return nil, err
	}
	root.treeCid = &c

	return root, nil
}

func (ns *BlockNodeStore) ReadTreeConfig(ctx context.Context, c cid.Cid) (*TreeConfig, error) {
	icfg, err := ns.lsys.Load(ipld.LinkContext{Ctx: ctx}, cidlink.Link{Cid: c}, ChunkConfigPrototype.Representation())
	if err != nil {
		return nil, err
	}

	cfg, err := UnwrapChunkConfig(icfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

func (ns *BlockNodeStore) WriteTreeConfig(ctx context.Context, cfg *TreeConfig, prefix *cid.Prefix) (cid.Cid, error) {
	var linkProto cidlink.LinkPrototype
	if prefix == nil {
		// default linkproto
		linkProto = DefaultLinkProto
	} else {
		linkProto = cidlink.LinkPrototype{Prefix: *prefix}
	}

	ipldNode, err := cfg.ToNode()
	if err != nil {
		return cid.Undef, err
	}
	lnk, err := ns.lsys.Store(ipld.LinkContext{Ctx: ctx}, linkProto, ipldNode)
	if err != nil {
		return cid.Undef, err
	}
	c := lnk.(cidlink.Link).Cid

	return c, nil
}

func (ns *BlockNodeStore) Close() {
}

func TestMemNodeStore() NodeStore {
	ds := datastore.NewMapDatastore()
	bs := blockstore.NewBlockstore(ds)
	ns, _ := NewBlockNodeStore(bs, &StoreConfig{CacheSize: 1 << 14})
	return ns
}

var _ NodeStore = &LinkSystemNodeStore{}

type LinkSystemNodeStore struct {
	lsys *linking.LinkSystem
}

func (ns *LinkSystemNodeStore) WriteProof(ctx context.Context, prf Proof, prefix *cid.Prefix) (cid.Cid, error) {
	var linkProto cidlink.LinkPrototype
	if prefix == nil {
		// default linkproto
		linkProto = DefaultLinkProto
	} else {
		linkProto = cidlink.LinkPrototype{Prefix: *prefix}
	}
	ipldNode, err := prf.ToNode()
	if err != nil {
		return cid.Undef, err
	}
	lnk, err := ns.lsys.Store(ipld.LinkContext{Ctx: ctx}, linkProto, ipldNode)
	if err != nil {
		return cid.Undef, err
	}
	c := lnk.(cidlink.Link).Cid

	return c, nil
}

func (ns *LinkSystemNodeStore) ReadProof(ctx context.Context, c cid.Cid) (Proof, error) {
	nd, err := ns.lsys.Load(ipld.LinkContext{Ctx: ctx}, cidlink.Link{Cid: c}, ProofPrototype.Representation())
	if err != nil {
		return nil, err
	}

	prf, err := UnwrapProof(nd)
	if err != nil {
		return nil, err
	}

	return *prf, nil
}

func (ns *LinkSystemNodeStore) LinkSystem() *ipld.LinkSystem {
	return ns.lsys
}

func NewLinkSystemNodeStore(lsys *linking.LinkSystem) *LinkSystemNodeStore {
	return &LinkSystemNodeStore{lsys: lsys}
}

func (ns *LinkSystemNodeStore) WriteNode(ctx context.Context, nd *ProllyNode, prefix *cid.Prefix) (cid.Cid, error) {
	var linkProto cidlink.LinkPrototype
	if prefix == nil {
		// default linkproto
		linkProto = DefaultLinkProto
	} else {
		linkProto = cidlink.LinkPrototype{Prefix: *prefix}
	}
	ipldNode, err := nd.ToNode()
	if err != nil {
		return cid.Undef, err
	}
	lnk, err := ns.lsys.Store(ipld.LinkContext{Ctx: ctx}, linkProto, ipldNode)
	if err != nil {
		return cid.Undef, err
	}
	c := lnk.(cidlink.Link).Cid

	return c, nil
}

func (ns *LinkSystemNodeStore) ReadNode(ctx context.Context, c cid.Cid) (*ProllyNode, error) {
	nd, err := ns.lsys.Load(ipld.LinkContext{Ctx: ctx}, cidlink.Link{Cid: c}, ProllyNodePrototype.Representation())
	if err != nil {
		return nil, err
	}

	inode, err := UnwrapProllyNode(nd)
	if err != nil {
		return nil, err
	}

	return inode, nil
}

func (ns *LinkSystemNodeStore) WriteTree(ctx context.Context, root *ProllyTree, prefix *cid.Prefix) (cid.Cid, error) {
	var linkProto cidlink.LinkPrototype
	if prefix == nil {
		// default linkproto
		linkProto = DefaultLinkProto
	} else {
		linkProto = cidlink.LinkPrototype{Prefix: *prefix}
	}
	ipldNode, err := root.ToNode()
	if err != nil {
		return cid.Undef, err
	}
	lnk, err := ns.lsys.Store(ipld.LinkContext{Ctx: ctx}, linkProto, ipldNode)
	if err != nil {
		return cid.Undef, err
	}
	c := lnk.(cidlink.Link).Cid

	return c, nil
}

func (ns *LinkSystemNodeStore) ReadTree(ctx context.Context, c cid.Cid) (*ProllyTree, error) {
	nd, err := ns.lsys.Load(ipld.LinkContext{Ctx: ctx}, cidlink.Link{Cid: c}, ProllyTreePrototype.Representation())
	if err != nil {
		return nil, err
	}

	root, err := UnwrapProllyTree(nd)
	if err != nil {
		return nil, err
	}
	root.treeCid = &c

	return root, nil
}

func (ns *LinkSystemNodeStore) ReadTreeConfig(ctx context.Context, c cid.Cid) (*TreeConfig, error) {
	icfg, err := ns.lsys.Load(ipld.LinkContext{Ctx: ctx}, cidlink.Link{Cid: c}, ChunkConfigPrototype.Representation())
	if err != nil {
		return nil, err
	}

	cfg, err := UnwrapChunkConfig(icfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

func (ns *LinkSystemNodeStore) WriteTreeConfig(ctx context.Context, cfg *TreeConfig, prefix *cid.Prefix) (cid.Cid, error) {
	var linkProto cidlink.LinkPrototype
	if prefix == nil {
		// default linkproto
		linkProto = DefaultLinkProto
	} else {
		linkProto = cidlink.LinkPrototype{Prefix: *prefix}
	}

	ipldNode, err := cfg.ToNode()
	if err != nil {
		return cid.Undef, err
	}
	lnk, err := ns.lsys.Store(ipld.LinkContext{Ctx: ctx}, linkProto, ipldNode)
	if err != nil {
		return cid.Undef, err
	}
	c := lnk.(cidlink.Link).Cid

	return c, nil
}

func (ns *LinkSystemNodeStore) Close() {
}
