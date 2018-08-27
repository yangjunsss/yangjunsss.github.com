package main
import "fmt"

type data struct{
  num int
}

func (p *data)rewritep() {
  p.num++
}

func (p data)rewritev(){
  p.num++
}

type rewrite interface{
  rewritep()
  rewritev()
}

func main(){
  d:=&data{0}
  d.rewritev()
  fmt.Println(d.num)
  d.rewritep()
  fmt.Println(d.num)
  var vdata data
  vdata.rewritep()
  fmt.Println(vdata.num)
  pdata:=new(data)
  pdata.rewritev()
  fmt.Println(pdata.num)
}
