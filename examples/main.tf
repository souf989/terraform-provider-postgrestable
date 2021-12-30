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