variable "db_user" {
  type    = string
  default = getenv("DB_USER")
}

variable "db_pass" {
  type    = string
  default = getenv("DB_PASSWORD")
}

variable "db_host" {
  type    = string
  default = getenv("DB_HOST")
}

variable "db_name" {
  type    = string
  default = getenv("DB_NAME")
}

variable "db_port" {
  type    = string
  default = getenv("DB_PORT")
}

env "docker-local" {
  diff {
    // By default, indexes are not added or dropped concurrently.
    concurrent_index {
      add  = true
      drop = true
    }
  }

  src = "file://db/schemas"

  url = "postgres://${var.db_user}:${var.db_pass}@${var.db_host}:${var.db_port}/${var.db_name}?search_path=public&sslmode=disable"
  // Define the URL of the Dev Database for this environment
  // See: https://atlasgo.io/concepts/dev-database
  dev = "postgres://${var.db_user}:${var.db_pass}@${var.db_host}:${var.db_port}/dev?search_path=public&sslmode=disable"

  migration {
    dir = "file://db/migrations"
  }

  format {
    migrate {
      {% raw %}
      diff = format("{{sql . \" \" }}")
      {% endraw %}
    }
  }
}

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
      {% raw %}
      diff = format("{{sql . \" \" }}")
      {% endraw %}
    }
  }
}
