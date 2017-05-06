package main

import "fmt"
import "github.com/Masterminds/semver"
import "github.com/srcclr/example-go-glide/sub"
import "github.com/mitchellh/go-homedir"
import "github.com/kylelemons/go-gypsy"
import "github.com/openshift/origin"
import "github.com/ant0ine/go-json-rest"
import "github.com/crowdmob/goamz"
import "github.com/simeji/jid"
type Options struct {
    Query   string `url:"q"`
    ShowAll bool   `url:"all"`
    Page    int    `url:"page"`
}

func main() {
    fmt.Println("Intro to Go!")
    opt := Options{ "foo", true, 2 }
    v, _ := query.Values(opt)
    fmt.Println(v.Encode()) // will output: "q=foo&all=true&page=2"
    
}
