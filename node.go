package main

type Node struct {
        Parent, FirstChild, LastChild, PrevSibling, NextSibling *Node

        Type      NodeType
        Data      string
        Attr      []Attribute
}

type Attribute struct{}