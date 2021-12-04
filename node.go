package main

import (
	"github.com/jonasknobloch/jinn/pkg/tree"
	"strings"
)

type NodeType int

const (
	MajorNode NodeType = iota + 1
	SubNode
	FinalNode
)

type Node struct {
	n     string
	r     string
	t     string
	p     []int
	tr    *tree.Tree
	f     []string
	k     int
	l     int
	nType NodeType
}

func (n *Node) Substring() string {
	return strings.Join(n.f[n.k:n.k+n.l], " ")
}
