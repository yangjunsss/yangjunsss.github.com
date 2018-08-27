package main
import (
  "fmt"
)

type Node struct{
  v int
  l *Node
  r *Node
}

func newBTree(depth int,curr int,root *Node) {
  root.v = curr
  if curr >= depth{
    root.l,root.r = nil,nil
    return
  }
  root.l,root.r = &Node{},&Node{}
  newBTree(depth,curr+1,root.l)
  newBTree(depth,curr+1,root.r)
}

func gothrough(root *Node){
  if root == nil {
    return
  }
  fmt.Printf("%d ",root.v)
  gothrough(root.l)
  gothrough(root.r)
}

func main(){
  root := &Node{v:0,l:nil,r:nil}
  newBTree(3,0,root)
  gothrough(root)
}
