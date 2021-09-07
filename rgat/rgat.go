package main

import (
	"fmt"
	"log"
	"os"
)

const HelpText = `usage: rgat <command> [<args>]

Avaliable commands:

'rgat help' list available subcommands`

func check(err error) {
    if err != nil {
        log.Fatal(err)
    }
}

func main() {
    if len(os.Args) < 2 {
        fmt.Println(HelpText)
    }

    switch os.Args[1] {

    case "help":
        if len(os.Args) != 2 {
            fmt.Printf("Unknow args for help: %v\n", os.Args[2])
        }
        fmt.Println(HelpText)
    default:
        fmt.Println(HelpText)
    }
}
