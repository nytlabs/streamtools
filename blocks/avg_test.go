package blocks

import "testing"
import "log"

func TestAvg(t *testing.T){

    BuildLibrary()
    b, err := NewBlock("avg", "testBlock")
    if err != nil{
        t.Error("failed to create avg block", err.Error())
    }
    log.Println("created", b.BlockType)
}



