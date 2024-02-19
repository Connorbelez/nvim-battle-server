package main



import(
	"fmt"
	"vbattle-server/utils"
	"vbattle-server/gameserver"
)

func main(){

	t:=utils.Test

	gameserver.Start()
	t()
	fmt.Println("hello world")
}
