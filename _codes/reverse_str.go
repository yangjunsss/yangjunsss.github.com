package main
import (
  "fmt"
)

func reverse(s string) (string,[]rune){
  r:=[]rune(s)
  for i,j:=0,len(r)-1;i<j;i,j=i+1,j-1 {
    r[i],r[j]=r[j],[i]
  }
  return string(r),r
}

func main(){
    _,r:=reverse("hello*the*world")
    r=append(r,'*')
    i,j,k:=0,len(r),0
    for ;i<j;i++{
      if r[i]=='*'{
        _,wr:=reverse(string(r[k:i]))
        r=append(r,wr...)
        r=append(r,'*')
        k = i+1
      }
    }
    fmt.Println(string(r[j:len(r)-1]))
}
