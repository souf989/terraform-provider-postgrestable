terraform {
  required_providers {
    postgrestable = {
      version = "0.1"
      source = "local.com/test/postgrestable"
    }
  }
}

provider "postgrestable" {
  database = "rds"
  scheme  = "postgres"
  host = "localhost"
  port = 5432
  username = "rds"
  password = "rds"
  sslmode  = "disable"
}



resource "postgrestable_table" "third_table" {
  table = "lola"
  schema = "test"
  columns {
    name = "first"
    type = "varchar"
  }
  columns {
    name = "second"
    type = "varchar"
  }
  columns {
    name = "third"
    type = "varchar(10)"
  }
  columns {
    name = "fourth"
    type = "varchar(10)"
  }

}

resource "postgrestable_table" "first_table" {
  table = "lola"
  schema = "test"
  columns {
    name = "first"
    type = "text"
  }
  columns {
    name = "second"
    type = "varchar"
  }


}