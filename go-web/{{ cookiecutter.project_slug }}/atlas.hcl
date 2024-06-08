env "local" {
  diff {
    // By default, indexes are not added or dropped concurrently.
    concurrent_index {
      add  = true
      drop = true
    }
  }

  src = "file://db/schemas"
  dev = "sqlite://file?mode=memory"

  migration {
    dir = "file://db/migrations"
  }

  format {
    migrate {
      diff = format("{{sql . \" \" }}")
    }
  }
}
