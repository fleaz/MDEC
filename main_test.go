package main

import (
	"testing"
)

func TestSplitEmailAddress(t *testing.T){
    testCases := []struct {
        input string
        userPart string
        domainPart string
        hasError bool
    }{
        {
            "foobar",
            "",
            "",
            true,
        },{
            "mail@fleaz.me",
            "mail",
            "fleaz.me",
            false,
        },{
            "valid@address@it.cyber.org",
            "valid@address",
            "it.cyber.org",
            false,
        },
    }
    
    for _,entry := range testCases{
        user,domain,err := splitEmailAddress(entry.input)

        if (user != entry.userPart) || (domain != entry.domainPart) || ((err!=nil) != entry.hasError){
            t.Errorf("splitEmailAddress failed for %q", entry.input)
        }
    }


}
