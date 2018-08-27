package main
import (
  "fmt"
)

type Node struct{
  v int
  next *Node
}

func newList(size int) *Node{
  var head *Node
  for i:=0;i<size;i++ {
    n:=&Node{v : i,next : head}
    head = n
  }
  return head
}

func reverse(head *Node) *Node{
  var last *Node
  for head != nil {
    tmp := head
    head = head.next
    tmp.next = last
    last = tmp
  }
  return last
}

func print(head *Node){
  for ;head!=nil; head=head.next{
    fmt.Print(head.v)
  }
}

func main(){
  head := newList(10)
  print(head)
  head = reverse(head)
  print(head)
}
