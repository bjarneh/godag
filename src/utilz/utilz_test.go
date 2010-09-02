package utilz_test

import (
    "testing"
    "utilz/stringset"
    "utilz/stringbuffer"
)

func TestStringSet(t *testing.T) {

    ss := stringset.New()

    ss.Add("en")

    if ss.Len() != 1 {
        t.Fatal("stringset.Len() != 1\n")
    }

    ss.Add("to")

    if ss.Len() != 2 {
        t.Fatal("stringset.Len() != 2\n")
    }

    if !ss.Contains("en") {
        t.Fatal("! stringset.Contains('en')\n")
    }

    if !ss.Contains("to") {
        t.Fatal("! stringset.Contains('to')\n")
    }

    if ss.Contains("not here") {
        t.Fatal(" stringset.Contains('not here')\n")
    }
}

func TestStringBuffer(t *testing.T) {

    ss := stringbuffer.New()
    ss.Add("en")
    if ss.String() != "en" {
        t.Fatal(" stringset.String() != 'en'\n")
    }
    ss.Add("to")
    if ss.String() != "ento" {
        t.Fatal(" stringset.String() != 'ento'\n")
    }
    if ss.Len() != 4 {
        t.Fatal(" stringset.Len() != 4\n")
    }
    ss.Add("øæå"); // utf-8 multi-byte fun
    if ss.Len() != 10 {
        t.Fatal(" stringset.Len() != 10\n");
    }
    if ss.String() != "entoøæå" {
        t.Fatal(" stringset.String() != 'entoøæå'\n");
    }
    ss.ClearSize(5)
    if ss.Len() != 0 {
        t.Fatal(" stringset.Len() != 0\n")
    }
}
