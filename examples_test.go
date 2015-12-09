package torc_test

import (
    "fmt"
)

func Example() {
    controller := NewController("tcp", "127.0.0.1:9051")

    if e := controller.Connect(); e != nil {
        // Handle error
    }

    response, e := controller.GetInfo([]string{"version"})
    if e != nil {
        // Handle error
    }

    fmt.Println(repsonse.ValueOf("version"))
}

func ExampleNewController() {
    controller := NewController("tcp", "127.0.0.1:9051")

    if e := controller.Connect(); e != nil {
        // Handle error
    }
}

